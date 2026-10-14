// Package session ties the embedded course to a learner's course directory:
// it creates exercise folders, runs checks, and records progress.
//
// A course directory looks like this:
//
//	.groundwork/progress.json         progress
//	exercises/01_basics/001_hello_go/  every exercise, each its own Go module
//	solutions/01_basics/001_hello_go/  reference solution, written once unlocked
//	guides/01_how_programs_run.md      the guides, kept in step with the course
package session

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/pjanderson/groundwork/internal/course"
	"github.com/pjanderson/groundwork/internal/progress"
	"github.com/pjanderson/groundwork/internal/workspace"
)

// ErrNoCourse means no course directory was found.
var ErrNoCourse = errors.New("no groundwork course here")

// Session is one learner's course directory.
type Session struct {
	Course *course.Course
	Root   string
	State  *progress.State

	// Now is the clock, replaceable in tests; nil means time.Now.
	Now func() time.Time
}

func (s *Session) now() time.Time {
	if s.Now != nil {
		return s.Now().UTC()
	}
	return time.Now().UTC()
}

// FindRoot walks up from start to the nearest directory holding a course.
func FindRoot(start string) (string, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	for {
		// The progress file, not just the directory, marks a course: a stray
		// empty .groundwork/ must not turn a home directory into one.
		if info, err := os.Stat(progress.Path(dir)); err == nil && info.Mode().IsRegular() {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", ErrNoCourse
		}
		dir = parent
	}
}

// Init creates a course directory at dir.
func Init(dir string) error {
	if _, err := os.Stat(progress.Path(dir)); err == nil {
		return fmt.Errorf("%s already contains a course", dir)
	}
	return progress.Save(dir, progress.New())
}

// Open loads the progress of the course at root.
func Open(root string, c *course.Course) (*Session, error) {
	state, err := progress.Load(root)
	if err != nil {
		return nil, err
	}
	return &Session{Course: c, Root: root, State: state}, nil
}

// Dir is the exercise's working folder.
func (s *Session) Dir(e *course.Exercise) string {
	return filepath.Join(s.Root, "exercises", filepath.FromSlash(e.Dir()))
}

// SolutionDir is where the reference solution is written once unlocked.
func (s *Session) SolutionDir(e *course.Exercise) string {
	return filepath.Join(s.Root, "solutions", filepath.FromSlash(e.Dir()))
}

// Done reports whether e has passed.
func (s *Session) Done(e *course.Exercise) bool {
	_, ok := s.State.Done[e.Name]
	return ok
}

// Revealed reports whether the learner looked at e's solution before passing.
func (s *Session) Revealed(e *course.Exercise) bool {
	_, ok := s.State.Revealed[e.Name]
	return ok
}

// Read reports whether the learner marked g as read.
func (s *Session) Read(g *course.Guide) bool {
	_, ok := s.State.Read[g.Name]
	return ok
}

// SetRead marks g as read or unread and saves progress.
func (s *Session) SetRead(g *course.Guide, read bool) error {
	if read {
		s.State.Read[g.Name] = time.Now().UTC()
	} else {
		delete(s.State.Read, g.Name)
	}
	return progress.Save(s.Root, s.State)
}

// NextGuide is the first unread guide for the current module or an earlier
// one, or nil when they are all read: the reading that should come next.
func (s *Session) NextGuide() *course.Guide {
	current := s.Current().Module.Number
	for _, g := range s.Course.Guides {
		if g.Module.Number > current {
			break
		}
		if !s.Read(g) {
			return g
		}
	}
	return nil
}

// DoneCount counts finished exercises that still exist in the course.
func (s *Session) DoneCount() int {
	n := 0
	for _, e := range s.Course.Exercises {
		if s.Done(e) {
			n++
		}
	}
	return n
}

// Current is the exercise the learner last worked on, else the first unfinished one.
func (s *Session) Current() *course.Exercise {
	if e, ok := s.Course.Find(s.State.Current); ok {
		return e
	}
	if e := s.NextPending(nil); e != nil {
		return e
	}
	return s.Course.Exercises[len(s.Course.Exercises)-1]
}

// NextPending returns the first unfinished exercise after `after` (or from
// the start when after is nil), wrapping around. It returns nil when every
// exercise is done.
func (s *Session) NextPending(after *course.Exercise) *course.Exercise {
	all := s.Course.Exercises
	start := 0
	if after != nil {
		start = after.Number // index of the exercise after `after`
	}
	for i := range all {
		e := all[(start+i)%len(all)]
		if !s.Done(e) {
			return e
		}
	}
	return nil
}

