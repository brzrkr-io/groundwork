package course

import (
	"strings"
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

func TestGuidesAreNumberedInModuleOrder(t *testing.T) {
	files := fstest.MapFS{
		"course.json": {Data: []byte(`{"modules":[
			{"dir":"one","guides":["state","debugging"],"exercises":[{"name":"a"}]},
			{"dir":"two","exercises":[{"name":"b"}]},
			{"dir":"three","guides":["concurrency"],"exercises":[{"name":"c"}]}]}`)},
		"guides/state.md":       {Data: []byte("# State\n\n**In one sentence:** state is what\nchanges.\n\n**Why it matters:** ...\n")},
		"guides/debugging.md":   {Data: []byte("# Debugging\n")},
		"guides/concurrency.md": {Data: []byte("# Concurrency\n")},
	}
	c, err := Load(files)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, g := range c.Guides {
		got = append(got, g.File()+" "+g.Module.Folder)
	}
	if want := "01_state.md one|02_debugging.md one|03_concurrency.md three"; strings.Join(got, "|") != want {
		t.Errorf("guides = %q, want %q", strings.Join(got, "|"), want)
	}
	if g := c.Guides[0]; g.Title != "State" || g.Summary != "state is what changes." {
		t.Errorf("title %q, summary %q", g.Title, g.Summary)
	}
	if len(c.Modules[1].Guides) != 0 || len(c.Modules[2].Guides) != 1 {
		t.Errorf("module guides = %d, %d; want 0, 1", len(c.Modules[1].Guides), len(c.Modules[2].Guides))
	}
	for _, query := range []string{"3", "03", "concurrency", "03_concurrency", "guides/03_concurrency.md", " Concurrency "} {
		if g, ok := c.FindGuide(query); !ok || g.Name != "concurrency" {
			t.Errorf("FindGuide(%q) = %v, %v; want concurrency", query, g, ok)
		}
	}
	for _, query := range []string{"0", "4", "02_state", "nope"} {
		if g, ok := c.FindGuide(query); ok {
			t.Errorf("FindGuide(%q) = %s, want no match", query, g.Name)
		}
	}
}

func TestLoadRejectsMissingDuplicateOrUntitledGuides(t *testing.T) {
	for name, files := range map[string]fstest.MapFS{
		"missing":   {"course.json": {Data: []byte(`{"modules":[{"dir":"m","guides":["gone"],"exercises":[{"name":"a"}]}]}`)}},
		"duplicate": {"course.json": {Data: []byte(`{"modules":[{"dir":"m","guides":["g"],"exercises":[{"name":"a"}]},{"dir":"n","guides":["g"],"exercises":[{"name":"b"}]}]}`)}, "guides/g.md": {Data: []byte("# G\n")}},
		"untitled":  {"course.json": {Data: []byte(`{"modules":[{"dir":"m","guides":["g"],"exercises":[{"name":"a"}]}]}`)}, "guides/g.md": {Data: []byte("G\n")}},
	} {
		if _, err := Load(files); err == nil {
			t.Errorf("%s guide: Load accepted it", name)
		}
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
