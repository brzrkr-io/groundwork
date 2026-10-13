package course

import (
	"testing"
	"testing/fstest"
)

func TestFindAcceptsNumberNameAndFolder(t *testing.T) {
	c, err := Load(fstest.MapFS{"course.json": {Data: []byte(`{"modules":[
		{"title":"m1","dir":"basics","exercises":[{"name":"vars"},{"name":"vars_2"}]},
		{"title":"m2","dir":"data","exercises":[{"name":"maps","kind":"practice"}]}]}`)}})
	if err != nil {
		t.Fatal(err)
	}
	for _, query := range []string{"2", "002", "vars_2", "002_vars", "01_basics/002_vars", "exercises/01_basics/002_vars/"} {
		e, ok := c.Find(query)
		if !ok || e.Name != "vars_2" {
			t.Errorf("Find(%q) = %v, %v; want vars_2", query, e, ok)
		}
	}
	for _, query := range []string{"0", "4", "nope", ""} {
		if e, ok := c.Find(query); ok {
			t.Errorf("Find(%q) = %v, want no match", query, e.Name)
		}
	}
	// A follow-up shares its topic's folder name; the number tells them apart.
	if e, _ := c.Find("vars_2"); e.Dir() != "01_basics/002_vars" {
		t.Errorf("vars_2 folder = %q, want 01_basics/002_vars", e.Dir())
	}
	if e, _ := c.Find("maps"); e.Module.Number != 2 || e.Kind != Practice || e.Dir() != "02_data/003_maps" {
		t.Errorf("maps = module %d, kind %q, folder %q", e.Module.Number, e.Kind, e.Dir())
	}
	if e, _ := c.Find("vars"); e.Kind != Concept {
		t.Errorf("default kind = %q, want concept", e.Kind)
	}
}

func TestLoadRejectsDuplicateNames(t *testing.T) {
	_, err := Load(fstest.MapFS{"course.json": {Data: []byte(`{"modules":[{"dir":"m","exercises":[{"name":"a"},{"name":"a"}]}]}`)}})
	if err == nil {
		t.Fatal("Load accepted two exercises named a")
	}
}

const twoTaskReadme = "# X\n\n## Your task\n\n### 1. Name a pod\n\ntext\n\n### 2. Build a target\n"

func TestParseHintsGroupsThreeLevelsPerTask(t *testing.T) {
	tasks, err := ParseHints(twoTaskReadme, "## Task 1: Name a pod\n\nLook at the dashes.\n\n### Closer\n\nUse +.\n\n### Answer\n\n```go\n## not a heading inside code\nreturn a + \"-\" + b\n```\n\n"+
		"## Task 2: Build a target\n\nnudge\n\n### Closer\n\ncloser\n\n### Answer\n\nanswer\n")
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 2 || tasks[0].Title != "Name a pod" || len(tasks[0].Hints) != 3 {
		t.Fatalf("tasks = %+v, want 2 tasks with 3 hints each", tasks)
	}
	if got := tasks[0].Hints[2].Body; got != "```go\n## not a heading inside code\nreturn a + \"-\" + b\n```" {
		t.Errorf("Task 1 Answer = %q; code blocks must stay whole", got)
	}
}

func TestParseHintsRejectsMismatchedOrIncompleteHints(t *testing.T) {
	for name, hints := range map[string]string{
		"missing task":  "## Task 1: Name a pod\nn\n### Closer\nc\n### Answer\na\n",
		"wrong title":   "## Task 1: Pods\nn\n### Closer\nc\n### Answer\na\n## Task 2: Build a target\nn\n### Closer\nc\n### Answer\na\n",
		"missing level": "## Task 1: Name a pod\nn\n### Answer\na\n## Task 2: Build a target\nn\n### Closer\nc\n### Answer\na\n",
		"old format":    "## Where to look\ntext\n",
	} {
		if _, err := ParseHints(twoTaskReadme, hints); err == nil {
			t.Errorf("%s: ParseHints accepted it", name)
		}
	}
}
