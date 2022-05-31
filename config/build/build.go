package main

import (
	"cx-im/config"
	"github.com/moeshin/go-errs"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	const name = "keys.go"
	dir, err := os.Getwd()
	errs.Print(err)
	base := filepath.Base(dir)
	if base == "build" {
		parent := filepath.Dir(base)
		if parent == "config" {
			err = os.Chdir("../..")
			errs.Print(err)
		}
	}
	err = os.Chdir("config")
	errs.Panic(err)
	file, err := os.Create(name)
	defer errs.Close(file)
	_, err = file.WriteString("package config\nconst (\n")
	for k := range config.KeyValues {
		_, err = file.WriteString(k)
		errs.Panic(err)
		_, err = file.WriteString("=\"")
		errs.Panic(err)
		_, err = file.WriteString(k)
		errs.Panic(err)
		_, err = file.WriteString("\"\n")
		errs.Panic(err)
	}
	_, err = file.WriteString(")")
	errs.Panic(exec.Command("gofmt", "-w", name).Run())
}
