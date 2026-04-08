package glap

import "slices"

// Command represents a CLI command or subcommand with its arguments and configuration.
type Command struct {
	name               string
	version            string
	author             string
	about              string
	args               []*Arg
	subcommands        []*Command
	argGroups          []*ArgGroup
	aliases            []string
	hidden               bool
	subcommandRequired   bool
	argRequiredElseHelp  bool
	allowNegativeNumbers bool
	parent               *Command
	longAbout            string
	displayOrder         int
	colorMode            ColorMode
	skipBinaryName       bool
	multicall            bool
}

// NewCommand creates a new Command with the given name.
func NewCommand(name string) *Command {
	return &Command{name: name}
}

// Version sets the version string shown by --version.
func (c *Command) Version(v string) *Command {
	c.version = v
	return c
}

// Author sets the author shown in help output.
func (c *Command) Author(a string) *Command {
	c.author = a
	return c
}

// About sets the short description shown in help output.
func (c *Command) About(text string) *Command {
	c.about = text
	return c
}

// Arg adds an argument definition to this command.
func (c *Command) Arg(arg *Arg) *Command {
	c.args = append(c.args, arg)
	return c
}

// Subcommand adds a subcommand to this command.
//
// In struct tags, subcommands are pointer-to-struct fields: glap:"name,subcommand,help=..."
func (c *Command) Subcommand(sub *Command) *Command {
	sub.parent = c
	c.subcommands = append(c.subcommands, sub)
	return c
}

// ArgGroup adds an argument group for mutual exclusion or co-occurrence constraints.
func (c *Command) ArgGroup(group *ArgGroup) *Command {
	c.argGroups = append(c.argGroups, group)
	return c
}

// Alias adds an alternative name for this command.
func (c *Command) Alias(alias string) *Command {
	c.aliases = append(c.aliases, alias)
	return c
}

// Hidden hides this command from the parent's help output.
func (c *Command) Hidden(b bool) *Command {
	c.hidden = b
	return c
}

// SubcommandRequired requires that a subcommand be provided.
func (c *Command) SubcommandRequired(b bool) *Command {
	c.subcommandRequired = b
	return c
}

// ArgRequiredElseHelp shows help when no arguments are provided.
func (c *Command) ArgRequiredElseHelp(b bool) *Command {
	c.argRequiredElseHelp = b
	return c
}

// AllowNegativeNumbers treats tokens like -1 and -3.14 as values instead of flags.
func (c *Command) AllowNegativeNumbers(b bool) *Command {
	c.allowNegativeNumbers = b
	return c
}

// LongAbout sets the detailed description shown by --help (as opposed to the
// short description shown by -h).
func (c *Command) LongAbout(text string) *Command {
	c.longAbout = text
	return c
}

// DisplayOrder controls the position of this subcommand in the parent's help output.
func (c *Command) DisplayOrder(n int) *Command {
	c.displayOrder = n
	return c
}

// SetColorMode controls ANSI color output in help and error messages.
func (c *Command) SetColorMode(mode ColorMode) *Command {
	c.colorMode = mode
	return c
}

// SkipBinaryName causes the parser to skip args[0], allowing raw os.Args to be passed.
func (c *Command) SkipBinaryName(b bool) *Command {
	c.skipBinaryName = b
	return c
}

// Multicall enables busybox-style dispatch where args[0] is treated as a subcommand name.
func (c *Command) Multicall(b bool) *Command {
	c.multicall = b
	return c
}

// GetName returns the command name.
func (c *Command) GetName() string { return c.name }

// GetVersion returns the version string.
func (c *Command) GetVersion() string { return c.version }

// GetAbout returns the short description.
func (c *Command) GetAbout() string { return c.about }

// GetAuthor returns the author string.
func (c *Command) GetAuthor() string { return c.author }

// GetArgs returns the command's argument definitions.
func (c *Command) GetArgs() []*Arg { return c.args }

// GetSubcommands returns the command's subcommands.
func (c *Command) GetSubcommands() []*Command { return c.subcommands }

// FindArg returns the arg with the given name, or nil.
func (c *Command) FindArg(name string) *Arg { return c.findArg(name) }

// FindSubcommand returns the subcommand with the given name or alias, or nil.
func (c *Command) FindSubcommand(name string) *Command { return c.findSubcommand(name) }

// MutArg finds an arg by name and applies fn to it.
func (c *Command) MutArg(name string, fn func(*Arg)) *Command {
	if a := c.findArg(name); a != nil {
		fn(a)
	}
	return c
}

// MutSubcommand finds a subcommand by name and applies fn to it.
func (c *Command) MutSubcommand(name string, fn func(*Command)) *Command {
	if sub := c.findSubcommand(name); sub != nil {
		fn(sub)
	}
	return c
}

// Parse parses the given arguments using the builder API and returns Matches.
func (c *Command) Parse(args []string) (*Matches, error) {
	c.injectHelpAndVersion()
	return parseCommand(c, args)
}

// GenerateCompletion generates a shell completion script for this command.
func (c *Command) GenerateCompletion(shell Shell) string {
	return GenerateCompletion(c, shell)
}

func (c *Command) injectHelpAndVersion() {
	hasHelp := false
	hasVersion := false
	for _, a := range c.args {
		if a.name == "help" {
			hasHelp = true
		}
		if a.name == "version" {
			hasVersion = true
		}
	}
	if !hasHelp {
		c.args = append(c.args, NewArg("help").Short('h').Long("help").Help("Print help").Action(SetTrue))
	}
	if !hasVersion && c.version != "" {
		c.args = append(c.args, NewArg("version").Short('V').Long("version").Help("Print version").Action(SetTrue))
	}
	for _, sub := range c.subcommands {
		sub.injectHelpAndVersion()
	}
}

func (c *Command) findSubcommand(name string) *Command {
	for _, sub := range c.subcommands {
		if sub.name == name {
			return sub
		}
		for _, alias := range sub.aliases {
			if alias == name {
				return sub
			}
		}
	}
	return nil
}

func (c *Command) findArg(name string) *Arg {
	for _, a := range c.args {
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

func (c *Command) findArgByShort(s rune) *Arg {
	for _, a := range c.args {
		if a.short == s {
			return a
		}
	}
	return nil
}

func (c *Command) positionalArgs() []*Arg {
	var indexed, unindexed []*Arg
	for _, a := range c.args {
		if !a.positional {
			continue
		}
		if a.index > 0 {
			indexed = append(indexed, a)
		} else {
			unindexed = append(unindexed, a)
		}
	}

	slices.SortStableFunc(indexed, func(a, b *Arg) int {
		return a.index - b.index
	})

	return append(indexed, unindexed...)
}

func (c *Command) globalArgs() []*Arg {
	var result []*Arg
	for _, a := range c.args {
		if a.global {
			result = append(result, a)
		}
	}
	return result
}
