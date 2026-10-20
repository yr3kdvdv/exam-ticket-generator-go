package main

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"io"
	"math/big"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"unicode/utf8"

	"golang.org/x/term"
)

const ticketCount = 20

var errExit = errors.New("exit requested")

func main() {
	file := flag.String("file", "journal.xlsx", "path to the Excel journal")
	flag.Parse()

	console, err := newConsole(os.Stdin, os.Stdout)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Ошибка настройки консоли:", err)
		os.Exit(1)
	}
	defer console.close()
	out := console.out
	errOut := io.Writer(os.Stderr)
	if console.raw && runtime.GOOS != "windows" {
		errOut = crlfWriter{os.Stderr}
	}

	interrupts := make(chan os.Signal, 1)
	signal.Notify(interrupts, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(interrupts)
	go func() {
		<-interrupts
		console.close()
		os.Exit(130)
	}()

	fmt.Fprintln(out, "Предмет: Proiectarea Sistemelor Informaționale")
	fmt.Fprintln(out, "Для выхода нажмите ESC.")
	for {
		lastName, err := console.readNonempty("Last name: ")
		if finishOnInputError(err, errOut) {
			return
		}
		firstName, err := console.readNonempty("First name: ")
		if finishOnInputError(err, errOut) {
			return
		}

		number, err := randomTicket()
		if err != nil {
			fmt.Fprintln(errOut, "Не удалось сгенерировать билет:", err)
			return
		}
		fmt.Fprintf(out, "Билет № %d.\nВопрос: %s\n", number, questions[number-1])
		for {
			err = appendJournal(*file, lastName, firstName, number)
			if err == nil {
				fmt.Fprintf(out, "Запись сохранена в %s.\n", *file)
				break
			}
			fmt.Fprintf(errOut, "Не удалось записать %s: %v\nЗакройте файл в Excel и проверьте права записи. Enter — повторить, ESC — выйти.\n", *file, err)
			if err = console.waitRetry(); err != nil {
				return
			}
		}
	}
}

func finishOnInputError(err error, errOut io.Writer) bool {
	if err == nil {
		return false
	}
	if !errors.Is(err, errExit) && !errors.Is(err, io.EOF) {
		fmt.Fprintln(errOut, "Ошибка ввода:", err)
	}
	return true
}

func randomTicket() (int, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(ticketCount))
	if err != nil {
		return 0, err
	}
	return int(n.Int64()) + 1, nil
}

type console struct {
	in       *bufio.Reader
	out      io.Writer
	stdin    *os.File
	rawState *term.State
	raw      bool
	mu       sync.Once
}

func newConsole(stdin *os.File, out io.Writer) (*console, error) {
	c := &console{in: bufio.NewReader(stdin), out: out, stdin: stdin}
	fd := int(stdin.Fd())
	if !term.IsTerminal(fd) {
		// A pipe has no terminal settings. It still supports line-based tests.
		return c, nil
	}
	state, err := term.MakeRaw(fd)
	if err != nil {
		return nil, err
	}
	c.rawState = state
	c.raw = true
	if runtime.GOOS != "windows" {
		c.out = crlfWriter{out}
	}
	return c, nil
}

func (c *console) close() {
	c.mu.Do(func() {
		if c.rawState == nil {
			return
		}
		_ = term.Restore(int(c.stdin.Fd()), c.rawState)
	})
}

type crlfWriter struct{ io.Writer }

func (w crlfWriter) Write(p []byte) (int, error) {
	converted := bytes.ReplaceAll(p, []byte("\n"), []byte("\r\n"))
	n, err := w.Writer.Write(converted)
	if err != nil {
		return 0, err
	}
	if n != len(converted) {
		return 0, io.ErrShortWrite
	}
	return len(p), nil
}

func (c *console) readNonempty(prompt string) (string, error) {
	for {
		fmt.Fprint(c.out, prompt)
		var input []rune
		for {
			r, size, err := c.in.ReadRune()
			if err != nil {
				return "", err
			}
			if size == 1 && r == 27 {
				fmt.Fprintln(c.out)
				return "", errExit
			}
			switch r {
			case '\r', '\n':
				fmt.Fprintln(c.out)
				value := strings.TrimSpace(string(input))
				if value != "" {
					return value, nil
				}
				fmt.Fprintln(c.out, "Поле не может быть пустым.")
				goto again
			case '\b', 127:
				if len(input) > 0 {
					input = input[:len(input)-1]
					if c.raw {
						fmt.Fprint(c.out, "\b \b")
					}
				}
			default:
				if r >= ' ' && r != utf8.RuneError {
					input = append(input, r)
					if c.raw {
						fmt.Fprint(c.out, string(r))
					}
				}
			}
		}
	again:
	}
}

func (c *console) waitRetry() error {
	for {
		r, _, err := c.in.ReadRune()
		if err != nil {
			return err
		}
		if r == 27 {
			fmt.Fprintln(c.out)
			return errExit
		}
		if r == '\r' || r == '\n' {
			fmt.Fprintln(c.out)
			return nil
		}
	}
}
