package glap

import (
	"errors"
	"testing"
)

func TestBasicStructTags(t *testing.T) {
	type CLI struct {
		Config  string `glap:"config,short=c,help=Path to config file"`
		Verbose bool   `glap:"verbose,short=v,help=Enable verbose output"`
		Port    int    `glap:"port,short=p,default=8080,help=Port to listen on"`
	}

	var cli CLI
	_, err := Parse(&cli, []string{"--config", "app.yaml", "-v", "--port", "3000"})
	if err != nil {
		t.Fatal(err)
	}
	if cli.Config != "app.yaml" {
		t.Errorf("Config = %q, want %q", cli.Config, "app.yaml")
	}
	if !cli.Verbose {
		t.Error("Verbose should be true")
	}
	if cli.Port != 3000 {
		t.Errorf("Port = %d, want %d", cli.Port, 3000)
	}
}

func TestDefaultValues(t *testing.T) {
	type CLI struct {
		Port   int    `glap:"port,default=8080"`
		Output string `glap:"output,default=text"`
	}

	var cli CLI
	_, err := Parse(&cli, []string{})
	if err != nil {
		t.Fatal(err)
	}
	if cli.Port != 8080 {
		t.Errorf("Port = %d, want %d", cli.Port, 8080)
	}
	if cli.Output != "text" {
		t.Errorf("Output = %q, want %q", cli.Output, "text")
	}
}

func TestEnvFallback(t *testing.T) {
	type CLI struct {
		Port int `glap:"port,env=GLAP_TEST_PORT"`
	}

	t.Setenv("GLAP_TEST_PORT", "9090")

	var cli CLI
	_, err := Parse(&cli, []string{})
	if err != nil {
		t.Fatal(err)
	}
	if cli.Port != 9090 {
		t.Errorf("Port = %d, want %d", cli.Port, 9090)
	}
}

func TestCLIOverridesEnv(t *testing.T) {
	type CLI struct {
		Port int `glap:"port,env=GLAP_TEST_PORT2"`
	}

	t.Setenv("GLAP_TEST_PORT2", "9090")

	var cli CLI
	_, err := Parse(&cli, []string{"--port", "3000"})
	if err != nil {
		t.Fatal(err)
	}
	if cli.Port != 3000 {
		t.Errorf("Port = %d, want %d (CLI should override env)", cli.Port, 3000)
	}
}

func TestEnvOverridesDefault(t *testing.T) {
	type CLI struct {
		Port int `glap:"port,default=8080,env=GLAP_TEST_PORT3"`
	}

	t.Setenv("GLAP_TEST_PORT3", "5000")

	var cli CLI
	_, err := Parse(&cli, []string{})
	if err != nil {
		t.Fatal(err)
	}
	if cli.Port != 5000 {
		t.Errorf("Port = %d, want %d (env should override default)", cli.Port, 5000)
	}
}

func TestLongEquals(t *testing.T) {
	type CLI struct {
		Config string `glap:"config"`
	}

	var cli CLI
	_, err := Parse(&cli, []string{"--config=app.yaml"})
	if err != nil {
		t.Fatal(err)
	}
	if cli.Config != "app.yaml" {
		t.Errorf("Config = %q, want %q", cli.Config, "app.yaml")
	}
}

func TestShortCombined(t *testing.T) {
	type CLI struct {
		A bool `glap:"aa,short=a"`
		B bool `glap:"bb,short=b"`
		C bool `glap:"cc,short=c"`
	}

	var cli CLI
	_, err := Parse(&cli, []string{"-abc"})
	if err != nil {
		t.Fatal(err)
	}
	if !cli.A || !cli.B || !cli.C {
		t.Errorf("A=%v B=%v C=%v, all should be true", cli.A, cli.B, cli.C)
	}
}

func TestShortWithValue(t *testing.T) {
	type CLI struct {
		Port int `glap:"port,short=p"`
	}

	var cli CLI
	_, err := Parse(&cli, []string{"-p3000"})
	if err != nil {
		t.Fatal(err)
	}
	if cli.Port != 3000 {
		t.Errorf("Port = %d, want %d", cli.Port, 3000)
	}
}

