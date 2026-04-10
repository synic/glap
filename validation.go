package glap

import "fmt"

func validate(cmd *Command, m *Matches, args []*Arg) error {
	if err := validateRequired(m, args); err != nil {
		return err
	}
	if err := validatePossibleValues(m, args); err != nil {
		return err
	}
	if err := validateConflicts(m, args); err != nil {
		return err
	}
	if err := validateRequires(m, args); err != nil {
		return err
	}
	if err := validateNumArgs(m, args); err != nil {
		return err
	}
	if err := validateConditionalRequired(m, args); err != nil {
		return err
	}
	if err := validateCustom(m, args); err != nil {
		return err
	}
	if err := validateGroups(cmd, m); err != nil {
		return err
	}
	return nil
}

func validateRequired(m *Matches, args []*Arg) error {
	for _, a := range args {
		if a.required {
			if _, ok := m.args[a.name]; !ok {
				return &MissingRequiredError{Arg: a.name}
			}
		}
	}
	return nil
}

func validatePossibleValues(m *Matches, args []*Arg) error {
	for _, a := range args {
		if len(a.possibleValues) == 0 {
			continue
		}
		ma, ok := m.args[a.name]
		if !ok {
			continue
		}
		for _, v := range ma.Values {
			valid := false
			for _, pv := range a.possibleValues {
				if v == pv {
					valid = true
					break
				}
			}
			if !valid {
				return &InvalidValueError{Arg: a.name, Value: v, Allowed: a.possibleValues}
			}
		}
	}
	return nil
}

func validateConflicts(m *Matches, args []*Arg) error {
	for _, a := range args {
		if _, ok := m.args[a.name]; !ok {
			continue
		}
		for _, conflict := range a.conflictsWith {
			if _, ok := m.args[conflict]; ok {
				return &ConflictError{Arg: a.name, Conflict: conflict}
			}
		}
	}
	return nil
}

func validateRequires(m *Matches, args []*Arg) error {
	for _, a := range args {
		if _, ok := m.args[a.name]; !ok {
			continue
		}
		for _, req := range a.requires {
			if _, ok := m.args[req]; !ok {
				return &MissingDependencyError{Arg: a.name, Requires: req}
			}
		}
	}
	return nil
}

func validateNumArgs(m *Matches, args []*Arg) error {
	for _, a := range args {
		if !a.numArgs.Set {
			continue
		}
		ma, ok := m.args[a.name]
		if !ok {
			continue
		}
		count := len(ma.Values)
		if count < a.numArgs.Min {
			return &TooFewValuesError{Arg: a.name, Min: a.numArgs.Min, Got: count}
		}
		if a.numArgs.Max >= 0 && count > a.numArgs.Max {
			return &TooManyValuesError{Arg: a.name, Max: a.numArgs.Max, Got: count}
		}
	}
	return nil
}

func validateConditionalRequired(m *Matches, args []*Arg) error {
	for _, a := range args {
		if _, ok := m.args[a.name]; ok {
			continue
		}

		for _, rule := range a.requiredIfEq {
			if ma, ok := m.args[rule.ArgName]; ok {
				if len(ma.Values) > 0 && ma.Values[0] == rule.Value {
					return &ConditionalRequiredError{
						Arg:       a.name,
						Condition: fmt.Sprintf("required when --%s=%s", rule.ArgName, rule.Value),
					}
				}
			}
		}

		if len(a.requiredUnless) > 0 {
			anyPresent := false
			for _, other := range a.requiredUnless {
				if _, ok := m.args[other]; ok {
					anyPresent = true
					break
				}
			}
			if !anyPresent {
				return &ConditionalRequiredError{
					Arg:       a.name,
					Condition: fmt.Sprintf("required unless --%s is present", a.requiredUnless[0]),
				}
			}
		}
	}
	return nil
}

func validateCustom(m *Matches, args []*Arg) error {
	for _, a := range args {
		if a.validator == nil {
			continue
		}
		ma, ok := m.args[a.name]
		if !ok {
			continue
		}
		for _, v := range ma.Values {
			if err := a.validator(v); err != nil {
				return &InvalidValueError{Arg: a.name, Value: v, Allowed: []string{err.Error()}}
			}
		}
	}
	return nil
}

func validateGroups(cmd *Command, m *Matches) error {
	// Collect group memberships declared on individual args via Arg.Group()
	// or the struct-tag `group=...`. These are merged with the names
	// explicitly added via ArgGroup.Arg(...).
	extra := make(map[string][]string)
	definedGroups := make(map[string]bool, len(cmd.argGroups))
	for _, g := range cmd.argGroups {
		definedGroups[g.name] = true
	}
	for _, a := range cmd.args {
		if a.groupID == "" {
			continue
		}
		if !definedGroups[a.groupID] {
			return &UndefinedGroupError{Arg: a.name, Group: a.groupID}
		}
		extra[a.groupID] = append(extra[a.groupID], a.name)
	}

	for _, g := range cmd.argGroups {
		members := make([]string, 0, len(g.args)+len(extra[g.name]))
		seen := make(map[string]bool)
		for _, argName := range g.args {
			if seen[argName] {
				continue
			}
			seen[argName] = true
			members = append(members, argName)
		}
		for _, argName := range extra[g.name] {
			if seen[argName] {
				continue
			}
			seen[argName] = true
			members = append(members, argName)
		}

		present := 0
		var presentArgs []string
		for _, argName := range members {
			if _, ok := m.args[argName]; ok {
				present++
				presentArgs = append(presentArgs, argName)
			}
		}

		if g.multiple {
			if g.required && present == 0 {
				return &GroupViolationError{Group: g.name, Message: "at least one argument is required"}
			}
		} else {
			if g.required && present == 0 {
				return &GroupViolationError{Group: g.name, Message: "one argument is required"}
			}
			if present > 1 {
				return &GroupViolationError{Group: g.name, Message: "arguments are mutually exclusive"}
			}
		}
	}
	return nil
}
