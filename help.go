package glap

import (
	"fmt"
	"sort"
	"strings"
)

func generateHelp(cmd *Command) string {
	return renderHelp(cmd, false)
}

func generateLongHelp(cmd *Command) string {
	return renderHelp(cmd, true)
}

func renderHelp(cmd *Command, long bool) string {
	var b strings.Builder
	s := newStyler(cmd.colorMode)
	argGroup := buildArgGroupMap(cmd)

	if cmd.name != "" {
		b.WriteString(s.bold(cmd.name))
		if cmd.version != "" {
			b.WriteString(" " + cmd.version)
		}
		b.WriteString("\n")
	}
	if cmd.author != "" {
		b.WriteString(cmd.author + "\n")
	}

	about := cmd.about
	if long && cmd.longAbout != "" {
		about = cmd.longAbout
	}
	if about != "" {
		b.WriteString(about + "\n")
	}

	b.WriteString("\n" + s.bold("USAGE:") + "\n")
	b.WriteString("    " + cmd.name)
	hasFlags := false
	for _, a := range cmd.args {
		if a.isFlag() && !a.hidden {
			hasFlags = true
			break
		}
	}
	if hasFlags {
		b.WriteString(" [OPTIONS]")
	}
	for _, a := range cmd.positionalArgs() {
		if !a.hidden {
			vn := a.valueName
			if vn == "" {
				vn = strings.ToUpper(a.name)
			}
			if a.required {
				b.WriteString(" <" + s.cyan(vn) + ">")
			} else {
				b.WriteString(" [" + s.cyan(vn) + "]")
			}
		}
	}
	if len(cmd.subcommands) > 0 {
		b.WriteString(" [SUBCOMMAND]")
	}
	b.WriteString("\n")

	type headingGroup struct {
		heading string
		args    []*Arg
	}
	groups := []headingGroup{{heading: "OPTIONS"}}
	groupMap := map[string]int{"OPTIONS": 0}

	for _, a := range cmd.args {
		if a.hidden || a.positional {
			continue
		}
		h := "OPTIONS"
		if a.heading != "" {
			h = a.heading
		}
		idx, ok := groupMap[h]
		if !ok {
			idx = len(groups)
			groupMap[h] = idx
			groups = append(groups, headingGroup{heading: h})
		}
		groups[idx].args = append(groups[idx].args, a)
	}

	for gi := range groups {
		sort.SliceStable(groups[gi].args, func(i, j int) bool {
			return groups[gi].args[i].displayOrder < groups[gi].args[j].displayOrder
		})
	}

	for _, g := range groups {
		if len(g.args) == 0 {
			continue
		}
		b.WriteString("\n" + s.bold(g.heading+":") + "\n")
		for _, a := range g.args {
			b.WriteString("    ")
			b.WriteString(formatArgHelpStyled(a, long, s, argGroup))
			b.WriteString("\n")
		}
	}

	var positionals []*Arg
	for _, a := range cmd.positionalArgs() {
		if a.hidden {
			continue
		}
		positionals = append(positionals, a)
	}
	if len(positionals) > 0 {
		b.WriteString("\n" + s.bold("ARGS:") + "\n")
		for _, a := range positionals {
			b.WriteString("    ")
			b.WriteString(formatPositionalHelpStyled(a, long, s, argGroup))
			b.WriteString("\n")
		}
	}

	if len(cmd.subcommands) > 0 {
		subs := make([]*Command, len(cmd.subcommands))
		copy(subs, cmd.subcommands)
		sort.SliceStable(subs, func(i, j int) bool {
			return subs[i].displayOrder < subs[j].displayOrder
		})

		b.WriteString("\n" + s.bold("SUBCOMMANDS:") + "\n")
		for _, sub := range subs {
			if sub.hidden {
				continue
			}
			label := sub.name
			if len(sub.aliases) > 0 {
				label += ", " + strings.Join(sub.aliases, ", ")
			}
			line := "    " + s.green(label)
			desc := sub.about
			if long && sub.longAbout != "" {
				desc = sub.longAbout
			}
			if desc != "" {
				pad := 20 - len(label)
				if pad < 4 {
					pad = 4
				}
				line += strings.Repeat(" ", pad) + desc
			}
			b.WriteString(line + "\n")
		}
	}

	return b.String()
}

