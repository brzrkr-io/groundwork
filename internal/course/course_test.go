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

func TestParseHintsSplitsOnLevelTwoHeadings(t *testing.T) {
	hints := ParseHints("intro ignored\n## First\nline one\n\n### not a split\n## Second\n```go\nx := 1\n```\n")
	if len(hints) != 2 {
		t.Fatalf("got %d hints, want 2: %+v", len(hints), hints)
	}
	if hints[0].Title != "First" || hints[0].Body != "line one\n\n### not a split" {
		t.Errorf("hint 1 = %+v", hints[0])
	}
	if hints[1].Title != "Second" || hints[1].Body != "```go\nx := 1\n```" {
		t.Errorf("hint 2 = %+v", hints[1])
	}
}
