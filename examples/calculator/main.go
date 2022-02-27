package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/mattmeyers/repl"
)

var matcher = regexp.MustCompile(`(-?\d+)\s*([\+\*\/\-x])\s*(-?\d+)`)

func main() {
	r := repl.Repl{
		Input:  bufio.NewReader(os.Stdin),
		Output: os.Stdout,

		Handlers: []repl.Handler{
			func(ctx *repl.Context) (string, error) {
				if strings.TrimSpace(ctx.Input) == "quit" {
					return "", repl.ErrExit
				}

				return "", repl.ErrNoMatch
			},
			func(ctx *repl.Context) (string, error) {
				matches := matcher.FindAllStringSubmatch(ctx.Input, -1)
				if len(matches) != 1 {
					return "", repl.NewError("That doesn't work.")
				}

				a, _ := strconv.Atoi(matches[0][1])
				b, _ := strconv.Atoi(matches[0][3])

				var res int
				switch matches[0][2] {
				case "+":
					res = a + b
				case "-":
					res = a - b
				case "*", "x":
					res = a * b
				case "/":
					if b == 0 {
						return "", repl.NewError("Cannot divide by zero")
					}

					res = a / b
				}

				return strconv.Itoa(res), nil
			},
		},

		PreRun:   func(ctx *repl.Context) (string, error) { return "Welcome!\n", nil },
		PreRead:  func(ctx *repl.Context) (string, error) { return "Reading...\n", nil },
		PostEval: func(ctx *repl.Context) (string, error) { return "Evaluated! Looping...\n", nil },
		PostRun:  func(ctx *repl.Context) (string, error) { return "Farewell!\n", nil },
		Prompt:   func(ctx *repl.Context) (string, error) { return ">> ", nil },
	}

	if err := r.Run(); err != nil {
		fmt.Println(err.Error())
	}
}
