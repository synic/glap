package glap

import "testing"

func TestGetInt64(t *testing.T) {
	app := NewCommand("myapp").
		Arg(NewArg("id"))

	m, err := app.Parse([]string{"--id", "9223372036854775807"})
	if err != nil {
		t.Fatal(err)
	}
	v, ok := m.GetInt64("id")
	if !ok || v != 9223372036854775807 {
		t.Errorf("id = %d, want max int64", v)
	}
}

func TestGetUint(t *testing.T) {
	app := NewCommand("myapp").
		Arg(NewArg("count"))

	m, err := app.Parse([]string{"--count", "42"})
	if err != nil {
		t.Fatal(err)
	}
	v, ok := m.GetUint("count")
	if !ok || v != 42 {
		t.Errorf("count = %d, want 42", v)
	}
}

func TestGetUint64(t *testing.T) {
	app := NewCommand("myapp").
		Arg(NewArg("big"))

	m, err := app.Parse([]string{"--big", "18446744073709551615"})
	if err != nil {
		t.Fatal(err)
	}
	v, ok := m.GetUint64("big")
	if !ok || v != 18446744073709551615 {
		t.Errorf("big = %d, want max uint64", v)
	}
}

func TestGetInt64Missing(t *testing.T) {
	app := NewCommand("myapp").
		Arg(NewArg("id"))

	m, err := app.Parse([]string{})
	if err != nil {
		t.Fatal(err)
	}
	_, ok := m.GetInt64("id")
	if ok {
		t.Error("should return false for missing arg")
	}
}

func TestScan(t *testing.T) {
	type Config struct {
		Host    string  `glap:"host"`
		Port    int     `glap:"port"`
		Verbose bool    `glap:"verbose"`
		Rate    float64 `glap:"rate"`
	}

	app := NewCommand("myapp").
		Arg(NewArg("host").Default("localhost")).
		Arg(NewArg("port").Default("3000")).
		Arg(NewArg("verbose").Action(SetTrue)).
		Arg(NewArg("rate").Default("1.5"))

	m, err := app.Parse([]string{"--verbose", "--host", "0.0.0.0"})
	if err != nil {
		t.Fatal(err)
	}

	var cfg Config
	if err := m.Scan(&cfg); err != nil {
		t.Fatal(err)
	}

	if cfg.Host != "0.0.0.0" {
		t.Errorf("Host = %q, want %q", cfg.Host, "0.0.0.0")
	}
	if cfg.Port != 3000 {
		t.Errorf("Port = %d, want 3000", cfg.Port)
	}
	if !cfg.Verbose {
		t.Error("Verbose should be true")
	}
	if cfg.Rate != 1.5 {
		t.Errorf("Rate = %f, want 1.5", cfg.Rate)
	}
}

func TestScanTypedSlices(t *testing.T) {
	type Config struct {
		Ports []int     `glap:"port"`
		Rates []float64 `glap:"rate"`
	}

	app := NewCommand("myapp").
		Arg(NewArg("port").Action(Append)).
		Arg(NewArg("rate").Action(Append))

	m, err := app.Parse([]string{"--port", "80", "--port", "443", "--rate", "1.5", "--rate", "2.5"})
	if err != nil {
		t.Fatal(err)
	}

	var cfg Config
	if err := m.Scan(&cfg); err != nil {
		t.Fatal(err)
	}

	if len(cfg.Ports) != 2 || cfg.Ports[0] != 80 || cfg.Ports[1] != 443 {
		t.Fatalf("Ports = %v, want [80 443]", cfg.Ports)
	}
	if len(cfg.Rates) != 2 || cfg.Rates[0] != 1.5 || cfg.Rates[1] != 2.5 {
		t.Fatalf("Rates = %v, want [1.5 2.5]", cfg.Rates)
	}
}

func TestScanNonPointer(t *testing.T) {
	type Config struct{}
	var cfg Config

	m := newMatches()
	err := m.Scan(cfg)
	if err == nil {
		t.Error("Scan should fail on non-pointer")
	}
}
