package main

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/posener/complete/v2"
)

var (
	predictBenchmark = funcPredict(regexp.MustCompile("^Benchmark"))
	predictTest      = funcPredict(regexp.MustCompile("^(Test|Example|Fuzz)"))
)

// predictTest predict test names.
// it searches in the current directory for all the go test files
// and then all the relevant function names.
// for test names use prefix of 'Test' or 'Example', and for benchmark
// test names use 'Benchmark'
func funcPredict(funcRegexp *regexp.Regexp) complete.Predictor {
	return complete.PredictFunc(func(prefix string) []string {
		return funcNames(".", funcRegexp)
	})
}

func dedupe(a []string) []string {
	if len(a) <= 1 {
		return a
	}
	sort.Strings(a)
	k := 1
	for i := 1; i < len(a); i++ {
		if a[k-1] != a[i] {
			a[k] = a[i]
			k++
		}
	}
	return a[:k]
}

// get all test names in current directory
func funcNames(dir string, re *regexp.Regexp) []string {

	des, _ := os.ReadDir(dir)
	k := 0
	for i := 0; i < len(des); i++ {
		d := des[i]
		if !d.IsDir() && strings.HasSuffix(d.Name(), "_test.go") {
			des[k] = d
			k++
		}
	}
	des = des[:k]

	switch len(des) {
	case 0:
		return nil
	case 1:
		tests := functionsInFile(filepath.Join(dir, des[0].Name()), re)
		return dedupe(tests)
	default:
		var wg sync.WaitGroup
		var mu sync.Mutex
		all := make([][]string, 0, len(des))
		ch := make(chan string, 4)
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for path := range ch {
					if names := functionsInFile(path, re); len(names) > 0 {
						mu.Lock()
						all = append(all, names)
						mu.Unlock()
					}
				}
			}()
		}
		for _, d := range des {
			ch <- filepath.Join(dir, d.Name())
		}
		close(ch)
		wg.Wait()

		n := 0
		for _, a := range all {
			n += len(a)
		}
		tests := make([]string, 0, n)
		for _, a := range all {
			tests = append(tests, a...)
		}
		return dedupe(tests)
	}
}
