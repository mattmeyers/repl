package repl

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

type Repl struct {
	Input  *bufio.Reader
	Output io.Writer

	Handlers []Handler

	PreRun  Hook
	PostRun Hook
	Prompt  Prompter
}

type Error struct {
	Message string
	Fatal   bool
}

func (e Error) Error() string {
	return e.Message
}

func NewError(message string) Error { return Error{Message: message, Fatal: false} }

func NewFatalError(message string) Error { return Error{Message: message, Fatal: true} }

var (
	ErrNoMatch = errors.New("no match")
	ErrExit    = errors.New("exit")
)

type Handler interface {
	Handle(string) (string, error)
}

type HandlerFunc func(string) (string, error)

func (f HandlerFunc) Handle(s string) (string, error) { return f(s) }

type Hook func() (string, error)

type Prompter func() (string, error)

func (r *Repl) Run() error {
	var err error

	err = r.runHook(r.PreRun)
	if err != nil {
		return err
	}

	err = r.runLoop()
	if err != nil {
		return err
	}

	err = r.runHook(r.PostRun)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repl) runHook(hook Hook) error {
	if hook == nil {
		return nil
	}

	s, err := hook()
	if err != nil {
		return err
	} else if s != "" {
		fmt.Fprint(r.Output, s)
	}

	return nil
}

func (r *Repl) runLoop() error {
	for {
		err := r.printPrompt()
		if err != nil {
			return err
		}

		input, err := r.readInput()
		if err != nil {
			return err
		}

		for _, h := range r.Handlers {
			output, err := h.Handle(input)

			var replErr Error
			if errors.Is(err, ErrNoMatch) {
				continue
			} else if errors.Is(err, ErrExit) {
				return nil
			} else if errors.As(err, &replErr) {
				if replErr.Fatal {
					return replErr
				}

				fmt.Fprintf(r.Output, "%v\n", replErr)
			} else if err != nil {
				return err
			} else if output != "" {
				fmt.Fprintf(r.Output, "%s\n", output)
			}
		}
	}
}

func (r *Repl) printPrompt() error {
	p, err := r.Prompt()
	if err != nil {
		return err
	}

	fmt.Fprint(r.Output, p)

	return nil
}

func (r *Repl) readInput() (string, error) {
	input, err := r.Input.ReadString('\n')
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(input), nil
}
