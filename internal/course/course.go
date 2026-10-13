// Package course loads the embedded curriculum: modules, exercises and the
// per-exercise packs (starter files, README, hints and reference solution).
package course

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"slices"
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
	Outcome   string      `json:"outcome"` // "After this module you can: ..."
	Exercises []*Exercise `json:"exercises"`

	Number int `json:"-"`
}

// Dir is the module's folder, for example "01_basics".
func (m *Module) Dir() string { return fmt.Sprintf("%02d_%s", m.Number, m.Folder) }

// Hint is one level of a task's hints: Nudge, Closer or Answer.
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

// Task is one numbered task of an exercise with its hints, gentlest first.
type Task struct {
	Number int
	Title  string
	Hints  []Hint // Nudge, Closer, Answer
}

// HintLevels names the three hints every task has, in reveal order.
var HintLevels = []string{"Nudge", "Closer", "Answer"}

var (
	readmeTask = regexp.MustCompile(`(?m)^### (\d+)\. (.+)$`)
	hintTask   = regexp.MustCompile(`^## Task (\d+): (.+)$`)
)

// Tasks returns e's tasks: titles from the README's "### N. Title" headings,
// hints from hints.md. See ParseHints for the format.
func (c *Course) Tasks(e *Exercise) ([]Task, error) {
	readme, err := c.README(e)
	if err != nil {
		return nil, err
	}
	raw, err := fs.ReadFile(c.packs, path.Join(e.Name, "hints.md"))
	if err != nil {
		return nil, err
	}
	return ParseHints(readme, string(raw))
}

// ParseHints matches hints.md to the README's tasks. hints.md has one
// section per task, in order, each with exactly three levels:
//
//	## Task 1: <same title as the README's "### 1. ...">
//	<nudge: what to look at, no code answer>
//	### Closer
//	<the syntax or API to use, a partial snippet>
//	### Answer
//	<near-complete code for this task>
func ParseHints(readme, hints string) ([]Task, error) {
	var tasks []Task
	for _, m := range readmeTask.FindAllStringSubmatch(readme, -1) {
		n, _ := strconv.Atoi(m[1])
		tasks = append(tasks, Task{Number: n, Title: strings.TrimSpace(m[2])})
	}
	var problems []error
	var cur *Task
	var level *Hint
	inCode := false
	for _, line := range strings.Split(hints, "\n") {
		if strings.HasPrefix(line, "```") {
			inCode = !inCode
		}
		if inCode || strings.HasPrefix(line, "```") {
			if level != nil {
				level.Body += line + "\n"
			}
			continue
		}
		if m := hintTask.FindStringSubmatch(line); m != nil {
			n, _ := strconv.Atoi(m[1])
			if n < 1 || n > len(tasks) {
				problems = append(problems, fmt.Errorf("hints.md has Task %d, but the README has %d tasks", n, len(tasks)))
				cur, level = nil, nil
				continue
			}
			cur = &tasks[n-1]
			if title := strings.TrimSpace(m[2]); title != cur.Title {
				problems = append(problems, fmt.Errorf("hints.md Task %d is %q, but the README calls it %q", n, title, cur.Title))
			}
			cur.Hints = append(cur.Hints, Hint{Title: HintLevels[0]})
			level = &cur.Hints[len(cur.Hints)-1]
			continue
		}
		if title, ok := strings.CutPrefix(line, "### "); ok && cur != nil {
			cur.Hints = append(cur.Hints, Hint{Title: strings.TrimSpace(title)})
			level = &cur.Hints[len(cur.Hints)-1]
			continue
		}
		if strings.HasPrefix(line, "## ") {
			problems = append(problems, fmt.Errorf("hints.md heading %q is not \"## Task N: Title\"", line))
			cur, level = nil, nil
			continue
		}
		if level != nil {
			level.Body += line + "\n"
		}
	}
	if len(tasks) == 0 {
		problems = append(problems, errors.New(`the README has no "### 1. ..." task headings`))
	}
	for i := range tasks {
		t := &tasks[i]
		var titles []string
		for j := range t.Hints {
			t.Hints[j].Body = strings.TrimSpace(t.Hints[j].Body)
			titles = append(titles, t.Hints[j].Title)
			if t.Hints[j].Body == "" {
				problems = append(problems, fmt.Errorf("Task %d's %s hint is empty", t.Number, t.Hints[j].Title))
			}
		}
		if !slices.Equal(titles, HintLevels) {
			problems = append(problems, fmt.Errorf("Task %d has hints %v, want exactly %v", t.Number, titles, HintLevels))
		}
	}
	return tasks, errors.Join(problems...)
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
