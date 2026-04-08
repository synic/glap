package glap

import "fmt"

type (
	// UnknownArgError is returned when an unrecognized argument is encountered.
	UnknownArgError struct {
		Arg string
	}

	// MissingRequiredError is returned when a required argument is not provided.
	MissingRequiredError struct {
		Arg string
	}

	// InvalidValueError is returned when an argument's value is not in the allowed set.
	InvalidValueError struct {
		Arg     string
		Value   string
		Allowed []string
	}

	// ConflictError is returned when two conflicting arguments are both provided.
	ConflictError struct {
		Arg      string
		Conflict string
	}

	// MissingDependencyError is returned when an argument's required dependency is missing.
	MissingDependencyError struct {
		Arg      string
		Requires string
	}

	// GroupViolationError is returned when an argument group constraint is violated.
	GroupViolationError struct {
		Group   string
		Message string
	}

	// TooFewValuesError is returned when an argument receives fewer values than required.
	TooFewValuesError struct {
		Arg string
		Min int
		Got int
	}

	// TooManyValuesError is returned when an argument receives more values than allowed.
	TooManyValuesError struct {
		Arg string
		Max int
		Got int
	}

	// HelpRequestedError is returned when --help or -h is invoked. Message contains
	// the formatted help text.
	HelpRequestedError struct {
		Message string
	}

	// VersionRequestedError is returned when --version or -V is invoked.
	VersionRequestedError struct {
		Message string
	}

	// ConditionalRequiredError is returned when a conditional requirement is not met
	// (e.g., RequiredIfEq or RequiredUnlessPresent).
	ConditionalRequiredError struct {
		Arg       string
		Condition string
	}

	// RequireEqualsError is returned when an argument requires --arg=value syntax
	// but was provided as --arg value.
	RequireEqualsError struct {
		Arg string
	}
)

func (e *UnknownArgError) Error() string {
	return fmt.Sprintf("unknown argument: %s", e.Arg)
}

func (e *MissingRequiredError) Error() string {
	return fmt.Sprintf("required argument missing: --%s", e.Arg)
}

func (e *InvalidValueError) Error() string {
	return fmt.Sprintf("invalid value '%s' for argument '--%s': allowed values are %v", e.Value, e.Arg, e.Allowed)
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("argument '--%s' conflicts with '--%s'", e.Arg, e.Conflict)
}

func (e *MissingDependencyError) Error() string {
	return fmt.Sprintf("argument '--%s' requires '--%s'", e.Arg, e.Requires)
}

func (e *GroupViolationError) Error() string {
	return fmt.Sprintf("group '%s': %s", e.Group, e.Message)
}

func (e *TooFewValuesError) Error() string {
	return fmt.Sprintf("argument '--%s' requires at least %d values, got %d", e.Arg, e.Min, e.Got)
}

func (e *TooManyValuesError) Error() string {
	return fmt.Sprintf("argument '--%s' accepts at most %d values, got %d", e.Arg, e.Max, e.Got)
}

func (e *HelpRequestedError) Error() string {
	return e.Message
}

func (e *VersionRequestedError) Error() string {
	return e.Message
}

func (e *ConditionalRequiredError) Error() string {
	return fmt.Sprintf("required argument missing: --%s (%s)", e.Arg, e.Condition)
}

func (e *RequireEqualsError) Error() string {
	return fmt.Sprintf("argument '--%s' requires '=' syntax (--%s=value)", e.Arg, e.Arg)
}