func TestDashDash(t *testing.T) {
	type CLI struct {
		Flag bool   `glap:"flag,short=f"`
		File string `glap:"file,positional"`
	}

	var cli CLI
	_, err := Parse(&cli, []string{"-f", "--", "--not-a-flag"})
	if err != nil {
		t.Fatal(err)
	}
	if !cli.Flag {
		t.Error("Flag should be true")
	}
	if cli.File != "--not-a-flag" {
		t.Errorf("File = %q, want %q", cli.File, "--not-a-flag")
	}
}

func TestPositionalArg(t *testing.T) {
	type CLI struct {
		File string `glap:"file,positional,required"`
	}

	var cli CLI
	_, err := Parse(&cli, []string{"hello.txt"})
	if err != nil {
		t.Fatal(err)
	}
	if cli.File != "hello.txt" {
		t.Errorf("File = %q, want %q", cli.File, "hello.txt")
	}
}

func TestFloatArg(t *testing.T) {
	type CLI struct {
		Rate float64 `glap:"rate,default=0.5"`
	}

	var cli CLI
	_, err := Parse(&cli, []string{"--rate", "1.5"})
	if err != nil {
		t.Fatal(err)
	}
	if cli.Rate != 1.5 {
		t.Errorf("Rate = %f, want %f", cli.Rate, 1.5)
	}
}

func TestUintArg(t *testing.T) {
	type CLI struct {
		Count uint `glap:"count,default=1"`
	}

	var cli CLI
	_, err := Parse(&cli, []string{"--count", "42"})
	if err != nil {
		t.Fatal(err)
	}
	if cli.Count != 42 {
		t.Errorf("Count = %d, want %d", cli.Count, 42)
	}
}

func TestEmptyArgs(t *testing.T) {
	type CLI struct {
		Port int `glap:"port,default=8080"`
	}

	var cli CLI
	_, err := Parse(&cli, []string{})
	if err != nil {
		t.Fatal(err)
	}
	if cli.Port != 8080 {
		t.Errorf("Port = %d, want %d", cli.Port, 8080)
	}
}

func TestNonPointerTarget(t *testing.T) {
	type CLI struct {
		Port int `glap:"port"`
	}

	var cli CLI
	_, err := Parse(cli, []string{}) // not a pointer
	if err == nil {
		t.Fatal("expected error for non-pointer target")
	}
}

func TestSubcommands(t *testing.T) {
	type ServeCLI struct {
		Port int `glap:"port,short=p,default=8080"`
	}
	type CLI struct {
		Verbose bool      `glap:"verbose,short=v,global"`
		Serve   *ServeCLI `glap:"serve,subcommand,help=Start the server"`
	}

	var cli CLI
	cmd, err := Parse(&cli, []string{"-v", "serve", "--port", "3000"})
	if err != nil {
		t.Fatal(err)
	}
	if cmd != "serve" {
		t.Errorf("cmd = %q, want %q", cmd, "serve")
	}
	if !cli.Verbose {
		t.Error("Verbose should be true")
	}
	if cli.Serve == nil {
		t.Fatal("Serve should not be nil")
	}
	if cli.Serve.Port != 3000 {
		t.Errorf("Serve.Port = %d, want %d", cli.Serve.Port, 3000)
	}
}

func TestNestedSubcommands(t *testing.T) {
	type RemoteAddCLI struct {
		Name string `glap:"name,positional,required"`
		URL  string `glap:"url,positional,required"`
	}
	type RemoteCLI struct {
		Add *RemoteAddCLI `glap:"add,subcommand,help=Add a remote"`
	}
	type CLI struct {
		Verbose bool       `glap:"verbose,short=v,global"`
		Remote  *RemoteCLI `glap:"remote,subcommand,help=Manage remotes"`
	}

	var cli CLI
	cmd, err := Parse(&cli, []string{"remote", "add", "origin", "https://example.com"})
	if err != nil {
		t.Fatal(err)
	}
	if cmd != "remote add" {
		t.Errorf("cmd = %q, want %q", cmd, "remote add")
	}
	if cli.Remote == nil {
		t.Fatal("Remote should not be nil")
	}
	if cli.Remote.Add == nil {
		t.Fatal("Remote.Add should not be nil")
	}
	if cli.Remote.Add.Name != "origin" {
		t.Errorf("Name = %q, want %q", cli.Remote.Add.Name, "origin")
	}
	if cli.Remote.Add.URL != "https://example.com" {
		t.Errorf("URL = %q, want %q", cli.Remote.Add.URL, "https://example.com")
	}
}

