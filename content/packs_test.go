package content_test

import (
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pjanderson/gopherlings/content"
	"github.com/pjanderson/gopherlings/internal/course"
	"github.com/pjanderson/gopherlings/internal/workspace"
)

// TestPacks proves every exercise is solvable and starts unsolved: the starter
// must fail its tests, and starter+solution must pass tests, vet and gofmt.
//
// Check one pack while authoring:
//
//	go test ./content -run 'TestPacks/^variables$' -v
//
// Packs with their own third-party dependencies are skipped under -short.
func TestPacks(t *testing.T) {
	c, err := course.Load(content.Files)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not on PATH")
	}
	for _, e := range c.Exercises {
		t.Run(e.Name, func(t *testing.T) {
			t.Parallel()
			pack, err := c.Pack(e)
			if err != nil {
				t.Fatal(err)
			}
			readme, err := c.README(e)
			if err != nil {
				t.Fatalf("missing README.md: %v", err)
			}
			if !strings.Contains(readme, "\n## Your task") {
				t.Error(`README.md has no "## Your task" section`)
			}
			hints, err := c.Hints(e)
			if err != nil {
				t.Fatalf("missing hints.md: %v", err)
			}
			if len(hints) < 2 {
				t.Errorf("hints.md has %d hints, want at least 2", len(hints))
			}
			for i, h := range hints {
				if h.Body == "" {
					t.Errorf("hint %d (%q) is empty", i+1, h.Title)
				}
			}
			starter := materialize(t, e, pack, false)
			solved := materialize(t, e, pack, true)
			for _, dir := range []string{starter, solved} {
				for _, p := range packagesWithoutDoc(t, dir) {
					t.Errorf("package %s has no package comment; editors flag this on every file", p)
				}
			}
			if workspace.OwnModule(pack) && testing.Short() {
				t.Skip("pack has its own module; run without -short")
			}

			if out, err := goTest(starter, e.Race); err == nil {
				t.Errorf("starter already passes; there is nothing to fix:\n%s", out)
			}
			if out := gofmt(t, starter); out != "" {
				t.Errorf("starter is not gofmt-clean:\n%s", out)
			}

			if out, err := goTest(solved, e.Race); err != nil {
				t.Fatalf("solution fails its tests:\n%s", out)
			}
			if out, err := run(solved, "go", "vet", "./..."); err != nil {
				t.Errorf("go vet fails on the solution:\n%s", out)
			}
			if out := gofmt(t, solved); out != "" {
				t.Errorf("solution is not gofmt-clean:\n%s", out)
			}
		})
	}
}

// packagesWithoutDoc lists the non-test packages under dir (relative paths)
// in which no file has a package doc comment.
func packagesWithoutDoc(t *testing.T, dir string) []string {
	t.Helper()
	documented := map[string]bool{}
	err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() && entry.Name() == "testdata" {
			return fs.SkipDir
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		f, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.PackageClauseOnly|parser.ParseComments)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(dir, filepath.Dir(path))
		documented[rel] = documented[rel] || f.Doc != nil
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	var missing []string
	for pkg, ok := range documented {
		if !ok {
			missing = append(missing, pkg)
		}
	}
	return missing
}

func materialize(t *testing.T, e *course.Exercise, pack fs.FS, solution bool) string {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, filepath.FromSlash(e.Dir()))
	if !workspace.OwnModule(pack) {
		if err := workspace.WriteGoMod(root, "gopherlings.local/exercises"); err != nil {
			t.Fatal(err)
		}
	}
	if err := workspace.Materialize(pack, dir, solution); err != nil {
		t.Fatal(err)
	}
	return dir
}

func goTest(dir string, race bool) (string, error) {
	args := []string{"test", "-count=1", "-timeout=120s"}
	if race {
		args = append(args, "-race")
	}
	return run(dir, "go", append(args, "./...")...)
}

// gofmt returns the unformatted files (or the parse errors) under dir.
func gofmt(t *testing.T, dir string) string {
	t.Helper()
	out, _ := run(dir, "gofmt", "-l", ".")
	return out
}

func run(dir, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOWORK=off")
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}
