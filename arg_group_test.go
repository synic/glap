package glap

import (
	"errors"
	"testing"
)

func TestMutuallyExclusiveGroup(t *testing.T) {
	app := NewCommand("myapp").
		Arg(NewArg("json").Action(SetTrue).Group("format")).
		Arg(NewArg("text").Action(SetTrue).Group("format")).
		ArgGroup(NewArgGroup("format").Arg("json").Arg("text"))

	_, err := app.Parse([]string{"--json", "--text"})
	if err == nil {
		t.Fatal("expected group violation error")
	}
	var gv *GroupViolationError
	if !errors.As(err, &gv) {
		t.Errorf("expected GroupViolationError, got %T: %v", err, err)
	}
}

func TestRequiredGroup(t *testing.T) {
	app := NewCommand("myapp").
		Arg(NewArg("json").Action(SetTrue)).
		Arg(NewArg("text").Action(SetTrue)).
		ArgGroup(NewArgGroup("format").Arg("json").Arg("text").Required(true))

	_, err := app.Parse([]string{})
	if err == nil {
		t.Fatal("expected group violation error for required group")
	}
	var gv *GroupViolationError
	if !errors.As(err, &gv) {
		t.Errorf("expected GroupViolationError, got %T: %v", err, err)
	}
}

func TestArgGroupMembershipViaArgGroup(t *testing.T) {
	app := NewCommand("myapp").
		Arg(NewArg("json").Action(SetTrue).Group("format")).
		Arg(NewArg("text").Action(SetTrue).Group("format")).
		ArgGroup(NewArgGroup("format"))

	_, err := app.Parse([]string{"--json", "--text"})
	if err == nil {
		t.Fatal("expected GroupViolationError for mutually exclusive args")
	}
	var gv *GroupViolationError
	if !errors.As(err, &gv) {
		t.Errorf("expected GroupViolationError, got %T: %v", err, err)
	}
}

func TestArgGroupRequiredViaArgGroup(t *testing.T) {
	app := NewCommand("myapp").
		Arg(NewArg("json").Action(SetTrue).Group("format")).
		Arg(NewArg("text").Action(SetTrue).Group("format")).
		ArgGroup(NewArgGroup("format").Required(true))

	_, err := app.Parse([]string{})
	if err == nil {
		t.Fatal("expected GroupViolationError for required group")
	}
	var gv *GroupViolationError
	if !errors.As(err, &gv) {
		t.Errorf("expected GroupViolationError, got %T: %v", err, err)
	}
}

func TestArgGroupMixedMembership(t *testing.T) {
	app := NewCommand("myapp").
		Arg(NewArg("json").Action(SetTrue)).
		Arg(NewArg("text").Action(SetTrue).Group("format")).
		ArgGroup(NewArgGroup("format").Arg("json"))

	_, err := app.Parse([]string{"--json", "--text"})
	if err == nil {
		t.Fatal("expected GroupViolationError — both members should be recognized")
	}
	var gv *GroupViolationError
	if !errors.As(err, &gv) {
		t.Errorf("expected GroupViolationError, got %T: %v", err, err)
	}
}

func TestArgGroupUndefinedGroupError(t *testing.T) {
	app := NewCommand("myapp").
		Arg(NewArg("json").Action(SetTrue).Group("nonexistent"))

	_, err := app.Parse([]string{"--json"})
	if err == nil {
		t.Fatal("expected UndefinedGroupError")
	}
	var ue *UndefinedGroupError
	if !errors.As(err, &ue) {
		t.Fatalf("expected UndefinedGroupError, got %T: %v", err, err)
	}
	if ue.Arg != "json" || ue.Group != "nonexistent" {
		t.Errorf("UndefinedGroupError fields = {Arg:%q, Group:%q}, want {json, nonexistent}", ue.Arg, ue.Group)
	}
}

func TestArgGroupStructTagMembership(t *testing.T) {
	type CLI struct {
		JSON bool `glap:"json,group=format"`
		Text bool `glap:"text,group=format"`
	}
	var cli CLI
	app := New(&cli).ArgGroup(NewArgGroup("format"))

	_, err := app.Parse([]string{"--json", "--text"})
	if err == nil {
		t.Fatal("expected GroupViolationError")
	}
	var gv *GroupViolationError
	if !errors.As(err, &gv) {
		t.Errorf("expected GroupViolationError, got %T: %v", err, err)
	}
}
