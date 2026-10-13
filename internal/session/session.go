// Package session ties the embedded course to a learner's course directory:
// it creates exercise folders, runs checks, and records progress.
//
// A course directory looks like this:
//
//	.gopherlings/progress.json         progress
//	exercises/01_basics/001_hello_go/  every exercise, each its own Go module
//	solutions/01_basics/001_hello_go/  reference solution, written once unlocked
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
	"slices"
	"strings"
	"time"

	"github.com/pjanderson/gopherlings/internal/course"
	"github.com/pjanderson/gopherlings/internal/progress"
	"github.com/pjanderson/gopherlings/internal/workspace"
)

// ErrNoCourse means no course directory was found.
var ErrNoCourse = errors.New("no gopherlings course here")

// Session is one learner's course directory.
type Session struct {
	Course *course.Course
	Root   string
	State  *progress.State
}

// FindRoot walks up from start to the nearest directory holding a course.
func FindRoot(start string) (string, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	for {
		// The progress file, not just the directory, marks a course: a stray
		// empty .gopherlings/ must not turn a home directory into one.
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
//     learner's work).
//
// Edited exercise files are never touched.
func (s *Session) Sync() error {
	if err := s.moveFolders(); err != nil {
		return err
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

func module(e *course.Exercise) string { return "gopherlings.local/" + e.Leaf() }

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

// Complete records e as done and writes its reference solution.
func (s *Session) Complete(e *course.Exercise) error {
	if !s.Done(e) {
		s.State.Done[e.Name] = time.Now().UTC()
	}
	return s.writeSolution(e)
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
// relative to the course root. Exercises are edited where the TODOs are.
func (s *Session) TodoFiles(e *course.Exercise) []string {
	var files []string
	dir := s.Dir(e)
	_ = filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || !strings.HasSuffix(path, ".go") {
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
