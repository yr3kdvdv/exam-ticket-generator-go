package main

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const sheetPath = "xl/worksheets/sheet1.xml"

var sheetDataEnd = regexp.MustCompile(`</(?:[A-Za-z_][A-Za-z0-9_.-]*:)?sheetData\s*>`)
var dimensionRef = regexp.MustCompile(`(<dimension\b[^>]*\bref=")A1:D[0-9]+(")`)

func appendJournal(path, lastName, firstName string, ticket int) error {
	if ticket < 1 || ticket > ticketCount {
		return fmt.Errorf("номер билета %d вне диапазона 1–%d", ticket, ticketCount)
	}
	if strings.TrimSpace(lastName) == "" || strings.TrimSpace(firstName) == "" {
		return errors.New("фамилия и имя не могут быть пустыми")
	}

	lockFile := filepath.Join(filepath.Dir(path), "~$"+filepath.Base(path))
	if _, err := os.Stat(lockFile); err == nil {
		return fmt.Errorf("файл занят Excel (%s)", lockFile)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	stamp := time.Now().Format("2006-01-02 15:04:05")
	book, err := zip.OpenReader(path)
	if errors.Is(err, os.ErrNotExist) {
		return writeNewJournal(path, lastName, firstName, ticket, stamp)
	}
	if err != nil {
		return fmt.Errorf("открыть журнал: %w", err)
	}
	defer book.Close()

	var original []byte
	for _, f := range book.File {
		if f.Name == sheetPath {
			r, err := f.Open()
			if err != nil {
				return err
			}
			original, err = io.ReadAll(r)
			closeErr := r.Close()
			if err != nil {
				return err
			}
			if closeErr != nil {
				return closeErr
			}
			break
		}
	}
	if original == nil {
		return fmt.Errorf("в %s не найден первый лист Excel", path)
	}

	next, err := nextRow(original)
	if err != nil {
		return fmt.Errorf("прочитать строки журнала: %w", err)
	}
	if next < 2 {
		return errors.New("в журнале отсутствует строка заголовков")
	}
	updated, err := addRow(original, next, lastName, firstName, ticket, stamp)
	if err != nil {
		return err
	}

	return writeAtomic(path, func(out *os.File) error {
		zw := zip.NewWriter(out)
		for _, f := range book.File {
			if f.Name == sheetPath {
				h := f.FileHeader
				w, err := zw.CreateHeader(&h)
				if err != nil {
					return err
				}
				if _, err := w.Write(updated); err != nil {
					return err
				}
				continue
			}
			if err := zw.Copy(f); err != nil {
				return err
			}
		}
		if err := zw.Close(); err != nil {
			return err
		}
		return book.Close()
	})
}

func nextRow(sheet []byte) (int, error) {
	dec := xml.NewDecoder(bytes.NewReader(sheet))
	maxRow := 0
	for {
		token, err := dec.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return 0, err
		}
		start, ok := token.(xml.StartElement)
		if !ok || start.Name.Local != "row" {
			continue
		}
		for _, attr := range start.Attr {
			if attr.Name.Local == "r" {
				n, err := strconv.Atoi(attr.Value)
				if err != nil || n < 1 {
					return 0, fmt.Errorf("некорректный номер строки %q", attr.Value)
				}
				if n > maxRow {
					maxRow = n
				}
				break
			}
		}
	}
	return maxRow + 1, nil
}

func addRow(sheet []byte, number int, lastName, firstName string, ticket int, stamp string) ([]byte, error) {
	row := rowXML(number, lastName, firstName, ticket, stamp)
	where := sheetDataEnd.FindIndex(sheet)
	if where == nil {
		return nil, errors.New("в первом листе отсутствует sheetData")
	}
	updated := make([]byte, 0, len(sheet)+len(row))
	updated = append(updated, sheet[:where[0]]...)
	updated = append(updated, row...)
	updated = append(updated, sheet[where[0]:]...)
	updated = dimensionRef.ReplaceAll(updated, []byte("${1}A1:D"+strconv.Itoa(number)+"${2}"))
	return updated, nil
}

func rowXML(number int, lastName, firstName string, ticket int, stamp string) []byte {
	var b strings.Builder
	fmt.Fprintf(&b, `<row xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" r="%d">`, number)
	inlineCell(&b, "A", number, lastName)
	inlineCell(&b, "B", number, firstName)
	fmt.Fprintf(&b, `<c r="C%d"><v>%d</v></c>`, number, ticket)
	inlineCell(&b, "D", number, stamp)
	b.WriteString(`</row>`)
	return []byte(b.String())
}

func inlineCell(b *strings.Builder, column string, row int, value string) {
	fmt.Fprintf(b, `<c r="%s%d" t="inlineStr"><is><t xml:space="preserve">`, column, row)
	_ = xml.EscapeText(b, []byte(value))
	b.WriteString(`</t></is></c>`)
}

func writeNewJournal(path, lastName, firstName string, ticket int, stamp string) error {
	entries := []struct {
		name string
		body string
	}{
		{"[Content_Types].xml", contentTypesXML},
		{"_rels/.rels", rootRelsXML},
		{"xl/workbook.xml", workbookXML},
		{"xl/_rels/workbook.xml.rels", workbookRelsXML},
		{"xl/styles.xml", stylesXML},
	}
	var sheet strings.Builder
	sheet.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?><worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData><row r="1">`)
	inlineCell(&sheet, "A", 1, "Last name")
	inlineCell(&sheet, "B", 1, "First name")
	inlineCell(&sheet, "C", 1, "Номер билета")
	inlineCell(&sheet, "D", 1, "Дата и время")
	sheet.WriteString(`</row>`)
	sheet.Write(rowXML(2, lastName, firstName, ticket, stamp))
	sheet.WriteString(`</sheetData></worksheet>`)

	return writeAtomic(path, func(out *os.File) error {
		zw := zip.NewWriter(out)
		for _, entry := range entries {
			w, err := zw.Create(entry.name)
			if err != nil {
				return err
			}
			if _, err := io.WriteString(w, entry.body); err != nil {
				return err
			}
		}
		w, err := zw.Create(sheetPath)
		if err != nil {
			return err
		}
		if _, err := io.WriteString(w, sheet.String()); err != nil {
			return err
		}
		return zw.Close()
	})
}

func writeAtomic(path string, write func(*os.File) error) error {
	dir := filepath.Dir(path)
	f, err := os.CreateTemp(dir, ".journal-*.xlsx")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if err := write(f); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := os.Rename(f.Name(), path); err != nil {
		return err
	}
	return nil
}

const contentTypesXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/><Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/><Override PartName="/xl/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.styles+xml"/></Types>`

const rootRelsXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/></Relationships>`

const workbookXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><sheets><sheet name="Journal" sheetId="1" r:id="rId1"/></sheets></workbook>`

const workbookRelsXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/><Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/></Relationships>`

const stylesXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><styleSheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><fonts count="1"><font><sz val="11"/><name val="Calibri"/></font></fonts><fills count="2"><fill><patternFill patternType="none"/></fill><fill><patternFill patternType="gray125"/></fill></fills><borders count="1"><border><left/><right/><top/><bottom/><diagonal/></border></borders><cellStyleXfs count="1"><xf numFmtId="0" fontId="0" fillId="0" borderId="0"/></cellStyleXfs><cellXfs count="1"><xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0"/></cellXfs><cellStyles count="1"><cellStyle name="Normal" xfId="0" builtinId="0"/></cellStyles></styleSheet>`
