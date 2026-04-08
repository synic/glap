package glap

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// App wraps a struct pointer with optional metadata for the tag-based API.
type App struct {
	target  any
	command *Command
}

// New creates a new App for the given struct pointer.
func New(target any) *App {
	return &App{
		target:  target,
		command: &Command{},
	}
}

// Name sets the application name shown in help and version output.
func (a *App) Name(name string) *App {
	a.command.name = name
	return a
}

// Version sets the version string shown by --version.
func (a *App) Version(v string) *App {
	a.command.version = v
	return a
}

// About sets the short description shown in help output.
func (a *App) About(text string) *App {
	a.command.about = text
	return a
}

// Author sets the author shown in help output.
func (a *App) Author(author string) *App {
	a.command.author = author
	return a
}

// ArgRequiredElseHelp shows help when no arguments are provided.
func (a *App) ArgRequiredElseHelp(b bool) *App {
	a.command.argRequiredElseHelp = b
	return a
}

// AllowNegativeNumbers treats tokens like -1 and -3.14 as values instead of flags.
func (a *App) AllowNegativeNumbers(b bool) *App {
	a.command.allowNegativeNumbers = b
	return a
}

// LongAbout sets the detailed description shown by --help.
func (a *App) LongAbout(text string) *App {
	a.command.longAbout = text
	return a
}

// SetColorMode controls ANSI color output in help and error messages.
func (a *App) SetColorMode(mode ColorMode) *App {
	a.command.colorMode = mode
	return a
}

// SkipBinaryName causes the parser to skip args[0], allowing raw os.Args to be passed.
func (a *App) SkipBinaryName(b bool) *App {
	a.command.skipBinaryName = b
	return a
}

// Multicall enables busybox-style dispatch where args[0] is treated as a subcommand name.
func (a *App) Multicall(b bool) *App {
	a.command.multicall = b
	return a
}

// Parse parses args using the App's metadata and populates the target struct.
func (a *App) Parse(args []string) (string, error) {
	cmd, err := buildCommand(a.command, a.target)
	if err != nil {
		return "", err
	}
	cmd.injectHelpAndVersion()

	matches, err := parseCommand(cmd, args)
	if err != nil {
		return "", err
	}

	if err := writeBack(a.target, matches, cmd); err != nil {
		return "", err
	}

	subcmd := resolveSubcommand(matches)
	return subcmd, nil
}

// Parse is the top-level convenience function for the struct tag API.
func Parse(target any, args []string) (string, error) {
	app := New(target)
	return app.Parse(args)
}

func buildCommand(cmd *Command, target any) (*Command, error) {
	v := reflect.ValueOf(target)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return nil, fmt.Errorf("glap: target must be a pointer to a struct")
	}
	v = v.Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("glap")
		if tag == "" || tag == "-" {
			continue
		}

		opts := parseTag(tag)

		if opts.has("subcommand") {
			sub := &Command{name: opts.name, parent: cmd}
			if h, ok := opts.get("help"); ok {
				sub.about = h
			}

			ft := field.Type
			if ft.Kind() == reflect.Ptr {
				ft = ft.Elem()
			}
			subTarget := reflect.New(ft).Interface()
			sub, err := buildCommand(sub, subTarget)
			if err != nil {
				return nil, fmt.Errorf("glap: subcommand %s: %w", opts.name, err)
			}
			cmd.subcommands = append(cmd.subcommands, sub)
			continue
		}

		arg := buildArgFromTag(field, opts)
		cmd.args = append(cmd.args, arg)
	}

	return cmd, nil
}

