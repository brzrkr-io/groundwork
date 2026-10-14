// Package cards turns the guides' "Check yourself" questions and "Key terms"
// tables into flashcards, written as a text file that Anki (desktop or
// AnkiMobile) imports.
package cards

import (
	"fmt"
	"html"
	"io"
	"regexp"
	"strings"

	"github.com/pjanderson/groundwork/internal/course"
)

// Card is one flashcard. Front and Back are Markdown.
type Card struct {
	ID    string // stable across exports, so re-importing updates the card
	Front string
	Back  string
	Guide *course.Guide
}

var (
	numbered = regexp.MustCompile(`(?m)^(\d+)\. `)
	tableRow = regexp.MustCompile(`(?m)^\|(.+)\|\s*$`)
)

// FromGuide extracts g's cards from its Markdown text: one per question
// (answered by the matching item under "## Answers"), and one per Key terms
// row (the term on the front).
func FromGuide(g *course.Guide, text string) ([]Card, error) {
	questions := items(section(text, "Check yourself"))
	answers := items(section(text, "Answers"))
	if len(questions) != len(answers) {
		return nil, fmt.Errorf("guide %s has %d questions but %d answers", g.Name, len(questions), len(answers))
	}
	var cards []Card
	for i, q := range questions {
		cards = append(cards, Card{
			ID:    fmt.Sprintf("groundwork-%s-q%d", g.Name, i+1),
			Front: q,
			Back:  answers[i],
			Guide: g,
		})
	}
	for _, row := range tableRow.FindAllStringSubmatch(section(text, "Key terms"), -1) {
		cells := strings.Split(row[1], "|")
		if len(cells) < 2 {
			continue
		}
		term, meaning := strings.TrimSpace(cells[0]), strings.TrimSpace(strings.Join(cells[1:], "|"))
		if term == "Term" || strings.Trim(term, "-: ") == "" {
			continue // header and separator rows
		}
		cards = append(cards, Card{
			ID:    "groundwork-" + g.Name + "-term-" + slug(term),
			Front: "What is **" + strings.Trim(term, "*") + "**?",
			Back:  meaning,
			Guide: g,
		})
	}
	return cards, nil
}

// section returns the body of the "## name" section, up to the next ##
// heading, ignoring headings inside code blocks.
func section(text, name string) string {
	var b strings.Builder
	in, fence := false, false
	for line := range strings.SplitSeq(text, "\n") {
		if strings.HasPrefix(line, "```") {
			fence = !fence
		}
		if !fence && strings.HasPrefix(line, "## ") {
			in = strings.TrimSpace(line[3:]) == name
			continue
		}
		if in {
			b.WriteString(line + "\n")
		}
	}
	return b.String()
}

// items splits a numbered list into its items, each with its continuation
// lines, without the "1. " markers.
func items(list string) []string {
	starts := numbered.FindAllStringIndex(list, -1)
	var out []string
	for i, s := range starts {
		end := len(list)
		if i+1 < len(starts) {
			end = starts[i+1][0]
		}
		out = append(out, strings.TrimSpace(list[s[1]:end]))
	}
	return out
}

func slug(s string) string {
	var b strings.Builder
	dash := false
	for _, r := range strings.ToLower(s) {
		switch {
		case r >= 'a' && r <= 'z' || r >= '0' && r <= '9':
			b.WriteRune(r)
			dash = false
		case !dash && b.Len() > 0:
			b.WriteByte('-')
			dash = true
		}
	}
	return strings.TrimSuffix(b.String(), "-")
}

// WriteAnki writes cards as an Anki text import: tab-separated, HTML fields,
// one subdeck per guide, with headers that preset everything so importing
// needs no choices. Re-importing an updated export updates existing cards in
// place (matched by ID) and keeps their review history.
func WriteAnki(w io.Writer, cards []Card) error {
	var b strings.Builder
	b.WriteString("#separator:tab\n#html:true\n#notetype:Basic\n" +
		"#columns:Front\tBack\tDeck\tTags\tGUID\n" +
		"#deck column:3\n#tags column:4\n#guid column:5\n")
	for _, c := range cards {
		g := c.Guide
		source := fmt.Sprintf("<br><br><small>Guide %d: %s</small>", g.Number, html.EscapeString(g.Title))
		fields := []string{
			toHTML(c.Front),
			toHTML(c.Back) + source,
			fmt.Sprintf("Groundwork::%02d %s", g.Number, deckName(g.Title)),
			fmt.Sprintf("groundwork guide::%s module::%02d", g.Name, g.Module.Number),
			c.ID,
		}
		b.WriteString(strings.Join(fields, "\t") + "\n")
	}
	_, err := io.WriteString(w, b.String())
	return err
}

// deckName keeps "::", which nests decks in Anki, out of a title.
func deckName(title string) string { return strings.ReplaceAll(title, "::", ":") }

var (
	inlineCode = regexp.MustCompile("`([^`]+)`")
	bold       = regexp.MustCompile(`\*\*([^*]+)\*\*`)
	italic     = regexp.MustCompile(`\*([^*\s][^*]*)\*`)
	link       = regexp.MustCompile(`\[([^\]]+)\]\((https?://[^)\s]+)\)`)
)

// toHTML converts the Markdown found in questions, answers and key terms
// (paragraphs, fenced code, inline code, bold, italics, links) into one line
// of HTML: no tabs or newlines, which would break the file's rows.
func toHTML(md string) string {
	var out []string
	lines := strings.Split(strings.TrimSpace(md), "\n")
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			var code []string
			for i++; i < len(lines) && !strings.HasPrefix(strings.TrimSpace(lines[i]), "```"); i++ {
				code = append(code, html.EscapeString(strings.ReplaceAll(strings.TrimPrefix(lines[i], "   "), "\t", "    ")))
			}
			out = append(out, "<pre><code>"+strings.Join(code, "<br>")+"</code></pre>")
			continue
		}
		out = append(out, inline(strings.TrimSpace(line)))
	}
	text := strings.Join(out, "<br>")
	return strings.ReplaceAll(text, "<br><br><pre>", "<br><pre>")
}

// inline converts one line's inline Markdown. Code spans are swapped out
// first so their contents stay literal.
func inline(line string) string {
	var spans []string
	line = inlineCode.ReplaceAllStringFunc(line, func(m string) string {
		spans = append(spans, "<code>"+html.EscapeString(m[1:len(m)-1])+"</code>")
		return fmt.Sprintf("\x00%d\x00", len(spans)-1)
	})
	line = html.EscapeString(strings.ReplaceAll(line, "\t", " "))
	line = link.ReplaceAllString(line, `<a href="$2">$1</a>`)
	line = bold.ReplaceAllString(line, "<b>$1</b>")
	line = italic.ReplaceAllString(line, "<i>$1</i>")
	for i, s := range spans {
		line = strings.Replace(line, fmt.Sprintf("\x00%d\x00", i), s, 1)
	}
	return line
}
