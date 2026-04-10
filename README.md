# glap

<img
src="https://github.com/egonelbre/gophers/raw/master/.thumb/vector/computer/music.png"
width="150" align="right" alt="Gopher">

A Go CLI argument parsing library inspired by Rust's
[clap](https://github.com/clap-rs/clap).

> **THIS LIBRARY IS BETA** — This library is under active development.
> The API may change before v1.0.

## Features

- **Struct tags** — define arguments declaratively on fields
- **Builder API** — construct arguments dynamically
- **Subcommands** — nested to arbitrary depth
- **Environment variable fallback** — CLI > Env > Default
- **Custom validators** — callback functions on arguments
- **Argument groups** — mutual exclusion / co-occurrence
- **Conflict and dependency rules** — `conflicts_with`,
  `requires`, `required_if_eq`, `required_unless`
- **Automatic help and version** — `--help` / `-h` and
  `--version` / `-V`, with short and long help variants
- **Shell completions** — Bash, Zsh, Fish, PowerShell
- **Colored output** — ANSI-styled help, configurable
- **Short flag combining** — `-abc` expands to `-a -b -c`
- **Positional arguments** — by field order or explicit index
- **Count and append actions** — `-vvv`, repeated `-f`
- **Value delimiters** — `--tag=a,b,c` splits into multiples
- **Trailing var args** — capture all remaining tokens
- **Negative number support** — `-1` as value, not flag
- **Multicall / busybox mode** — dispatch on binary name

## Install

```bash
go get github.com/synic/glap
```

## Quick Start

### Struct Tags

```go
type CLI struct {
    Config  string `glap:"config,short=c,required,help=Config"`
    Verbose bool   `glap:"verbose,short=v,help=Verbose"`
    Port    int    `glap:"port,short=p,default=8080,env=PORT"`
    Output  string `glap:"output,short=o,possible=json|text|yaml"`
}

var cli CLI
_, err := glap.Parse(&cli, os.Args[1:])
```

### Builder API

```go
app := glap.NewCommand("myapp").
    Version("1.0.0").
    Arg(glap.NewArg("config").
        Short('c').Required(true).Help("Config file")).
    Arg(glap.NewArg("verbose").
        Short('v').Action(glap.SetTrue))

matches, err := app.Parse(os.Args[1:])
config, _ := matches.GetString("config")
verbose, _ := matches.GetBool("verbose")
```

### App Metadata

```go
var cli CLI
app := glap.New(&cli).
    Name("myapp").
    Version("1.0.0").
    About("My cool app").
    Author("Your Name")

cmd, err := app.Parse(os.Args[1:])
```

## Struct Tag Reference

`glap:"<name>[,key=value|flag]..."`

| Key | Example | Description |
| --- | ------- | ----------- |
| *(first)* | `config` | Arg name (`--name`) |
| `short` | `short=c` | Short flag |
| `env` | `env=MY_VAR` | Env var fallback |
| `default` | `default=8080` | Default value |
| `required` | `required` | Mark required |
| `help` | `help=Desc` | Help text |
| `long_help` | `long_help=...` | `--help` detail |
| `action` | `action=count` | set/append/set_true/set_false/count |
| `possible` | `possible=a\|b` | Allowed values |
| `positional` | `positional` | Positional arg |
| `index` | `index=1` | Explicit position (1-based) |
| `hidden` | `hidden` | Hide from help |
| `global` | `global` | Propagate to subcommands |
| `alias` | `alias=cfg` | Long alias |
| `conflicts_with` | `conflicts_with=x` | Conflict |
| `requires` | `requires=x` | Dependency |
| `required_if_eq` | `required_if_eq=a:v` | Conditional required |
| `required_unless` | `required_unless=x` | Required unless present |
| `default_if` | `default_if=a:v:d` | Conditional default |
| `overrides_with` | `overrides_with=x` | Removes other when set |
| `group` | `group=grp` | Arg group |
| `value_name` | `value_name=FILE` | Help placeholder |
| `value_hint` | `value_hint=file_path` | Completion hint |
| `heading` | `heading=Advanced` | Help section |
| `display_order` | `display_order=1` | Help ordering |
| `num_args` | `num_args=1..3` | Value count |
| `delimiter` | `delimiter=comma` | Split values |
| `require_equals` | `require_equals` | Force `--a=v` syntax |
| `hide_default_value` | `hide_default_value` | Hide default in help |
| `allow_hyphen_values` | `allow_hyphen_values` | Accept `-` values |
| `trailing_var_arg` | `trailing_var_arg` | Capture remaining |
| `subcommand` | `subcommand` | Subcommand field |

### Subcommand-specific tag keys

On a field tagged `subcommand`, the following additional keys configure
the nested command:

| Key | Example | Description |
| --- | ------- | ----------- |
| `help` | `help=Desc` | Short description |
| `long_help` | `long_help=...` | `--help` detail |
| `version` | `version=1.0.0` | Subcommand version |
| `author` | `author=Name` | Subcommand author |
| `alias` | `alias=s` | Alternate invocation name |
| `hidden` | `hidden` | Hide from parent help |
| `display_order` | `display_order=2` | Subcommand help ordering |
| `subcommand_required` | `subcommand_required` | Require a nested subcommand |

## Subcommands

Pointer-to-struct fields. Nesting works to arbitrary depth.

```go
type ServeCLI struct {
    Port int `glap:"port,short=p,default=8080"`
}

type CLI struct {
    Verbose bool      `glap:"verbose,short=v,global"`
    Serve   *ServeCLI `glap:"serve,subcommand,help=Server"`
}

var cli CLI
cmd, err := glap.Parse(&cli, os.Args[1:])
// cmd == "serve", cli.Serve is non-nil when matched
```

## Documentation

Full API documentation is available on
[pkg.go.dev](https://pkg.go.dev/github.com/synic/glap).

Every exported type, method, and constant includes its
corresponding struct tag syntax where applicable.

## Examples

See [examples/](examples/). Build them all with:

```bash
make build
```

Binaries are written to `build/`. Each example has a `test.sh` script
that exercises its behavior end-to-end. Run them all with
`make test-examples`, or as part of the full check suite with
`make check`.

| Example | Description |
| ------- | ----------- |
| [basic](examples/basic) | Tags, defaults, delimiters |
| [env-override](examples/env-override) | Env var fallback |
| [validator](examples/validator) | Custom validators |
| [subcommands](examples/subcommands) | Nested subcommands |
| [builder](examples/builder) | Builder API |
| [positional](examples/positional) | Positional args |
| [conflicts-requires](examples/conflicts-requires) | Conflicts, conditionals |
| [groups](examples/groups) | Mutual exclusion |
| [trailing-var-arg](examples/trailing-var-arg) | Trailing args |
| [completions](examples/completions) | Shell completions |
