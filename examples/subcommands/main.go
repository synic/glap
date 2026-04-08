package main

import (
	"fmt"
	"os"

	"github.com/synic/glap"
)

type ServeCLI struct {
	Port int    `glap:"port,short=p,default=8080,env=PORT,help=Port to listen on"`
	Host string `glap:"host,short=H,default=localhost,help=Bind address"`
}

type InitCLI struct {
	Name     string `glap:"name,positional,required,help=Project name"`
	Template string `glap:"template,short=t,default=default,possible=default|minimal|full,help=Project template"`
}

type RemoteAddCLI struct {
	Name string `glap:"name,positional,required,help=Remote name"`
	URL  string `glap:"url,positional,required,help=Remote URL"`
}

type RemoteCLI struct {
	Add *RemoteAddCLI `glap:"add,subcommand,help=Add a new remote"`
}

type CLI struct {
	Verbose bool       `glap:"verbose,short=v,global,help=Enable verbose output"`
	Serve   *ServeCLI  `glap:"serve,subcommand,help=Start the server"`
	Init    *InitCLI   `glap:"init,subcommand,help=Initialize a new project"`
	Remote  *RemoteCLI `glap:"remote,subcommand,help=Manage remotes"`
}

func main() {
	var cli CLI
	app := glap.New(&cli).
		Name("myapp").
		Version("2.0.0").
		About("Demonstrates subcommands, including nested ones")

	cmd, err := app.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("Verbose: %v\n", cli.Verbose)

	switch cmd {
	case "serve":
		fmt.Printf("Serving on %s:%d\n", cli.Serve.Host, cli.Serve.Port)
	case "init":
		fmt.Printf("Initializing project %q with template %q\n", cli.Init.Name, cli.Init.Template)
	case "remote add":
		fmt.Printf("Adding remote %q -> %s\n", cli.Remote.Add.Name, cli.Remote.Add.URL)
	default:
		fmt.Println("No subcommand given. Try --help.")
	}
}
