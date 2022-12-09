package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"regexp"
)

var fset = token.NewFileSet() // global FileSet

func functionsInFile(path string, regexp *regexp.Regexp) []string {
	f, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		log.Printf("Failed parsing %s: %s", path, err)
		return nil
	}
	var names []string
	for _, d := range f.Decls {
		if f, ok := d.(*ast.FuncDecl); ok && f != nil && f.Name != nil {
			name := f.Name.String()
			if regexp == nil || regexp.MatchString(name) {
				names = append(names, name)
			}
		}
	}
	return names
}
