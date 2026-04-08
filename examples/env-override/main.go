package main

import (
	"fmt"
	"os"

	"github.com/synic/glap"
)

// Precedence: CLI flags > environment variables > defaults.
// Try: APP_PORT=9090 APP_HOST=0.0.0.0 go run . --port 3000
// Port will be 3000 (CLI wins), host will be 0.0.0.0 (env wins over default).
type CLI struct {
	Host    string `glap:"host,short=H,default=localhost,env=APP_HOST,help=Bind address"`
	Port    int    `glap:"port,short=p,default=8080,env=APP_PORT,help=Port to listen on"`
	Debug   bool   `glap:"debug,env=APP_DEBUG,help=Enable debug mode"`
	Secret  string `glap:"secret,env=APP_SECRET,help=API secret key,value_name=KEY"`
}

func main() {
	var cli CLI
	app := glap.New(&cli).
		Name("env-override").
		About("Demonstrates environment variable overrides")

	_, err := app.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("Host:   %s\n", cli.Host)
	fmt.Printf("Port:   %d\n", cli.Port)
	fmt.Printf("Debug:  %v\n", cli.Debug)
	fmt.Printf("Secret: %s\n", cli.Secret)
}
