package content_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/pjanderson/groundwork/content"
	"github.com/pjanderson/groundwork/internal/cards"
	"github.com/pjanderson/groundwork/internal/course"
)

// guideEnd are the closing sections every guide has, in this order, after
// its teaching sections. See AUTHORING.md.
var guideEnd = []string{"Common mistakes", "Check yourself", "Answers", "Key terms", "Watch", "Go deeper"}

// TestGuides checks every guide follows the layout in AUTHORING.md, so the
// learner always finds the summary at the top and the same closing sections.
func TestGuides(t *testing.T) {
	c, err := course.Load(content.Files)
	if err != nil {
		t.Fatal(err)
	}
	files, err := content.Files.ReadDir("guides")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != len(c.Guides) {
		t.Errorf("guides/ has %d files but course.json lists %d guides; every guide belongs to a module", len(files), len(c.Guides))
	}
	codeBlock := regexp.MustCompile("(?s)```.*?```")
	heading := regexp.MustCompile(`(?m)^(#{1,6}) (.+)$`)
	numbered := regexp.MustCompile(`(?m)^\d+\. `)
	video := regexp.MustCompile(`(?m)^- \[[^\]]+\]\(https://[^)\s]+\) · [^·\n]+ · \d+ min: \S`)
	link := regexp.MustCompile(`\]\((https?://[^)\s]+)\)`)
	for _, g := range c.Guides {
		t.Run(g.Name, func(t *testing.T) {
			raw, err := c.GuideText(g)
			if err != nil {
				t.Fatal(err)
			}
			text := codeBlock.ReplaceAllString(raw, "")
			lines := strings.Split(raw, "\n")
			if len(lines) < 3 || lines[1] != "" || !strings.HasPrefix(lines[2], "**In one sentence:** ") {
				t.Error(`the line after the title must be "**In one sentence:** ..."`)
			}
			if !strings.Contains(text, "\n**Why it matters:** ") {
				t.Error(`missing the "**Why it matters:** ..." paragraph`)
			}
			if !strings.Contains(text, "\n**You'll learn:**\n") {
				t.Error(`missing the "**You'll learn:**" list`)
			}
			var h2 []string
			for _, h := range heading.FindAllStringSubmatch(text, -1) {
				switch level := len(h[1]); {
				case level == 1 && h[2] != g.Title:
					t.Errorf("second title %q: a guide has one # heading", h[2])
				case level > 3:
					t.Errorf("heading %q is level %d; guides use ## and ### only", h[2], level)
				case level == 2:
					h2 = append(h2, h[2])
				}
			}
			if len(h2) < len(guideEnd)+2 {
				t.Fatalf("only %d ## sections; a guide teaches in at least two before %v", len(h2), guideEnd)
			}
			if got := h2[len(h2)-len(guideEnd):]; strings.Join(got, "|") != strings.Join(guideEnd, "|") {
				t.Errorf("the last ## sections are %q, want %q", got, guideEnd)
			}
			videos := video.FindAllString(section(text, "Watch"), -1)
			if items := strings.Count(section(text, "Watch"), "\n- "); len(videos) < 2 || len(videos) > 5 || len(videos) != items {
				t.Errorf("Watch has %d well-formed videos in %d items; want 2 to 5 items, each `- [Title](https://...) · Channel or speaker · N min: why`", len(videos), items)
			}
			if links := link.FindAllString(section(text, "Go deeper"), -1); len(links) < 2 {
				t.Errorf("Go deeper has %d links, want at least 2", len(links))
			}
			if deck, err := cards.FromGuide(g, raw); err != nil || len(deck) < 6 {
				t.Errorf("flashcards: %d cards, error %v; want at least 3 questions and 3 key terms", len(deck), err)
			}
			questions := len(numbered.FindAllString(section(text, "Check yourself"), -1))
			answers := len(numbered.FindAllString(section(text, "Answers"), -1))
			if questions < 3 || questions != answers {
				t.Error("Check yourself and Answers must be numbered lists of the same length, with at least three questions")
			}
		})
	}
}

// section returns the text of the "## name" section, up to the next ## heading.
func section(text, name string) string {
	_, rest, ok := strings.Cut(text, "\n## "+name+"\n")
	if !ok {
		return ""
	}
	if i := strings.Index(rest, "\n## "); i >= 0 {
		rest = rest[:i]
	}
	return "\n" + rest
}
