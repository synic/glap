package main

import (
	"fmt"
	"os"

	"github.com/synic/glap"
)

type ServeCLI struct {
	Port int    `glap:"port,short=p,default=8080,help=Port to listen on"`
	Host string `glap:"host,short=H,default=localhost,help=Bind address"`
}

type CLI struct {
	Verbose bool      `glap:"verbose,short=v,global,help=Enable verbose output"`
	Output  string    `glap:"output,short=o,possible=json|text|yaml,default=text,help=Output format"`
	Serve   *ServeCLI `glap:"serve,subcommand,help=Start the server"`
}

// Generate completions by setting COMPLETE to a shell name:
//
//	COMPLETE=bash myapp >> ~/.bashrc
//	COMPLETE=zsh myapp > ~/.zfunc/_myapp
//	COMPLETE=fish myapp > ~/.config/fish/completions/myapp.fish
func main() {
	var cli CLI
	app := glap.New(&cli).
		Name("myapp").
		Version("1.0.0").
		About("Demonstrates shell completion generation")

	if glap.CompleteApp(app, os.Stdout) {
		return
	}

	cmd, err := app.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("Command: %s, Verbose: %v, Output: %s\n",
		cmd, cli.Verbose, cli.Output)
}