func buildArgGroupMap(cmd *Command) map[string]string {
	m := make(map[string]string)
	for _, a := range cmd.args {
		if a.groupID != "" {
			m[a.name] = a.groupID
		}
	}
	for _, g := range cmd.argGroups {
		for _, name := range g.args {
			if _, ok := m[name]; !ok {
				m[name] = g.name
			}
		}
	}
	return m
}

func formatArgHelpStyled(a *Arg, long bool, s styler, argGroup map[string]string) string {
	var b strings.Builder

	if a.short != 0 {
		b.WriteString(s.green(fmt.Sprintf("-%c", a.short)) + ", ")
	} else {
		b.WriteString("    ")
	}
	b.WriteString(s.green("--" + a.long))
	for _, alias := range a.aliases {
		b.WriteString(", " + s.green("--"+alias))
	}

	if a.action.takesValue() {
		vn := a.valueName
		if vn == "" {
			vn = strings.ToUpper(a.name)
		}
		b.WriteString(" " + s.cyan("<"+vn+">"))
	}

	left := stripANSI(b.String())
	pad := 30 - len(left)
	if pad < 4 {
		pad = 4
	}
	b.WriteString(strings.Repeat(" ", pad))

	help := a.help
	if long && a.longHelp != "" {
		help = a.longHelp
	}
	if help != "" {
		b.WriteString(help)
	}

	annotations := argAnnotations(a, argGroup)
	if len(annotations) > 0 {
		b.WriteString(" " + s.dim("["+strings.Join(annotations, "] [")+"]"))
	}

	return b.String()
}

func formatPositionalHelpStyled(a *Arg, long bool, s styler, argGroup map[string]string) string {
	var b strings.Builder

	vn := a.valueName
	if vn == "" {
		vn = strings.ToUpper(a.name)
	}
	if a.required {
		b.WriteString(s.cyan("<" + vn + ">"))
	} else {
		b.WriteString(s.cyan("[" + vn + "]"))
	}

	left := stripANSI(b.String())
	pad := 30 - len(left)
	if pad < 4 {
		pad = 4
	}
	b.WriteString(strings.Repeat(" ", pad))

	help := a.help
	if long && a.longHelp != "" {
		help = a.longHelp
	}
	if help != "" {
		b.WriteString(help)
	}

	annotations := argAnnotations(a, argGroup)
	if len(annotations) > 0 {
		b.WriteString(" " + s.dim("["+strings.Join(annotations, "] [")+"]"))
	}

	return b.String()
}

func argAnnotations(a *Arg, argGroup map[string]string) []string {
	var annotations []string
	if a.envVar != "" {
		annotations = append(annotations, fmt.Sprintf("env: %s", a.envVar))
	}
	if a.defaultValue != "" && !a.hideDefaultValue {
		annotations = append(annotations, fmt.Sprintf("default: %s", a.defaultValue))
	}
	if len(a.possibleValues) > 0 {
		annotations = append(annotations, fmt.Sprintf("possible values: %s", strings.Join(a.possibleValues, ", ")))
	}
	if a.required {
		annotations = append(annotations, "required")
	}
	if g, ok := argGroup[a.name]; ok {
		annotations = append(annotations, fmt.Sprintf("group: %s", g))
	}
	return annotations
}

func stripANSI(s string) string {
	var b strings.Builder
	inEsc := false
	for _, r := range s {
		if r == '\033' {
			inEsc = true
			continue
		}
		if inEsc {
			if r == 'm' {
				inEsc = false
			}
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
