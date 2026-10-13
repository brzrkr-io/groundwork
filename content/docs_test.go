package content_test

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/pjanderson/gopherlings/content"
	"github.com/pjanderson/gopherlings/internal/course"
)

var update = flag.Bool("update", false, "rewrite the generated parts of LESSON_PLAN.md")

const lessonPlan = "../LESSON_PLAN.md"

// TestLessonPlan keeps the syllabus tables in LESSON_PLAN.md in step with
// course.json. After changing the course, regenerate them with:
//
//	go test ./content -run TestLessonPlan -args -update
func TestLessonPlan(t *testing.T) {
	c, err := course.Load(content.Files)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(lessonPlan)
	if err != nil {
		t.Fatal(err)
	}
	doc := string(raw)
	doc = replaceSection(t, doc, "summary", planSummary(c))
	doc = replaceSection(t, doc, "modules", planModules(c))
	if doc == string(raw) {
		return
	}
	if *update {
		if err := os.WriteFile(lessonPlan, []byte(doc), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	t.Error("LESSON_PLAN.md is out of date with course.json; run: go test ./content -run TestLessonPlan -args -update")
}

// replaceSection swaps the text between <!-- begin NAME --> and
// <!-- end NAME --> markers.
func replaceSection(t *testing.T, doc, name, body string) string {
	t.Helper()
	begin, end := "<!-- begin "+name+" -->\n", "<!-- end "+name+" -->"
	i, j := strings.Index(doc, begin), strings.Index(doc, end)
	if i < 0 || j < i {
		t.Fatalf("LESSON_PLAN.md lacks the %q and %q markers", strings.TrimSpace(begin), end)
	}
	return doc[:i+len(begin)] + body + doc[j:]
}

func minutes(es []*course.Exercise) int {
	total := 0
	for _, e := range es {
		total += e.Minutes
	}
	return total
}

func planSummary(c *course.Course) string {
	return fmt.Sprintf("%d exercises in %d modules, about %d hours of hands-on work.\n",
		len(c.Exercises), len(c.Modules), (minutes(c.Exercises)+30)/60)
}

func planModules(c *course.Course) string {
	var b strings.Builder
	phase := ""
	for _, m := range c.Modules {
		if m.Phase != phase {
			phase = m.Phase
			var inPhase []*course.Exercise
			for _, other := range c.Modules {
				if other.Phase == phase {
					inPhase = append(inPhase, other.Exercises...)
				}
			}
			fmt.Fprintf(&b, "## Phase: %s (about %d hours)\n\n", phase, (minutes(inPhase)+30)/60)
		}
		fmt.Fprintf(&b, "### Module %d: %s\n\n", m.Number, m.Title)
		fmt.Fprintf(&b, "*About %.1f hours, in `exercises/%s/`.* %s\n\n", float64(minutes(m.Exercises))/60, m.Dir(), m.Summary)
		fmt.Fprintf(&b, "**After this module you can:** %s\n\n", m.Outcome)
		if len(m.Guides) > 0 {
			var links []string
			for _, g := range m.Guides {
				links = append(links, fmt.Sprintf("[%s](content/guides/%s.md)", g.Title, g.Name))
			}
			fmt.Fprintf(&b, "**Guides:** %s\n\n", strings.Join(links, " · "))
		}
		b.WriteString("| # | Exercise | Kind | Min |\n|---:|---|---|---:|\n")
		for _, e := range m.Exercises {
			kind := string(e.Kind)
			if e.Kind != course.Concept {
				kind = "**" + kind + "**"
			}
			if e.Race {
				kind += " · race"
			}
			fmt.Fprintf(&b, "| %03d | %s | %s | %d |\n", e.Number, e.Title, kind, e.Minutes)
		}
		b.WriteString("\n")
	}
	return b.String()
}

// TestCrossReferences checks that READMEs, hints and guides which point to
// another exercise ("as in *Slices*", "from *Wrapping errors*") name one that
// exists, by its title, its concept (the part before ":"), its module or a
// guide's title, or name a section heading of the file's own README or guide.
func TestCrossReferences(t *testing.T) {
	c, err := course.Load(content.Files)
	if err != nil {
		t.Fatal(err)
	}
	known := map[string]bool{}
	for _, m := range c.Modules {
		known[m.Title] = true
	}
	for _, e := range c.Exercises {
		known[e.Title] = true
		concept, _, _ := strings.Cut(e.Title, ":")
		known[concept] = true
	}
	for _, g := range c.Guides {
		known[g.Title] = true
	}
	codeBlock := regexp.MustCompile("(?s)```.*?```")
	reference := regexp.MustCompile(`\b(?:in|from|see|See|like|to|after|since|of|In|From|Like|After)\s+\*([A-Z][^*\n]{1,80}?)\*`)
	heading := regexp.MustCompile(`(?m)^#{2,4} (.+)$`)
	for _, e := range c.Exercises {
		readme, _ := content.Files.ReadFile("exercises/" + e.Name + "/README.md")
		sections := map[string]bool{}
		for _, h := range heading.FindAllStringSubmatch(string(readme), -1) {
			sections[strings.TrimSpace(h[1])] = true
		}
		for _, file := range []string{"README.md", "hints.md"} {
			raw, err := content.Files.ReadFile("exercises/" + e.Name + "/" + file)
			if err != nil {
				continue
			}
			text := codeBlock.ReplaceAllString(string(raw), "")
			for _, m := range reference.FindAllStringSubmatch(text, -1) {
				if !known[m[1]] && !sections[m[1]] {
					t.Errorf("%s/%s refers to *%s*, which is no exercise, concept, module, guide or README section", e.Name, file, m[1])
				}
			}
		}
	}
	for _, g := range c.Guides {
		raw, err := c.GuideText(g)
		if err != nil {
			t.Fatal(err)
		}
		text := codeBlock.ReplaceAllString(raw, "")
		sections := map[string]bool{}
		for _, h := range heading.FindAllStringSubmatch(text, -1) {
			sections[strings.TrimSpace(h[1])] = true
		}
		for _, m := range reference.FindAllStringSubmatch(text, -1) {
			if !known[m[1]] && !sections[m[1]] {
				t.Errorf("guides/%s.md refers to *%s*, which is no exercise, concept, module, guide or section of the guide", g.Name, m[1])
			}
		}
	}
}
