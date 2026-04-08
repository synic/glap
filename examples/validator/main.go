package main

import (
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/synic/glap"
)

func validatePort(val string) error {
	n, err := strconv.Atoi(val)
	if err != nil {
		return fmt.Errorf("must be a number")
	}
	if n < 1 || n > 65535 {
		return fmt.Errorf("must be between 1 and 65535")
	}
	return nil
}

func validateIP(val string) error {
	if net.ParseIP(val) == nil {
		return fmt.Errorf("must be a valid IP address")
	}
	return nil
}

func main() {
	app := glap.NewCommand("validator").
		About("Demonstrates custom validator callbacks on arguments").
		Arg(glap.NewArg("port").
			Short('p').
			Default("8080").
			Help("Port to listen on").
			Validator(validatePort)).
		Arg(glap.NewArg("bind").
			Short('b').
			Default("127.0.0.1").
			Help("IP address to bind to").
			Validator(validateIP))

	matches, err := app.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	port, _ := matches.GetString("port")
	bind, _ := matches.GetString("bind")
	fmt.Printf("Listening on %s:%s\n", bind, port)
}