func TestNoSubcommandSelected(t *testing.T) {
	type ServeCLI struct {
		Port int `glap:"port,default=8080"`
	}
	type CLI struct {
		Verbose bool      `glap:"verbose,short=v"`
		Serve   *ServeCLI `glap:"serve,subcommand"`
	}

	var cli CLI
	cmd, err := Parse(&cli, []string{"-v"})
	if err != nil {
		t.Fatal(err)
	}
	if cmd != "" {
		t.Errorf("cmd = %q, want empty", cmd)
	}
	if cli.Serve != nil {
		t.Error("Serve should be nil when not selected")
	}
}

func TestSubcommandAlias(t *testing.T) {
	app := NewCommand("myapp").
		Subcommand(NewCommand("serve").Alias("s").
			Arg(NewArg("port").Default("8080")))

	m, err := app.Parse([]string{"s", "--port", "3000"})
	if err != nil {
		t.Fatal(err)
	}
	if m.SubcommandName() != "serve" {
		t.Errorf("subcommand = %q, want %q", m.SubcommandName(), "serve")
	}
}

func TestGlobalArgPropagation(t *testing.T) {
	type ServeCLI struct {
		Port int `glap:"port,default=8080"`
	}
	type CLI struct {
		Verbose bool      `glap:"verbose,short=v,global"`
		Serve   *ServeCLI `glap:"serve,subcommand"`
	}

	var cli CLI
	cmd, err := Parse(&cli, []string{"serve", "-v", "--port", "3000"})
	if err != nil {
		t.Fatal(err)
	}
	if cmd != "serve" {
		t.Errorf("cmd = %q, want %q", cmd, "serve")
	}
	if cli.Serve == nil {
		t.Fatal("Serve should not be nil")
	}
	if cli.Serve.Port != 3000 {
		t.Errorf("Port = %d, want %d", cli.Serve.Port, 3000)
	}
}

func TestBuilderBasic(t *testing.T) {
	app := NewCommand("myapp").
		Arg(NewArg("config").Short('c').Help("Config file").Required(true)).
		Arg(NewArg("verbose").Short('v').Action(SetTrue))

	m, err := app.Parse([]string{"-c", "app.yaml", "-v"})
	if err != nil {
		t.Fatal(err)
	}

	config, ok := m.GetString("config")
	if !ok || config != "app.yaml" {
		t.Errorf("config = %q, want %q", config, "app.yaml")
	}

	verbose, ok := m.GetBool("verbose")
	if !ok || !verbose {
		t.Error("verbose should be true")
	}
}

func TestBuilderSubcommand(t *testing.T) {
	app := NewCommand("myapp").
		Subcommand(NewCommand("serve").
			Arg(NewArg("port").Short('p').Default("8080")))

	m, err := app.Parse([]string{"serve", "--port", "3000"})
	if err != nil {
		t.Fatal(err)
	}

	if m.SubcommandName() != "serve" {
		t.Errorf("subcommand = %q, want %q", m.SubcommandName(), "serve")
	}
	port, ok := m.SubcommandMatches().GetInt("port")
	if !ok || port != 3000 {
		t.Errorf("port = %d, want %d", port, 3000)
	}
}

func TestBuilderCount(t *testing.T) {
	app := NewCommand("myapp").
		Arg(NewArg("verbose").Short('v').Action(Count))

	m, err := app.Parse([]string{"-v", "-v", "-v"})
	if err != nil {
		t.Fatal(err)
	}

	if m.GetOccurrences("verbose") != 3 {
		t.Errorf("occurrences = %d, want %d", m.GetOccurrences("verbose"), 3)
	}
}

func TestBuilderAppend(t *testing.T) {
	app := NewCommand("myapp").
		Arg(NewArg("file").Short('f').Action(Append))

	m, err := app.Parse([]string{"-f", "a.txt", "-f", "b.txt"})
	if err != nil {
		t.Fatal(err)
	}

	files, ok := m.GetStringSlice("file")
	if !ok || len(files) != 2 {
		t.Errorf("files = %v, want [a.txt b.txt]", files)
	}
}

