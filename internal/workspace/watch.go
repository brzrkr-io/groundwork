package workspace

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Watcher reports (debounced) saves of any file under a directory: Go code,
// and the scripts, SQL and configs that non-Go exercises are written in.
// Editor swap and backup files, and the scratch repos under sandbox/, are
// ignored so a save triggers exactly one check.
type Watcher struct {
	Changes <-chan struct{}

	fs   *fsnotify.Watcher
	done chan struct{}
}

const debounce = 150 * time.Millisecond

// Watch starts watching root recursively.
func Watch(root string) (*Watcher, error) {
	fw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("start file watcher: %w", err)
	}
	changes := make(chan struct{}, 1)
	w := &Watcher{Changes: changes, fs: fw, done: make(chan struct{})}
	if err := w.addTree(root); err != nil {
		fw.Close()
		return nil, err
	}
	go w.loop(changes)
	return w, nil
}

// Close stops the watcher and closes Changes.
func (w *Watcher) Close() error {
	close(w.done)
	return w.fs.Close()
}

func (w *Watcher) addTree(root string) error {
	return filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || !entry.IsDir() {
			return err
		}
		return w.fs.Add(path)
	})
}

func (w *Watcher) loop(changes chan<- struct{}) {
	defer close(changes)
	timer := time.NewTimer(time.Hour)
	timer.Stop()
	for {
		select {
		case ev, ok := <-w.fs.Events:
			if !ok {
				return
			}
			if ev.Has(fsnotify.Create) {
				_ = w.addTree(ev.Name) // new directories are watched too
			}
			if relevant(ev.Name) {
				timer.Reset(debounce)
			}
		case _, ok := <-w.fs.Errors:
			if !ok {
				return
			}
		case <-timer.C:
			select {
			case changes <- struct{}{}:
			default: // a change is already pending
			}
		case <-w.done:
			timer.Stop()
			return
		}
	}
}

func relevant(path string) bool {
	base := filepath.Base(path)
	switch {
	case strings.HasPrefix(base, "."), strings.HasSuffix(base, "~"),
		strings.HasSuffix(base, ".swp"), strings.HasSuffix(base, ".swx"),
		base == "4913": // vim probes whether it may write the directory
		return false
	}
	slash := "/" + filepath.ToSlash(path)
	return !strings.Contains(slash, "/sandbox/") && !strings.Contains(slash, "/.git/")
}
