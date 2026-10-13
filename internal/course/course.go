// Package course loads the embedded curriculum: modules, exercises and the
// per-exercise packs (starter files, README, hints and reference solution).
package course

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"strconv"
	"strings"
)

// Kind classifies an exercise.
type Kind string

const (
	Concept  Kind = "concept"  // introduces one idea
	Practice Kind = "practice" // combines the module's ideas
	Project  Kind = "project"  // multi-file, multi-task build
)

// Exercise is one lesson. Number and Module are derived from its position in
// course.json, so reordering the course never requires renaming packs.
type Exercise struct {
	Name    string `json:"name"`
	Title   string `json:"title"`
	Minutes int    `json:"minutes"`
	Kind    Kind   `json:"kind"`
	Race    bool   `json:"race"`

	Number int     `json:"-"`
	Module *Module `json:"-"`
}

// Dir is the learner-facing folder, relative to exercises/ and using
// forward slashes: the module folder, then the exercise number and topic,
// for example "01_basics/003_variables". Follow-up packs such as
// "variables_2" share their topic's name; the number tells them apart.
func (e *Exercise) Dir() string { return e.Module.Dir() + "/" + e.Leaf() }

// Leaf is the exercise's own folder name, for example "003_variables".
func (e *Exercise) Leaf() string { return fmt.Sprintf("%03d_%s", e.Number, e.Topic()) }

// Topic is the pack name without its follow-up suffix: "variables_2" and
// "variables" are both about "variables".
func (e *Exercise) Topic() string {
	if i := strings.LastIndexByte(e.Name, '_'); i > 0 {
		if _, err := strconv.Atoi(e.Name[i+1:]); err == nil {
			return e.Name[:i]
		}
	}
	return e.Name
}

// Module groups related exercises.
type Module struct {
	Title     string      `json:"title"`
	Folder    string      `json:"dir"`
	Phase     string      `json:"phase"`
	Summary   string      `json:"summary"`
	Exercises []*Exercise `json:"exercises"`

	Number int `json:"-"`
}

// Dir is the module's folder, for example "01_basics".
func (m *Module) Dir() string { return fmt.Sprintf("%02d_%s", m.Number, m.Folder) }

// Hint is one progressively stronger hint from hints.md.
type Hint struct {
	Title string
	Body  string
}

// Course is the loaded curriculum.
type Course struct {
	Title     string    `json:"title"`
	Modules   []*Module `json:"modules"`
	Exercises []*Exercise

	packs fs.FS
}

// Load reads course.json from files and validates every referenced pack.
func Load(files fs.FS) (*Course, error) {
	raw, err := fs.ReadFile(files, "course.json")
	if err != nil {
		return nil, fmt.Errorf("read course.json: %w", err)
	}
	var c Course
	if err := json.Unmarshal(raw, &c); err != nil {
		return nil, fmt.Errorf("decode course.json: %w", err)
	}
	if c.packs, err = fs.Sub(files, "exercises"); err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	var problems []error
	for mi, m := range c.Modules {
		m.Number = mi + 1
		if m.Folder == "" {
			problems = append(problems, fmt.Errorf("module %d (%s) has no dir", m.Number, m.Title))
		}
		for _, e := range m.Exercises {
			e.Number = len(c.Exercises) + 1
			e.Module = m
			if e.Kind == "" {
				e.Kind = Concept
			}
			c.Exercises = append(c.Exercises, e)
			if seen[e.Name] {
				problems = append(problems, fmt.Errorf("duplicate exercise name %q", e.Name))
			}
			seen[e.Name] = true
		}
	}
	if len(c.Exercises) == 0 {
		problems = append(problems, errors.New("course has no exercises"))
	}
	return &c, errors.Join(problems...)
}

// Pack returns the exercise's immutable content pack.
func (c *Course) Pack(e *Exercise) (fs.FS, error) { return fs.Sub(c.packs, e.Name) }

// README returns the exercise instructions.
func (c *Course) README(e *Exercise) (string, error) {
	raw, err := fs.ReadFile(c.packs, path.Join(e.Name, "README.md"))
	return string(raw), err
}

// Hints parses hints.md: every "## " heading starts a new hint.
func (c *Course) Hints(e *Exercise) ([]Hint, error) {
	raw, err := fs.ReadFile(c.packs, path.Join(e.Name, "hints.md"))
	if err != nil {
		return nil, err
	}
	return ParseHints(string(raw)), nil
}

// ParseHints splits markdown into hints at each level-two heading.
func ParseHints(markdown string) []Hint {
	var hints []Hint
	for _, line := range strings.Split(markdown, "\n") {
		if title, ok := strings.CutPrefix(line, "## "); ok {
			hints = append(hints, Hint{Title: strings.TrimSpace(title)})
			continue
		}
		if len(hints) > 0 {
			hints[len(hints)-1].Body += line + "\n"
		}
	}
	for i := range hints {
		hints[i].Body = strings.TrimSpace(hints[i].Body)
	}
	return hints
}

// Find resolves an exercise number ("3", "003"), a pack name ("variables_2"),
// or a folder ("003_variables", "01_basics/003_variables", with or without a
// trailing slash or leading "exercises/").
func (c *Course) Find(query string) (*Exercise, bool) {
	query = strings.Trim(strings.TrimSpace(strings.ToLower(query)), "/")
	query = strings.TrimPrefix(query, "exercises/")
	if n, err := strconv.Atoi(query); err == nil {
		if n >= 1 && n <= len(c.Exercises) {
			return c.Exercises[n-1], true
		}
		return nil, false
	}
	for _, e := range c.Exercises {
		if e.Name == query || e.Leaf() == query || e.Dir() == query {
			return e, true
		}
	}
	return nil, false
}
