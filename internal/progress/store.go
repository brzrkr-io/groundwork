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
	version = 5
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
	s.Version = version // formats 3 and 4 only lack the newer, optional fields
	s.init()
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
