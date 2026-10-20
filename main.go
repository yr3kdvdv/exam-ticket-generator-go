package main

import (
	"bufio"
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"io"
	"math/big"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"unicode/utf8"
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

	interrupts := make(chan os.Signal, 1)
	signal.Notify(interrupts, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(interrupts)
	go func() {
		<-interrupts
		console.close()
		os.Exit(130)
	}()

	fmt.Fprintln(os.Stdout, "Предмет: Proiectarea Sistemelor Informaționale")
	fmt.Fprintln(os.Stdout, "Для выхода нажмите ESC.")
	for {
		lastName, err := console.readNonempty("Last name: ")
		if finishOnInputError(err) {
			return
		}
		firstName, err := console.readNonempty("First name: ")
		if finishOnInputError(err) {
			return
		}

		number, err := randomTicket()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Не удалось сгенерировать билет:", err)
			return
		}
		fmt.Fprintf(os.Stdout, "Билет № %d.\nВопрос: %s\n", number, questions[number-1])
		for {
			err = appendJournal(*file, lastName, firstName, number)
			if err == nil {
				fmt.Fprintf(os.Stdout, "Запись сохранена в %s.\n", *file)
				break
			}
			fmt.Fprintf(os.Stderr, "Не удалось записать %s: %v\nЗакройте файл в Excel и проверьте права записи. Enter — повторить, ESC — выйти.\n", *file, err)
			if err = console.waitRetry(); err != nil {
				return
			}
		}
	}
}

func finishOnInputError(err error) bool {
	if err == nil {
		return false
	}
	if !errors.Is(err, errExit) && !errors.Is(err, io.EOF) {
		fmt.Fprintln(os.Stderr, "Ошибка ввода:", err)
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
	in      *bufio.Reader
	out     io.Writer
	stdin   *os.File
	sttyOld string
	mu      sync.Once
}

func newConsole(stdin *os.File, out io.Writer) (*console, error) {
	c := &console{in: bufio.NewReader(stdin), out: out, stdin: stdin}
	old := exec.Command("stty", "-g")
	old.Stdin = stdin
	state, err := old.Output()
	if err != nil {
		// A pipe has no terminal settings. It still supports line-based tests.
		return c, nil
	}
	c.sttyOld = strings.TrimSpace(string(state))
	raw := exec.Command("stty", "-echo", "-icanon", "min", "1", "time", "0")
	raw.Stdin = stdin
	if err := raw.Run(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *console) close() {
	c.mu.Do(func() {
		if c.sttyOld == "" {
			return
		}
		cmd := exec.Command("stty", c.sttyOld)
		cmd.Stdin = c.stdin
		_ = cmd.Run()
	})
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
					if c.sttyOld != "" {
						fmt.Fprint(c.out, "\b \b")
					}
				}
			default:
				if r >= ' ' && r != utf8.RuneError {
					input = append(input, r)
					if c.sttyOld != "" {
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
