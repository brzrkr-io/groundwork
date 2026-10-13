// Package progress persists which exercises a learner has opened and finished.
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
	Dir     = ".gopherlings"
	file    = "progress.json"
	version = 3
)

// State is keyed by exercise name, so reordering the course keeps progress.
type State struct {
	Version int                  `json:"version"`
	Current string               `json:"current,omitempty"`
	Done    map[string]time.Time `json:"done"`
	// Starters records the SHA-256 of every starter file as last written,
	// keyed "exercise/path". A file that still has that hash is untouched by
	// the learner, so course updates may replace it.
	Starters map[string]string `json:"starters,omitempty"`
}

// New returns empty progress.
func New() *State {
	return &State{Version: version, Done: map[string]time.Time{}, Starters: map[string]string{}}
}

// Path is the progress file of the course at root.
func Path(root string) string { return filepath.Join(root, Dir, file) }

// Load reads root/.gopherlings/progress.json. A missing file is empty progress.
func Load(root string) (*State, error) {
	raw, err := os.ReadFile(Path(root))
	if errors.Is(err, fs.ErrNotExist) {
		return New(), nil
	}
	if err != nil {
		return nil, fmt.Errorf("read progress: %w", err)
	}
	s := New()
	if err := json.Unmarshal(raw, s); err != nil {
		return nil, fmt.Errorf("decode %s: %w", filepath.Join(Dir, file), err)
	}
	if s.Version != version {
		return nil, fmt.Errorf("%s has version %d, this gopherlings understands version %d; move it aside to start fresh",
			filepath.Join(Dir, file), s.Version, version)
	}
	if s.Done == nil {
		s.Done = map[string]time.Time{}
	}
	if s.Starters == nil {
		s.Starters = map[string]string{}
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