func TestBuilderGetFloat(t *testing.T) {
	app := NewCommand("myapp").
		Arg(NewArg("rate"))

	m, err := app.Parse([]string{"--rate", "3.14"})
	if err != nil {
		t.Fatal(err)
	}
	v, ok := m.GetFloat("rate")
	if !ok || v != 3.14 {
		t.Errorf("rate = %f, want %f", v, 3.14)
	}
}

func TestBuilderContains(t *testing.T) {
	app := NewCommand("myapp").
		Arg(NewArg("verbose").Short('v').Action(SetTrue))

	m, err := app.Parse([]string{"-v"})
	if err != nil {
		t.Fatal(err)
	}
	if !m.Contains("verbose") {
		t.Error("should contain verbose")
	}
	if m.Contains("nonexistent") {
		t.Error("should not contain nonexistent")
	}
}

func TestBuilderGetSource(t *testing.T) {
	app := NewCommand("myapp").
		Arg(NewArg("port").Default("8080"))

	m, err := app.Parse([]string{})
	if err != nil {
		t.Fatal(err)
	}
	src, ok := m.GetSource("port")
	if !ok || src != SourceDefault {
		t.Errorf("source = %v, want SourceDefault", src)
	}
}

func TestAllowNegativeNumbers(t *testing.T) {
	type CLI struct {
		Value int `glap:"value,positional"`
	}
	var cli CLI
	app := New(&cli).Name("myapp").AllowNegativeNumbers(true)
	_, err := app.Parse([]string{"-1"})
	if err != nil {
		t.Fatal(err)
	}
	if cli.Value != -1 {
		t.Errorf("Value = %d, want -1", cli.Value)
	}
}

func TestAllowNegativeNumbersFloat(t *testing.T) {
	app := NewCommand("myapp").
		AllowNegativeNumbers(true).
		Arg(NewArg("val").Positional(true))

	m, err := app.Parse([]string{"-3.14"})
	if err != nil {
		t.Fatal(err)
	}
	v, _ := m.GetFloat("val")
	if v != -3.14 {
		t.Errorf("val = %f, want -3.14", v)
	}
}

func TestNegativeNumberWithoutAllow(t *testing.T) {
	app := NewCommand("myapp").
		Arg(NewArg("val").Positional(true))

	_, err := app.Parse([]string{"-1"})
	if err == nil {
		t.Fatal("expected error without AllowNegativeNumbers")
	}
}

func TestAllowHyphenValues(t *testing.T) {
	app := NewCommand("myapp").
		Arg(NewArg("pattern").Positional(true).AllowHyphenValues(true))

	m, err := app.Parse([]string{"-E"})
	if err != nil {
		t.Fatal(err)
	}
	v, _ := m.GetString("pattern")
	if v != "-E" {
		t.Errorf("pattern = %q, want %q", v, "-E")
	}
}

func TestAllowHyphenValuesStructTag(t *testing.T) {
	type CLI struct {
		Pattern string `glap:"pattern,positional,allow_hyphen_values"`
	}
	var cli CLI
	_, err := Parse(&cli, []string{"-v"})
	if err != nil {
		t.Fatal(err)
	}
	if cli.Pattern != "-v" {
		t.Errorf("Pattern = %q, want %q", cli.Pattern, "-v")
	}
}

func TestTrailingVarArg(t *testing.T) {
	app := NewCommand("myapp").
		Arg(NewArg("flag").Short('f').Action(SetTrue)).
		Arg(NewArg("rest").TrailingVarArg(true))

	m, err := app.Parse([]string{"-f", "a", "b", "-c", "--d"})
	if err != nil {
		t.Fatal(err)
	}
	if v, _ := m.GetBool("flag"); !v {
		t.Error("flag should be true")
	}
	vals, _ := m.GetStringSlice("rest")
	if len(vals) != 4 || vals[0] != "a" || vals[1] != "b" || vals[2] != "-c" || vals[3] != "--d" {
		t.Errorf("rest = %v, want [a b -c --d]", vals)
	}
}

