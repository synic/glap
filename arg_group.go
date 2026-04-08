package glap

// ArgGroup defines a group of arguments with mutual exclusion or co-occurrence constraints.
type ArgGroup struct {
	name     string
	args     []string
	required bool
	multiple bool // false = mutually exclusive, true = all required together
}

// NewArgGroup creates a new ArgGroup with the given name.
func NewArgGroup(name string) *ArgGroup {
	return &ArgGroup{name: name}
}

// Arg adds an argument name to this group.
func (g *ArgGroup) Arg(name string) *ArgGroup {
	g.args = append(g.args, name)
	return g
}

// Required makes the group require at least one argument to be present.
func (g *ArgGroup) Required(b bool) *ArgGroup {
	g.required = b
	return g
}

// Multiple switches the group from mutual exclusion (default) to co-occurrence mode,
// where all arguments in the group must be provided together.
func (g *ArgGroup) Multiple(b bool) *ArgGroup {
	g.multiple = b
	return g
}
