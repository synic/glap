package glap

import (
	"bytes"
	"strings"
	"testing"
)

func testCommand() *Command {
	cmd := NewCommand("myapp").
		Version("1.0.0").
		Arg(NewArg("config").Short('c').Help("Config file").Required(true)).
		Arg(NewArg("verbose").Short('v').Help("Enable verbose output").Action(SetTrue)).
		Arg(NewArg("output").Short('o').Help("Output format").PossibleValues("json", "text", "yaml")).
		Subcommand(NewCommand("serve").About("Start the server").
			Arg(NewArg("port").Short('p').Help("Port number").Default("8080")))
	cmd.injectHelpAndVersion()
	return cmd
}

func TestBashCompletion(t *testing.T) {
	cmd := testCommand()
	out := GenerateCompletion(cmd, Bash)

	checks := []string{
		"_myapp()",
		"complete -F _myapp myapp",
		"--config",
		"--verbose",
		"-c",
		"-v",
		"serve",
		"compgen",
		"COMPREPLY",
		"json text yaml",
	}
	for _, check := range checks {
		if !strings.Contains(out, check) {
			t.Errorf("bash completion missing %q", check)
		}
	}
}

func TestZshCompletion(t *testing.T) {
	cmd := testCommand()
	out := GenerateCompletion(cmd, Zsh)

	checks := []string{
		"#compdef myapp",
		"_myapp()",
		"_arguments",
		"--config",
		"--verbose",
		"serve",
		"_describe",
		"_myapp_serve",
	}
	for _, check := range checks {
		if !strings.Contains(out, check) {
			t.Errorf("zsh completion missing %q", check)
		}
	}
}

func TestFishCompletion(t *testing.T) {
	cmd := testCommand()
	out := GenerateCompletion(cmd, Fish)

	checks := []string{
		"complete -c myapp",
		"-s c",
		"-l config",
		"-l verbose",
		"__fish_use_subcommand",
		"serve",
		"__fish_seen_subcommand_from serve",
		"-l port",
	}
	for _, check := range checks {
		if !strings.Contains(out, check) {
			t.Errorf("fish completion missing %q", check)
		}
	}
}

func TestPowerShellCompletion(t *testing.T) {
	cmd := testCommand()
	out := GenerateCompletion(cmd, PowerShell)

	checks := []string{
		"Register-ArgumentCompleter",
		"-CommandName myapp",
		"CompletionResult",
		"--config",
		"--verbose",
		"-c",
		"serve",
	}
	for _, check := range checks {
		if !strings.Contains(out, check) {
			t.Errorf("powershell completion missing %q", check)
		}
	}
}

func TestCompletionPossibleValues(t *testing.T) {
	cmd := testCommand()
	out := GenerateCompletion(cmd, Fish)

	if !strings.Contains(out, "json text yaml") {
		t.Error("fish completion should include possible values for output")
	}
}

func TestValueHintBashCompletion(t *testing.T) {
	cmd := NewCommand("myapp").
		Arg(NewArg("config").SetValueHint(HintFilePath)).
		Arg(NewArg("output-dir").SetValueHint(HintDirPath))
	cmd.injectHelpAndVersion()

	out := GenerateCompletion(cmd, Bash)
	if !strings.Contains(out, "compgen -f") {
		t.Error("bash completion should use compgen -f for FilePath hint")
	}
	if !strings.Contains(out, "compgen -d") {
		t.Error("bash completion should use compgen -d for DirPath hint")
	}
}

func TestValueHintZshCompletion(t *testing.T) {
	cmd := NewCommand("myapp").
		Arg(NewArg("config").SetValueHint(HintFilePath)).
		Arg(NewArg("host").SetValueHint(HintHostname))
	cmd.injectHelpAndVersion()

	out := GenerateCompletion(cmd, Zsh)
	if !strings.Contains(out, "_files") {
		t.Error("zsh completion should use _files for FilePath hint")
	}
	if !strings.Contains(out, "_hosts") {
		t.Error("zsh completion should use _hosts for Hostname hint")
	}
}

func TestValueHintFishCompletion(t *testing.T) {
	cmd := NewCommand("myapp").
		Arg(NewArg("config").SetValueHint(HintFilePath)).
		Arg(NewArg("user").SetValueHint(HintUsername))
	cmd.injectHelpAndVersion()

	out := GenerateCompletion(cmd, Fish)
	if !strings.Contains(out, "-F") {
		t.Error("fish completion should use -F for FilePath hint")
	}
	if !strings.Contains(out, "__fish_complete_users") {
		t.Error("fish completion should use __fish_complete_users for Username hint")
	}
}

func TestValueHintStructTag(t *testing.T) {
	type CLI struct {
		Config string `glap:"config,value_hint=file_path"`
	}
	var cli CLI
	app := New(&cli).Name("myapp")
	out, err := app.GenerateCompletion(Bash)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "compgen -f") {
		t.Error("bash completion should include file hint from struct tag")
	}
}

func TestCompleteCommand(t *testing.T) {
	t.Setenv("COMPLETE", "bash")

	cmd := NewCommand("myapp").
		Arg(NewArg("config").Short('c'))
	var buf bytes.Buffer
	ok := CompleteCommand(cmd, &buf)
	if !ok {
		t.Fatal("CompleteCommand should return true when COMPLETE is set")
	}
	if !strings.Contains(buf.String(), "_myapp()") {
		t.Error("should contain bash completion function")
	}
}

func TestCompleteCommandNotSet(t *testing.T) {
	cmd := NewCommand("myapp")
	var buf bytes.Buffer
	ok := CompleteCommand(cmd, &buf)
	if ok {
		t.Error("CompleteCommand should return false when COMPLETE is not set")
	}
	if buf.Len() != 0 {
		t.Error("should not write anything")
	}
}

func TestCompleteCommandInvalidShell(t *testing.T) {
	t.Setenv("COMPLETE", "nushell")

	cmd := NewCommand("myapp")
	var buf bytes.Buffer
	ok := CompleteCommand(cmd, &buf)
	if ok {
		t.Error("CompleteCommand should return false for unknown shell")
	}
}

func TestCompleteApp(t *testing.T) {
	t.Setenv("COMPLETE", "zsh")

	type CLI struct {
		Config string `glap:"config,short=c"`
	}
	var cli CLI
	app := New(&cli).Name("myapp")
	var buf bytes.Buffer
	ok := CompleteApp(app, &buf)
	if !ok {
		t.Fatal("CompleteApp should return true when COMPLETE is set")
	}
	if !strings.Contains(buf.String(), "#compdef myapp") {
		t.Error("should contain zsh compdef header")
	}
}

func TestCompleteAppNotSet(t *testing.T) {
	type CLI struct{}
	var cli CLI
	app := New(&cli).Name("myapp")
	var buf bytes.Buffer
	ok := CompleteApp(app, &buf)
	if ok {
		t.Error("CompleteApp should return false when COMPLETE is not set")
	}
}
