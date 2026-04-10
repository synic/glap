package glap

// ValueHint provides context to shell completion generators about what kind of value an arg expects.
//
// Struct tag: value_hint=file_path, value_hint=dir_path, etc.
type ValueHint int

const (
	// HintNone provides no completion hint.
	HintNone ValueHint = iota
	// HintFilePath suggests file path completion. Struct tag value: file_path.
	HintFilePath
	// HintDirPath suggests directory path completion. Struct tag value: dir_path.
	HintDirPath
	// HintExecutablePath suggests executable path completion. Struct tag value: executable_path.
	HintExecutablePath
	// HintCommandName suggests command name completion. Struct tag value: command_name.
	HintCommandName
	// HintUsername suggests system username completion. Struct tag value: username.
	HintUsername
	// HintHostname suggests hostname completion. Struct tag value: hostname.
	HintHostname
	// HintUrl suggests URL completion. Struct tag value: url.
	HintUrl
	// HintEmailAddress suggests email address completion. Struct tag value: email_address.
	HintEmailAddress
)

type conditionalRule struct {
	ArgName string
	Value   string
}

type conditionalDefault struct {
	ArgName      string
	Value        string
	DefaultValue string
}

// Arg represents a single command-line argument definition.
//
// Arguments can be configured via the builder API (NewArg + chained methods) or
// declaratively via struct tags with the "glap" key. See each method's documentation
// for the corresponding struct tag syntax.
type Arg struct {
	name              string
	short             rune
	long              string
	help              string
	envVar            string
	defaultValue      string
	required          bool
	hidden            bool
	global            bool
	action            ArgAction
	possibleValues    []string
	numArgs           NumArgs
	valueName         string
	aliases           []string
	conflictsWith     []string
	requires          []string
	groupID           string
	heading           string
	positional        bool
	validator         func(string) error
	valueDelimiter    string
	allowHyphenValues bool
	trailingVarArg    bool
	requiredIfEq      []conditionalRule
	requiredUnless    []string
	defaultValueIfs   []conditionalDefault
	longHelp          string
	displayOrder      int
	overridesWith     []string
	requireEquals     bool
	index             int
	hideDefaultValue  bool
	valueHint         ValueHint
}

// NumArgs specifies how many values an argument accepts.
//
// Struct tag: num_args=N or num_args=MIN..MAX (e.g., num_args=1..3).
type NumArgs struct {
	Min int
	Max int // -1 means unlimited
	Set bool
}

// NewArg creates a new Arg with the given name. The long flag name defaults to the arg name.
//
// In struct tags, the name is the first value: glap:"myarg,...".
func NewArg(name string) *Arg {
	return &Arg{
		name: name,
		long: name,
	}
}

// Short sets the single-character short flag (e.g., 'v' for -v).
//
// Struct tag: short=v
func (a *Arg) Short(c rune) *Arg {
	a.short = c
	return a
}

// Long sets the long flag name (e.g., "verbose" for --verbose). Defaults to the arg name.
func (a *Arg) Long(name string) *Arg {
	a.long = name
	return a
}

// Help sets the short help text shown in -h output.
//
// Struct tag: help=description text
func (a *Arg) Help(text string) *Arg {
	a.help = text
	return a
}

// Env sets the environment variable name used as a fallback value source.
// Precedence: CLI > Env > Default.
//
// Struct tag: env=MY_VAR
func (a *Arg) Env(varName string) *Arg {
	a.envVar = varName
	return a
}

// Default sets the default value used when neither CLI nor env provides one.
//
// Struct tag: default=value
func (a *Arg) Default(val string) *Arg {
	a.defaultValue = val
	return a
}

// Required marks the argument as required. Parsing fails if it is not provided.
//
// Struct tag: required
func (a *Arg) Required(b bool) *Arg {
	a.required = b
	return a
}

// Hidden hides the argument from help output.
//
// Struct tag: hidden
func (a *Arg) Hidden(b bool) *Arg {
	a.hidden = b
	return a
}

// Global propagates this argument to all subcommands.
//
// Struct tag: global
func (a *Arg) Global(b bool) *Arg {
	a.global = b
	return a
}

