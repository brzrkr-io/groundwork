package tui

import (
	"strings"
	"testing"
	"testing/fstest"

	tea "github.com/charmbracelet/bubbletea"
)

func TestSectionKeysJumpThroughTheOutline(t *testing.T) {
	readme := "# A\n\n## Learn\n\n" + strings.Repeat("text\n\n", 30) + "### Recap\n\n" + strings.Repeat("more\n\n", 30) + "## Your task\n\n### 1. First\n\n" + strings.Repeat("do it\n\n", 30)
	files := fstest.MapFS{
		"course.json":                       {Data: []byte(`{"modules":[{"title":"m","dir":"one","exercises":[{"name":"a","title":"A"}]}]}`)},
		"packs/a/README.md":                 {Data: []byte(readme)},
		"packs/a/exercise.go.tmpl":          {Data: []byte("package x\n")},
		"packs/a/solution/exercise.go.tmpl": {Data: []byte("package x\n")},
	}
	m, _ := modelFor(t, files)
	m.Update(tea.WindowSizeMsg{Width: 140, Height: 40})
	press(m, "i")
	var titles []string
	for _, it := range m.outline {
		titles = append(titles, it.title)
	}
	if got := strings.Join(titles, "|"); got != "Learn|Recap|Your task|1. First" {
		t.Fatalf("outline = %q", got)
	}
	for _, want := range []string{"Recap", "Your task", "1. First"} {
		press(m, "]")
		if cur := m.outline[currentSection(m.outline, m.body.YOffset)].title; cur != want {
			t.Fatalf("after ], current section = %q (offset %d), want %q", cur, m.body.YOffset, want)
		}
	}
	press(m, "[")
	if cur := m.outline[currentSection(m.outline, m.body.YOffset)].title; cur != "Your task" {
		t.Errorf("after [, current section = %q, want Your task", cur)
	}
	if view := m.View(); !strings.Contains(view, "On this page") || !strings.Contains(view, "╰") {
		t.Errorf("framed view missing its frame or outline:\n%s", view)
	}
}
