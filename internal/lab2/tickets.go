package lab2

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type Ticket struct {
	Number    int
	Questions [3]string
}

var headingPattern = regexp.MustCompile(`(?i)^\s*Билет\s+(\d+)\s*$`)
var questionPattern = regexp.MustCompile(`^\s*\d+[.)]\s*(.+)$`)

// LoadTickets accepts the visible paragraphs in a Word document. Broken blocks
// are skipped so that one bad ticket does not make the entire bank unusable.
func LoadTickets(path string) (map[int]Ticket, []string, error) {
	z, err := zip.OpenReader(path)
	if err != nil {
		return nil, nil, fmt.Errorf("не удалось открыть tickets.docx: %w", err)
	}
	defer z.Close()
	var body io.ReadCloser
	for _, file := range z.File {
		if file.Name == "word/document.xml" {
			body, err = file.Open()
			if err != nil {
				return nil, nil, err
			}
			break
		}
	}
	if body == nil {
		return nil, nil, fmt.Errorf("в tickets.docx отсутствует word/document.xml")
	}
	defer body.Close()
	paragraphs, err := readParagraphs(body)
	if err != nil {
		return nil, nil, fmt.Errorf("ошибка чтения tickets.docx: %w", err)
	}
	return ParseTickets(paragraphs)
}

func readParagraphs(r io.Reader) ([]string, error) {
	d := xml.NewDecoder(r)
	var paragraphs []string
	var current strings.Builder
	insideParagraph := false
	insideText := false
	for {
		tok, err := d.Token()
		if err == io.EOF {
			return paragraphs, nil
		}
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "p":
				insideParagraph = true
				current.Reset()
			case "t":
				insideText = insideParagraph
			case "tab":
				if insideParagraph {
					current.WriteByte(' ')
				}
			}
		case xml.CharData:
			if insideText {
				current.Write(t)
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "t":
				insideText = false
			case "p":
				if insideParagraph {
					paragraphs = append(paragraphs, strings.TrimSpace(current.String()))
				}
				insideParagraph = false
			}
		}
	}
}

func ParseTickets(paragraphs []string) (map[int]Ticket, []string, error) {
	tickets := make(map[int]Ticket)
	var warnings []string
	var number int
	var questions []string
	active := false
	finish := func() {
		if !active {
			return
		}
		if len(questions) < 3 {
			warnings = append(warnings, fmt.Sprintf("билет %d пропущен: найдено %d вопросов вместо трёх", number, len(questions)))
			return
		}
		if _, exists := tickets[number]; exists {
			warnings = append(warnings, fmt.Sprintf("повтор номера билета %d пропущен", number))
			return
		}
		tickets[number] = Ticket{Number: number, Questions: [3]string{questions[0], questions[1], questions[2]}}
	}
	for _, raw := range paragraphs {
		line := strings.TrimSpace(raw)
		if m := headingPattern.FindStringSubmatch(line); m != nil {
			finish()
			number, _ = strconv.Atoi(m[1])
			active = number > 0
			questions = nil
			if !active {
				warnings = append(warnings, "билет с нулевым номером пропущен")
			}
			continue
		}
		if active {
			if m := questionPattern.FindStringSubmatch(line); m != nil {
				questions = append(questions, strings.TrimSpace(m[1]))
			}
		}
	}
	finish()
	if len(tickets) == 0 {
		return nil, warnings, fmt.Errorf("в tickets.docx не найдено ни одного корректного билета с тремя вопросами")
	}
	return tickets, warnings, nil
}

func SortedTicketNumbers(tickets map[int]Ticket) []int {
	ids := make([]int, 0, len(tickets))
	for id := range tickets {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	return ids
}
