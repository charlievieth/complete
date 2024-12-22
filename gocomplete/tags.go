package main

import (
	"go/build"
	"strings"

	complete "github.com/posener/complete/v2"
)

func predictTags() complete.Predictor {
	return complete.PredictFunc(func(prefix string) []string {
		pkg, err := build.ImportDir(".", build.AllowBinary|build.ImportComment)
		if err != nil {
			return nil
		}
		if len(prefix) == 0 {
			return pkg.AllTags
		}
		a := pkg.AllTags[:0]
		for _, s := range pkg.AllTags {
			if strings.HasPrefix(s, prefix) {
				a = append(a, s)
			}
		}
		return a
	})
}
