package glap

import (
	"fmt"
	"reflect"
	"testing"
)

type parityBasicCLI struct {
	Config  string   `glap:"config,required"`
	Verbose int      `glap:"verbose,short=v,action=count"`
	Port    int      `glap:"port,short=p,default=8080,env=PARITY_PORT"`
	Tags    []string `glap:"tag,short=t,action=append,delimiter=comma"`
	Offset  int      `glap:"offset,positional"`
}

type parityNumArgsCLI struct {
	Pair []string `glap:"pair,num_args=2"`
}

type parityServeCLI struct {
	Port int `glap:"port,default=8080"`
}

type parityRootCLI struct {
	Verbose bool            `glap:"verbose,short=v,global"`
	Serve   *parityServeCLI `glap:"serve,subcommand"`
}

type parityGroupCLI struct {
	JSON bool `glap:"json,group=format"`
	YAML bool `glap:"yaml,group=format"`
}

type parityRequireEqualsCLI struct {
	Color string `glap:"color,require_equals"`
}

type parityHelpCLI struct {
	Config string `glap:"config,short=c,alias=cfg,help=Config,long_help=Detailed config help,default=app.yaml,hide_default_value"`
}

type parityValidatorCLI struct {
	Port int `glap:"port"`
}

type parityConflictCLI struct {
	A bool `glap:"a,conflicts_with=b,conflicts_with=c"`
	B bool `glap:"b"`
	C bool `glap:"c"`
}

type parityScenario struct {
	name         string
	argv         []string
	env          map[string]string
	buildBuilder func() *Command
	newTarget    func() any
	buildApp     func(any) *App
	wantTarget   any
}

type matchesSnapshot struct {
	Args       map[string]matchedArgSnapshot
	Subcommand string
	Child      *matchesSnapshot
}

type matchedArgSnapshot struct {
	Values      []string
	Source      ValueSource
	Occurrences int
}

func TestAPIParity(t *testing.T) {
	validatorFn := func(v string) error {
		if v == "0" {
			return fmt.Errorf("port must be non-zero")
		}
		return nil
	}

	scenarios := []parityScenario{
		{
			name: "basic parsing parity",
			argv: []string{"--config", "app.yaml", "-vv", "--tag=a,b", "--tag", "c", "42"},
			env: map[string]string{
				"PARITY_PORT": "9090",
			},
			buildBuilder: func() *Command {
				return NewCommand("myapp").
					Arg(NewArg("config").Required(true)).
					Arg(NewArg("verbose").Short('v').Action(Count)).
					Arg(NewArg("port").Short('p').Default("8080").Env("PARITY_PORT")).
					Arg(NewArg("tag").Short('t').Action(Append).ValueDelimiter(",")).
					Arg(NewArg("offset").Positional(true))
			},
			newTarget: func() any { return &parityBasicCLI{} },
			buildApp:  func(target any) *App { return New(target).Name("myapp") },
			wantTarget: parityBasicCLI{
				Config:  "app.yaml",
				Verbose: 2,
				Port:    9090,
				Tags:    []string{"a", "b", "c"},
				Offset:  42,
			},
		},
		{
			name: "num args parity",
			argv: []string{"--pair", "left", "right"},
			buildBuilder: func() *Command {
				return NewCommand("myapp").
					Arg(NewArg("pair").SetNumArgs(2, 2))
			},
			newTarget: func() any { return &parityNumArgsCLI{} },
			buildApp:  func(target any) *App { return New(target).Name("myapp") },
			wantTarget: parityNumArgsCLI{
				Pair: []string{"left", "right"},
			},
		},
		{
			name: "subcommands and globals parity",
			argv: []string{"serve", "-v", "--port", "3000"},
			buildBuilder: func() *Command {
				return NewCommand("myapp").
					Arg(NewArg("verbose").Short('v').Action(SetTrue).Global(true)).
					Subcommand(NewCommand("serve").
						Arg(NewArg("port").Default("8080")))
			},
			newTarget: func() any { return &parityRootCLI{} },
			buildApp:  func(target any) *App { return New(target).Name("myapp") },
			wantTarget: parityRootCLI{
				Verbose: false,
				Serve:   &parityServeCLI{Port: 3000},
			},
		},
		{
			name: "arg group error parity",
			argv: []string{"--json", "--yaml"},
			buildBuilder: func() *Command {
				return NewCommand("myapp").
					Arg(NewArg("json").Action(SetTrue).Group("format")).
					Arg(NewArg("yaml").Action(SetTrue).Group("format")).
					ArgGroup(NewArgGroup("format"))
			},
			newTarget: func() any { return &parityGroupCLI{} },
			buildApp: func(target any) *App {
				return New(target).
					Name("myapp").
					ArgGroup(NewArgGroup("format"))
			},
		},
		{
			name: "require equals error parity",
			argv: []string{"--color", "blue"},
			buildBuilder: func() *Command {
				return NewCommand("myapp").
					Arg(NewArg("color").RequireEquals(true))
			},
			newTarget: func() any { return &parityRequireEqualsCLI{} },
			buildApp:  func(target any) *App { return New(target).Name("myapp") },
		},
		{
			name: "help output parity",
			argv: []string{"--help"},
			buildBuilder: func() *Command {
				return NewCommand("myapp").
					About("Short description").
					LongAbout("Long description").
					Arg(NewArg("config").
						Short('c').
						Alias("cfg").
						Help("Config").
						LongHelp("Detailed config help").
						Default("app.yaml").
						HideDefaultValue(true))
			},
			newTarget: func() any { return &parityHelpCLI{} },
			buildApp: func(target any) *App {
				return New(target).
					Name("myapp").
					About("Short description").
					LongAbout("Long description")
			},
		},
		{
			name: "validator error parity",
			argv: []string{"--port", "0"},
			buildBuilder: func() *Command {
				return NewCommand("myapp").
					Arg(NewArg("port").Validator(validatorFn))
			},
			newTarget: func() any { return &parityValidatorCLI{} },
			buildApp: func(target any) *App {
				return New(target).
					Name("myapp").
					Validator("port", validatorFn)
			},
		},
		{
			name: "repeated constraint parity",
			argv: []string{"--a", "--c"},
			buildBuilder: func() *Command {
				return NewCommand("myapp").
					Arg(NewArg("a").Action(SetTrue).ConflictsWith("b", "c")).
					Arg(NewArg("b").Action(SetTrue)).
					Arg(NewArg("c").Action(SetTrue))
			},
			newTarget: func() any { return &parityConflictCLI{} },
			buildApp:  func(target any) *App { return New(target).Name("myapp") },
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			runParityScenario(t, scenario)
		})
	}
}

