package main

import (
	"fmt"
	"os"

	"github.com/synic/glap"
)

type CLI struct {
	Force  bool   `glap:"force,short=f,help=Overwrite existing files"`
	Source string `glap:"source,positional,required,value_name=SRC,help=Source file"`
	Dest   string `glap:"dest,positional,required,value_name=DST,help=Destination file"`
}

func main() {
	var cli CLI
	app := glap.New(&cli).
		Name("cp-lite").
		About("Demonstrates positional arguments mixed with flags")

	_, err := app.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	action := "Copying"
	if cli.Force {
		action = "Force copying"
	}
	fmt.Printf("%s %s -> %s\n", action, cli.Source, cli.Dest)
}
