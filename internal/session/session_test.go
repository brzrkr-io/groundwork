package session

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/pjanderson/groundwork/internal/course"
)

const starter = "package x\n\n// TODO(1): fix\nfunc F() int { return 0 }\n"

const threeExercises = `{"title":"t","modules":[
	{"title":"m1","dir":"one","exercises":[{"name":"a","title":"A"},{"name":"b","title":"B"}]},
	{"title":"m2","dir":"two","exercises":[{"name":"c","title":"C"}]}]}`

func testCourse(t *testing.T) *course.Course {
	t.Helper()
	return loadCourse(t, threeExercises, starter)
}

func courseWithStarter(t *testing.T, starter string) *course.Course {
	t.Helper()
	return loadCourse(t, threeExercises, starter)
}

// loadCourse builds a course from courseJSON with a small pack for every
// exercise named in it.
func loadCourse(t *testing.T, courseJSON, starter string) *course.Course {
	t.Helper()
	files := fstest.MapFS{"course.json": {Data: []byte(courseJSON)}}
	for _, name := range []string{"a", "a_2", "b", "c", "new"} {
		files["exercises/"+name+"/README.md"] = &fstest.MapFile{Data: []byte("# " + name + "\n\n## Your task\n")}
		files["exercises/"+name+"/hints.md"] = &fstest.MapFile{Data: []byte("## One\na\n## Two\nb\n")}
		files["exercises/"+name+"/exercise.go.tmpl"] = &fstest.MapFile{Data: []byte(starter)}
		files["exercises/"+name+"/solution/exercise.go.tmpl"] = &fstest.MapFile{Data: []byte("package x\n\nfunc F() int { return 1 }\n")}
	}
	c, err := course.Load(files)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func newSession(t *testing.T) *Session {
	t.Helper()
	root := t.TempDir()
	if err := Init(root); err != nil {
		t.Fatal(err)
	}
	s, err := Open(root, testCourse(t))
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestNextPendingWrapsAndStopsWhenAllDone(t *testing.T) {
	s := newSession(t)
	a, b, c := s.Course.Exercises[0], s.Course.Exercises[1], s.Course.Exercises[2]
	s.State.Done[b.Name] = time.Now()
	if got := s.NextPending(a); got != c {
		t.Errorf("NextPending(a) with b done = %v, want c", got.Name)
	}
	if got := s.NextPending(c); got != a {
		t.Errorf("NextPending(c) = %v, want it to wrap to a", got.Name)
	}
	s.State.Done[a.Name], s.State.Done[c.Name] = time.Now(), time.Now()
	if got := s.NextPending(a); got != nil {
		t.Errorf("NextPending with everything done = %v, want nil", got.Name)
	}
}

func TestSyncCreatesEveryExerciseAndUpdatesOnlyUneditedFiles(t *testing.T) {
	s := newSession(t)
	if err := s.Sync(); err != nil {
		t.Fatal(err)
	}
	for _, e := range s.Course.Exercises {
		if _, err := os.Stat(filepath.Join(s.Dir(e), "exercise.go")); err != nil {
			t.Fatalf("Sync did not create %s: %v", e.Dir(), err)
		}
	}
	edited := filepath.Join(s.Dir(s.Course.Exercises[0]), "exercise.go")
	untouched := filepath.Join(s.Dir(s.Course.Exercises[1]), "exercise.go")
	if err := os.WriteFile(edited, []byte("package x\n// my work\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// The course ships a new version of the starter.
	const v2 = "package x\n\n// TODO(1): fix, now clearer\nfunc F() int { return 0 }\n"
	updated, err := Open(s.Root, courseWithStarter(t, v2))
	if err != nil {
		t.Fatal(err)
	}
	if err := updated.Sync(); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(edited); string(got) != "package x\n// my work\n" {
		t.Errorf("Sync overwrote the learner's edit: %q", got)
	}
	if got, _ := os.ReadFile(untouched); string(got) != v2 {
		t.Errorf("unedited exercise.go = %q, want the course's new version", got)
	}
}

func TestSyncMovesWorkWhenTheCourseIsReordered(t *testing.T) {
	s := newSession(t)
	if err := s.Sync(); err != nil {
		t.Fatal(err)
	}
	c := s.Course.Exercises[2] // "c", module two, number 3
	if err := s.Complete(c); err != nil {
		t.Fatal(err)
	}
	mine := filepath.Join(s.Dir(c), "exercise.go")
	if err := os.WriteFile(mine, []byte("package x\n// my work\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// A new exercise is inserted before c, so c moves from 003_c to 004_c.
	reordered := loadCourse(t, `{"title":"t","modules":[
		{"title":"m1","dir":"one","exercises":[{"name":"a","title":"A"},{"name":"b","title":"B"}]},
		{"title":"m2","dir":"two","exercises":[{"name":"new","title":"New"},{"name":"c","title":"C"}]}]}`, starter)
	moved, err := Open(s.Root, reordered)
	if err != nil {
		t.Fatal(err)
	}
	if err := moved.Sync(); err != nil {
		t.Fatal(err)
	}
	nc, _ := reordered.Find("c")
	if got, _ := os.ReadFile(filepath.Join(moved.Dir(nc), "exercise.go")); string(got) != "package x\n// my work\n" {
		t.Errorf("work at the new folder %s = %q, want the learner's file moved there", nc.Dir(), got)
	}
	if _, err := os.Stat(filepath.Join(moved.SolutionDir(nc), "exercise.go")); err != nil {
		t.Errorf("solution not moved with the exercise: %v", err)
	}
	if _, err := os.Stat(mine); err == nil {
		t.Errorf("old folder %s still exists after the move", c.Dir())
	}
}

func TestSyncShiftsSameTopicFoldersWhenAnExerciseIsInserted(t *testing.T) {
	root := t.TempDir()
	if err := Init(root); err != nil {
		t.Fatal(err)
	}
	// a and a_2 share the topic "a": folders 001_a and 002_a.
	before := loadCourse(t, `{"modules":[{"title":"m","dir":"one","exercises":[{"name":"a"},{"name":"a_2"}]}]}`, starter)
	s, err := Open(root, before)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Sync(); err != nil {
		t.Fatal(err)
	}
	for name, body := range map[string]string{"a": "// work on a\n", "a_2": "// work on a_2\n"} {
		e, _ := before.Find(name)
		if err := os.WriteFile(filepath.Join(s.Dir(e), "exercise.go"), []byte("package x\n"+body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// Inserting "new" first moves a to 002_a, which a_2 occupies until it
	// moves on to 003_a.
	after := loadCourse(t, `{"modules":[{"title":"m","dir":"one","exercises":[{"name":"new"},{"name":"a"},{"name":"a_2"}]}]}`, starter)
	s2, err := Open(root, after)
	if err != nil {
		t.Fatal(err)
	}
	if err := s2.Sync(); err != nil {
		t.Fatal(err)
	}
	for name, want := range map[string]string{"a": "// work on a\n", "a_2": "// work on a_2\n"} {
		e, _ := after.Find(name)
		got, _ := os.ReadFile(filepath.Join(s2.Dir(e), "exercise.go"))
		if string(got) != "package x\n"+want {
			t.Errorf("%s in %s = %q, want the learner's work %q", name, e.Dir(), got, want)
		}
	}
	if leftovers, _ := filepath.Glob(filepath.Join(root, "exercises", ".moving-*")); len(leftovers) > 0 {
		t.Errorf("temporary folders left behind: %v", leftovers)
	}
}

func TestPrepareKeepsLearnerEdits(t *testing.T) {
	s := newSession(t)
	a := s.Course.Exercises[0]
	if err := s.Prepare(a); err != nil {
		t.Fatal(err)
	}
	src := filepath.Join(s.Dir(a), "exercise.go")
	if err := os.WriteFile(src, []byte("package x\n\nfunc F() int { return 1 }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	reopened, err := Open(s.Root, s.Course)
	if err != nil {
		t.Fatal(err)
	}
	if err := reopened.Prepare(a); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(src); string(got) != "package x\n\nfunc F() int { return 1 }\n" {
		t.Fatalf("Prepare overwrote the learner's edit: %q", got)
	}
	if reopened.Current() != a {
		t.Errorf("Current() = %v, want the prepared exercise", reopened.Current().Name)
	}
}

func TestCompleteWritesSolutionAndResetRestoresStarter(t *testing.T) {
	s := newSession(t)
	b := s.Course.Exercises[1]
	if err := s.Prepare(b); err != nil {
		t.Fatal(err)
	}
	if err := s.Complete(b); err != nil {
		t.Fatal(err)
	}
	if !s.Done(b) || s.DoneCount() != 1 {
		t.Fatalf("Done(b) = %v, DoneCount = %d; want b done", s.Done(b), s.DoneCount())
	}
	solution, err := os.ReadFile(filepath.Join(s.SolutionDir(b), "exercise.go"))
	if err != nil || string(solution) != "package x\n\nfunc F() int { return 1 }\n" {
		t.Fatalf("solution file = %q, %v", solution, err)
	}
	src := filepath.Join(s.Dir(b), "exercise.go")
	if err := os.WriteFile(src, []byte("package x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Reset(b); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(src); string(got) != "package x\n\n// TODO(1): fix\nfunc F() int { return 0 }\n" {
		t.Fatalf("after Reset exercise.go = %q, want the starter", got)
	}
	if files := s.TodoFiles(b); len(files) != 1 {
		t.Errorf("TodoFiles after reset = %v, want exercise.go", files)
	}
}

func TestFindRootWalksUpFromExerciseFolders(t *testing.T) {
	s := newSession(t)
	a := s.Course.Exercises[0]
	if err := s.Prepare(a); err != nil {
		t.Fatal(err)
	}
	got, err := FindRoot(s.Dir(a))
	if err != nil || got != s.Root {
		t.Fatalf("FindRoot(exercise dir) = %q, %v; want %q", got, err, s.Root)
	}
	stray := t.TempDir()
	if err := os.Mkdir(filepath.Join(stray, ".groundwork"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := FindRoot(stray); err != ErrNoCourse {
		t.Fatalf("FindRoot(dir with an empty .groundwork) error = %v, want ErrNoCourse", err)
	}
}

func TestSyncWritesGuidesAndDropsStaleOnes(t *testing.T) {
	files := fstest.MapFS{
		"course.json":                           {Data: []byte(`{"modules":[{"title":"M","dir":"m","guides":["state"],"exercises":[{"name":"a","title":"A"}]}]}`)},
		"exercises/a/README.md":                 {Data: []byte("# A\n")},
		"exercises/a/exercise.go.tmpl":          {Data: []byte(starter)},
		"exercises/a/solution/exercise.go.tmpl": {Data: []byte(starter)},
		"guides/state.md":                       {Data: []byte("# State\n\n**In one sentence:** state changes.\n")},
	}
	c, err := course.Load(files)
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	if err := Init(root); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(root, "guides")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for name, text := range map[string]string{
		"01_state.md": "# State\n\nan edit the course replaces\n",
		"02_old.md":   "# A guide the course dropped\n",
		"my-notes.md": "mine\n",
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	s, err := Open(root, c)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Sync(); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(s.GuidePath(c.Guides[0])); string(got) != "# State\n\n**In one sentence:** state changes.\n" {
		t.Errorf("01_state.md = %q, want the course's text", got)
	}
	if _, err := os.Stat(filepath.Join(dir, "02_old.md")); err == nil {
		t.Error("a guide the course no longer has was left behind")
	}
	if got, _ := os.ReadFile(filepath.Join(dir, "my-notes.md")); string(got) != "mine\n" {
		t.Error("the learner's own file in guides/ was touched")
	}
	if index, _ := os.ReadFile(filepath.Join(dir, "README.md")); !strings.Contains(string(index), "- **[State](01_state.md)**: state changes.") {
		t.Errorf("README.md index lacks the guide:\n%s", index)
	}
}

func TestNextGuideIsTheFirstUnreadUpToTheCurrentModule(t *testing.T) {
	files := fstest.MapFS{
		"course.json": {Data: []byte(`{"modules":[
			{"title":"M1","dir":"one","guides":["g1","g2"],"exercises":[{"name":"a","title":"A"}]},
			{"title":"M2","dir":"two","guides":["g3"],"exercises":[{"name":"b","title":"B"}]}]}`)},
		"guides/g1.md": {Data: []byte("# G1\n")},
		"guides/g2.md": {Data: []byte("# G2\n")},
		"guides/g3.md": {Data: []byte("# G3\n")},
	}
	for _, name := range []string{"a", "b"} {
		files["exercises/"+name+"/README.md"] = &fstest.MapFile{Data: []byte("# " + name + "\n")}
		files["exercises/"+name+"/exercise.go.tmpl"] = &fstest.MapFile{Data: []byte(starter)}
	}
	c, err := course.Load(files)
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	if err := Init(root); err != nil {
		t.Fatal(err)
	}
	s, err := Open(root, c)
	if err != nil {
		t.Fatal(err)
	}
	g1, g2 := c.Guides[0], c.Guides[1]
	if got := s.NextGuide(); got != g1 {
		t.Fatalf("NextGuide() = %v, want g1", got)
	}
	if err := s.SetRead(g1, true); err != nil {
		t.Fatal(err)
	}
	if got := s.NextGuide(); got != g2 {
		t.Fatalf("after reading g1, NextGuide() = %v, want g2", got)
	}
	if err := s.SetRead(g2, true); err != nil {
		t.Fatal(err)
	}
	if got := s.NextGuide(); got != nil {
		t.Errorf("module one's guides are read; NextGuide() = %s, want nil (g3 is for a later module)", got.Name)
	}
	reopened, err := Open(root, c)
	if err != nil {
		t.Fatal(err)
	}
	if !reopened.Read(g1) {
		t.Error("read marks weren't saved")
	}
	if err := reopened.SetRead(g1, false); err != nil || reopened.NextGuide() != g1 {
		t.Error("unmarking g1 didn't make it next again")
	}
}

func TestTodoFilesFindsScriptsButNotFixturesOrDocs(t *testing.T) {
	s := newSession(t)
	a := s.Course.Exercises[0]
	if err := s.Prepare(a); err != nil {
		t.Fatal(err)
	}
	for name, text := range map[string]string{
		"exercise.go":        "package x\n",
		"count_errors.sh":    "# TODO(1): count the ERROR lines\n",
		"queries/report.sql": "-- TODO(2): join the tables\n",
		"testdata/app.log":   "TODO(1) is just log text here\n",
		"sandbox/notes.txt":  "TODO(3) in the learner's scratch repo\n",
		"README.md":          "Each `TODO(n)` marks a task.\n",
		"exercise_test.go":   "package x // TODO(1) mentioned in a test\n",
	} {
		path := filepath.Join(s.Dir(a), filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	var got []string
	for _, f := range s.TodoFiles(a) {
		got = append(got, filepath.ToSlash(f))
	}
	want := "exercises/" + a.Dir() + "/count_errors.sh exercises/" + a.Dir() + "/queries/report.sql"
	if strings.Join(got, " ") != want {
		t.Errorf("TodoFiles = %v, want %s", got, want)
	}
}
