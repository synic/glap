package glap

import "testing"

func TestClone(t *testing.T) {
	orig := NewArg("port").
		Short('p').
		Default("8080").
		Help("Port number").
		PossibleValues("80", "443", "8080").
		ConflictsWith("socket").
		Requires("host")

	clone := orig.Clone()

	if clone.GetName() != "port" {
		t.Errorf("name = %q", clone.GetName())
	}
	if clone.GetShort() != 'p' {
		t.Errorf("short = %c", clone.GetShort())
	}
	if clone.GetDefault() != "8080" {
		t.Errorf("default = %q", clone.GetDefault())
	}

	clone.Default("3000")
	if orig.GetDefault() != "8080" {
		t.Error("mutating clone should not affect original")
	}

	clone.possibleValues[0] = "changed"
	if orig.possibleValues[0] != "80" {
		t.Error("mutating clone slice should not affect original")
	}
}

func TestCloneIndependentCommands(t *testing.T) {
	shared := NewArg("port").Short('p').Default("8080")

	app := NewCommand("myapp").
		Subcommand(NewCommand("add").Arg(shared.Clone())).
		Subcommand(NewCommand("update").Arg(shared.Clone()))

	m, err := app.Parse([]string{"add", "--port", "3000"})
	if err != nil {
		t.Fatal(err)
	}
	port, _ := m.SubcommandMatches().GetInt("port")
	if port != 3000 {
		t.Errorf("port = %d, want 3000", port)
	}

	m, err = app.Parse([]string{"update"})
	if err != nil {
		t.Fatal(err)
	}
	port, _ = m.SubcommandMatches().GetInt("port")
	if port != 8080 {
		t.Errorf("port = %d, want 8080 (default)", port)
	}
}