func runParityScenario(t *testing.T, scenario parityScenario) {
	t.Helper()

	for key, value := range scenario.env {
		t.Setenv(key, value)
	}

	builderCmd := scenario.buildBuilder()
	builderCmd.injectHelpAndVersion()
	builderMatches, builderErr := parseCommand(builderCmd, scenario.argv)

	target := scenario.newTarget()
	app := scenario.buildApp(target)
	tagCmd, err := buildCommandWithValidators(app.command, app.target, app.validators)
	if err != nil {
		t.Fatalf("buildCommandWithValidators failed: %v", err)
	}
	tagCmd.injectHelpAndVersion()
	tagMatches, tagErr := parseCommand(tagCmd, scenario.argv)
	if tagErr == nil {
		if err := writeBack(target, tagMatches, tagCmd); err != nil {
			tagErr = err
		}
	}

	compareErrors(t, builderErr, tagErr)
	if builderErr != nil || tagErr != nil {
		return
	}

	builderSnapshot := snapshotMatches(builderMatches)
	tagSnapshot := snapshotMatches(tagMatches)
	if !reflect.DeepEqual(builderSnapshot, tagSnapshot) {
		t.Fatalf("match snapshots differ\nbuilder: %#v\ntags: %#v", builderSnapshot, tagSnapshot)
	}

	if scenario.wantTarget != nil {
		gotTarget := reflect.ValueOf(target).Elem().Interface()
		if !reflect.DeepEqual(gotTarget, scenario.wantTarget) {
			t.Fatalf("target mismatch\ngot:  %#v\nwant: %#v", gotTarget, scenario.wantTarget)
		}
	}
}

func compareErrors(t *testing.T, builderErr, tagErr error) {
	t.Helper()

	if (builderErr == nil) != (tagErr == nil) {
		t.Fatalf("error mismatch\nbuilder: %v\ntags: %v", builderErr, tagErr)
	}
	if builderErr == nil {
		return
	}
	if reflect.TypeOf(builderErr) != reflect.TypeOf(tagErr) {
		t.Fatalf("error type mismatch\nbuilder: %T\ntags: %T", builderErr, tagErr)
	}
	if builderErr.Error() != tagErr.Error() {
		t.Fatalf("error message mismatch\nbuilder: %q\ntags: %q", builderErr.Error(), tagErr.Error())
	}
}

func snapshotMatches(m *Matches) *matchesSnapshot {
	if m == nil {
		return nil
	}

	snapshot := &matchesSnapshot{
		Args:       make(map[string]matchedArgSnapshot, len(m.args)),
		Subcommand: m.subcommandName,
	}
	for name, arg := range m.args {
		snapshot.Args[name] = matchedArgSnapshot{
			Values:      append([]string(nil), arg.Values...),
			Source:      arg.Source,
			Occurrences: arg.Occurrences,
		}
	}
	snapshot.Child = snapshotMatches(m.subcommandMatches)
	return snapshot
}
