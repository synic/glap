package glap

import (
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
