package main

import (
	"fmt"
	"os"

	"github.com/synic/glap"
)

// Mutually exclusive group: exactly one of --json, --text, --yaml is required.
func main() {
	app := glap.NewCommand("groups").
		About("Demonstrates argument groups for mutual exclusion").
		Arg(glap.NewArg("json").
			Help("Output as JSON").
			Action(glap.SetTrue).
			Group("format")).
		Arg(glap.NewArg("text").
			Help("Output as plain text").
			Action(glap.SetTrue).
			Group("format")).
		Arg(glap.NewArg("yaml").
			Help("Output as YAML").
			Action(glap.SetTrue).
			Group("format")).
		Arg(glap.NewArg("input").
			Short('i').
			Help("Input file").
			Required(true).
			ValueName("FILE")).
		ArgGroup(glap.NewArgGroup("format").Required(true))

	matches, err := app.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	input, _ := matches.GetString("input")
	format := "unknown"
	for _, f := range []string{"json", "text", "yaml"} {
		if matches.Contains(f) {
			format = f
			break
		}
	}

	fmt.Printf("Processing %s with %s output\n", input, format)
}
