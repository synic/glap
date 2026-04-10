package glap

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestAppWrapper(t *testing.T) {
	type CLI struct {
		Port int `glap:"port,default=8080"`
	}

	var cli CLI
	app := New(&cli).
		Name("myapp").
		Version("1.0.0").
		About("My cool app").
		Author("Adam Olsen")

	_, err := app.Parse([]string{})
	if err != nil {
		t.Fatal(err)
	}
	if cli.Port != 8080 {
		t.Errorf("Port = %d, want %d", cli.Port, 8080)
	}
}

func TestAppSubcommandRequired(t *testing.T) {
	type CLI struct{}
	var cli CLI
	app := New(&cli).
		Name("myapp").
		SubcommandRequired(true).
		Subcommand(NewCommand("sub"))

	_, err := app.Parse([]string{})
	if err == nil {
		t.Fatal("expected HelpRequestedError")
	}
	var helpErr *HelpRequestedError
	if !errors.As(err, &helpErr) {
		t.Errorf("expected HelpRequestedError, got %T: %v", err, err)
	}

	sub, err := app.Parse([]string{"sub"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if sub != "sub" {
		t.Errorf("subcommand name = %q, want %q", sub, "sub")
	}
}

func TestSubcommandStructTagOptions(t *testing.T) {
	type ServeCLI struct {
		Port int `glap:"port,default=8080"`
	}
	type CLI struct {
		Serve *ServeCLI `glap:"serve,subcommand,alias=s,hidden,display_order=2,help=Run server,long_help=Run the HTTP server,version=2.0.0,author=Dev Team,subcommand_required"`
	}

	var cli CLI
	app := New(&cli).Name("myapp")

	cmd, err := buildCommand(app.command, app.target)
	if err != nil {
		t.Fatalf("buildCommand failed: %v", err)
	}

	sub := cmd.findSubcommand("serve")
	if sub == nil {
		t.Fatal("serve subcommand not found")
	}
	if sub.about != "Run server" {
		t.Errorf("about = %q, want %q", sub.about, "Run server")
	}
	if sub.longAbout != "Run the HTTP server" {
		t.Errorf("longAbout = %q, want %q", sub.longAbout, "Run the HTTP server")
	}
	if sub.version != "2.0.0" {
		t.Errorf("version = %q, want %q", sub.version, "2.0.0")
	}
	if sub.author != "Dev Team" {
		t.Errorf("author = %q, want %q", sub.author, "Dev Team")
	}
	if !sub.hidden {
		t.Error("hidden should be true")
	}
	if sub.displayOrder != 2 {
		t.Errorf("displayOrder = %d, want %d", sub.displayOrder, 2)
	}
	if !sub.subcommandRequired {
		t.Error("subcommandRequired should be true")
	}
	if len(sub.aliases) == 0 {
		t.Fatal("expected aliases to be set")
	}
	foundS := false
	for _, a := range sub.aliases {
		if a == "s" {
			foundS = true
		}
	}
	if !foundS {
		t.Errorf("expected alias 's' in %v", sub.aliases)
	}

	if cmd.findSubcommand("s") == nil {
		t.Error("subcommand not reachable via alias 's'")
	}
}

func TestAppValidator(t *testing.T) {
	type CLI struct {
		Port int `glap:"port"`
	}

	var cli CLI
	app := New(&cli).
		Name("myapp").
		Validator("port", func(v string) error {
			if v == "0" {
				return fmt.Errorf("port must be non-zero")
			}
			return nil
		})

	_, err := app.Parse([]string{"--port", "0"})
	if err == nil {
		t.Fatal("expected validator error")
	}
	if !strings.Contains(err.Error(), "port must be non-zero") {
		t.Fatalf("validator error = %v", err)
	}
}

func TestAppParseDoesNotAccumulateReflectedArgs(t *testing.T) {
	type CLI struct {
		Verbose bool `glap:"verbose,short=v"`
	}

	var cli CLI
	app := New(&cli).Name("myapp")

	_, err := app.Parse([]string{"-v"})
	if err != nil {
		t.Fatal(err)
	}
	if !cli.Verbose {
		t.Fatal("verbose should be true after first parse")
	}

	cli.Verbose = false
	_, err = app.Parse([]string{})
	if err != nil {
		t.Fatal(err)
	}
	if cli.Verbose {
		t.Fatal("verbose should remain false on second parse")
	}
	if len(app.command.args) != 0 {
		t.Fatalf("template command should not retain reflected args, got %d", len(app.command.args))
	}
}

func TestStructTagAppendIntoTypedSlice(t *testing.T) {
	type CLI struct {
		Ports []int `glap:"port,action=append"`
	}

	var cli CLI
	_, err := Parse(&cli, []string{"--port", "80", "--port", "443"})
	if err != nil {
		t.Fatal(err)
	}
	if len(cli.Ports) != 2 || cli.Ports[0] != 80 || cli.Ports[1] != 443 {
		t.Fatalf("Ports = %v, want [80 443]", cli.Ports)
	}
}