func buildArgFromTag(field reflect.StructField, opts tagOpts) *Arg {
	a := &Arg{
		name: opts.name,
		long: opts.name,
	}

	if s, ok := opts.get("short"); ok && len(s) > 0 {
		a.short = rune(s[0])
	}
	if h, ok := opts.get("help"); ok {
		a.help = h
	}
	if e, ok := opts.get("env"); ok {
		a.envVar = e
	}
	if d, ok := opts.get("default"); ok {
		a.defaultValue = d
	}
	if vn, ok := opts.get("value_name"); ok {
		a.valueName = vn
	}
	if al, ok := opts.get("alias"); ok {
		a.aliases = append(a.aliases, al)
	}
	if c, ok := opts.get("conflicts_with"); ok {
		a.conflictsWith = append(a.conflictsWith, c)
	}
	if r, ok := opts.get("requires"); ok {
		a.requires = append(a.requires, r)
	}
	if g, ok := opts.get("group"); ok {
		a.groupID = g
	}
	if h, ok := opts.get("heading"); ok {
		a.heading = h
	}
	if p, ok := opts.get("possible"); ok {
		a.possibleValues = strings.Split(p, "|")
	}
	if na, ok := opts.get("num_args"); ok {
		a.numArgs = parseNumArgs(na)
	}

	if d, ok := opts.get("delimiter"); ok {
		a.valueDelimiter = resolveDelimiterName(d)
	}
	if rie, ok := opts.get("required_if_eq"); ok {
		parts := strings.SplitN(rie, ":", 2)
		if len(parts) == 2 {
			a.requiredIfEq = append(a.requiredIfEq, conditionalRule{ArgName: parts[0], Value: parts[1]})
		}
	}
	if ru, ok := opts.get("required_unless"); ok {
		a.requiredUnless = append(a.requiredUnless, ru)
	}
	if di, ok := opts.get("default_if"); ok {
		parts := strings.SplitN(di, ":", 3)
		if len(parts) == 3 {
			a.defaultValueIfs = append(a.defaultValueIfs, conditionalDefault{ArgName: parts[0], Value: parts[1], DefaultValue: parts[2]})
		}
	}

	if lh, ok := opts.get("long_help"); ok {
		a.longHelp = lh
	}
	if do, ok := opts.get("display_order"); ok {
		a.displayOrder, _ = strconv.Atoi(do)
	}
	if ow, ok := opts.get("overrides_with"); ok {
		a.overridesWith = append(a.overridesWith, ow)
	}
	if idx, ok := opts.get("index"); ok {
		a.index, _ = strconv.Atoi(idx)
		a.positional = true
	}
	a.requireEquals = opts.has("require_equals")
	a.hideDefaultValue = opts.has("hide_default_value")

	if vh, ok := opts.get("value_hint"); ok {
		a.valueHint = parseValueHint(vh)
	}

	a.required = opts.has("required")
	a.hidden = opts.has("hidden")
	a.global = opts.has("global")
	if opts.has("positional") {
		a.positional = true
	}
	a.allowHyphenValues = opts.has("allow_hyphen_values")

	if opts.has("trailing_var_arg") {
		a.trailingVarArg = true
		a.positional = true
		a.action = Append
	}

	if act, ok := opts.get("action"); ok {
		switch act {
		case "append":
			a.action = Append
		case "set_true":
			a.action = SetTrue
		case "count":
			a.action = Count
		case "set_false":
			a.action = SetFalse
		default:
			a.action = Set
		}
	} else {
		switch field.Type.Kind() {
		case reflect.Bool:
			a.action = SetTrue
		case reflect.Slice:
			a.action = Append
		default:
			a.action = Set
		}
	}

	return a
}

func parseNumArgs(s string) NumArgs {
	parts := strings.Split(s, "..")
	if len(parts) == 2 {
		min, _ := strconv.Atoi(parts[0])
		max := -1
		if parts[1] != "" {
			max, _ = strconv.Atoi(parts[1])
		}
		return NumArgs{Min: min, Max: max, Set: true}
	}
	n, _ := strconv.Atoi(s)
	return NumArgs{Min: n, Max: n, Set: true}
}

