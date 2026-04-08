package glap

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// Shell represents a shell type for completion generation.
type Shell int

const (
	Bash Shell = iota
	Zsh
	Fish
	PowerShell
)

func parseShellName(s string) (Shell, bool) {
	switch strings.ToLower(s) {
	case "bash":
		return Bash, true
	case "zsh":
		return Zsh, true
	case "fish":
		return Fish, true
	case "powershell", "pwsh":
		return PowerShell, true
	default:
		return 0, false
	}
}

// CompleteCommand checks the COMPLETE environment variable and, if set,
// writes the appropriate shell completion script to w and returns true.
// Call this before Parse. If it returns true, the program should exit.
//
//	app := glap.NewCommand("myapp").
//	    Arg(glap.NewArg("config").Short('c'))
//	if glap.CompleteCommand(app, os.Stdout) {
//	    return
//	}
//	matches, err := app.Parse(os.Args[1:])
//
// Generate completions by setting COMPLETE to a shell name:
//
//	COMPLETE=bash myapp >> ~/.bashrc
//	COMPLETE=zsh myapp > ~/.zfunc/_myapp
func CompleteCommand(cmd *Command, w io.Writer) bool {
	val, ok := os.LookupEnv("COMPLETE")
	if !ok || val == "" {
		return false
	}
	shell, ok := parseShellName(val)
	if !ok {
		return false
	}
	cmd.injectHelpAndVersion()
	fmt.Fprint(w, GenerateCompletion(cmd, shell))
	return true
}

// CompleteApp checks the COMPLETE environment variable and, if set,
// builds the command tree from the App's struct tags, writes the
// appropriate shell completion script to w, and returns true.
// Call this before Parse. If it returns true, the program should exit.
//
//	app := glap.New(&cli).Name("myapp")
//	if glap.CompleteApp(app, os.Stdout) {
//	    return
//	}
//	cmd, err := app.Parse(os.Args[1:])
//
// Generate completions by setting COMPLETE to a shell name:
//
//	COMPLETE=bash myapp >> ~/.bashrc
//	COMPLETE=zsh myapp > ~/.zfunc/_myapp
func CompleteApp(app *App, w io.Writer) bool {
	val, ok := os.LookupEnv("COMPLETE")
	if !ok || val == "" {
		return false
	}
	shell, ok := parseShellName(val)
	if !ok {
		return false
	}
	out, err := app.GenerateCompletion(shell)
	if err != nil {
		return false
	}
	fmt.Fprint(w, out)
	return true
}

// GenerateCompletion generates a shell completion script for the given command.
func GenerateCompletion(cmd *Command, shell Shell) string {
	switch shell {
	case Bash:
		return generateBashCompletion(cmd)
	case Zsh:
		return generateZshCompletion(cmd)
	case Fish:
		return generateFishCompletion(cmd)
	case PowerShell:
		return generatePowerShellCompletion(cmd)
	default:
		return ""
	}
}

// GenerateCompletion builds the command tree from the App's struct tags and generates
// a shell completion script.
func (a *App) GenerateCompletion(shell Shell) (string, error) {
	cmd, err := buildCommand(a.command, a.target)
	if err != nil {
		return "", err
	}
	cmd.injectHelpAndVersion()
	return GenerateCompletion(cmd, shell), nil
}