func TestTrailingVarArgStructTag(t *testing.T) {
	type CLI struct {
		Verbose bool     `glap:"verbose,short=v"`
		Rest    []string `glap:"rest,trailing_var_arg"`
	}
	var cli CLI
	_, err := Parse(&cli, []string{"-v", "one", "two", "--three"})
	if err != nil {
		t.Fatal(err)
	}
	if !cli.Verbose {
		t.Error("Verbose should be true")
	}
	if len(cli.Rest) != 3 || cli.Rest[0] != "one" || cli.Rest[1] != "two" || cli.Rest[2] != "--three" {
		t.Errorf("Rest = %v, want [one two --three]", cli.Rest)
	}
}

func TestSetFalseBuilder(t *testing.T) {
	app := NewCommand("myapp").
		Arg(NewArg("no-color").Action(SetFalse))

	m, err := app.Parse([]string{"--no-color"})
	if err != nil {
		t.Fatal(err)
	}
	v, ok := m.GetBool("no-color")
	if !ok || v != false {
		t.Errorf("no-color = %v, want false", v)
	}
}

func TestSetFalseStructTag(t *testing.T) {
	type CLI struct {
		NoColor bool `glap:"no-color,action=set_false"`
	}
	var cli CLI
	cli.NoColor = true
	_, err := Parse(&cli, []string{"--no-color"})
	if err != nil {
		t.Fatal(err)
	}
	if cli.NoColor != false {
		t.Error("NoColor should be false")
	}
}

func TestCountAction(t *testing.T) {
	type CLI struct {
		Verbose int `glap:"verbose,short=v,action=count"`
	}

	var cli CLI
	_, err := Parse(&cli, []string{"-v", "-v", "-v"})
	if err != nil {
		t.Fatal(err)
	}
	if cli.Verbose != 3 {
		t.Errorf("Verbose = %d, want %d", cli.Verbose, 3)
	}
}

func TestAppendAction(t *testing.T) {
	type CLI struct {
		Files []string `glap:"file,short=f,action=append"`
	}

	var cli CLI
	_, err := Parse(&cli, []string{"-f", "a.txt", "-f", "b.txt"})
	if err != nil {
		t.Fatal(err)
	}
	if len(cli.Files) != 2 || cli.Files[0] != "a.txt" || cli.Files[1] != "b.txt" {
		t.Errorf("Files = %v, want [a.txt b.txt]", cli.Files)
	}
}

func TestValueDelimiterBuilder(t *testing.T) {
	app := NewCommand("myapp").
		Arg(NewArg("tag").Short('t').Action(Append).ValueDelimiter(","))

	m, err := app.Parse([]string{"--tag=a,b,c"})
	if err != nil {
		t.Fatal(err)
	}
	vals, ok := m.GetStringSlice("tag")
	if !ok || len(vals) != 3 || vals[0] != "a" || vals[1] != "b" || vals[2] != "c" {
		t.Errorf("tag = %v, want [a b c]", vals)
	}
}

func TestValueDelimiterCombined(t *testing.T) {
	app := NewCommand("myapp").
		Arg(NewArg("tag").Short('t').Action(Append).ValueDelimiter(","))

	m, err := app.Parse([]string{"--tag=a,b", "--tag=c"})
	if err != nil {
		t.Fatal(err)
	}
	vals, _ := m.GetStringSlice("tag")
	if len(vals) != 3 {
		t.Errorf("tag = %v, want [a b c]", vals)
	}
}

func TestValueDelimiterStructTag(t *testing.T) {
	type CLI struct {
		Tags []string `glap:"tag,short=t,action=append,delimiter=comma"`
	}
	var cli CLI
	_, err := Parse(&cli, []string{"--tag=x,y,z"})
	if err != nil {
		t.Fatal(err)
	}
	if len(cli.Tags) != 3 || cli.Tags[0] != "x" || cli.Tags[1] != "y" || cli.Tags[2] != "z" {
		t.Errorf("Tags = %v, want [x y z]", cli.Tags)
	}
}

