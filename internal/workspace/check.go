package workspace

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
	"time"
)

// Status is the overall outcome of checking an exercise.
type Status int

const (
	Passed      Status = iota
	Failed             // tests or vet failed
	BuildFailed        // the code (or its tests) does not compile
)

// Test is one test function or subtest and its failure messages.
type Test struct {
	Name     string // display name, e.g. "TestLabel" or "production"
	Passed   bool
	Skipped  bool
	Output   []string
	Subtests []*Test

	pkg  string
	full string
}

// Result describes one check run.
type Result struct {
	Status   Status
	Tests    []*Test  // top-level tests in run order
	Build    string   // compiler or go command output when Status is BuildFailed
	Problems []string // failures not tied to a test: panics, timeouts, vet
	Notes    []string // non-blocking advice
	Duration time.Duration
}

// Counts returns how many top-level tests passed and failed.
func (r *Result) Counts() (passed, failed int) {
	for _, t := range r.Tests {
		switch {
		case t.Passed || t.Skipped:
			passed++
		default:
			failed++
		}
	}
	return passed, failed
}

// Check runs go test (and, if that passes, go vet) in dir. It returns an
// error only when the check could not run at all or ctx was cancelled.
func Check(ctx context.Context, dir string, race bool) (*Result, error) {
	start := time.Now()
	r, err := check(ctx, dir, race)
	if err != nil {
		return nil, err
	}
	if race && strings.Contains(r.Build, "requires cgo") {
		r, err = check(ctx, dir, false)
		if err != nil {
			return nil, err
		}
		r.Notes = append(r.Notes, "The race detector needs cgo and a C compiler, which aren't available, so this exercise ran without -race.")
	}
	r.Duration = time.Since(start)
	return r, nil
}

func check(ctx context.Context, dir string, race bool) (*Result, error) {
	args := []string{"test", "-json", "-count=1", "-timeout=60s"}
	if race {
		args = append(args, "-race")
	}
	stdout, stderr, err := run(ctx, dir, "go", append(args, "./...")...)
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if errors.As(err, new(*exec.Error)) {
		return nil, err
	}
	r := parse(stdout)
	if stderr = cleanStderr(stderr); stderr != "" && (r.Build != "" || len(r.Tests) == 0) {
		r.Build = strings.TrimSpace(r.Build + "\n" + stderr)
	}
	switch {
	case r.Build != "":
		r.Status = BuildFailed
	case err != nil || r.failed():
		r.Status = Failed
	case len(r.Tests) == 0:
		r.Status = Failed
		r.Problems = append(r.Problems, "No tests ran. Did a test file get deleted or renamed?")
	default:
		if out, errOut, err := run(ctx, dir, "go", "vet", "./..."); err != nil {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			problems, unreachable := splitVet(string(out) + errOut)
			for _, l := range unreachable {
				r.Notes = append(r.Notes, "Code that can never run: "+l+". Delete it (often a leftover placeholder return below your new code).")
			}
			if len(problems) > 0 {
				r.Status = Failed
				r.Problems = append(r.Problems, "go vet found a problem:\n"+strings.Join(problems, "\n"))
			}
		}
	}
	if out, _, err := run(ctx, dir, "gofmt", "-l", "."); err == nil && len(bytes.TrimSpace(out)) > 0 {
		files := strings.Fields(string(out))
		r.Notes = append(r.Notes, "Not gofmt-formatted: "+strings.Join(files, ", ")+". Run gofmt -w . or turn on format-on-save.")
	}
	return r, nil
}

// splitVet separates vet's "unreachable code" reports, a harmless leftover
// that gets a note instead of failing the exercise, from real problems.
// It returns the location of each unreachable block.
func splitVet(output string) (problems, unreachable []string) {
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		switch {
		case line == "" || strings.HasPrefix(line, "# "):
		case strings.HasSuffix(line, ": unreachable code"):
			unreachable = append(unreachable, strings.TrimSuffix(strings.TrimPrefix(line, "./"), ": unreachable code"))
		default:
			problems = append(problems, line)
		}
	}
	return problems, unreachable
}

func (r *Result) failed() bool {
	if len(r.Problems) > 0 {
		return true
	}
	_, failed := r.Counts()
	return failed > 0
}

func run(ctx context.Context, dir, name string, args ...string) ([]byte, string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOWORK=off")
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	return stdout.Bytes(), stderr.String(), err
}

// cleanStderr drops progress chatter such as "go: downloading ...".
func cleanStderr(s string) string {
	var keep []string
	for _, line := range strings.Split(s, "\n") {
		if strings.HasPrefix(line, "go: downloading ") || strings.TrimSpace(line) == "" {
			continue
		}
		keep = append(keep, line)
	}
	return strings.Join(keep, "\n")
}