func generateBashCompletion(cmd *Command) string {
	var b strings.Builder
	name := sanitizeName(cmd.name)

	b.WriteString(fmt.Sprintf("_%s() {\n", name))
	b.WriteString("    local cur prev opts cmds\n")
	b.WriteString("    COMPREPLY=()\n")
	b.WriteString("    cur=\"${COMP_WORDS[COMP_CWORD]}\"\n")
	b.WriteString("    prev=\"${COMP_WORDS[COMP_CWORD-1]}\"\n")
	b.WriteString("\n")

	var flags []string
	for _, a := range cmd.args {
		if a.hidden || a.positional {
			continue
		}
		if a.short != 0 {
			flags = append(flags, fmt.Sprintf("-%c", a.short))
		}
		flags = append(flags, "--"+a.long)
	}

	var subcmds []string
	for _, sub := range cmd.subcommands {
		if !sub.hidden {
			subcmds = append(subcmds, sub.name)
		}
	}

	b.WriteString(fmt.Sprintf("    opts=\"%s\"\n", strings.Join(flags, " ")))
	if len(subcmds) > 0 {
		b.WriteString(fmt.Sprintf("    cmds=\"%s\"\n", strings.Join(subcmds, " ")))
	}
	b.WriteString("\n")

	writeBashValueCompletions(&b, cmd)

	if len(subcmds) > 0 {
		b.WriteString("    if [[ ${cur} != -* ]]; then\n")
		b.WriteString("        COMPREPLY=( $(compgen -W \"${cmds}\" -- \"${cur}\") )\n")
		b.WriteString("        return 0\n")
		b.WriteString("    fi\n\n")
	}

	b.WriteString("    COMPREPLY=( $(compgen -W \"${opts}\" -- \"${cur}\") )\n")
	b.WriteString("    return 0\n")
	b.WriteString("}\n\n")
	b.WriteString(fmt.Sprintf("complete -F _%s %s\n", name, cmd.name))

	for _, sub := range cmd.subcommands {
		if sub.hidden {
			continue
		}
		b.WriteString("\n")
		subCopy := *sub
		subCopy.name = cmd.name + "_" + sub.name
		b.WriteString(generateBashSubcommandFunc(&subCopy, cmd.name+" "+sub.name))
	}

	return b.String()
}

func writeBashValueCompletions(b *strings.Builder, cmd *Command) {
	for _, a := range cmd.args {
		if a.positional || (!a.action.takesValue()) {
			continue
		}

		var compgen string
		if len(a.possibleValues) > 0 {
			compgen = fmt.Sprintf("compgen -W \"%s\" -- \"${cur}\"", strings.Join(a.possibleValues, " "))
		} else if hint := bashHintCompgen(a.valueHint); hint != "" {
			compgen = hint
		} else {
			continue
		}

		if a.short != 0 {
			b.WriteString(fmt.Sprintf("    if [[ \"${prev}\" == \"-%c\" || \"${prev}\" == \"--%s\" ]]; then\n", a.short, a.long))
		} else {
			b.WriteString(fmt.Sprintf("    if [[ \"${prev}\" == \"--%s\" ]]; then\n", a.long))
		}
		b.WriteString(fmt.Sprintf("        COMPREPLY=( $(%s) )\n", compgen))
		b.WriteString("        return 0\n")
		b.WriteString("    fi\n\n")
	}
}

func bashHintCompgen(hint ValueHint) string {
	switch hint {
	case HintFilePath:
		return "compgen -f -- \"${cur}\""
	case HintDirPath:
		return "compgen -d -- \"${cur}\""
	case HintExecutablePath, HintCommandName:
		return "compgen -c -- \"${cur}\""
	case HintUsername:
		return "compgen -u -- \"${cur}\""
	case HintHostname:
		return "compgen -A hostname -- \"${cur}\""
	default:
		return ""
	}
}

func generateBashSubcommandFunc(cmd *Command, _ string) string {
	var b strings.Builder
	name := sanitizeName(cmd.name)

	b.WriteString(fmt.Sprintf("_%s() {\n", name))
	b.WriteString("    local cur prev opts\n")
	b.WriteString("    COMPREPLY=()\n")
	b.WriteString("    cur=\"${COMP_WORDS[COMP_CWORD]}\"\n")
	b.WriteString("    prev=\"${COMP_WORDS[COMP_CWORD-1]}\"\n\n")

	var flags []string
	for _, a := range cmd.args {
		if a.hidden || a.positional {
			continue
		}
		if a.short != 0 {
			flags = append(flags, fmt.Sprintf("-%c", a.short))
		}
		flags = append(flags, "--"+a.long)
	}

	b.WriteString(fmt.Sprintf("    opts=\"%s\"\n\n", strings.Join(flags, " ")))
	writeBashValueCompletions(&b, cmd)
	b.WriteString("    COMPREPLY=( $(compgen -W \"${opts}\" -- \"${cur}\") )\n")
	b.WriteString("    return 0\n")
	b.WriteString("}\n")

	return b.String()
}

