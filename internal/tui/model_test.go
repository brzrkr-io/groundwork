package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/pjanderson/gopherlings/internal/course"
	"github.com/pjanderson/gopherlings/internal/session"
	"github.com/pjanderson/gopherlings/internal/workspace"
)

func newModel(t *testing.T) (*Model, *session.Session) {
	t.Helper()
	files := fstest.MapFS{"course.json": {Data: []byte(`{"modules":[{"title":"m","dir":"one","exercises":[{"name":"a","title":"A"},{"name":"b","title":"B"}]}]}`)}}
	for _, name := range []string{"a", "b"} {
		files["exercises/"+name+"/README.md"] = &fstest.MapFile{Data: []byte("# " + name + "\n\n## Your task\n")}
		files["exercises/"+name+"/hints.md"] = &fstest.MapFile{Data: []byte("## First\none\n## Second\ntwo\n")}
		files["exercises/"+name+"/exercise.go.tmpl"] = &fstest.MapFile{Data: []byte("package x\n\n// TODO(1): fix\n")}
		files["exercises/"+name+"/solution/exercise.go.tmpl"] = &fstest.MapFile{Data: []byte("package x\n")}
	}
	c, err := course.Load(files)
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	if err := session.Init(root); err != nil {
		t.Fatal(err)
	}
	s, err := session.Open(root, c)
	if err != nil {
		t.Fatal(err)
	}
	m := New(s)
	m.Init() // opens the first exercise; the returned commands (check, watch) are not run
	t.Cleanup(m.Close)
	return m, s
}

func press(m *Model, key string) {
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
}

func TestPassingCheckCompletesExerciseAndNextAdvances(t *testing.T) {
	m, s := newModel(t)
	a := s.Course.Exercises[0]
	if m.ex != a || m.pane != paneResult {
		t.Fatalf("opened %v in pane %v, want a in the result pane", m.ex.Name, m.pane)
	}

	m.Update(checkedMsg{seq: m.checkSeq, result: &workspace.Result{Status: workspace.Passed, Tests: []*workspace.Test{{Name: "TestA", Passed: true}}}})
	if !s.Done(a) {
		t.Fatal("a passing check did not mark the exercise done")
	}
	if _, err := os.Stat(filepath.Join(s.SolutionDir(a), "exercise.go")); err != nil {
		t.Fatalf("solution not written: %v", err)
	}
	if view := m.View(); !strings.Contains(view, "Exercise complete") {
		t.Errorf("view after passing lacks the completion message:\n%s", view)
	}

	press(m, "n")
	if m.ex != s.Course.Exercises[1] {
		t.Fatalf("n opened %v, want b", m.ex.Name)
	}
}

func TestStaleCheckResultsAreIgnored(t *testing.T) {
	m, s := newModel(t)
	stale := m.checkSeq
	m.runCheck() // a save started a newer check
	m.Update(checkedMsg{seq: stale, result: &workspace.Result{Status: workspace.Passed}})
	if s.Done(s.Course.Exercises[0]) || !m.checking {
		t.Fatal("an outdated check result was applied")
	}
}

func TestHintsRevealOneAtATime(t *testing.T) {
	m, _ := newModel(t)
	press(m, "h")
	if m.pane != paneHints || m.hintsShown != 1 {
		t.Fatalf("after h: pane %v, %d hints shown; want hints pane with 1", m.pane, m.hintsShown)
	}
	press(m, "h")
	press(m, "h")
	if m.hintsShown != 2 {
		t.Fatalf("hints shown = %d, want it capped at the 2 available", m.hintsShown)
	}
}

func TestSolutionUnlocksOnlyAfterEveryHintAndConfirmation(t *testing.T) {
	m, s := newModel(t)
	a := s.Course.Exercises[0]
	press(m, "s")
	if m.pane == paneSolution || s.Revealed(a) {
		t.Fatal("s showed the solution before any hint was read")
	}
	press(m, "h")
	press(m, "h") // both hints seen
	press(m, "s")
	press(m, "i") // anything but s cancels
	if s.Revealed(a) {
		t.Fatal("the solution was revealed without confirming")
	}
	press(m, "s")
	press(m, "s")
	if m.pane != paneSolution || !s.Revealed(a) || s.Done(a) {
		t.Fatalf("after s s: pane %v, revealed %v, done %v; want the solution shown, revealed, not done", m.pane, s.Revealed(a), s.Done(a))
	}
	if _, err := os.Stat(filepath.Join(s.SolutionDir(a), "exercise.go")); err != nil {
		t.Fatalf("revealed solution not written: %v", err)
	}
}

func TestResetNeedsConfirmation(t *testing.T) {
	m, s := newModel(t)
	src := filepath.Join(s.Dir(m.ex), "exercise.go")
	if err := os.WriteFile(src, []byte("package x\n// edited\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	press(m, "x")
	press(m, "i") // anything but x cancels
	if got, _ := os.ReadFile(src); !strings.Contains(string(got), "edited") {
		t.Fatal("reset ran without a second x")
	}
	press(m, "x")
	press(m, "x")
	if got, _ := os.ReadFile(src); strings.Contains(string(got), "edited") {
		t.Fatal("x x did not reset the exercise")
	}
}
