package glap

// ArgAction determines how an argument's value is handled during parsing.
type ArgAction int

const (
	// Set stores the value, replacing any previous value.
	Set ArgAction = iota
	// Append adds the value to a list, allowing repeated use.
	Append
	// SetTrue sets a boolean flag to true. No value is consumed.
	SetTrue
	// Count increments a counter each time the flag appears.
	Count
	// SetFalse sets a boolean flag to false. No value is consumed.
	SetFalse
)

// String returns the string representation of the action (e.g., "set", "append", "count").
func (a ArgAction) String() string {
	switch a {
	case Set:
		return "set"
	case Append:
		return "append"
	case SetTrue:
		return "set_true"
	case Count:
		return "count"
	case SetFalse:
		return "set_false"
	default:
		return "unknown"
	}
}

func (a ArgAction) takesValue() bool {
	return a == Set || a == Append
}

func (a ArgAction) acceptsValue() bool {
	return a.takesValue()
}

func (a ArgAction) acceptsMultipleValues() bool {
	return a == Append
}