func generateZshCompletion(cmd *Command) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("#compdef %s\n\n", cmd.name))
	b.WriteString(fmt.Sprintf("_%s() {\n", sanitizeName(cmd.name)))

	if len(cmd.subcommands) > 0 {
		b.WriteString("    local -a subcommands\n")
		b.WriteString("    subcommands=(\n")
		for _, sub := range cmd.subcommands {
			if sub.hidden {
				continue
			}
			desc := sub.about
			if desc == "" {
				desc = sub.name
			}
			b.WriteString(fmt.Sprintf("        '%s:%s'\n", sub.name, zshEscape(desc)))
		}
		b.WriteString("    )\n\n")
	}

	b.WriteString("    _arguments -s \\\n")
	for _, a := range cmd.args {
		if a.hidden || a.positional {
			continue
		}
		b.WriteString("        ")
		b.WriteString(zshArgSpec(a))
		b.WriteString(" \\\n")
	}

	if len(cmd.subcommands) > 0 {
		b.WriteString("        '1:subcommand:->subcmd' \\\n")
		b.WriteString("        '*::arg:->args'\n\n")
		b.WriteString("    case $state in\n")
		b.WriteString("    subcmd)\n")
		b.WriteString("        _describe 'subcommand' subcommands\n")
		b.WriteString("        ;;\n")
		b.WriteString("    args)\n")
		b.WriteString("        case $words[1] in\n")
		for _, sub := range cmd.subcommands {
			if sub.hidden {
				continue
			}
			b.WriteString(fmt.Sprintf("        %s)\n", sub.name))
			b.WriteString(fmt.Sprintf("            _%s_%s\n", sanitizeName(cmd.name), sanitizeName(sub.name)))
			b.WriteString("            ;;\n")
		}
		b.WriteString("        esac\n")
		b.WriteString("        ;;\n")
		b.WriteString("    esac\n")
	} else {
		b.WriteString("        '*:'\n")
	}

	b.WriteString("}\n")

	for _, sub := range cmd.subcommands {
		if sub.hidden {
			continue
		}
		b.WriteString("\n")
		b.WriteString(generateZshSubcommandFunc(cmd.name, sub))
	}

	b.WriteString(fmt.Sprintf("\n_%s\n", sanitizeName(cmd.name)))

	return b.String()
}

func generateZshSubcommandFunc(parentName string, cmd *Command) string {
	var b strings.Builder
	funcName := fmt.Sprintf("_%s_%s", sanitizeName(parentName), sanitizeName(cmd.name))

	b.WriteString(fmt.Sprintf("%s() {\n", funcName))
	b.WriteString("    _arguments -s \\\n")
	for _, a := range cmd.args {
		if a.hidden || a.positional {
			continue
		}
		b.WriteString("        ")
		b.WriteString(zshArgSpec(a))
		b.WriteString(" \\\n")
	}
	b.WriteString("        '*:'\n")
	b.WriteString("}\n")

	return b.String()
}

func zshArgSpec(a *Arg) string {
	var spec string

	desc := zshEscape(a.help)
	if desc == "" {
		desc = a.name
	}

	if a.short != 0 {
		if a.action.takesValue() {
			vn := a.valueName
			if vn == "" {
				vn = strings.ToUpper(a.name)
			}
			action := zshCompletionAction(a)
			spec = fmt.Sprintf("'(-%c --%s)'{-%c,--%s}'[%s]:%s:%s'",
				a.short, a.long, a.short, a.long, desc, vn, action)
		} else {
			spec = fmt.Sprintf("'(-%c --%s)'{-%c,--%s}'[%s]'",
				a.short, a.long, a.short, a.long, desc)
		}
	} else {
		if a.action.takesValue() {
			vn := a.valueName
			if vn == "" {
				vn = strings.ToUpper(a.name)
			}
			action := zshCompletionAction(a)
			spec = fmt.Sprintf("'--%s[%s]:%s:%s'", a.long, desc, vn, action)
		} else {
			spec = fmt.Sprintf("'--%s[%s]'", a.long, desc)
		}
	}

	return spec
}

func zshCompletionAction(a *Arg) string {
	if len(a.possibleValues) > 0 {
		return "(" + strings.Join(a.possibleValues, " ") + ")"
	}
	switch a.valueHint {
	case HintFilePath:
		return "_files"
	case HintDirPath:
		return "_directories"
	case HintExecutablePath, HintCommandName:
		return "_command_names"
	case HintUsername:
		return "_users"
	case HintHostname:
		return "_hosts"
	case HintUrl:
		return "_urls"
	default:
		return ""
	}
}

func zshEscape(s string) string {
	s = strings.ReplaceAll(s, "'", "'\\''")
	s = strings.ReplaceAll(s, "[", "\\[")
	s = strings.ReplaceAll(s, "]", "\\]")
	return s
}