// Sync makes the course folder match the course:
//   - an exercise whose folder moved (the course was reordered) is moved,
//     with the learner's work and solution;
//   - missing exercise folders are created;
//   - files the learner never edited are refreshed when the course changes;
//   - unlocked solutions are rewritten from the course (they're never the
//     learner's work);
//   - guides/ is rewritten to match the course's guides.
//
// Edited exercise files are never touched.
func (s *Session) Sync() error {
	if err := s.moveFolders(); err != nil {
		return err
	}
	if err := s.writeGuides(); err != nil {
		return fmt.Errorf("update guides/: %w", err)
	}
	for _, e := range s.Course.Exercises {
		if _, err := s.sync(e, false); err != nil {
			return fmt.Errorf("update %s: %w", e.Dir(), err)
		}
		if s.Done(e) || s.Revealed(e) {
			if err := s.writeSolutionFiles(e); err != nil {
				return fmt.Errorf("update the solution of %s: %w", e.Dir(), err)
			}
		}
	}
	return progress.Save(s.Root, s.State)
}

// moveFolders relocates exercise and solution folders the course moved
// since the last Sync. Inserting an exercise shifts the numbers of the ones
// after it, and follow-ups share their topic's name (003_variables,
// 004_variables), so a folder's new name is often still taken by the
// exercise that is about to move on. Every moving folder is therefore first
// renamed aside, then put in place.
func (s *Session) moveFolders() error {
	type move struct{ from, aside, to string }
	var moves []move
	for _, e := range s.Course.Exercises {
		previous := s.State.Folders[e.Name]
		s.State.Folders[e.Name] = e.Dir()
		if previous == "" || previous == e.Dir() {
			continue
		}
		for _, top := range []string{"exercises", "solutions"} {
			from := filepath.Join(s.Root, top, filepath.FromSlash(previous))
			if _, err := os.Stat(from); err != nil {
				continue
			}
			moves = append(moves, move{
				from:  from,
				aside: filepath.Join(s.Root, top, ".moving-"+e.Name),
				to:    filepath.Join(s.Root, top, filepath.FromSlash(e.Dir())),
			})
		}
	}
	for _, m := range moves {
		if err := os.Rename(m.from, m.aside); err != nil {
			return fmt.Errorf("move %s: %w", m.from, err)
		}
	}
	for _, m := range moves {
		if _, err := os.Stat(m.to); err == nil {
			// Something unexpected already lives there; leave the work under
			// its aside name rather than overwrite anything.
			return fmt.Errorf("can't move your work to %s: that folder already exists (your files are safe in %s)", m.to, m.aside)
		}
		if err := os.MkdirAll(filepath.Dir(m.to), 0o755); err != nil {
			return err
		}
		if err := os.Rename(m.aside, m.to); err != nil {
			return fmt.Errorf("move %s: %w", m.to, err)
		}
		os.Remove(filepath.Dir(m.from)) // drop an old module folder once empty
	}
	return nil
}

// GuidePath is where the learner can read g in their editor.
func (s *Session) GuidePath(g *course.Guide) string {
	return filepath.Join(s.Root, "guides", g.File())
}

// guideFile matches the files writeGuides owns: "07_data_structures.md".
var guideFile = regexp.MustCompile(`^\d{2}_[a-z0-9_]+\.md$`)

// writeGuides makes guides/ hold every guide plus a README.md index. Guide
// files the course no longer has (renamed or renumbered) are removed; other
// files the learner put there are left alone.
func (s *Session) writeGuides() error {
	dir := filepath.Join(s.Root, "guides")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	want := map[string][]byte{"README.md": []byte(GuideIndex(s.Course))}
	for _, g := range s.Course.Guides {
		text, err := s.Course.GuideText(g)
		if err != nil {
			return err
		}
		want[g.File()] = []byte(text)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if _, keep := want[entry.Name()]; !keep && guideFile.MatchString(entry.Name()) {
			if err := os.Remove(filepath.Join(dir, entry.Name())); err != nil {
				return err
			}
		}
	}
	for name, text := range want {
		path := filepath.Join(dir, name)
		if have, err := os.ReadFile(path); err == nil && string(have) == string(text) {
			continue
		}
		if err := os.WriteFile(path, text, 0o644); err != nil {
			return err
		}
	}
	return nil
}

