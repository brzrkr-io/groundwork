package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/pjanderson/groundwork/internal/course"
	"github.com/pjanderson/groundwork/internal/session"
	"github.com/pjanderson/groundwork/internal/workspace"
)

func newModel(t *testing.T) (*Model, *session.Session) {
	t.Helper()
	files := fstest.MapFS{"course.json": {Data: []byte(`{"modules":[{"title":"m","dir":"one","exercises":[{"name":"a","title":"A"},{"name":"b","title":"B"}]}]}`)}}
	for _, name := range []string{"a", "b"} {
		files["exercises/"+name+"/README.md"] = &fstest.MapFile{Data: []byte("# " + name + "\n\n## Your task\n\n### 1. First\n\n### 2. Second\n")}
		files["exercises/"+name+"/hints.md"] = &fstest.MapFile{Data: []byte(
			"## Task 1: First\nnudge one\n### Closer\ncloser one\n### Answer\nanswer one\n" +
				"## Task 2: Second\nnudge two\n### Closer\ncloser two\n### Answer\nanswer two\n")}
		files["exercises/"+name+"/exercise.go.tmpl"] = &fstest.MapFile{Data: []byte("package x\n\n// TODO(1): fix\n")}
		files["exercises/"+name+"/solution/exercise.go.tmpl"] = &fstest.MapFile{Data: []byte("package x\n")}
	}
	return modelFor(t, files)
}

// modelFor opens a model on a fresh course directory for the course in files.
func modelFor(t *testing.T, files fstest.MapFS) (*Model, *session.Session) {
	t.Helper()
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

func TestHintsOpenOnTheFailingTaskAndGetCloserOneLevelAtATime(t *testing.T) {
	m, _ := newModel(t)
	// Task 1's test passes and task 2's fails, so h starts at task 2.
	m.Update(checkedMsg{seq: m.checkSeq, result: &workspace.Result{Status: workspace.Failed, Tests: []*workspace.Test{
		{Name: "TestFirst", Passed: true}, {Name: "TestSecond"},
	}}})
	press(m, "h")
	if m.pane != paneHints || m.hintTask != 1 || m.hintLevels[1] != 1 {
		t.Fatalf("after h: pane %v, task %d, level %d; want hints for task 2 at level 1", m.pane, m.hintTask+1, m.hintLevels[1])
	}
	press(m, "h")
	press(m, "h")
	press(m, "h")
	if m.hintLevels[1] != 3 {
		t.Fatalf("task 2 level = %d, want it capped at Nudge, Closer, Answer (3)", m.hintLevels[1])
	}
	press(m, "1")
	if m.hintTask != 0 || m.hintLevels[0] != 1 {
		t.Fatalf("after 1: task %d level %d, want task 1 at its first level", m.hintTask+1, m.hintLevels[0])
	}
}

func TestSolutionUnlocksOnlyAfterAnAnswerHintAndConfirmation(t *testing.T) {
	m, s := newModel(t)
	a := s.Course.Exercises[0]
	press(m, "s")
	if m.pane == paneSolution || s.Revealed(a) {
		t.Fatal("s showed the solution before any hint was read")
	}
	press(m, "h")
	press(m, "h")
	press(m, "s")
	if s.Revealed(a) || m.confirm != "" {
		t.Fatal("s offered the solution before the task's Answer hint")
	}
	press(m, "h") // the Answer
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

func TestGuidesOpenOnTheNextUnreadAndEscGoesBackOneLevel(t *testing.T) {
	files := fstest.MapFS{
		"course.json": {Data: []byte(`{"modules":[
			{"title":"One","dir":"one","guides":["state","debugging"],"exercises":[{"name":"a","title":"A"}]},
			{"title":"Two","dir":"two","guides":["concurrency"],"exercises":[{"name":"b","title":"B"}]}]}`)},
		"guides/state.md":       {Data: []byte("# State\n\n**In one sentence:** state is what changes.\n")},
		"guides/debugging.md":   {Data: []byte("# Debugging\n\n**In one sentence:** a method, not luck.\n")},
		"guides/concurrency.md": {Data: []byte("# Concurrency\n\n**In one sentence:** many tasks making progress.\n\nGoroutines are cheap.\n")},
	}
	for _, name := range []string{"a", "b"} {
		files["exercises/"+name+"/README.md"] = &fstest.MapFile{Data: []byte("# " + name + "\n\n## Your task\n")}
		files["exercises/"+name+"/exercise.go.tmpl"] = &fstest.MapFile{Data: []byte("package x\n")}
		files["exercises/"+name+"/solution/exercise.go.tmpl"] = &fstest.MapFile{Data: []byte("package x\n")}
	}
	m, s := modelFor(t, files)
	m.Update(tea.KeyMsg{Type: tea.KeyRight})
	if m.ex != s.Course.Exercises[1] {
		t.Fatalf("on %s, want b (module Two)", m.ex.Name)
	}

	press(m, "i")
	if view := m.View(); !strings.Contains(view, "Background for this module") || !strings.Contains(view, "Concurrency") {
		t.Errorf("instructions don't point to the module's guide:\n%s", view)
	}

	press(m, "g")
	if m.pane != paneGuides || s.Course.Guides[m.guideIdx].Name != "state" {
		t.Fatalf("g: pane %v, cursor on %s; want the guide list on the first unread guide", m.pane, s.Course.Guides[m.guideIdx].Name)
	}
	press(m, "j")
	press(m, "j")
	if view := m.View(); !strings.Contains(view, "many tasks making progress") || strings.Contains(view, "a method, not luck") {
		t.Errorf("only the guide under the cursor shows its summary:\n%s", view)
	}
	press(m, "k")
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.pane != paneGuide || m.guide.Name != "debugging" {
		t.Fatalf("k, enter: pane %v, guide %v; want the Debugging guide open", m.pane, m.guide)
	}
	press(m, "m")
	if !s.Read(m.guide) {
		t.Fatal("m in a guide didn't mark it read")
	}
	if view := m.View(); !strings.Contains(view, "✓ read") {
		t.Errorf("the guide doesn't show it's read:\n%s", view)
	}
	m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if m.pane != paneGuides {
		t.Fatalf("esc in a guide: pane %v, want back to the guide list", m.pane)
	}
	m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if m.pane != paneResult {
		t.Fatalf("esc in the guide list: pane %v, want the results", m.pane)
	}
}
