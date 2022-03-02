package main

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/mattmeyers/repl"
)

func main() {
	r := repl.New().
		WithCommand(repl.Command{
			Name:   "quit",
			Usage:  "Exit the REPL",
			Match:  repl.OneOfMatcher("quit", "q"),
			Handle: handleQuit,
		}).
		WithCommand(repl.Command{
			Match:  repl.AlwaysMatcher(),
			Handle: handleCalculate,
		}).
		WithPrompt(func(ctx *repl.Context) (string, error) { return ">> ", nil }).
		WithPreRunHook(func(ctx *repl.Context) (string, error) { return "Welcome!\n", nil }).
		WithPreReadHook(func(ctx *repl.Context) (string, error) { return "Reading...\n", nil }).
		WithPostEvalHook(func(ctx *repl.Context) (string, error) { return "Evaluated! Looping...\n", nil }).
		WithPostRunHook(func(ctx *repl.Context) (string, error) { return "Farewell!\n", nil })

	if err := r.Run(); err != nil {
		fmt.Println(err.Error())
	}
}

func handleQuit(ctx *repl.Context) (string, error) {
	return "", repl.ErrExit
}

var matcher = regexp.MustCompile(`(-?\d+)\s*([\+\*\/\-x])\s*(-?\d+)`)

func handleCalculate(ctx *repl.Context) (string, error) {
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
}