type event struct {
	Action     string
	Package    string
	ImportPath string
	Test       string
	Output     string
}

var noise = regexp.MustCompile(`^\s*(=== (RUN|PAUSE|CONT|NAME)|--- (PASS|FAIL|SKIP):)`)

func parse(stdout []byte) *Result {
	r := &Result{}
	tests := map[string]*Test{} // package + "\x00" + full name
	packages := map[string]bool{}
	pkgOutput := map[string][]string{}
	pkgFailed := map[string]bool{}
	var build []string

	scanner := bufio.NewScanner(bytes.NewReader(stdout))
	scanner.Buffer(make([]byte, 1024*1024), 16*1024*1024)
	for scanner.Scan() {
		var ev event
		if json.Unmarshal(scanner.Bytes(), &ev) != nil {
			continue
		}
		switch ev.Action {
		case "build-output":
			build = append(build, strings.TrimRight(ev.Output, "\n"))
			continue
		case "build-fail":
			continue
		}
		if ev.Test == "" {
			switch ev.Action {
			case "output":
				line := strings.TrimRight(ev.Output, "\n")
				if keepPackageLine(line) {
					pkgOutput[ev.Package] = append(pkgOutput[ev.Package], line)
				}
			case "fail":
				pkgFailed[ev.Package] = true
			}
			continue
		}
		key := ev.Package + "\x00" + ev.Test
		t := tests[key]
		if t == nil {
			t = &Test{pkg: ev.Package, full: ev.Test, Name: ev.Test}
			tests[key] = t
			if parent := findParent(tests, ev.Package, ev.Test); parent != nil {
				parent.Subtests = append(parent.Subtests, t)
				t.Name = strings.TrimPrefix(ev.Test, parent.full+"/")
			} else {
				packages[ev.Package] = true
				r.Tests = append(r.Tests, t)
			}
		}
		switch ev.Action {
		case "output":
			line := strings.TrimRight(ev.Output, "\n")
			if !noise.MatchString(line) {
				// go test -json indents every t.Log/t.Error line by four
				// spaces whatever the nesting; continuation lines get more.
				t.Output = append(t.Output, strings.TrimPrefix(line, "    "))
			}
		case "pass":
			t.Passed = true
		case "skip":
			t.Skipped = true
		}
	}

	for _, t := range r.Tests {
		if len(packages) > 1 {
			t.Name = path.Base(t.pkg) + ": " + t.Name
		}
		hidePassingOutput(t)
	}
	r.Build = strings.TrimSpace(strings.Join(dropHeaders(build), "\n"))
	for pkg, failed := range pkgFailed {
		if failed && len(pkgOutput[pkg]) > 0 && !hasFailedTest(r.Tests) {
			r.Problems = append(r.Problems, strings.Join(limit(pkgOutput[pkg], 25), "\n"))
		}
	}
	return r
}

// findParent returns the deepest already-seen test whose name prefixes full.
func findParent(tests map[string]*Test, pkg, full string) *Test {
	for i := strings.LastIndex(full, "/"); i > 0; i = strings.LastIndex(full[:i], "/") {
		if t := tests[pkg+"\x00"+full[:i]]; t != nil {
			return t
		}
	}
	return nil
}

// hidePassingOutput drops log lines from passing tests; only failures matter.
// Tests that never reported pass (killed by a panic or timeout) stay failed.
func hidePassingOutput(t *Test) {
	for _, s := range t.Subtests {
		hidePassingOutput(s)
	}
	if t.Passed || t.Skipped {
		t.Output = nil
	}
}

func hasFailedTest(tests []*Test) bool {
	for _, t := range tests {
		if !t.Passed && !t.Skipped {
			return true
		}
	}
	return false
}

func keepPackageLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	switch {
	case trimmed == "", trimmed == "PASS", trimmed == "FAIL",
		strings.HasPrefix(trimmed, "ok "), strings.HasPrefix(trimmed, "ok\t"),
		strings.HasPrefix(trimmed, "FAIL\t"), strings.HasPrefix(trimmed, "?"),
		noise.MatchString(line):
		return false
	}
	return true
}

// dropHeaders removes the "# package" lines go prints above compiler errors.
func dropHeaders(lines []string) []string {
	var keep []string
	for _, l := range lines {
		if strings.HasPrefix(l, "# ") {
			continue
		}
		keep = append(keep, l)
	}
	return keep
}

func limit(lines []string, n int) []string {
	if len(lines) <= n {
		return lines
	}
	return append(lines[:n:n], "…")
}
