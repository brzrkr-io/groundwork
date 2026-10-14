package cards

import (
	"strings"
	"testing"

	"github.com/pjanderson/groundwork/internal/course"
)

const guide = "# State\n\n**In one sentence:** state changes.\n\n" +
	"## Check yourself\n\n" +
	"1. What does `x := 1` print\n   after `x++`?\n" +
	"2. Is this a heading?\n\n   ```go\n   ## not a heading\n   fmt.Println(\"a\tb\")\n   ```\n\n" +
	"## Answers\n\n" +
	"1. **2**: `x++` adds one.\n" +
	"2. No, it's *inside* code. See [the spec](https://go.dev/ref/spec).\n\n" +
	"## Key terms\n\n| Term | Meaning |\n|---|---|\n| state | The values of all variables, like `<x>` |\n| trace table | A table of each step |\n\n" +
	"## Go deeper\n\n- [x](https://go.dev)\n"

func TestFromGuidePairsQuestionsWithAnswersAndTerms(t *testing.T) {
	g := &course.Guide{Name: "state", Title: "State", Number: 4, Module: &course.Module{Number: 2}}
	cards, err := FromGuide(g, guide)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, c := range cards {
		got = append(got, c.ID+" => "+c.Back)
	}
	want := []string{
		"groundwork-state-q1 => **2**: `x++` adds one.",
		"groundwork-state-q2 => No, it's *inside* code. See [the spec](https://go.dev/ref/spec).",
		"groundwork-state-term-state => The values of all variables, like `<x>`",
		"groundwork-state-term-trace-table => A table of each step",
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("cards:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
	if cards[0].Front != "What does `x := 1` print\n   after `x++`?" {
		t.Errorf("multi-line question = %q", cards[0].Front)
	}

	var b strings.Builder
	if err := WriteAnki(&b, cards); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSuffix(b.String(), "\n"), "\n")
	var rows []string
	for _, l := range lines {
		if !strings.HasPrefix(l, "#") {
			rows = append(rows, l)
		}
	}
	if len(rows) != 4 {
		t.Fatalf("%d rows, want one per card (a newline or tab inside a field would split rows):\n%s", len(rows), b.String())
	}
	for _, r := range rows {
		if n := strings.Count(r, "\t"); n != 4 {
			t.Errorf("row has %d tabs, want 4 (5 columns): %q", n, r)
		}
	}
	q2 := strings.Split(rows[1], "\t")
	if !strings.Contains(q2[0], "<pre><code>## not a heading<br>fmt.Println(&#34;a    b&#34;)</code></pre>") {
		t.Errorf("code block not kept literal and tab-free: %q", q2[0])
	}
	if want := `No, it&#39;s <i>inside</i> code. See <a href="https://go.dev/ref/spec">the spec</a>.`; !strings.HasPrefix(q2[1], want) {
		t.Errorf("answer HTML = %q, want prefix %q", q2[1], want)
	}
	if q2[2] != "Groundwork::04 State" || q2[3] != "groundwork guide::state module::02" {
		t.Errorf("deck %q, tags %q", q2[2], q2[3])
	}
	term := strings.Split(rows[2], "\t")
	if term[0] != "What is <b>state</b>?" || !strings.HasPrefix(term[1], "The values of all variables, like <code>&lt;x&gt;</code>") {
		t.Errorf("term card = %q / %q", term[0], term[1])
	}
}

func TestFromGuideRejectsUnmatchedAnswers(t *testing.T) {
	g := &course.Guide{Name: "g"}
	if _, err := FromGuide(g, "## Check yourself\n\n1. a?\n2. b?\n\n## Answers\n\n1. a\n"); err == nil {
		t.Error("two questions with one answer was accepted")
	}
}
