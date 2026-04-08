package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/synic/glap"
)

// A wrapper that passes remaining args to another command.
// Try: go run . --verbose ls -la /tmp
type CLI struct {
	Verbose bool     `glap:"verbose,short=v,help=Show the command being run"`
	Cmd     []string `glap:"cmd,trailing_var_arg,help=Command and arguments to run"`
}

func main() {
	var cli CLI
	app := glap.New(&cli).
		Name("runner").
		About("Demonstrates trailing_var_arg to capture remaining arguments")

	_, err := app.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if len(cli.Cmd) == 0 {
		fmt.Fprintln(os.Stderr, "no command specified")
		os.Exit(1)
	}

	if cli.Verbose {
		fmt.Printf("Running: %v\n", cli.Cmd)
	}

	cmd := exec.Command(cli.Cmd[0], cli.Cmd[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		os.Exit(1)
	}
}