func generateFishCompletion(cmd *Command) string {
	var b strings.Builder

	for _, a := range cmd.args {
		if a.hidden || a.positional {
			continue
		}
		b.WriteString(fmt.Sprintf("complete -c %s", cmd.name))
		if a.short != 0 {
			b.WriteString(fmt.Sprintf(" -s %c", a.short))
		}
		b.WriteString(fmt.Sprintf(" -l %s", a.long))
		if a.help != "" {
			b.WriteString(fmt.Sprintf(" -d %q", a.help))
		}
		writeFishArgCompletion(&b, a)
		b.WriteString("\n")
	}

	for _, sub := range cmd.subcommands {
		if sub.hidden {
			continue
		}
		b.WriteString(fmt.Sprintf("complete -c %s -n '__fish_use_subcommand' -a %s",
			cmd.name, sub.name))
		if sub.about != "" {
			b.WriteString(fmt.Sprintf(" -d %q", sub.about))
		}
		b.WriteString("\n")

		for _, a := range sub.args {
			if a.hidden || a.positional {
				continue
			}
			b.WriteString(fmt.Sprintf("complete -c %s -n '__fish_seen_subcommand_from %s'",
				cmd.name, sub.name))
			if a.short != 0 {
				b.WriteString(fmt.Sprintf(" -s %c", a.short))
			}
			b.WriteString(fmt.Sprintf(" -l %s", a.long))
			if a.help != "" {
				b.WriteString(fmt.Sprintf(" -d %q", a.help))
			}
			if a.action.takesValue() {
				writeFishArgCompletion(&b, a)
			} else {
				b.WriteString(" -f")
			}
			b.WriteString("\n")
		}
	}

	return b.String()
}

func writeFishArgCompletion(b *strings.Builder, a *Arg) {
	if !a.action.takesValue() {
		b.WriteString(" -f")
		return
	}
	b.WriteString(" -r")
	if len(a.possibleValues) > 0 {
		b.WriteString(fmt.Sprintf(" -f -a %q", strings.Join(a.possibleValues, " ")))
	} else {
		switch a.valueHint {
		case HintFilePath:
			b.WriteString(" -F")
		case HintDirPath:
			b.WriteString(" -a '(__fish_complete_directories)'")
		case HintUsername:
			b.WriteString(" -a '(__fish_complete_users)'")
		case HintHostname:
			b.WriteString(" -a '(__fish_print_hostnames)'")
		case HintCommandName, HintExecutablePath:
			b.WriteString(" -a '(__fish_complete_command)'")
		}
	}
}

func generatePowerShellCompletion(cmd *Command) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("Register-ArgumentCompleter -CommandName %s -ScriptBlock {\n", cmd.name))
	b.WriteString("    param($wordToComplete, $commandAst, $cursorPosition)\n\n")
	b.WriteString("    $commands = $commandAst.ToString().Split(' ')\n")
	b.WriteString("    $completions = @()\n\n")

	if len(cmd.subcommands) > 0 {
		b.WriteString("    if ($commands.Count -le 2) {\n")
		for _, sub := range cmd.subcommands {
			if sub.hidden {
				continue
			}
			desc := sub.about
			if desc == "" {
				desc = sub.name
			}
			b.WriteString(fmt.Sprintf("        $completions += [System.Management.Automation.CompletionResult]::new('%s', '%s', 'ParameterValue', '%s')\n",
				sub.name, sub.name, psEscape(desc)))
		}
		b.WriteString("    }\n\n")
	}

	for _, a := range cmd.args {
		if a.hidden || a.positional {
			continue
		}
		desc := a.help
		if desc == "" {
			desc = a.name
		}
		flag := "--" + a.long
		b.WriteString(fmt.Sprintf("    $completions += [System.Management.Automation.CompletionResult]::new('%s', '%s', 'ParameterName', '%s')\n",
			flag, flag, psEscape(desc)))
		if a.short != 0 {
			sflag := fmt.Sprintf("-%c", a.short)
			b.WriteString(fmt.Sprintf("    $completions += [System.Management.Automation.CompletionResult]::new('%s', '%s', 'ParameterName', '%s')\n",
				sflag, sflag, psEscape(desc)))
		}
	}

	b.WriteString("\n    $completions | Where-Object { $_.CompletionText -like \"$wordToComplete*\" }\n")
	b.WriteString("}\n")

	return b.String()
}

func psEscape(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

func sanitizeName(name string) string {
	return strings.NewReplacer("-", "_", ".", "_", " ", "_").Replace(name)
}