// GuideIndex is guides/README.md: every guide with its one-sentence summary,
// grouped by the module it goes with.
func GuideIndex(c *course.Course) string {
	var b strings.Builder
	b.WriteString("# Guides\n\n" +
		"The computer science underneath the exercises, one idea per guide. Read\n" +
		"each one alongside its module, or whenever an idea feels shaky. In the\n" +
		"watcher, press `g`; from a shell, run `groundwork guides`.\n\n" +
		"Groundwork rewrites the numbered files here whenever the course changes,\n" +
		"so keep your own notes in another file.\n")
	for _, m := range c.Modules {
		if len(m.Guides) == 0 {
			continue
		}
		fmt.Fprintf(&b, "\n## Module %d: %s\n\n", m.Number, m.Title)
		for _, g := range m.Guides {
			fmt.Fprintf(&b, "- **[%s](%s)**: %s\n", g.Title, g.File(), g.Summary)
		}
	}
	return b.String()
}

func module(e *course.Exercise) string { return "groundwork.local/" + e.Leaf() }

// sync writes e's starter files that are missing, unedited and out of date,
// or all of them with force. It returns the paths it wrote.
func (s *Session) sync(e *course.Exercise, force bool) ([]string, error) {
	pack, err := s.Course.Pack(e)
	if err != nil {
		return nil, err
	}
	files, err := workspace.Starter(pack, module(e))
	if err != nil {
		return nil, err
	}
	var written []string
	for _, rel := range slices.Sorted(maps.Keys(files)) {
		want, key := files[rel], e.Name+"/"+rel
		target := filepath.Join(s.Dir(e), filepath.FromSlash(rel))
		have, err := os.ReadFile(target)
		missing := errors.Is(err, fs.ErrNotExist)
		if err != nil && !missing {
			return nil, err
		}
		switch {
		case string(have) == string(want) && !missing:
			// already current
		case force, missing,
			rel == "README.md", // instructions, never the learner's work
			s.State.Starters[key] == digest(have):
			if err := workspace.Write(s.Dir(e), map[string][]byte{rel: want}); err != nil {
				return nil, err
			}
			written = append(written, rel)
		default:
			continue // the learner changed it: leave it alone
		}
		s.State.Starters[key] = digest(want)
	}
	return written, nil
}

