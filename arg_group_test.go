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