// Action sets the parsing action (Set, Append, SetTrue, SetFalse, or Count).
//
// Struct tag: action=set|append|set_true|set_false|count
func (a *Arg) Action(action ArgAction) *Arg {
	a.action = action
	return a
}

// PossibleValues restricts the argument to the given set of allowed values.
//
// Struct tag: possible=a|b|c
func (a *Arg) PossibleValues(vals ...string) *Arg {
	a.possibleValues = vals
	return a
}

// ValueName sets the placeholder name shown in help output (e.g., "FILE" in --config <FILE>).
//
// Struct tag: value_name=FILE
func (a *Arg) ValueName(name string) *Arg {
	a.valueName = name
	return a
}

// Alias adds a long alias for this argument (e.g., --cfg as an alias for --config).
//
// Struct tag: alias=cfg
func (a *Arg) Alias(alias string) *Arg {
	a.aliases = append(a.aliases, alias)
	return a
}

// ConflictsWith declares that this argument cannot be used together with the named arguments.
//
// Struct tag: conflicts_with=other
func (a *Arg) ConflictsWith(names ...string) *Arg {
	a.conflictsWith = append(a.conflictsWith, names...)
	return a
}

// Requires declares that the named arguments must also be present when this argument is used.
//
// Struct tag: requires=other
func (a *Arg) Requires(names ...string) *Arg {
	a.requires = append(a.requires, names...)
	return a
}

// Group assigns this argument to the named argument group for mutual exclusion or co-occurrence.
//
// Struct tag: group=mygroup
func (a *Arg) Group(name string) *Arg {
	a.groupID = name
	return a
}

// Heading sets the help section heading this argument appears under (default: "OPTIONS").
//
// Struct tag: heading=Advanced
func (a *Arg) Heading(h string) *Arg {
	a.heading = h
	return a
}

// Positional marks this argument as positional rather than a flag.
//
// Struct tag: positional
func (a *Arg) Positional(b bool) *Arg {
	a.positional = b
	return a
}

// Validator sets a custom validation function called on each parsed value.
// Return a non-nil error to reject the value. Builder API only.
func (a *Arg) Validator(fn func(string) error) *Arg {
	a.validator = fn
	return a
}

// ValueDelimiter sets a character that splits a single value into multiple entries.
// For example, ValueDelimiter(",") causes --tag=a,b,c to produce ["a", "b", "c"].
//
// Struct tag: delimiter=comma (also: colon, semicolon, pipe, space, or any single char)
func (a *Arg) ValueDelimiter(d string) *Arg {
	a.valueDelimiter = d
	return a
}

// AllowHyphenValues permits this positional argument to accept values starting with "-".
//
// Struct tag: allow_hyphen_values
func (a *Arg) AllowHyphenValues(b bool) *Arg {
	a.allowHyphenValues = b
	return a
}

// TrailingVarArg marks this as the final positional argument that captures all remaining
// tokens, including ones that look like flags. Implies Positional(true) and Action(Append).
//
// Struct tag: trailing_var_arg
func (a *Arg) TrailingVarArg(b bool) *Arg {
	a.trailingVarArg = b
	a.positional = true
	a.action = Append
	return a
}

// RequiredIfEq makes this argument required when otherArg has the given value.
//
// Struct tag: required_if_eq=other:value
func (a *Arg) RequiredIfEq(otherArg, value string) *Arg {
	a.requiredIfEq = append(a.requiredIfEq, conditionalRule{ArgName: otherArg, Value: value})
	return a
}

// RequiredUnlessPresent makes this argument required unless otherArg is present.
//
// Struct tag: required_unless=other
func (a *Arg) RequiredUnlessPresent(otherArg string) *Arg {
	a.requiredUnless = append(a.requiredUnless, otherArg)
	return a
}

// DefaultValueIf sets a conditional default: when otherArg equals value, this arg
// defaults to defaultVal.
//
// Struct tag: default_if=other:value:default
func (a *Arg) DefaultValueIf(otherArg, value, defaultVal string) *Arg {
	a.defaultValueIfs = append(a.defaultValueIfs, conditionalDefault{ArgName: otherArg, Value: value, DefaultValue: defaultVal})
	return a
}

