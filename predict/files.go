package predict

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// Dirs returns a predictor that predict directory paths. If a non-empty pattern is given, the
// predicted paths will match that pattern.
func Dirs(pattern string) FilesPredictor {
	return FilesPredictor{pattern: pattern, includeFiles: false}
}

// Dirs returns a predictor that predict file or directory paths. If a non-empty pattern is given,
// the predicted paths will match that pattern.
func Files(pattern string) FilesPredictor {
	return FilesPredictor{pattern: pattern, includeFiles: true}
}

type FilesPredictor struct {
	pattern      string
	includeFiles bool
}

// Predict searches for files according to the given prefix.
// If the only predicted path is a single directory, the search will continue another recursive
// layer into that directory.
func (f FilesPredictor) Predict(prefix string) (options []string) {
	options = f.predictFiles(prefix)

	// If the number of prediction is not 1, we either have many results or have no results, so we
	// return it.
	if len(options) != 1 {
		return
	}

	// Only try deeper, if the one item is a directory.
	if stat, err := os.Stat(options[0]); err != nil || !stat.IsDir() {
		return
	}

	return dedupe(f.predictFiles(options[0]))
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

func (f FilesPredictor) predictFiles(prefix string) []string {
	if strings.HasSuffix(prefix, "/..") {
		return nil
	}

	dir := directory(prefix)
	files := f.listFiles(dir)

	// Add dir if match.
	files = append(files, dir)

	return FilesSet(files).Predict(prefix)
}

func (f FilesPredictor) listFiles(dir string) []string {
	var list []string
	des, _ := os.ReadDir(dir)
	for _, e := range des {
		typ := e.Type()
		name := e.Name()
		// Resolve the file type of the symlinks target
		if typ&os.ModeSymlink != 0 {
			fi, err := os.Stat(dir + string(os.PathSeparator) + name)
			if err != nil {
				continue
			}
			typ = fi.Mode()
		}
		if f.includeFiles && typ.IsRegular() {
			if ok, _ := filepath.Match(f.pattern, name); ok {
				list = append(list, filepath.Join(dir, name))
			}
		}
		if typ.IsDir() {
			list = append(list, filepath.Join(dir, name))
		}
	}
	return list
}

// directory gives the directory of the given partial path in case that it is not, we fall back to
// the current directory.
func directory(path string) string {
	if info, err := os.Stat(path); err == nil && info.IsDir() {
		return fixPathForm(path, path)
	}
	dir := filepath.Dir(path)
	if info, err := os.Stat(dir); err == nil && info.IsDir() {
		return fixPathForm(path, dir)
	}
	return "./"
}

// FilesSet predict according to file rules to a given fixed set of file names.
type FilesSet []string

func (s FilesSet) Predict(prefix string) []string {
	var matches []string
	// add all matching files to prediction
	for _, f := range s {
		f = fixPathForm(prefix, f)

		// test matching of file to the argument
		if matchFile(f, prefix) {
			matches = append(matches, f)
		}
	}
	if len(matches) == 0 {
		return s
	}
	return matches
}

// MatchFile returns true if prefix can match the file
func matchFile(file, prefix string) bool {
	// special case for current directory completion
	if file == "./" && (prefix == "." || prefix == "") {
		return true
	}
	if prefix == "." && strings.HasPrefix(file, ".") {
		return true
	}

	file = strings.TrimPrefix(file, "./")
	prefix = strings.TrimPrefix(prefix, "./")

	return strings.HasPrefix(file, prefix)
}

var _wdOnce struct {
	sync.Once
	wd  string
	err error
}

func getwd() (string, error) {
	_wdOnce.Do(func() {
		_wdOnce.wd, _wdOnce.err = os.Getwd()
	})
	return _wdOnce.wd, _wdOnce.err
}

// fixPathForm changes a file name to a relative name
func fixPathForm(last, file string) string {
	// Get wording directory for relative name.
	workDir, err := getwd()
	if err != nil {
		return file
	}

	abs, err := filepath.Abs(file)
	if err != nil {
		return file
	}

	// If last is absolute, return path as absolute.
	if filepath.IsAbs(last) {
		return fixDirPath(abs)
	}

	rel, err := filepath.Rel(workDir, abs)
	if err != nil {
		return file
	}

	// Fix ./ prefix of path.
	if rel != "." && strings.HasPrefix(last, ".") {
		rel = "./" + rel
	}

	return fixDirPath(rel)
}

func fixDirPath(path string) string {
	info, err := os.Stat(path)
	if err == nil && info.IsDir() && !strings.HasSuffix(path, "/") {
		path += "/"
	}
	return path
}
