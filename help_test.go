package glap

import (
	"errors"
	"strings"
	"testing"
)

func TestHelpRequested(t *testing.T) {
	type CLI struct {
		Port int `glap:"port,short=p,default=8080"`
	}

	var cli CLI
	_, err := Parse(&cli, []string{"--help"})
	if err == nil {
		t.Fatal("expected HelpRequestedError")
	}
	var helpErr *HelpRequestedError
	if !errors.As(err, &helpErr) {
		t.Errorf("expected HelpRequestedError, got %T: %v", err, err)
	}
}

func TestHelpShort(t *testing.T) {
	type CLI struct {
		Port int `glap:"port,short=p"`
	}

	var cli CLI
	_, err := Parse(&cli, []string{"-h"})
	var helpErr *HelpRequestedError
	if !errors.As(err, &helpErr) {
		t.Errorf("expected HelpRequestedError, got %T: %v", err, err)
	}
}

func TestVersionRequested(t *testing.T) {
	type CLI struct {
		Port int `glap:"port"`
	}

	var cli CLI
	app := New(&cli).Name("myapp").Version("1.0.0")
	_, err := app.Parse([]string{"--version"})
	if err == nil {
		t.Fatal("expected VersionRequestedError")
	}
	var vErr *VersionRequestedError
	if !errors.As(err, &vErr) {
		t.Errorf("expected VersionRequestedError, got %T: %v", err, err)
	}
	if !strings.Contains(vErr.Message, "1.0.0") {
		t.Errorf("version message = %q, should contain 1.0.0", vErr.Message)
	}
}

func TestHelpOutput(t *testing.T) {
	app := NewCommand("myapp").
		Version("1.0.0").
		Author("Test Author").
		About("A test app").
		Arg(NewArg("config").Short('c').Help("Path to config file").Env("MYAPP_CONFIG").Required(true)).
		Arg(NewArg("verbose").Short('v').Help("Enable verbose output").Action(SetTrue)).
		Subcommand(NewCommand("serve").About("Start the server"))

	_, err := app.Parse([]string{"--help"})
	var helpErr *HelpRequestedError
	if !errors.As(err, &helpErr) {
		t.Fatalf("expected HelpRequestedError, got %T: %v", err, err)
	}

	msg := stripANSI(helpErr.Message)
	checks := []string{
		"myapp 1.0.0",
		"Test Author",
		"A test app",
		"USAGE:",
		"OPTIONS:",
		"-c, --config",
		"-v, --verbose",
		"[env: MYAPP_CONFIG]",
		"[required]",
		"SUBCOMMANDS:",
		"serve",
	}
	for _, check := range checks {
		if !strings.Contains(msg, check) {
			t.Errorf("help output missing %q:\n%s", check, msg)
		}
	}
}

func TestHiddenArg(t *testing.T) {
	app := NewCommand("myapp").
		Arg(NewArg("secret").Hidden(true))

	_, err := app.Parse([]string{"--help"})
	var helpErr *HelpRequestedError
	if !errors.As(err, &helpErr) {
		t.Fatal("expected help error")
	}
	if strings.Contains(stripANSI(helpErr.Message), "secret") {
		t.Error("hidden arg should not appear in help")
	}
}

func TestLongHelpVsShortHelp(t *testing.T) {
	app := NewCommand("myapp").
		About("Short description").
		LongAbout("This is the long, detailed description of the app.").
		Arg(NewArg("config").Short('c').Help("Config file").LongHelp("Path to the configuration file. Supports YAML and JSON formats."))

	_, err := app.Parse([]string{"-h"})
	var helpErr *HelpRequestedError
	if !errors.As(err, &helpErr) {
		t.Fatal("expected HelpRequestedError from -h")
	}
	shortMsg := stripANSI(helpErr.Message)
	if strings.Contains(shortMsg, "long, detailed") {
		t.Error("-h should show short help, not long")
	}
	if !strings.Contains(shortMsg, "Short description") {
		t.Error("-h should contain short about")
	}

	_, err = app.Parse([]string{"--help"})
	if !errors.As(err, &helpErr) {
		t.Fatal("expected HelpRequestedError from --help")
	}
	longMsg := stripANSI(helpErr.Message)
	if !strings.Contains(longMsg, "long, detailed") {
		t.Error("--help should show long about")
	}
	if !strings.Contains(longMsg, "Supports YAML and JSON") {
		t.Error("--help should show long help for args")
	}
}

func TestDisplayOrder(t *testing.T) {
	app := NewCommand("myapp").
		Arg(NewArg("zebra").Action(SetTrue).DisplayOrder(3)).
		Arg(NewArg("alpha").Action(SetTrue).DisplayOrder(1)).
		Arg(NewArg("middle").Action(SetTrue).DisplayOrder(2))

	_, err := app.Parse([]string{"--help"})
	var helpErr *HelpRequestedError
	if !errors.As(err, &helpErr) {
		t.Fatal("expected HelpRequestedError")
	}

	msg := stripANSI(helpErr.Message)
	alphaIdx := strings.Index(msg, "--alpha")
	middleIdx := strings.Index(msg, "--middle")
	zebraIdx := strings.Index(msg, "--zebra")
	if alphaIdx > middleIdx || middleIdx > zebraIdx {
		t.Errorf("args not in display order: alpha=%d middle=%d zebra=%d", alphaIdx, middleIdx, zebraIdx)
	}
}

