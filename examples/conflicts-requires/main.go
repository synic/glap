package main

import (
	"fmt"
	"os"

	"github.com/synic/glap"
)

// --json and --text conflict with each other.
// --output requires --format to also be set.
// --filename is required when --format=file.
// --input is required unless --stdin is present.
type CLI struct {
	JSON     bool   `glap:"json,conflicts_with=text,help=Output as JSON"`
	Text     bool   `glap:"text,conflicts_with=json,help=Output as plain text"`
	Output   string `glap:"output,short=o,requires=format,help=Output file path,value_name=FILE"`
	Format   string `glap:"format,short=f,help=Output format for file"`
	Filename string `glap:"filename,required_if_eq=format:file,help=Filename (required when format=file)"`
	Stdin    bool   `glap:"stdin,help=Read from stdin"`
	Input    string `glap:"input,short=i,required_unless=stdin,help=Input file (required unless --stdin)"`
}

func main() {
	var cli CLI
	app := glap.New(&cli).
		Name("conflicts-requires").
		About("Demonstrates conflict/dependency/conditional relationships between arguments")

	_, err := app.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if cli.JSON {
		fmt.Println("Using JSON output")
	} else if cli.Text {
		fmt.Println("Using text output")
	}

	if cli.Output != "" {
		fmt.Printf("Writing to %s as %s\n", cli.Output, cli.Format)
	}

	if cli.Stdin {
		fmt.Println("Reading from stdin")
	} else {
		fmt.Printf("Reading from %s\n", cli.Input)
	}
}
