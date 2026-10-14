// Package progress persists which exercises a learner has opened and
// finished, and which guides they have read.
package progress

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

const (
	// Dir marks a course root; Load and Save use Dir/progress.json.
	Dir  = ".groundwork"
	file = "progress.json"

	// version is bumped whenever the course folder's layout or this file
	// changes in a way an older groundwork would get wrong. Older binaries
	// then refuse the folder instead of rebuilding it their way.
	version = 7
)

// ErrNewer means the course folder was set up by a newer groundwork.
var ErrNewer = errors.New("this course folder was set up by a newer groundwork")

// State is keyed by exercise name, so reordering the course keeps progress.
type State struct {
	Version int                  `json:"version"`
	Current string               `json:"current,omitempty"`
	Done    map[string]time.Time `json:"done"`
	// Revealed lists exercises whose solution was shown before they passed.
	Revealed map[string]time.Time `json:"revealed,omitempty"`
	// Folders records each exercise's folder (relative to exercises/), so a
	// reordered course can move the learner's work instead of losing it.
	Folders map[string]string `json:"folders,omitempty"`
	// Starters records the SHA-256 of every starter file as last written,
	// keyed "exercise/path". A file that still has that hash is untouched by
	// the learner, so course updates may replace it.
	Starters map[string]string `json:"starters,omitempty"`
	// Read records the guides the learner marked as read, keyed by guide name.
	Read map[string]time.Time `json:"read,omitempty"`
	// Reviews schedules exercises to be solved again from scratch, keyed by
	// exercise name.
	Reviews map[string]Review `json:"reviews,omitempty"`
	// Activity counts what the learner did each day, keyed by local date
	// ("2026-10-13"), for streaks and stats.
	Activity map[string]Day `json:"activity,omitempty"`
}

// Day is one day's activity.
type Day struct {
	Checks  int `json:"checks,omitempty"`  // test runs: every save counts as practice
	Passed  int `json:"passed,omitempty"`  // exercises passed for the first time
	Reviews int `json:"reviews,omitempty"` // reviews passed
	Guides  int `json:"guides,omitempty"`  // guides marked read
}

// DateKey is the Activity key for t: its date in the local time zone, so a
// late-night session counts for the day it felt like.
func DateKey(t time.Time) string { return t.In(time.Local).Format(time.DateOnly) }

// Review is the spaced-repetition schedule of one exercise.
type Review struct {
	Due    time.Time `json:"due"`
	Step   int       `json:"step"`             // how many reviews have passed
	Active bool      `json:"active,omitempty"` // reset and being solved again now
	Lapsed bool      `json:"lapsed,omitempty"` // the solution or an Answer hint was used during it
}

// New returns empty progress.
func New() *State {
	s := &State{Version: version}
	s.init()
	return s
}

func (s *State) init() {
	if s.Done == nil {
		s.Done = map[string]time.Time{}
	}
	if s.Revealed == nil {
		s.Revealed = map[string]time.Time{}
	}
	if s.Folders == nil {
		s.Folders = map[string]string{}
	}
	if s.Starters == nil {
		s.Starters = map[string]string{}
	}
	if s.Read == nil {
		s.Read = map[string]time.Time{}
	}
	if s.Reviews == nil {
		s.Reviews = map[string]Review{}
	}
	if s.Activity == nil {
		s.Activity = map[string]Day{}
	}
}

// Path is the progress file of the course at root.
func Path(root string) string { return filepath.Join(root, Dir, file) }

// Load reads root/.groundwork/progress.json. A missing file is empty progress.
func Load(root string) (*State, error) {
	raw, err := os.ReadFile(Path(root))
	if errors.Is(err, fs.ErrNotExist) {
		return New(), nil
	}
	if err != nil {
		return nil, fmt.Errorf("read progress: %w", err)
	}
	s := &State{}
	if err := json.Unmarshal(raw, s); err != nil {
		return nil, fmt.Errorf("decode %s: %w", filepath.Join(Dir, file), err)
	}
	switch {
	case s.Version > version:
		return nil, fmt.Errorf("%w (format %d, this one reads %d); update it by running go install . in the groundwork source folder", ErrNewer, s.Version, version)
	case s.Version < 3:
		return nil, fmt.Errorf("%s has an old format (%d) that this groundwork can't read; move it aside to start fresh", filepath.Join(Dir, file), s.Version)
	}
	backfill := s.Version < 7
	s.Version = version // formats 3 to 6 only lack the newer, optional fields
	s.init()
	if backfill {
		// Activity is new in format 7: rebuild past days from the dates
		// already recorded, so the history starts with what was done.
		for _, at := range s.Done {
			d := s.Activity[DateKey(at)]
			d.Passed++
			s.Activity[DateKey(at)] = d
		}
		for _, at := range s.Read {
			d := s.Activity[DateKey(at)]
			d.Guides++
			s.Activity[DateKey(at)] = d
		}
	}
	return s, nil
}

// Save writes progress atomically (temp file + rename).
func Save(root string, s *State) error {
	dir := filepath.Join(root, Dir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create %s: %w", dir, err)
	}
	raw, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, "progress-*.tmp")
	if err != nil {
		return fmt.Errorf("save progress: %w", err)
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(append(raw, '\n')); err != nil {
		tmp.Close()
		return fmt.Errorf("save progress: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("save progress: %w", err)
	}
	if err := os.Rename(tmp.Name(), filepath.Join(dir, file)); err != nil {
		return fmt.Errorf("save progress: %w", err)
	}
	return nil
}
