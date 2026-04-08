package glap

import (
	"testing"
)

func TestAppWrapper(t *testing.T) {
	type CLI struct {
		Port int `glap:"port,default=8080"`
	}

	var cli CLI
	app := New(&cli).
		Name("myapp").
		Version("1.0.0").
		About("My cool app").
		Author("Adam Olsen")

	_, err := app.Parse([]string{})
	if err != nil {
		t.Fatal(err)
	}
	if cli.Port != 8080 {
		t.Errorf("Port = %d, want %d", cli.Port, 8080)
	}
}