func TestArgRequiredElseHelp(t *testing.T) {
	app := NewCommand("myapp").
		ArgRequiredElseHelp(true).
		Arg(NewArg("config").Required(true))

	_, err := app.Parse([]string{})
	if err == nil {
		t.Fatal("expected HelpRequestedError")
	}
	var helpErr *HelpRequestedError
	if !errors.As(err, &helpErr) {
		t.Errorf("expected HelpRequestedError, got %T: %v", err, err)
	}
}

func TestArgRequiredElseHelpWithArgs(t *testing.T) {
	app := NewCommand("myapp").
		ArgRequiredElseHelp(true).
		Arg(NewArg("config"))

	_, err := app.Parse([]string{"--config", "file.yaml"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestSubcommandRequiredEmptyArgs(t *testing.T) {
	app := NewCommand("myapp").
		SubcommandRequired(true).
		Subcommand(NewCommand("sub"))

	_, err := app.Parse([]string{})
	if err == nil {
		t.Fatal("expected HelpRequestedError")
	}
	var helpErr *HelpRequestedError
	if !errors.As(err, &helpErr) {
		t.Errorf("expected HelpRequestedError, got %T: %v", err, err)
	}
}

func TestSubcommandRequiredWithSubcommand(t *testing.T) {
	app := NewCommand("myapp").
		SubcommandRequired(true).
		Subcommand(NewCommand("sub"))

	m, err := app.Parse([]string{"sub"})
	if err != nil {
		t.Fatal(err)
	}
	if m.SubcommandName() != "sub" {
		t.Errorf("SubcommandName = %q, want %q", m.SubcommandName(), "sub")
	}
}

func TestSubcommandNotRequiredEmptyArgs(t *testing.T) {
	app := NewCommand("myapp").
		Subcommand(NewCommand("sub"))

	m, err := app.Parse([]string{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if m.SubcommandName() != "" {
		t.Errorf("SubcommandName = %q, want empty", m.SubcommandName())
	}
}

func TestSubcommandRequiredWithRootArgOnly(t *testing.T) {
	app := NewCommand("myapp").
		SubcommandRequired(true).
		Arg(NewArg("flag").Long("flag").Action(SetTrue)).
		Subcommand(NewCommand("sub"))

	_, err := app.Parse([]string{"--flag"})
	if err == nil {
		t.Fatal("expected HelpRequestedError")
	}
	var helpErr *HelpRequestedError
	if !errors.As(err, &helpErr) {
		t.Errorf("expected HelpRequestedError, got %T: %v", err, err)
	}
}

func TestSubcommandRequiredMissingRequiredArgTakesPrecedence(t *testing.T) {
	app := NewCommand("myapp").
		SubcommandRequired(true).
		Arg(NewArg("config").Long("config").Required(true)).
		Subcommand(NewCommand("sub"))

	_, err := app.Parse([]string{})
	if err == nil {
		t.Fatal("expected error")
	}
	var missingErr *MissingRequiredError
	if !errors.As(err, &missingErr) {
		t.Errorf("expected MissingRequiredError, got %T: %v", err, err)
	}
}

func TestSkipBinaryName(t *testing.T) {
	app := NewCommand("myapp").
		SkipBinaryName(true).
		Arg(NewArg("flag").Action(SetTrue))

	m, err := app.Parse([]string{"myapp", "--flag"})
	if err != nil {
		t.Fatal(err)
	}
	if v, _ := m.GetBool("flag"); !v {
		t.Error("flag should be true after skipping binary name")
	}
}

func TestMulticall(t *testing.T) {
	app := NewCommand("busybox").
		Multicall(true).
		Subcommand(NewCommand("ls").
			Arg(NewArg("all").Short('a').Action(SetTrue))).
		Subcommand(NewCommand("cat").
			Arg(NewArg("number").Short('n').Action(SetTrue)))

	m, err := app.Parse([]string{"ls", "-a"})
	if err != nil {
		t.Fatal(err)
	}
	if m.SubcommandName() != "ls" {
		t.Errorf("subcommand = %q, want %q", m.SubcommandName(), "ls")
	}
	if v, _ := m.SubcommandMatches().GetBool("all"); !v {
		t.Error("all should be true")
	}
}

func TestMulticallWithPath(t *testing.T) {
	app := NewCommand("busybox").
		Multicall(true).
		Subcommand(NewCommand("ls").
			Arg(NewArg("all").Short('a').Action(SetTrue)))

	m, err := app.Parse([]string{"/usr/bin/ls", "-a"})
	if err != nil {
		t.Fatal(err)
	}
	if m.SubcommandName() != "ls" {
		t.Errorf("subcommand = %q, want %q", m.SubcommandName(), "ls")
	}
}

func TestMulticallFallthrough(t *testing.T) {
	app := NewCommand("myapp").
		Multicall(true).
		Arg(NewArg("flag").Action(SetTrue)).
		Subcommand(NewCommand("sub"))

	m, err := app.Parse([]string{"myapp", "--flag"})
	if err != nil {
		t.Fatal(err)
	}
	if v, _ := m.GetBool("flag"); !v {
		t.Error("flag should be true when multicall falls through to self")
	}
}

func TestMulticallRootAlias(t *testing.T) {
	app := NewCommand("myapp").
		Multicall(true).
		Alias("alt").
		Arg(NewArg("flag").Action(SetTrue)).
		Subcommand(NewCommand("sub"))

	m, err := app.Parse([]string{"alt", "--flag"})
	if err != nil {
		t.Fatal(err)
	}
	if v, _ := m.GetBool("flag"); !v {
		t.Error("flag should be true when multicall falls through via root alias")
	}
}

func TestRequireEquals(t *testing.T) {
	app := NewCommand("myapp").
		Arg(NewArg("color").RequireEquals(true))

	_, err := app.Parse([]string{"--color=red"})
	if err != nil {
		t.Fatal(err)
	}

	_, err = app.Parse([]string{"--color", "red"})
	if err == nil {
		t.Fatal("expected RequireEqualsError")
	}
	var reqErr *RequireEqualsError
	if !errors.As(err, &reqErr) {
		t.Errorf("expected RequireEqualsError, got %T: %v", err, err)
	}
}

func TestRequireEqualsStructTag(t *testing.T) {
	type CLI struct {
		Color string `glap:"color,require_equals"`
	}
	var cli CLI
	_, err := Parse(&cli, []string{"--color=blue"})
	if err != nil {
		t.Fatal(err)
	}
	if cli.Color != "blue" {
		t.Errorf("Color = %q, want %q", cli.Color, "blue")
	}

	_, err = Parse(&cli, []string{"--color", "blue"})
	if err == nil {
		t.Fatal("expected RequireEqualsError")
	}
}

func TestOverridesWith(t *testing.T) {
	app := NewCommand("myapp").
		Arg(NewArg("json").Action(SetTrue).OverridesWith("text")).
		Arg(NewArg("text").Action(SetTrue).OverridesWith("json"))

	m, err := app.Parse([]string{"--json", "--text"})
	if err != nil {
		t.Fatal(err)
	}
	if m.Contains("json") {
		t.Error("json should have been overridden by text")
	}
	if !m.Contains("text") {
		t.Error("text should be present")
	}
}

func TestExplicitPositionalIndex(t *testing.T) {
	app := NewCommand("myapp").
		Arg(NewArg("dest").Index(2)).
		Arg(NewArg("source").Index(1))

	m, err := app.Parse([]string{"src.txt", "dst.txt"})
	if err != nil {
		t.Fatal(err)
	}
	src, _ := m.GetString("source")
	dst, _ := m.GetString("dest")
	if src != "src.txt" {
		t.Errorf("source = %q, want %q", src, "src.txt")
	}
	if dst != "dst.txt" {
		t.Errorf("dest = %q, want %q", dst, "dst.txt")
	}
}

func TestExplicitPositionalIndexStructTag(t *testing.T) {
	type CLI struct {
		Dest   string `glap:"dest,index=2"`
		Source string `glap:"source,index=1"`
	}
	var cli CLI
	_, err := Parse(&cli, []string{"src.txt", "dst.txt"})
	if err != nil {
		t.Fatal(err)
	}
	if cli.Source != "src.txt" {
		t.Errorf("Source = %q, want %q", cli.Source, "src.txt")
	}
	if cli.Dest != "dst.txt" {
		t.Errorf("Dest = %q, want %q", cli.Dest, "dst.txt")
	}
}
