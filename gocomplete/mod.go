package main

import (
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"

	complete "github.com/posener/complete/v2"
	"golang.org/x/mod/modfile"
)

func findModFile(dir string) string {
	dir = filepath.Clean(dir)
	for {
		mod := filepath.Join(dir, "go.mod")
		if fi, err := os.Stat(mod); err == nil && fi.Mode().IsRegular() {
			return mod
		}
		parent := filepath.Dir(dir)
		if len(parent) >= len(dir) {
			break
		}
		dir = parent
	}
	return ""
}

func predictModule(prefix string, indirect bool) (modules []string, _ error) {
	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	name := findModFile(pwd)
	if name == "" {
		return nil, nil
	}
	data, err := os.ReadFile(name)
	if err != nil {
		return nil, err
	}
	file, err := modfile.ParseLax(name, data, nil)
	if err != nil {
		return nil, err
	}
	for _, r := range file.Require {
		if (prefix == "" || strings.HasPrefix(r.Mod.Path, prefix)) &&
			(indirect || !r.Indirect) {
			modules = append(modules, r.Mod.Path)
		}
	}
	slices.Sort(modules)
	return slices.Compact(modules), nil
}

func newModulePredictFunc(indirect bool) complete.PredictFunc {
	return func(prefix string) []string {
		modules, err := predictModule(prefix, indirect)
		if err != nil {
			log.Println("Error:", err)
		}
		return modules
	}
}