func writeBack(target any, m *Matches, cmd *Command) error {
	v := reflect.ValueOf(target).Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fv := v.Field(i)
		tag := field.Tag.Get("glap")
		if tag == "" || tag == "-" {
			continue
		}

		opts := parseTag(tag)

		if opts.has("subcommand") {
			if m.subcommandName == opts.name {
				sub := cmd.findSubcommand(opts.name)
				if sub == nil {
					continue
				}
				ft := field.Type
				isPtr := ft.Kind() == reflect.Ptr
				if isPtr {
					ft = ft.Elem()
				}
				newVal := reflect.New(ft)
				if err := writeBack(newVal.Interface(), m.subcommandMatches, sub); err != nil {
					return err
				}
				if isPtr {
					fv.Set(newVal)
				} else {
					fv.Set(newVal.Elem())
				}
			}
			continue
		}

		ma, ok := m.args[opts.name]
		if !ok {
			continue
		}

		if err := setFieldValue(fv, field.Type, ma); err != nil {
			return fmt.Errorf("glap: field %s: %w", field.Name, err)
		}
	}

	return nil
}

func setFieldValue(fv reflect.Value, ft reflect.Type, ma *MatchedArg) error {
	switch ft.Kind() {
	case reflect.String:
		if len(ma.Values) > 0 {
			fv.SetString(ma.Values[0])
		}
	case reflect.Bool:
		if len(ma.Values) > 0 {
			b, err := strconv.ParseBool(ma.Values[0])
			if err != nil {
				return err
			}
			fv.SetBool(b)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if len(ma.Values) > 0 {
			n, err := strconv.ParseInt(ma.Values[0], 10, 64)
			if err != nil {
				return err
			}
			fv.SetInt(n)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if len(ma.Values) > 0 {
			n, err := strconv.ParseUint(ma.Values[0], 10, 64)
			if err != nil {
				return err
			}
			fv.SetUint(n)
		}
	case reflect.Float32, reflect.Float64:
		if len(ma.Values) > 0 {
			f, err := strconv.ParseFloat(ma.Values[0], 64)
			if err != nil {
				return err
			}
			fv.SetFloat(f)
		}
	case reflect.Slice:
		if ft.Elem().Kind() == reflect.String {
			fv.Set(reflect.ValueOf(ma.Values))
		}
	}
	return nil
}

func resolveSubcommand(m *Matches) string {
	if m.subcommandName == "" {
		return ""
	}
	sub := resolveSubcommand(m.subcommandMatches)
	if sub != "" {
		return m.subcommandName + " " + sub
	}
	return m.subcommandName
}

// tagOpts parses the glap struct tag format.
type tagOpts struct {
	name  string
	pairs map[string]string
	flags map[string]bool
}

func parseTag(tag string) tagOpts {
	opts := tagOpts{
		pairs: make(map[string]string),
		flags: make(map[string]bool),
	}

	parts := splitTag(tag)
	if len(parts) == 0 {
		return opts
	}

	opts.name = parts[0]
	for _, p := range parts[1:] {
		if k, v, ok := strings.Cut(p, "="); ok {
			opts.pairs[k] = v
		} else {
			opts.flags[p] = true
		}
	}

	return opts
}

func (o tagOpts) get(key string) (string, bool) {
	v, ok := o.pairs[key]
	return v, ok
}

func (o tagOpts) has(key string) bool {
	if o.flags[key] {
		return true
	}
	_, ok := o.pairs[key]
	return ok
}

// splitTag splits a tag string by commas, respecting that values may not contain commas
// in this format since we use | for multi-value fields like possible values.
func parseValueHint(s string) ValueHint {
	switch s {
	case "file_path":
		return HintFilePath
	case "dir_path":
		return HintDirPath
	case "executable_path":
		return HintExecutablePath
	case "command_name":
		return HintCommandName
	case "username":
		return HintUsername
	case "hostname":
		return HintHostname
	case "url":
		return HintUrl
	case "email_address":
		return HintEmailAddress
	default:
		return HintNone
	}
}

func resolveDelimiterName(s string) string {
	switch s {
	case "comma":
		return ","
	case "colon":
		return ":"
	case "semicolon":
		return ";"
	case "pipe":
		return "|"
	case "space":
		return " "
	default:
		return s
	}
}

func splitTag(tag string) []string {
	return strings.Split(tag, ",")
}
