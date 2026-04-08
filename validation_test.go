package glap

import (
	"errors"
	"testing"
)

func TestRequiredArg(t *testing.T) {
	type CLI struct {
		Config string `glap:"config,required"`
	}

	var cli CLI
	_, err := Parse(&cli, []string{})
	if err == nil {
		t.Fatal("expected error for missing required arg")
	}
	var missing *MissingRequiredError
	if !errors.As(err, &missing) {
		t.Errorf("expected MissingRequiredError, got %T: %v", err, err)
	}
}

func TestPossibleValues(t *testing.T) {
	type CLI struct {
		Output string `glap:"output,possible=json|text|yaml,default=text"`
	}

	var cli CLI
	_, err := Parse(&cli, []string{"--output", "json"})
	if err != nil {
		t.Fatal(err)
	}
	if cli.Output != "json" {
		t.Errorf("Output = %q, want %q", cli.Output, "json")
	}

	_, err = Parse(&cli, []string{"--output", "xml"})
	if err == nil {
		t.Fatal("expected error for invalid possible value")
	}
	var invalid *InvalidValueError
	if !errors.As(err, &invalid) {
		t.Errorf("expected InvalidValueError, got %T: %v", err, err)
	}
}

func TestConflictsWith(t *testing.T) {
	type CLI struct {
		JSON bool `glap:"json,conflicts_with=text"`
		Text bool `glap:"text,conflicts_with=json"`
	}

	var cli CLI
	_, err := Parse(&cli, []string{"--json", "--text"})
	if err == nil {
		t.Fatal("expected conflict error")
	}
	var conflict *ConflictError
	if !errors.As(err, &conflict) {
		t.Errorf("expected ConflictError, got %T: %v", err, err)
	}
}

func TestRequires(t *testing.T) {
	type CLI struct {
		Output string `glap:"output,requires=format"`
		Format string `glap:"format"`
	}

	var cli CLI
	_, err := Parse(&cli, []string{"--output", "file.txt"})
	if err == nil {
		t.Fatal("expected missing dependency error")
	}
	var dep *MissingDependencyError
	if !errors.As(err, &dep) {
		t.Errorf("expected MissingDependencyError, got %T: %v", err, err)
	}

	_, err = Parse(&cli, []string{"--output", "file.txt", "--format", "json"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestRequiredIfEq(t *testing.T) {
	app := NewCommand("myapp").
		Arg(NewArg("format").Default("text")).
		Arg(NewArg("output").RequiredIfEq("format", "json"))

	_, err := app.Parse([]string{"--format", "json"})
	if err == nil {
		t.Fatal("expected ConditionalRequiredError")
	}
	var condErr *ConditionalRequiredError
	if !errors.As(err, &condErr) {
		t.Errorf("expected ConditionalRequiredError, got %T: %v", err, err)
	}

	_, err = app.Parse([]string{"--format", "text"})
	if err != nil {
		t.Fatal(err)
	}

	_, err = app.Parse([]string{"--format", "json", "--output", "file.json"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestRequiredUnlessPresent(t *testing.T) {
	app := NewCommand("myapp").
		Arg(NewArg("stdin").Action(SetTrue)).
		Arg(NewArg("input").RequiredUnlessPresent("stdin"))

	_, err := app.Parse([]string{})
	if err == nil {
		t.Fatal("expected ConditionalRequiredError")
	}
	var condErr *ConditionalRequiredError
	if !errors.As(err, &condErr) {
		t.Errorf("expected ConditionalRequiredError, got %T: %v", err, err)
	}

	_, err = app.Parse([]string{"--stdin"})
	if err != nil {
		t.Fatal(err)
	}

	_, err = app.Parse([]string{"--input", "file.txt"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestDefaultValueIf(t *testing.T) {
	app := NewCommand("myapp").
		Arg(NewArg("mode")).
		Arg(NewArg("port").DefaultValueIf("mode", "server", "8080"))

	m, err := app.Parse([]string{"--mode", "server"})
	if err != nil {
		t.Fatal(err)
	}
	port, ok := m.GetInt("port")
	if !ok || port != 8080 {
		t.Errorf("port = %d, want 8080", port)
	}

	m, err = app.Parse([]string{"--mode", "client"})
	if err != nil {
		t.Fatal(err)
	}
	if m.Contains("port") {
		t.Error("port should not be set when mode != server")
	}
}

func TestRequiredIfEqStructTag(t *testing.T) {
	type CLI struct {
		Format string `glap:"format,default=text"`
		Output string `glap:"output,required_if_eq=format:json"`
	}
	var cli CLI
	_, err := Parse(&cli, []string{"--format", "json"})
	if err == nil {
		t.Fatal("expected ConditionalRequiredError")
	}

	_, err = Parse(&cli, []string{"--format", "json", "--output", "out.json"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestUnknownArg(t *testing.T) {
	type CLI struct {
		Port int `glap:"port"`
	}

	var cli CLI
	_, err := Parse(&cli, []string{"--unknown"})
	if err == nil {
		t.Fatal("expected unknown arg error")
	}
	var unknown *UnknownArgError
	if !errors.As(err, &unknown) {
		t.Errorf("expected UnknownArgError, got %T: %v", err, err)
	}
}