func TestHideDefaultValue(t *testing.T) {
	app := NewCommand("myapp").
		Arg(NewArg("port").Default("8080").HideDefaultValue(true).Help("Port number")).
		Arg(NewArg("host").Default("localhost").Help("Host name"))

	_, err := app.Parse([]string{"--help"})
	var helpErr *HelpRequestedError
	if !errors.As(err, &helpErr) {
		t.Fatal("expected HelpRequestedError")
	}

	msg := stripANSI(helpErr.Message)
	if strings.Contains(msg, "default: 8080") {
		t.Error("port default should be hidden")
	}
	if !strings.Contains(msg, "default: localhost") {
		t.Error("host default should be visible")
	}
}

func TestHideDefaultValueStructTag(t *testing.T) {
	type CLI struct {
		Port int `glap:"port,default=8080,hide_default_value,help=Port"`
	}
	var cli CLI
	app := New(&cli).Name("myapp")
	_, err := app.Parse([]string{"--help"})
	var helpErr *HelpRequestedError
	if !errors.As(err, &helpErr) {
		t.Fatal("expected HelpRequestedError")
	}
	if strings.Contains(stripANSI(helpErr.Message), "default: 8080") {
		t.Error("port default should be hidden")
	}
}

func TestHelpShowsPositionalArgs(t *testing.T) {
	app := NewCommand("myapp").
		Arg(NewArg("input").Positional(true).Required(true).Help("Path to input file")).
		Arg(NewArg("output").Positional(true).Help("Path to output file"))

	_, err := app.Parse([]string{"--help"})
	var helpErr *HelpRequestedError
	if !errors.As(err, &helpErr) {
		t.Fatal("expected HelpRequestedError")
	}

	msg := stripANSI(helpErr.Message)
	checks := []string{
		"ARGS:",
		"<INPUT>",
		"Path to input file",
		"[OUTPUT]",
		"Path to output file",
	}
	for _, check := range checks {
		if !strings.Contains(msg, check) {
			t.Errorf("help output missing %q:\n%s", check, msg)
		}
	}
}

func TestHelpHidesHiddenPositional(t *testing.T) {
	app := NewCommand("myapp").
		Arg(NewArg("visible").Positional(true).Help("Visible positional")).
		Arg(NewArg("secret").Positional(true).Hidden(true).Help("Secret positional"))

	_, err := app.Parse([]string{"--help"})
	var helpErr *HelpRequestedError
	if !errors.As(err, &helpErr) {
		t.Fatal("expected HelpRequestedError")
	}

	msg := stripANSI(helpErr.Message)
	if !strings.Contains(msg, "Visible positional") {
		t.Error("visible positional should appear in help")
	}
	if strings.Contains(msg, "Secret positional") {
		t.Error("hidden positional should not appear in help")
	}
}

func TestHelpShowsArgAliases(t *testing.T) {
	app := NewCommand("myapp").
		Arg(NewArg("config").Short('c').Alias("conf").Alias("cfg").Help("Config file"))

	_, err := app.Parse([]string{"--help"})
	var helpErr *HelpRequestedError
	if !errors.As(err, &helpErr) {
		t.Fatal("expected HelpRequestedError")
	}

	msg := stripANSI(helpErr.Message)
	if !strings.Contains(msg, "--config") {
		t.Errorf("help missing --config:\n%s", msg)
	}
	if !strings.Contains(msg, "--conf") {
		t.Errorf("help missing --conf alias:\n%s", msg)
	}
	if !strings.Contains(msg, "--cfg") {
		t.Errorf("help missing --cfg alias:\n%s", msg)
	}
}

func TestHelpShowsSubcommandAliases(t *testing.T) {
	app := NewCommand("myapp").
		Subcommand(NewCommand("serve").Alias("s").Alias("server").About("Start the server"))

	_, err := app.Parse([]string{"--help"})
	var helpErr *HelpRequestedError
	if !errors.As(err, &helpErr) {
		t.Fatal("expected HelpRequestedError")
	}

	msg := stripANSI(helpErr.Message)
	if !strings.Contains(msg, "serve") {
		t.Errorf("help missing serve:\n%s", msg)
	}
	if !strings.Contains(msg, ", s") {
		t.Errorf("help missing 's' alias:\n%s", msg)
	}
	if !strings.Contains(msg, "server") {
		t.Errorf("help missing 'server' alias:\n%s", msg)
	}
}

func TestHelpShowsArgGroup(t *testing.T) {
	app := NewCommand("myapp").
		Arg(NewArg("json").Action(SetTrue).Group("format").Help("JSON output")).
		Arg(NewArg("text").Action(SetTrue).Group("format").Help("Text output")).
		ArgGroup(NewArgGroup("format"))

	_, err := app.Parse([]string{"--help"})
	var helpErr *HelpRequestedError
	if !errors.As(err, &helpErr) {
		t.Fatal("expected HelpRequestedError")
	}

	msg := stripANSI(helpErr.Message)
	if strings.Count(msg, "group: format") != 2 {
		t.Errorf("expected 'group: format' annotation on both args, got:\n%s", msg)
	}
}

func TestStructTagEscapedCommaInHelp(t *testing.T) {
	type CLI struct {
		Name string `glap:"name,help=Last\\, First"`
	}

	var cli CLI
	_, err := Parse(&cli, []string{"--help"})
	var helpErr *HelpRequestedError
	if !errors.As(err, &helpErr) {
		t.Fatal("expected HelpRequestedError")
	}

	msg := stripANSI(helpErr.Message)
	if !strings.Contains(msg, "Last, First") {
		t.Fatalf("help missing escaped comma value:\n%s", msg)
	}
}
