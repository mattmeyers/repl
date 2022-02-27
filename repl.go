package repl

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

// Repl holds all of the REPL dependencies. Unless overwritten, the input will default
// to stdin and the output will default to stdout.
type Repl struct {
	Input  *bufio.Reader
	Output io.Writer

	Handlers []Handler

	Prompt Prompter

	PreRun   Hook
	PreRead  Hook
	PostEval Hook
	PostRun  Hook
}

// Handler represents a function that can handle a command. Handlers are expected to
// ensure the command is appropriate for the handler function. If not, the Handler
// must return ErrNoMatch. Any non-empty string returned from the function will be
// printed.
type Handler func(string) (string, error)

// Hook is a function that can be run at certain execution points in the REPL lifecycle.
// Any error returned from a hook function will be treated as a fatal error. Any
// non-empty string returned will be printed.
type Hook func() (string, error)

// Prompter is a function that can be used to dynamically build the REPL prompt. Any
// error returned will be treated as fatal.
type Prompter func() (string, error)

// Error represents a REPL error. Because a REPL command can result in a non fatal
// error that keeps the REPL alive, a special error cosntruct must be used to
// encode this information. REPL errors must always be used to keep the REPL running
// after a failed command. All other errors will be treated as fatal errors.
type Error struct {
	Message string
	Fatal   bool
}

func (e Error) Error() string {
	return e.Message
}

// NewError creates a new non fatal REPL error. When one of these errors is returned
// from a Handler, the message will be displayed and the REPL will loop.
func NewError(message string) Error {
	return Error{Message: message, Fatal: false}
}

// NewFatalError creates a new fatal REPL error. When one of these errors is returned
// from a Handler, the message will be displayed and the REPL will exit.
func NewFatalError(message string) Error {
	return Error{Message: message, Fatal: true}
}

var (
	// ErrNoMatch signals that an entered command does not match a handler. This error
	// must be returned from any Handler that cannot handle the provided command.
	ErrNoMatch = errors.New("no match")
	// ErrExit signals that the REPL should cleanly exit.
	ErrExit = errors.New("exit")
)

// Run starts the REPL. The lifecycle of the REPL is as follows
//
//		1. Pre run hook
//		2. Loop until exit
//			a. Pre read hook
//			b. Print prompt
//			c. Handle input
//			d. Post eval hook
//		3. Post run hook
//
// During the execution of the REPL, errors may occur. Most errors will be non fatal.
// These errors will result in their message being printed, and then the loop continuing.
// If at some point a non recoverable error occurs, the error message will be printed
// and the REPL will exit. The exit will occur immediately. Any hooks that have not yet
// run will be skipped.
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
		err := r.runHook(r.PreRead)
		if err != nil {
			return err
		}

		err = r.printPrompt()
		if err != nil {
			return err
		}

		input, err := r.readInput()
		if err != nil {
			return err
		}

		for _, handler := range r.Handlers {
			output, err := handler(input)

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

			break
		}

		err = r.runHook(r.PostEval)
		if err != nil {
			return err
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
