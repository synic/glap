//go:build mage

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

var buildDir = "build"

func gowork() string {
	abs, _ := filepath.Abs(filepath.Join("examples", "workspace", "go.work"))
	return abs
}

func Build() error {
	mg.Deps(Clean)

	if err := os.MkdirAll(buildDir, 0o755); err != nil {
		return err
	}

	examples, err := filepath.Glob("examples/*/main.go")
	if err != nil {
		return err
	}

	env := map[string]string{"GOWORK": gowork()}

	for _, example := range examples {
		dir := filepath.Dir(example)
		name := filepath.Base(dir)
		if name == "workspace" {
			continue
		}
		out := filepath.Join(buildDir, name)
		fmt.Printf("Building %s -> %s\n", dir, out)
		if err := sh.RunWith(env, "go", "build", "-o", out, "./"+dir); err != nil {
			return err
		}
	}

	return nil
}

func Clean() error {
	return os.RemoveAll(buildDir)
}

func Test() error {
	return sh.RunV("go", "test", "-v", "./...")
}

func Vet() error {
	if err := sh.RunV("go", "vet", "./..."); err != nil {
		return err
	}

	env := map[string]string{"GOWORK": gowork()}
	examples, err := filepath.Glob("examples/*/main.go")
	if err != nil {
		return err
	}
	for _, example := range examples {
		dir := filepath.Dir(example)
		name := filepath.Base(dir)
		if name == "workspace" {
			continue
		}
		fmt.Printf("Vetting %s\n", dir)
		if err := sh.RunWith(env, "go", "vet", "./"+dir); err != nil {
			return fmt.Errorf("vet %s: %w", name, err)
		}
	}
	return nil
}

func Check() {
	mg.SerialDeps(Vet, Test)
}
