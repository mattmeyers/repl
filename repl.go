package repl

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

// Repl holds all of the REPL dependencies. Unless overwritten, the input will default
// to stdin and the output will default to stdout.
type Repl struct {
	Input  *bufio.Reader
	Output io.Writer

	Commands []Command

	Prompt Prompter

	PreRun   Hook
	PreRead  Hook
	PostEval Hook
	PostRun  Hook

	ctx *Context
}

// New instantiates a new Repl that can be further built and run.
func New() *Repl {
	return &Repl{
		Input:  bufio.NewReader(os.Stdin),
		Output: os.Stdout,
		ctx:    &Context{ctx: context.Background()},
	}
}

// WithCommand appends another Command to the Repl's command chain.
func (r *Repl) WithCommand(c Command) *Repl {
	r.Commands = append(r.Commands, c)
	return r
}

// WithPrompt sets the function to use to generate the REPL prompt.
func (r *Repl) WithPrompt(p Prompter) *Repl {
	r.Prompt = p
	return r
}

// WithPreRunHook sets the function to call when the pre run hook is run.
func (r *Repl) WithPreRunHook(h Hook) *Repl {
	r.PreRun = h
	return r
}

// WithPreReadHook sets the function to call when the pre read hook is run.
func (r *Repl) WithPreReadHook(h Hook) *Repl {
	r.PreRead = h
	return r
}

// WithPostEvalHook sets the function to call when the post eval hook is run.
func (r *Repl) WithPostEvalHook(h Hook) *Repl {
	r.PostEval = h
	return r
}

// WithPostRunHook sets the function to call when the post run hook is run.
func (r *Repl) WithPostRunHook(h Hook) *Repl {
	r.PostRun = h
	return r
}

// WithContext sets the context.Context within the Repl Context object.
func (r *Repl) WithContext(ctx context.Context) *Repl {
	r.ctx.ctx = ctx
	return r
}

// Context holds the current context of the REPL. This object can be used to access the
// user's input.
type Context struct {
	ctx context.Context

	Input string
}

// Context returns the context.Context held within the Repl's Context.
func (c *Context) Context() context.Context {
	return c.ctx
}

// Command represents a single REPL command. If the Matcher returns no error when passed
// the user's input, then the Handler is run.
type Command struct {
	Name   string
	Usage  string
	Match  Matcher
	Handle Handler
}

// Matcher is a function that takes the user's input and determines if the command's
// Handler should be run. If the input does not match, then the function must return
// ErrNoMatch.
type Matcher func(string) error

// StringMatcher checks if the user input perfectly matches the provided string.
func StringMatcher(s string) Matcher {
	return func(input string) error {
		if s != input {
			return ErrNoMatch
		}

		return nil
	}
}

// StringPrefixMatcher checks of the provided string is the prefix to the user input.
func StringPrefixMatcher(s string) Matcher {
	return func(input string) error {
		if !strings.HasPrefix(input, s) {
			return ErrNoMatch
		}

		return nil
	}
}

// OneOfMatcher matches if at least one of the provided strings perfectly matches the
// user input.
func OneOfMatcher(strs ...string) Matcher {
	return func(s string) error {
		for _, str := range strs {
			if s == str {
				return nil
			}
		}

		return ErrNoMatch
	}
}

// RegexMatcher matches the user input against the provided regex pattern. The regular
// expression is lazily compiled when the matcher is called for the first time. If the
// regular expression cannot be compiled, then a fatal error will be returned.
func RegexMatcher(pattern string) Matcher {
	var matcher *regexp.Regexp
	return func(s string) error {
		if matcher == nil {
			var err error
			matcher, err = regexp.Compile(pattern)
			if err != nil {
				return NewFatalError("invalid regular expression")
			}
		}

		if !matcher.MatchString(s) {
			return ErrNoMatch
		}

		return nil
	}
}

// AlwaysMatcher matches all user input. This can be used for a catch-all command.
func AlwaysMatcher() Matcher {
	return func(s string) error { return nil }
}

// NeverMatcher never matches the user input.
func NeverMatcher() Matcher {
	return func(s string) error { return ErrNoMatch }
}

// Handler represents a function that can handle a command.
type Handler func(*Context) (string, error)

// Hook is a function that can be run at certain execution points in the REPL lifecycle.
// Any error returned from a hook function will be treated as a fatal error. Any
// non-empty string returned will be printed.
type Hook func(*Context) (string, error)

// Prompter is a function that can be used to dynamically build the REPL prompt. Any
// error returned will be treated as fatal.
type Prompter func(*Context) (string, error)

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
// from a Command, the message will be displayed and the REPL will loop.
func NewError(message string) Error {
	return Error{Message: message, Fatal: false}
}

// NewFatalError creates a new fatal REPL error. When one of these errors is returned
// from a Command, the message will be displayed and the REPL will exit.
func NewFatalError(message string) Error {
	return Error{Message: message, Fatal: true}
}

var (
	// ErrNoMatch signals that an entered command does not match a handler. This error
	// must be returned from any Command that cannot handle the provided command.
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

	r.ctx = &Context{ctx: context.Background()}

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

	s, err := hook(r.ctx)
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

		r.ctx.Input, err = r.readInput()
		if err != nil {
			return err
		}

		for _, command := range r.Commands {
			var replErr Error
			err := command.Match(r.ctx.Input)
			if errors.Is(err, ErrNoMatch) {
				continue
			} else if errors.As(err, &replErr) {
				if replErr.Fatal {
					return replErr
				}

				fmt.Fprintf(r.Output, "%v\n", replErr)
			} else if err != nil {
				return err
			}

			output, err := command.Handle(r.ctx)
			if errors.Is(err, ErrExit) {
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
	p, err := r.Prompt(r.ctx)
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
