package content_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/brzrkr-io/groundwork/content"
	"github.com/brzrkr-io/groundwork/internal/course"
	"github.com/brzrkr-io/groundwork/internal/workspace"
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
			for _, problem := range readmeStructure(readme) {
				t.Error(problem)
			}
			if _, err := c.Tasks(e); err != nil {
				t.Errorf("hints.md doesn't match the README's tasks:\n%v", err)
			}
			starter := materialize(t, e, pack, false)
			solved := materialize(t, e, pack, true)
			for _, dir := range []string{starter, solved} {
				for _, p := range packagesWithoutDoc(t, dir) {
					t.Errorf("package %s has no package comment; editors flag this on every file", p)
				}
			}
			for _, name := range missingDocComments(t, e.Name, starter, solved) {
				t.Errorf("the solution adds exported %s, but the starter has no ready-made \"// %s ...\" comment where it goes; "+
					"learners type under the TODO and their editor flags the TODO as a wrong doc comment", name, name)
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
	dir := filepath.Join(t.TempDir(), filepath.FromSlash(e.Dir()))
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

// readmeStructure checks the scannable layout from AUTHORING.md: a
// "**New here:**" line, and every task with a "**Goal:**" line and a
// "**Think it through:**" list.
func readmeStructure(readme string) []string {
	var problems []string
	if !strings.Contains(readme, "\n**New here:**") {
		problems = append(problems, `README.md has no "**New here:**" line under the title`)
	}
	tasks := regexp.MustCompile(`(?m)^### (\d+)\. `).FindAllStringSubmatchIndex(readme, -1)
	for i, m := range tasks {
		end := len(readme)
		if i+1 < len(tasks) {
			end = tasks[i+1][0]
		} else if j := strings.Index(readme[m[0]:], "\n## "); j >= 0 {
			end = m[0] + j
		}
		section, n := readme[m[0]:end], readme[m[2]:m[3]]
		if !strings.Contains(section, "\n**Goal:**") {
			problems = append(problems, "README task "+n+` has no "**Goal:**" line`)
		}
		if !strings.Contains(section, "\n**Think it through:**") {
			problems = append(problems, "README task "+n+` has no "**Think it through:**" list`)
		}
	}
	return problems
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

// renamedByLearner lists declarations the learner creates by renaming an
// existing one. Their doc comment is already there, under the old name, and
// the task text says to rename it too.
var renamedByLearner = map[string][]string{
	"exported_names":    {"NewCredential", "Credential.Masked"},
	"exported_names_2":  {"New"},
	"interfaces_2":      {"HTTPCheck.Check"},
	"internal_packages": {"Make"},
}

// missingDocComments lists exported declarations the learner must add (in
// the solution, not in the starter) whose doc comment the starter doesn't
// already provide. Go expects a declaration's comment to start with its
// name. Without one ready, the TODO above becomes the doc comment and the
// editor warns about it.
func missingDocComments(t *testing.T, pack, starter, solved string) []string {
	t.Helper()
	starterNames, ready := exportedDecls(t, starter)
	solvedNames, _ := exportedDecls(t, solved)
	var missing []string
	for name, ownDoc := range solvedNames {
		short := name[strings.LastIndex(name, ".")+1:]
		if ownDoc && !starterNames[name] && !ready[short] && !slices.Contains(renamedByLearner[pack], name) {
			missing = append(missing, name)
		}
	}
	slices.Sort(missing)
	return missing
}

// exportedDecls returns the exported top-level names and methods
// ("Type.Method") in the non-test Go files under dir (true when it needs its
// own doc comment, false inside a parenthesised block), and the first words of
// top-level comments not attached to any declaration: ready-made doc
// comments waiting for the learner's code.
func exportedDecls(t *testing.T, dir string) (names, ready map[string]bool) {
	t.Helper()
	names, ready = map[string]bool{}, map[string]bool{}
	err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return err
		}
		src, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		f, err := parser.ParseFile(token.NewFileSet(), path, src, parser.SkipObjectResolution|parser.ParseComments)
		if err != nil {
			return nil // the starter may not compile on purpose
		}
		attached := map[*ast.CommentGroup]bool{f.Doc: true}
		for _, d := range f.Decls {
			switch d := d.(type) {
			case *ast.FuncDecl:
				attached[d.Doc] = true
				if !d.Name.IsExported() {
					continue
				}
				name := d.Name.Name
				if d.Recv != nil && len(d.Recv.List) > 0 {
					name = receiverName(d.Recv.List[0].Type) + "." + name
				}
				names[name] = true
			case *ast.GenDecl:
				grouped := d.Lparen.IsValid()
				attached[d.Doc] = true
				for _, s := range d.Specs {
					switch s := s.(type) {
					case *ast.TypeSpec:
						attached[s.Doc] = true
						if s.Name.IsExported() {
							names[s.Name.Name] = !grouped
						}
					case *ast.ValueSpec:
						attached[s.Doc] = true
						for _, n := range s.Names {
							if n.IsExported() {
								names[n.Name] = !grouped
							}
						}
					}
				}
			}
		}
		for _, g := range f.Comments {
			if !attached[g] {
				if words := strings.Fields(g.Text()); len(words) > 0 {
					ready[words[0]] = true
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return names, ready
}

func receiverName(e ast.Expr) string {
	switch e := e.(type) {
	case *ast.StarExpr:
		return receiverName(e.X)
	case *ast.IndexExpr:
		return receiverName(e.X)
	case *ast.IndexListExpr:
		return receiverName(e.X)
	case *ast.Ident:
		return e.Name
	}
	return ""
}
