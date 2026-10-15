package tui

import (
	"strings"
	"testing"
)

func TestLinksBecomeClickableTextThatSurvivesWrapping(t *testing.T) {
	md := "- [Rubber Duck Debugging - How to Solve a Problem](https://www.youtube.com/watch?v=NBgIHOrjSxs) · 6 min\n" +
		"- [State](https://en.wikipedia.org/wiki/State_(computer_science)) and `[not](https://a.link)`\n" +
		"```\n[in code](https://example.com)\n```\n"
	marked, urls := markLinks(md)
	want := []string{"https://www.youtube.com/watch?v=NBgIHOrjSxs", "https://en.wikipedia.org/wiki/State_(computer_science)"}
	if strings.Join(urls, " ") != strings.Join(want, " ") {
		t.Fatalf("urls = %q, want %q (links in code stay literal)", urls, want)
	}
	if strings.Contains(marked, "youtube") || !strings.Contains(marked, "[in code](https://example.com)") {
		t.Fatalf("marked text:\n%s", marked)
	}

	r, err := newMarkdown(30) // narrow, so the first link's text wraps
	if err != nil {
		t.Fatal(err)
	}
	out, err := renderMarkdown(r, md)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(out, "\n")
	opened := 0
	for _, l := range lines {
		starts := strings.Count(l, "\x1b]8;;https")
		ends := strings.Count(l, "\x1b]8;;\x1b\\")
		if starts != ends {
			t.Errorf("line opens %d links but closes %d; every line must be complete: %q", starts, ends, l)
		}
		if strings.Contains(l, "\x1b]8;;https://www.youtube.com/watch?v=NBgIHOrjSxs") {
			opened++
		}
	}
	if opened < 2 {
		t.Errorf("the wrapped link should be clickable on each of its lines, found it on %d:\n%s", opened, out)
	}
	if strings.Contains(stripLinks(out), "youtube.com") {
		t.Errorf("the URL is printed as text:\n%s", out)
	}
}

// stripLinks removes OSC 8 sequences, leaving the visible text.
func stripLinks(s string) string {
	for {
		i := strings.Index(s, "\x1b]8;;")
		if i < 0 {
			return s
		}
		j := strings.Index(s[i:], "\x1b\\")
		s = s[:i] + s[i+j+2:]
	}
}
