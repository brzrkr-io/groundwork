package workspace

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func writeModule(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/check\n\ngo "+GoVersion+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestCheckReportsFailingSubtestsUnderTheirTest(t *testing.T) {
	dir := writeModule(t, map[string]string{
		"x.go": "package x\n\nfunc Double(n int) int { return n + n }\n\nfunc Half(n int) int { return n }\n",
		"x_test.go": `package x

import "testing"

func TestDouble(t *testing.T) {
	if Double(2) != 4 {
		t.Error("bad")
	}
}

func TestHalf(t *testing.T) {
	for _, n := range []int{0, 4} {
		t.Run(map[int]string{0: "zero", 4: "four"}[n], func(t *testing.T) {
			if got := Half(n); got != n/2 {
				t.Errorf("Half(%d) = %d, want %d", n, got, n/2)
			}
		})
	}
}
`,
	})
	r, err := Check(context.Background(), dir, false)
	if err != nil {
		t.Fatal(err)
	}
	if r.Status != Failed {
		t.Fatalf("Status = %v, want Failed; build output %q", r.Status, r.Build)
	}
	if passed, failed := r.Counts(); passed != 1 || failed != 1 {
		t.Fatalf("Counts() = %d passed, %d failed, want 1 and 1", passed, failed)
	}
	half := r.Tests[1]
	if half.Name != "TestHalf" || len(half.Subtests) != 2 {
		t.Fatalf("second test = %q with %d subtests, want TestHalf with 2", half.Name, len(half.Subtests))
	}
	zero, four := half.Subtests[0], half.Subtests[1]
	if !zero.Passed || zero.Name != "zero" || four.Passed || four.Name != "four" {
		t.Fatalf("subtests = %+v, %+v; want zero passing and four failing", zero, four)
	}
	if !slices.ContainsFunc(four.Output, func(l string) bool { return strings.Contains(l, "Half(4) = 4, want 2") }) {
		t.Errorf("failing subtest output %q lacks the assertion message", four.Output)
	}
	if slices.ContainsFunc(four.Output, func(l string) bool { return strings.Contains(l, "--- FAIL") }) {
		t.Errorf("failing subtest output %q still contains go test framing", four.Output)
	}
}

func TestCheckReportsCompileErrors(t *testing.T) {
	dir := writeModule(t, map[string]string{
		"x.go":      "package x\n\nfunc Broken() int { return \"nope\" }\n",
		"x_test.go": "package x\n\nimport \"testing\"\n\nfunc TestBroken(t *testing.T) { Broken() }\n",
	})
	r, err := Check(context.Background(), dir, false)
	if err != nil {
		t.Fatal(err)
	}
	if r.Status != BuildFailed || !strings.Contains(r.Build, "x.go:3") {
		t.Fatalf("Status = %v, Build = %q; want BuildFailed pointing at x.go:3", r.Status, r.Build)
	}
}

func TestCheckRunsVetAfterTestsPass(t *testing.T) {
	dir := writeModule(t, map[string]string{
		"x.go":      "package x\n\nimport \"fmt\"\n\nfunc Label(n int) string { return fmt.Sprintf(\"%d\", n) }\n\nfunc Wrong() string { return fmt.Sprintf(\"%d\", \"text\") }\n",
		"x_test.go": "package x\n\nimport \"testing\"\n\nfunc TestLabel(t *testing.T) {\n\tif Label(1) != \"1\" {\n\t\tt.Fail()\n\t}\n}\n",
	})
	r, err := Check(context.Background(), dir, false)
	if err != nil {
		t.Fatal(err)
	}
	// go test runs a subset of vet itself, so the printf mistake can surface
	// as a build failure or as a vet problem; it must never pass.
	if r.Status == Passed {
		t.Fatal("Status = Passed, want the printf mistake to fail the check")
	}
}

func TestCheckPasses(t *testing.T) {
	dir := writeModule(t, map[string]string{
		"x.go":      "package x\n\nfunc One() int { return 1 }\n",
		"x_test.go": "package x\n\nimport \"testing\"\n\nfunc TestOne(t *testing.T) {\n\tif One() != 1 {\n\t\tt.Fail()\n\t}\n}\n",
	})
	r, err := Check(context.Background(), dir, false)
	if err != nil {
		t.Fatal(err)
	}
	if r.Status != Passed || len(r.Tests) != 1 || !r.Tests[0].Passed {
		t.Fatalf("Status = %v, Tests = %+v; want one passing test", r.Status, r.Tests)
	}
}
