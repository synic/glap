package glap

import (
	"fmt"
	"testing"
)

func TestArgGetters(t *testing.T) {
	a := NewArg("config").Short('c').Help("Config file").Env("CFG").
		Default("app.yaml").Action(Set).PossibleValues("a", "b").
		Required(true).Hidden(false).Global(true)

	if a.GetName() != "config" {
		t.Errorf("GetName = %q", a.GetName())
	}
	if a.GetShort() != 'c' {
		t.Errorf("GetShort = %c", a.GetShort())
	}
	if a.GetHelp() != "Config file" {
		t.Errorf("GetHelp = %q", a.GetHelp())
	}
	if a.GetEnv() != "CFG" {
		t.Errorf("GetEnv = %q", a.GetEnv())
	}
	if a.GetDefault() != "app.yaml" {
		t.Errorf("GetDefault = %q", a.GetDefault())
	}
	if a.GetAction() != Set {
		t.Errorf("GetAction = %v", a.GetAction())
	}
	if len(a.GetPossibleValues()) != 2 {
		t.Errorf("GetPossibleValues = %v", a.GetPossibleValues())
	}
	if !a.IsRequired() {
		t.Error("IsRequired should be true")
	}
	if !a.IsGlobal() {
		t.Error("IsGlobal should be true")
	}
}

func TestCommandGetters(t *testing.T) {
	cmd := NewCommand("myapp").Version("1.0").About("desc").Author("me").
		Arg(NewArg("flag").Action(SetTrue)).
		Subcommand(NewCommand("sub"))

	if cmd.GetName() != "myapp" {
		t.Errorf("GetName = %q", cmd.GetName())
	}
	if cmd.GetVersion() != "1.0" {
		t.Errorf("GetVersion = %q", cmd.GetVersion())
	}
	if len(cmd.GetArgs()) != 1 {
		t.Errorf("GetArgs len = %d", len(cmd.GetArgs()))
	}
	if len(cmd.GetSubcommands()) != 1 {
		t.Errorf("GetSubcommands len = %d", len(cmd.GetSubcommands()))
	}
	if cmd.FindArg("flag") == nil {
		t.Error("FindArg should find 'flag'")
	}
	if cmd.FindSubcommand("sub") == nil {
		t.Error("FindSubcommand should find 'sub'")
	}
}

func TestMutArg(t *testing.T) {
	app := NewCommand("myapp").
		Arg(NewArg("port").Default("8080")).
		MutArg("port", func(a *Arg) {
			a.Default("3000")
		})

	m, err := app.Parse([]string{})
	if err != nil {
		t.Fatal(err)
	}
	port, _ := m.GetInt("port")
	if port != 3000 {
		t.Errorf("port = %d, want 3000", port)
	}
}

func TestMutSubcommand(t *testing.T) {
	app := NewCommand("myapp").
		Subcommand(NewCommand("serve").About("old")).
		MutSubcommand("serve", func(c *Command) {
			c.About("new description")
		})

	if app.FindSubcommand("serve").GetAbout() != "new description" {
		t.Error("MutSubcommand should have updated about")
	}
}

func TestRunCallback(t *testing.T) {
	var called string
	app := NewCommand("myapp").
		Subcommand(NewCommand("serve").
			Arg(NewArg("port").Default("8080")).
			Run(func(m *Matches) error {
				port, _ := m.GetString("port")
				called = "serve:" + port
				return nil
			}))

	_, err := app.Parse([]string{"serve", "--port", "3000"})
	if err != nil {
		t.Fatal(err)
	}
	if called != "serve:3000" {
		t.Errorf("called = %q, want %q", called, "serve:3000")
	}
}

func TestRunCallbackNested(t *testing.T) {
	var called string
	app := NewCommand("myapp").
		Subcommand(NewCommand("remote").
			Subcommand(NewCommand("add").
				Arg(NewArg("name").Positional(true)).
				Run(func(m *Matches) error {
					name, _ := m.GetString("name")
					called = "remote-add:" + name
					return nil
				})))

	_, err := app.Parse([]string{"remote", "add", "origin"})
	if err != nil {
		t.Fatal(err)
	}
	if called != "remote-add:origin" {
		t.Errorf("called = %q, want %q", called, "remote-add:origin")
	}
}

func TestRunCallbackError(t *testing.T) {
	app := NewCommand("myapp").
		Run(func(m *Matches) error {
			return fmt.Errorf("handler failed")
		})

	_, err := app.Parse([]string{})
	if err == nil || err.Error() != "handler failed" {
		t.Errorf("err = %v, want 'handler failed'", err)
	}
}

func TestRunCallbackCalledForParent(t *testing.T) {
	parentCalled := false
	app := NewCommand("myapp").
		Run(func(m *Matches) error {
			parentCalled = true
			return nil
		}).
		Subcommand(NewCommand("sub").
			Run(func(m *Matches) error { return nil }))

	_, err := app.Parse([]string{"sub"})
	if err != nil {
		t.Fatal(err)
	}
	if !parentCalled {
		t.Error("parent Run should be called before subcommand Run")
	}
}

func TestAppArg(t *testing.T) {
	type CLI struct {
		Verbose bool `glap:"verbose,short=v"`
	}
	var cli CLI
	app := New(&cli).
		Name("myapp").
		Arg(NewArg("extra").Default("hello"))

	_, err := app.Parse([]string{"-v"})
	if err != nil {
		t.Fatal(err)
	}
	if !cli.Verbose {
		t.Error("Verbose should be true")
	}
}

func TestAppSubcommand(t *testing.T) {
	type CLI struct {
		Verbose bool `glap:"verbose,short=v"`
	}
	var cli CLI

	var called bool
	app := New(&cli).
		Name("myapp").
		Subcommand(NewCommand("dynamic").
			Arg(NewArg("flag").Action(SetTrue)).
			Run(func(m *Matches) error {
				called = true
				return nil
			}))

	cmd, err := app.Parse([]string{"dynamic", "--flag"})
	if err != nil {
		t.Fatal(err)
	}
	if cmd != "dynamic" {
		t.Errorf("cmd = %q, want %q", cmd, "dynamic")
	}
	if !called {
		t.Error("dynamic subcommand Run should have been called")
	}
}
