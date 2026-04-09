package glap

import (
	"fmt"
	"reflect"
	"strconv"
)

// ValueSource indicates where a matched argument's value came from.
type ValueSource int

const (
	// SourceCLI indicates the value was provided on the command line.
	SourceCLI ValueSource = iota
	// SourceEnv indicates the value came from an environment variable.
	SourceEnv
	// SourceDefault indicates the value is the configured default.
	SourceDefault
)

// MatchedArg holds the parsed result for a single argument.
type MatchedArg struct {
	Values      []string
	Source      ValueSource
	Occurrences int
}

// Matches holds the complete result of parsing via the builder API.
type Matches struct {
	args              map[string]*MatchedArg
	subcommandName    string
	subcommandMatches *Matches
}

func newMatches() *Matches {
	return &Matches{args: make(map[string]*MatchedArg)}
}

// SubcommandName returns the name of the matched subcommand, or empty if none.
func (m *Matches) SubcommandName() string {
	return m.subcommandName
}

// SubcommandMatches returns the Matches for the matched subcommand, or nil if none.
func (m *Matches) SubcommandMatches() *Matches {
	return m.subcommandMatches
}

// Contains reports whether the named argument was matched.
func (m *Matches) Contains(name string) bool {
	_, ok := m.args[name]
	return ok
}

// GetString returns the first value of the named argument as a string.
func (m *Matches) GetString(name string) (string, bool) {
	ma, ok := m.args[name]
	if !ok || len(ma.Values) == 0 {
		return "", false
	}
	return ma.Values[0], true
}

// GetBool returns the first value of the named argument parsed as a bool.
func (m *Matches) GetBool(name string) (bool, bool) {
	ma, ok := m.args[name]
	if !ok || len(ma.Values) == 0 {
		return false, false
	}
	b, err := strconv.ParseBool(ma.Values[0])
	if err != nil {
		return false, false
	}
	return b, true
}

// GetInt returns the first value of the named argument parsed as an int.
func (m *Matches) GetInt(name string) (int, bool) {
	ma, ok := m.args[name]
	if !ok || len(ma.Values) == 0 {
		return 0, false
	}
	v, err := strconv.Atoi(ma.Values[0])
	if err != nil {
		return 0, false
	}
	return v, true
}

// GetFloat returns the first value of the named argument parsed as a float64.
func (m *Matches) GetFloat(name string) (float64, bool) {
	ma, ok := m.args[name]
	if !ok || len(ma.Values) == 0 {
		return 0, false
	}
	v, err := strconv.ParseFloat(ma.Values[0], 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

// GetInt64 returns the first value of the named argument parsed as an int64.
func (m *Matches) GetInt64(name string) (int64, bool) {
	ma, ok := m.args[name]
	if !ok || len(ma.Values) == 0 {
		return 0, false
	}
	v, err := strconv.ParseInt(ma.Values[0], 10, 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

// GetUint returns the first value of the named argument parsed as a uint.
func (m *Matches) GetUint(name string) (uint, bool) {
	ma, ok := m.args[name]
	if !ok || len(ma.Values) == 0 {
		return 0, false
	}
	v, err := strconv.ParseUint(ma.Values[0], 10, 64)
	if err != nil {
		return 0, false
	}
	return uint(v), true
}

// GetUint64 returns the first value of the named argument parsed as a uint64.
func (m *Matches) GetUint64(name string) (uint64, bool) {
	ma, ok := m.args[name]
	if !ok || len(ma.Values) == 0 {
		return 0, false
	}
	v, err := strconv.ParseUint(ma.Values[0], 10, 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

// GetStringSlice returns all values of the named argument.
func (m *Matches) GetStringSlice(name string) ([]string, bool) {
	ma, ok := m.args[name]
	if !ok {
		return nil, false
	}
	return ma.Values, true
}

// GetOccurrences returns the number of times the named argument appeared.
func (m *Matches) GetOccurrences(name string) int {
	ma, ok := m.args[name]
	if !ok {
		return 0
	}
	return ma.Occurrences
}

// GetSource returns the ValueSource indicating where the named argument's value came from.
func (m *Matches) GetSource(name string) (ValueSource, bool) {
	ma, ok := m.args[name]
	if !ok {
		return 0, false
	}
	return ma.Source, true
}

// Scan populates a struct from the parsed matches using glap struct tags.
// This allows the builder API to write results into a struct without
// needing the full struct-tag parsing pipeline.
//
//	var cfg Config
//	matches.Scan(&cfg)
func (m *Matches) Scan(target any) error {
	v := reflect.ValueOf(target)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("glap: Scan target must be a pointer to a struct")
	}
	v = v.Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("glap")
		if tag == "" || tag == "-" {
			continue
		}
		name := tag
		if idx := indexOf(tag, ','); idx >= 0 {
			name = tag[:idx]
		}
		ma, ok := m.args[name]
		if !ok {
			continue
		}
		if err := setFieldValue(v.Field(i), field.Type, ma); err != nil {
			return fmt.Errorf("glap: field %s: %w", field.Name, err)
		}
	}
	return nil
}

func indexOf(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}

func (m *Matches) set(name string, value string, source ValueSource) {
	ma, ok := m.args[name]
	if !ok {
		ma = &MatchedArg{Source: source}
		m.args[name] = ma
	}
	ma.Values = []string{value}
	ma.Occurrences++
	ma.Source = source
}

func (m *Matches) appendValue(name string, value string, source ValueSource) {
	ma, ok := m.args[name]
	if !ok {
		ma = &MatchedArg{Source: source}
		m.args[name] = ma
	}
	ma.Values = append(ma.Values, value)
	ma.Occurrences++
	ma.Source = source
}

func (m *Matches) increment(name string) {
	ma, ok := m.args[name]
	if !ok {
		ma = &MatchedArg{Source: SourceCLI}
		m.args[name] = ma
	}
	ma.Occurrences++
	ma.Values = []string{strconv.Itoa(ma.Occurrences)}
}
