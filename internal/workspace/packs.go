package workspace

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Pack layout (see content/AUTHORING.md):
//
//	README.md            instructions, copied to the workspace
//	hints.md             hints, shown only inside gopherlings
//	*.tmpl               starter files; the suffix keeps Go tooling away from them
//	solution/...         files that replace or add to the starter
const (
	hintsFile   = "hints.md"
	solutionDir = "solution"
	tmplSuffix  = ".tmpl"
)

// Materialize writes the starter files of pack into dst, overwriting files
// with the same name. With solution set, the solution overlay is applied on
// top so dst holds a complete, passing program.
func Materialize(pack fs.FS, dst string, solution bool) error {
	if err := copyTree(pack, ".", dst, func(p string) bool {
		return p == hintsFile || p == solutionDir || strings.HasPrefix(p, solutionDir+"/")
	}); err != nil {
		return err
	}
	if !solution {
		return nil
	}
	if _, err := fs.Stat(pack, solutionDir); err != nil {
		return fmt.Errorf("pack has no solution: %w", err)
	}
	return copyTree(pack, solutionDir, dst, nil)
}

// OwnModule reports whether the pack ships its own go.mod (multi-package
// exercises and exercises with third-party dependencies).
func OwnModule(pack fs.FS) bool {
	_, err := fs.Stat(pack, "go.mod"+tmplSuffix)
	return err == nil
}

func copyTree(src fs.FS, root, dst string, skip func(string) bool) error {
	return fs.WalkDir(src, root, func(p string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel := strings.TrimPrefix(strings.TrimPrefix(p, root), "/")
		if root == "." {
			rel = p
		}
		if rel == "." || rel == "" {
			return os.MkdirAll(dst, 0o755)
		}
		if skip != nil && skip(rel) {
			if entry.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		target := filepath.Join(dst, filepath.FromSlash(strings.TrimSuffix(rel, tmplSuffix)))
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := fs.ReadFile(src, p)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
}

// Starter returns the files a fresh exercise folder holds, keyed by
// slash-separated path relative to the folder (".tmpl" already stripped).
func Starter(pack fs.FS) (map[string][]byte, error) {
	files := map[string][]byte{}
	err := fs.WalkDir(pack, ".", func(p string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if p == solutionDir {
			return fs.SkipDir
		}
		if entry.IsDir() || p == hintsFile {
			return nil
		}
		data, err := fs.ReadFile(pack, p)
		if err != nil {
			return err
		}
		files[strings.TrimSuffix(p, tmplSuffix)] = data
		return nil
	})
	return files, err
}

// Exists reports whether path exists.
func Exists(p string) bool {
	_, err := os.Stat(p)
	return !errors.Is(err, fs.ErrNotExist)
}

// GoVersion is the go directive for learner modules. client-go and the AWS
// SDK packs need at least this version too.
const GoVersion = "1.26"

// WriteGoMod creates dir/go.mod for a module holding stdlib-only exercises.
func WriteGoMod(dir, module string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	contents := "module " + module + "\n\ngo " + GoVersion + "\n"
	return os.WriteFile(filepath.Join(dir, "go.mod"), []byte(contents), 0o644)
}