// LongHelp sets the detailed help text shown in --help output (as opposed to the
// short help shown by -h).
//
// Struct tag: long_help=detailed description
func (a *Arg) LongHelp(text string) *Arg {
	a.longHelp = text
	return a
}

// DisplayOrder controls the position of this argument in help output. Lower values appear first.
//
// Struct tag: display_order=N
func (a *Arg) DisplayOrder(n int) *Arg {
	a.displayOrder = n
	return a
}

// OverridesWith declares that when this argument is set, the named arguments are removed
// from the parsed results.
//
// Struct tag: overrides_with=other
func (a *Arg) OverridesWith(names ...string) *Arg {
	a.overridesWith = append(a.overridesWith, names...)
	return a
}

// RequireEquals forces the --arg=value syntax; --arg value is rejected.
//
// Struct tag: require_equals
func (a *Arg) RequireEquals(b bool) *Arg {
	a.requireEquals = b
	return a
}

// Index sets the sort order for this positional argument. Implies Positional(true).
// Positionals with lower Index values consume values first; positionals without
// an explicit index are filled after indexed ones. This controls the order in
// which positionals consume values, not the absolute position in argv: values
// are still consumed from whatever non-flag tokens come next, in the sorted order.
//
// Struct tag: index=N
func (a *Arg) Index(n int) *Arg {
	a.index = n
	a.positional = true
	return a
}

// HideDefaultValue hides the [default: ...] annotation from help output.
//
// Struct tag: hide_default_value
func (a *Arg) HideDefaultValue(b bool) *Arg {
	a.hideDefaultValue = b
	return a
}

// SetValueHint provides a hint to shell completion generators about the expected value type.
//
// Struct tag: value_hint=file_path (see [ValueHint] constants for all values)
func (a *Arg) SetValueHint(hint ValueHint) *Arg {
	a.valueHint = hint
	return a
}

// SetNumArgs sets the minimum and maximum number of values this argument accepts.
// Use -1 for max to allow unlimited values.
//
// Struct tag: num_args=N or num_args=MIN..MAX
func (a *Arg) SetNumArgs(min, max int) *Arg {
	a.numArgs = NumArgs{Min: min, Max: max, Set: true}
	return a
}

// Clone returns a deep copy of this Arg, safe to add to multiple commands.
func (a *Arg) Clone() *Arg {
	c := *a
	c.aliases = append([]string(nil), a.aliases...)
	c.conflictsWith = append([]string(nil), a.conflictsWith...)
	c.requires = append([]string(nil), a.requires...)
	c.possibleValues = append([]string(nil), a.possibleValues...)
	c.overridesWith = append([]string(nil), a.overridesWith...)
	c.requiredUnless = append([]string(nil), a.requiredUnless...)
	c.requiredIfEq = append([]conditionalRule(nil), a.requiredIfEq...)
	c.defaultValueIfs = append([]conditionalDefault(nil), a.defaultValueIfs...)
	return &c
}

func (a *Arg) isFlag() bool {
	return !a.positional
}

// GetName returns the argument's name.
func (a *Arg) GetName() string { return a.name }

// GetShort returns the short flag character, or 0 if unset.
func (a *Arg) GetShort() rune { return a.short }

// GetLong returns the long flag name.
func (a *Arg) GetLong() string { return a.long }

// GetHelp returns the help text.
func (a *Arg) GetHelp() string { return a.help }

// GetEnv returns the environment variable name, or empty if unset.
func (a *Arg) GetEnv() string { return a.envVar }

// GetDefault returns the default value, or empty if unset.
func (a *Arg) GetDefault() string { return a.defaultValue }

// GetAction returns the argument's parsing action.
func (a *Arg) GetAction() ArgAction { return a.action }

// GetPossibleValues returns the allowed values, or nil if unrestricted.
func (a *Arg) GetPossibleValues() []string { return a.possibleValues }

// IsRequired reports whether the argument is required.
func (a *Arg) IsRequired() bool { return a.required }

// IsHidden reports whether the argument is hidden from help.
func (a *Arg) IsHidden() bool { return a.hidden }

// IsPositional reports whether the argument is positional.
func (a *Arg) IsPositional() bool { return a.positional }

// IsGlobal reports whether the argument propagates to subcommands.
func (a *Arg) IsGlobal() bool { return a.global }
