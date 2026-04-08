package main

import (
	"fmt"
	"os"

	"github.com/synic/glap"
)

type CLI struct {
	Config  string   `glap:"config,short=c,required,help=Path to config file"`
	Verbose bool     `glap:"verbose,short=v,help=Enable verbose output"`
	NoColor bool     `glap:"no-color,action=set_false,help=Disable colored output"`
	Port    int      `glap:"port,short=p,default=8080,help=Port to listen on"`
	Output  string   `glap:"output,short=o,possible=json|text|yaml,default=text,help=Output format"`
	Tags    []string `glap:"tag,short=t,action=append,delimiter=comma,help=Tags (comma-separated)"`
	Offset  int      `glap:"offset,positional"`
}

func main() {
	var cli CLI
	app := glap.New(&cli).
		Name("basic").
		Version("1.0.0").
		About("Demonstrates basic glap usage with struct tags").
		ArgRequiredElseHelp(true).
		AllowNegativeNumbers(true)

	_, err := app.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("Config:  %s\n", cli.Config)
	fmt.Printf("Verbose: %v\n", cli.Verbose)
	fmt.Printf("NoColor: %v\n", cli.NoColor)
	fmt.Printf("Port:    %d\n", cli.Port)
	fmt.Printf("Output:  %s\n", cli.Output)
	fmt.Printf("Tags:    %v\n", cli.Tags)
	fmt.Printf("Offset:  %d\n", cli.Offset)
}
