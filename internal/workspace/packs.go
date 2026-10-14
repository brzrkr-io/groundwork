package workspace

import (
	"io/fs"
	"maps"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
)

// Pack layout (see content/AUTHORING.md):
//
//	README.md            instructions, copied to the workspace
//	hints.md             hints, shown only inside groundwork
//	*.tmpl               starter files; the suffix keeps Go tooling away from them
//	go.mod.tmpl          optional; packs without one get a generated go.mod
//	solution/...         files that replace or add to the starter
const (
	hintsFile   = "hints.md"
	solutionDir = "solution"
	tmplSuffix  = ".tmpl"
)

// GoVersion is the go directive for generated exercise modules. The
// client-go and AWS SDK packs need at least this version too.
const GoVersion = "1.26"

// Starter returns the files of a fresh exercise folder, keyed by
// slash-separated path (".tmpl" stripped). Every exercise is its own Go
// module, so editors and tools only ever see one exercise at a time; packs
// without a go.mod get one for module.
func Starter(pack fs.FS, module string) (map[string][]byte, error) {
	files := map[string][]byte{}
	err := fs.WalkDir(pack, ".", func(p string, entry fs.DirEntry, err error) error {
		switch {
		case err != nil:
			return err
		case p == solutionDir:
			return fs.SkipDir
		case entry.IsDir() || p == hintsFile:
			return nil
		}
		data, err := fs.ReadFile(pack, p)
		files[strings.TrimSuffix(p, tmplSuffix)] = data
		return err
	})
	if _, ok := files["go.mod"]; !ok && err == nil {
		files["go.mod"] = []byte("module " + module + "\n\ngo " + GoVersion + "\n")
	}
	return files, err
}

// Solution returns the reference solution's files: the starter with the
// pack's solution/ overlay applied.
func Solution(pack fs.FS, module string) (map[string][]byte, error) {
	files, err := Starter(pack, module)
	if err != nil {
		return nil, err
	}
	err = fs.WalkDir(pack, solutionDir, func(p string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		data, err := fs.ReadFile(pack, p)
		files[strings.TrimSuffix(strings.TrimPrefix(p, solutionDir+"/"), tmplSuffix)] = data
		return err
	})
	return files, err
}

// SolutionChanges lists the files the solution adds or rewrites, which are
// the ones worth reading.
func SolutionChanges(pack fs.FS) ([]string, error) {
	var changed []string
	err := fs.WalkDir(pack, solutionDir, func(p string, entry fs.DirEntry, err error) error {
		if err == nil && !entry.IsDir() {
			changed = append(changed, strings.TrimSuffix(strings.TrimPrefix(p, solutionDir+"/"), tmplSuffix))
		}
		return err
	})
	slices.Sort(changed)
	return changed, err
}

// OwnModule reports whether the pack ships its own go.mod (multi-package
// exercises and exercises with third-party dependencies).
func OwnModule(pack fs.FS) bool {
	_, err := fs.Stat(pack, "go.mod"+tmplSuffix)
	return err == nil
}

// Write creates files under dir, replacing any that exist.
func Write(dir string, files map[string][]byte) error {
	for _, rel := range slices.Sorted(maps.Keys(files)) {
		target := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(target, files[rel], 0o644); err != nil {
			return err
		}
	}
	return nil
}

// Materialize writes a complete starter (or, with solution, solved) exercise
// into dir, as module "groundwork.local/<base of dir>".
func Materialize(pack fs.FS, dir string, solution bool) error {
	module := "groundwork.local/" + path.Base(filepath.ToSlash(dir))
	assemble := Starter
	if solution {
		assemble = Solution
	}
	files, err := assemble(pack, module)
	if err != nil {
		return err
	}
	return Write(dir, files)
}
