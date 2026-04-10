package glap

import (
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

func parseCommand(cmd *Command, args []string) (*Matches, error) {
	if cmd.parent == nil && cmd.skipBinaryName && len(args) > 0 {
		args = args[1:]
	}

	if cmd.parent == nil && cmd.multicall && len(args) > 0 {
		binName := filepath.Base(args[0])
		if sub := cmd.findSubcommand(binName); sub != nil {
			m := newMatches()
			subMatches, err := parseCommand(sub, args[1:])
			if err != nil {
				return nil, err
			}
			m.subcommandName = sub.name
			m.subcommandMatches = subMatches
			return m, nil
		}
		matched := binName == cmd.name
		if !matched {
			for _, alias := range cmd.aliases {
				if alias == binName {
					matched = true
					break
				}
			}
		}
		if matched {
			args = args[1:]
		}
	}

	m := newMatches()
	positionals := cmd.positionalArgs()
	posIdx := 0
	dashDash := false

	globals := collectGlobals(cmd)
	allArgs := make([]*Arg, len(cmd.args))
	copy(allArgs, cmd.args)
	for _, g := range globals {
		if cmd.findArg(g.name) == nil {
			allArgs = append(allArgs, g)
		}
	}

	if cmd.argRequiredElseHelp && len(args) == 0 {
		return nil, &HelpRequestedError{Message: generateHelp(cmd)}
	}

	i := 0
	for i < len(args) {
		token := args[i]

		if dashDash {
			if posIdx < len(positionals) {
				arg := positionals[posIdx]
				handlePositionalValue(m, arg, token)
				if !canAcceptMore(arg, m) {
					posIdx++
				}
			}
			i++
			continue
		}

		if token == "--" {
			dashDash = true
			i++
			continue
		}

		if posIdx == 0 && !strings.HasPrefix(token, "-") {
			if sub := cmd.findSubcommand(token); sub != nil {
				subMatches, err := parseCommand(sub, args[i+1:])
				if err != nil {
					return nil, err
				}
				m.subcommandName = sub.name
				m.subcommandMatches = subMatches
				if err := fillEnvAndDefaults(m, allArgs); err != nil {
					return nil, err
				}
				return m, nil
			}
		}

		if strings.HasPrefix(token, "-") && len(token) > 1 && posIdx < len(positionals) && positionals[posIdx].allowHyphenValues {
			arg := positionals[posIdx]
			handlePositionalValue(m, arg, token)
			if !canAcceptMore(arg, m) {
				posIdx++
			}
			i++
			continue
		}

		if strings.HasPrefix(token, "--") {
			name, value, hasValue := parseEquals(token[2:])
			a := findInArgs(allArgs, name)
			if a == nil {
				return nil, &UnknownArgError{Arg: token}
			}

			if a.name == "help" && a.action == SetTrue {
				return nil, &HelpRequestedError{Message: generateLongHelp(cmd)}
			}
			if a.name == "version" && a.action == SetTrue {
				return nil, &VersionRequestedError{Message: formatVersion(cmd)}
			}

			if a.action.takesValue() {
				if hasValue {
					setArgValue(m, a, value, SourceCLI)
				} else if a.requireEquals {
					return nil, &RequireEqualsError{Arg: a.name}
				} else {
					i++
					if i >= len(args) {
						return nil, &TooFewValuesError{Arg: a.name, Min: 1, Got: 0}
					}
					setArgValue(m, a, args[i], SourceCLI)
				}
			} else {
				handleNoValueArg(m, a)
			}
		} else if strings.HasPrefix(token, "-") && len(token) > 1 && !(cmd.allowNegativeNumbers && isNegativeNumber(token)) {
			runes := []rune(token[1:])
			for j, r := range runes {
				a := findInArgsByShort(allArgs, r)
				if a == nil {
					return nil, &UnknownArgError{Arg: string([]rune{'-', r})}
				}

				if a.name == "help" && a.action == SetTrue {
					return nil, &HelpRequestedError{Message: generateHelp(cmd)}
				}
				if a.name == "version" && a.action == SetTrue {
					return nil, &VersionRequestedError{Message: formatVersion(cmd)}
				}

				if a.action.takesValue() {
					rest := string(runes[j+1:])
					if rest != "" {
						setArgValue(m, a, rest, SourceCLI)
					} else {
						i++
						if i >= len(args) {
							return nil, &TooFewValuesError{Arg: a.name, Min: 1, Got: 0}
						}
						setArgValue(m, a, args[i], SourceCLI)
					}
					break
				}
				handleNoValueArg(m, a)
			}
		} else {
			if posIdx < len(positionals) {
				arg := positionals[posIdx]
				handlePositionalValue(m, arg, token)
				if arg.trailingVarArg {
					for i++; i < len(args); i++ {
						handlePositionalValue(m, arg, args[i])
					}
					break
				}
				if !canAcceptMore(arg, m) {
					posIdx++
				}
			} else {
				return nil, &UnknownArgError{Arg: token}
			}
		}

		i++
	}

	if err := fillEnvAndDefaults(m, allArgs); err != nil {
		return nil, err
	}

	fillConditionalDefaults(m, allArgs)

	if err := validate(cmd, m, allArgs); err != nil {
		return nil, err
	}

	if cmd.subcommandRequired && m.subcommandName == "" {
		return nil, &HelpRequestedError{Message: generateHelp(cmd)}
	}

	return m, nil
}

func parseEquals(s string) (name, value string, hasValue bool) {
	idx := strings.IndexByte(s, '=')
	if idx == -1 {
		return s, "", false
	}
	return s[:idx], s[idx+1:], true
}

func findInArgs(args []*Arg, name string) *Arg {
	for _, a := range args {
		if a.name == name || a.long == name {
			return a
		}
		for _, alias := range a.aliases {
			if alias == name {
				return a
			}
		}
	}
	return nil
}

func findInArgsByShort(args []*Arg, s rune) *Arg {
	for _, a := range args {
		if a.short == s {
			return a
		}
	}
	return nil
}

func setArgValue(m *Matches, a *Arg, value string, source ValueSource) {
	if a.valueDelimiter != "" {
		parts := strings.Split(value, a.valueDelimiter)
		for _, p := range parts {
			m.appendValue(a.name, p, source)
		}
		applyOverrides(m, a)
		return
	}
	switch a.action {
	case Append:
		m.appendValue(a.name, value, source)
	default:
		m.set(a.name, value, source)
	}
	applyOverrides(m, a)
}

func applyOverrides(m *Matches, a *Arg) {
	for _, name := range a.overridesWith {
		delete(m.args, name)
	}
}

func handleNoValueArg(m *Matches, a *Arg) {
	switch a.action {
	case SetTrue:
		m.set(a.name, "true", SourceCLI)
	case SetFalse:
		m.set(a.name, "false", SourceCLI)
	case Count:
		m.increment(a.name)
	default:
		m.set(a.name, "true", SourceCLI)
	}
	applyOverrides(m, a)
}

func handlePositionalValue(m *Matches, a *Arg, value string) {
	switch a.action {
	case Append:
		m.appendValue(a.name, value, SourceCLI)
	default:
		m.set(a.name, value, SourceCLI)
	}
}

func canAcceptMore(a *Arg, m *Matches) bool {
	if a.action == Append {
		if a.numArgs.Set && a.numArgs.Max > 0 {
			ma, ok := m.args[a.name]
			if ok && len(ma.Values) >= a.numArgs.Max {
				return false
			}
		}
		return true
	}
	return false
}

func fillEnvAndDefaults(m *Matches, args []*Arg) error {
	for _, a := range args {
		if _, ok := m.args[a.name]; ok {
			continue
		}

		if a.envVar != "" {
			if val, ok := os.LookupEnv(a.envVar); ok {
				setArgValue(m, a, val, SourceEnv)
				continue
			}
		}

		if a.defaultValue != "" {
			setArgValue(m, a, a.defaultValue, SourceDefault)
		}
	}
	return nil
}

func collectGlobals(cmd *Command) []*Arg {
	var globals []*Arg
	for p := cmd.parent; p != nil; p = p.parent {
		globals = append(globals, p.globalArgs()...)
	}
	return globals
}

func fillConditionalDefaults(m *Matches, args []*Arg) {
	for _, a := range args {
		if _, ok := m.args[a.name]; ok {
			continue
		}
		for _, cd := range a.defaultValueIfs {
			if ma, ok := m.args[cd.ArgName]; ok {
				if len(ma.Values) > 0 && ma.Values[0] == cd.Value {
					setArgValue(m, a, cd.DefaultValue, SourceDefault)
					break
				}
			}
		}
	}
}

func isNegativeNumber(s string) bool {
	if len(s) < 2 || s[0] != '-' {
		return false
	}
	for i, r := range s[1:] {
		if r == '.' && i > 0 {
			continue
		}
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func formatVersion(cmd *Command) string {
	if cmd.version != "" {
		return cmd.name + " " + cmd.version
	}
	return cmd.name
}
