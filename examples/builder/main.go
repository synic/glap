package main

import (
	"fmt"
	"os"

	"github.com/synic/glap"
)

func main() {
	app := glap.NewCommand("builder").
		Version("1.0.0").
		About("Demonstrates the builder API for dynamic argument construction").
		Arg(glap.NewArg("config").
			Short('c').
			Help("Path to config file").
			Env("BUILDER_CONFIG").
			Required(true).
			ValueName("FILE")).
		Arg(glap.NewArg("verbose").
			Short('v').
			Help("Increase verbosity").
			Action(glap.Count)).
		Arg(glap.NewArg("tag").
			Short('t').
			Help("Add tags (repeatable)").
			Action(glap.Append)).
		Subcommand(glap.NewCommand("deploy").
			About("Deploy the application").
			Arg(glap.NewArg("target").
				Short('T').
				Help("Deploy target").
				PossibleValues("staging", "production").
				Required(true)).
			Arg(glap.NewArg("dry-run").
				Short('n').
				Help("Simulate without changes").
				Action(glap.SetTrue)))

	matches, err := app.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	config, _ := matches.GetString("config")
	verbosity := matches.GetOccurrences("verbose")
	tags, _ := matches.GetStringSlice("tag")

	fmt.Printf("Config:    %s\n", config)
	fmt.Printf("Verbosity: %d\n", verbosity)
	fmt.Printf("Tags:      %v\n", tags)

	if matches.SubcommandName() == "deploy" {
		sub := matches.SubcommandMatches()
		target, _ := sub.GetString("target")
		dryRun, _ := sub.GetBool("dry-run")
		fmt.Printf("Deploying to %s (dry-run: %v)\n", target, dryRun)
	}
}