func digest(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// Prepare records e as the current exercise, creating its folder if needed.
func (s *Session) Prepare(e *course.Exercise) error {
	if _, err := s.sync(e, false); err != nil {
		return fmt.Errorf("create %s: %w", e.Dir(), err)
	}
	s.State.Current = e.Name
	return progress.Save(s.Root, s.State)
}

// Check runs the exercise's tests.
func (s *Session) Check(ctx context.Context, e *course.Exercise) (*workspace.Result, error) {
	return workspace.Check(ctx, s.Dir(e), e.Race)
}

// Complete records e as done and writes its reference solution. Passing an
// exercise that needs reviews schedules the first one; passing a review
// schedules the next.
func (s *Session) Complete(e *course.Exercise) error {
	now := s.now()
	switch r, reviewing := s.State.Reviews[e.Name]; {
	case !s.Done(e):
		s.State.Done[e.Name] = now
		if e.Kind == course.Practice || e.Kind == course.Project || s.Revealed(e) {
			s.State.Reviews[e.Name] = progress.Review{Due: now.Add(ReviewIntervals[0])}
		}
	case reviewing && r.Active:
		s.State.Reviews[e.Name] = nextReview(r, now)
		if r.Step+1 == len(ReviewIntervals) && !r.Lapsed {
			delete(s.State.Reviews, e.Name) // remembered for good
		}
	}
	return s.writeSolution(e)
}

// ReviewIntervals are the waits before each spaced review of an exercise:
// solving it again from scratch, just as it starts to fade, is what makes it
// stick. Practice and project exercises are reviewed, and any exercise whose
// solution was shown before it passed.
var ReviewIntervals = []time.Duration{2 * day, 7 * day, 30 * day}

const day = 24 * time.Hour

// nextReview is the schedule after passing review r: the next interval, or
// back to the first one if the learner needed the solution or an Answer.
func nextReview(r progress.Review, now time.Time) progress.Review {
	step := r.Step + 1
	if r.Lapsed {
		step = 0
	}
	return progress.Review{Step: step, Due: now.Add(ReviewIntervals[min(step, len(ReviewIntervals)-1)])}
}

// DueReviews lists the exercises whose next review is due, oldest first.
func (s *Session) DueReviews() []*course.Exercise {
	now := s.now()
	var due []*course.Exercise
	for _, e := range s.Course.Exercises {
		if r, ok := s.State.Reviews[e.Name]; ok && !r.Active && !r.Due.After(now) {
			due = append(due, e)
		}
	}
	slices.SortStableFunc(due, func(a, b *course.Exercise) int {
		return s.State.Reviews[a.Name].Due.Compare(s.State.Reviews[b.Name].Due)
	})
	return due
}

// Review returns e's review schedule, if it has one.
func (s *Session) Review(e *course.Exercise) (progress.Review, bool) {
	r, ok := s.State.Reviews[e.Name]
	return r, ok
}

// StartReview saves a copy of the learner's current files for e under
// .groundwork/attempts/, resets e to its starter code, and marks the review
// as in progress. It returns where the copy went.
func (s *Session) StartReview(e *course.Exercise) (string, error) {
	r, ok := s.State.Reviews[e.Name]
	if !ok {
		return "", fmt.Errorf("%s has no review scheduled", e.Title)
	}
	backup := filepath.Join(s.Root, progress.Dir, "attempts", e.Name, s.now().Format("20060102-150405"))
	if err := copyWork(s.Dir(e), backup); err != nil {
		return "", fmt.Errorf("save your previous code: %w", err)
	}
	if _, err := s.sync(e, true); err != nil {
		return "", err
	}
	r.Active, r.Lapsed = true, false
	s.State.Reviews[e.Name] = r
	s.State.Current = e.Name
	return backup, progress.Save(s.Root, s.State)
}

// Lapse records that the learner used the solution or an Answer hint during
// e's review, so passing it restarts the schedule instead of advancing it.
func (s *Session) Lapse(e *course.Exercise) error {
	r, ok := s.State.Reviews[e.Name]
	if !ok || !r.Active || r.Lapsed {
		return nil
	}
	r.Lapsed = true
	s.State.Reviews[e.Name] = r
	return progress.Save(s.Root, s.State)
}

// copyWork copies the files under dir to backup, skipping scratch repos.
func copyWork(dir, backup string) error {
	return filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if name := entry.Name(); name == "sandbox" || name == ".git" {
				return fs.SkipDir
			}
			return nil
		}
		if !entry.Type().IsRegular() {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		target := filepath.Join(backup, rel)
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
}

// Reveal writes e's reference solution before the learner has passed, and
// remembers that they looked.
func (s *Session) Reveal(e *course.Exercise) error {
	if !s.Revealed(e) && !s.Done(e) {
		s.State.Revealed[e.Name] = time.Now().UTC()
	}
	return s.writeSolution(e)
}

func (s *Session) writeSolution(e *course.Exercise) error {
	if err := progress.Save(s.Root, s.State); err != nil {
		return err
	}
	return s.writeSolutionFiles(e)
}

func (s *Session) writeSolutionFiles(e *course.Exercise) error {
	pack, err := s.Course.Pack(e)
	if err != nil {
		return err
	}
	files, err := workspace.Solution(pack, module(e))
	if err != nil {
		return err
	}
	return workspace.Write(s.SolutionDir(e), files)
}

// File is one file of a reference solution.
type File struct {
	Path    string // relative to the exercise folder
	Content string
}

// SolutionFiles returns the files the reference solution adds or changes,
// the ones worth reading next to your own code.
func (s *Session) SolutionFiles(e *course.Exercise) ([]File, error) {
	pack, err := s.Course.Pack(e)
	if err != nil {
		return nil, err
	}
	changed, err := workspace.SolutionChanges(pack)
	if err != nil {
		return nil, err
	}
	files, err := workspace.Solution(pack, module(e))
	if err != nil {
		return nil, err
	}
	out := make([]File, 0, len(changed))
	for _, p := range changed {
		out = append(out, File{Path: p, Content: string(files[p])})
	}
	return out, nil
}

// Reset overwrites e's starter files with fresh copies and returns their
// paths. Files the learner added are left alone.
func (s *Session) Reset(e *course.Exercise) ([]string, error) {
	written, err := s.sync(e, true)
	if err != nil {
		return nil, err
	}
	return written, progress.Save(s.Root, s.State)
}

// TodoFiles lists the files in e's folder that still contain a TODO marker,
// relative to the course root: Go code, or the shell scripts, SQL files,
// Dockerfiles and so on that non-Go exercises are written in. Exercises are
// edited where the TODOs are.
func (s *Session) TodoFiles(e *course.Exercise) []string {
	var files []string
	dir := s.Dir(e)
	_ = filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		switch name := entry.Name(); {
		case err != nil:
			return nil
		case entry.IsDir() && (name == "testdata" || name == "sandbox" || name == ".git"):
			return fs.SkipDir // fixtures and scratch repos, not the learner's work
		case entry.IsDir(), strings.HasSuffix(name, "_test.go"), strings.HasSuffix(name, ".md"),
			name == "go.mod", name == "go.sum":
			return nil
		}
		if raw, err := os.ReadFile(path); err == nil && strings.Contains(string(raw), "TODO(") {
			rel, _ := filepath.Rel(s.Root, path)
			files = append(files, rel)
		}
		return nil
	})
	return files
}
