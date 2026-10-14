package tui

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// Documents (instructions, hints, guides, the solution, keys) are shown in a
// frame that fills the window's width, with the title in its top edge. The
// frame stays put and only the text inside scrolls, like a pane in lazygit
// or k9s. Text keeps a comfortable reading width on the left; on a wide
// window the space to its right holds an "On this page" outline that follows
// the scroll position.

const (
	pagePad       = 2  // spaces between the frame and the text
	outlineMin    = 24 // narrowest useful outline column
	outlineMax    = 44
	frameRowsUsed = 2 // the frame's top and bottom edges
)

// docLayout is how a framed document divides the window.
type docLayout struct {
	width   int // whole frame
	text    int // text column
	outline int // outline column, 0 when there's no room or no outline
}

// layoutDoc splits a frame w columns wide. With an outline, the text keeps
// the reading width and the outline takes the rest.
func layoutDoc(w int, withOutline bool) docLayout {
	inner := w - 2 - 2*pagePad
	if withOutline && inner >= readingWidth+3+outlineMin {
		ow := min(outlineMax, inner-readingWidth-3)
		return docLayout{width: w, text: inner - 3 - ow, outline: ow}
	}
	// No outline: the frame hugs the reading width, left-aligned.
	text := max(20, min(inner, readingWidth))
	return docLayout{width: text + 2 + 2*pagePad, text: text}
}

// outlineItem is one heading of a document and the line it rendered on.
type outlineItem struct {
	title string
	level int // 2 for ## sections, 3 for ### subsections and tasks
	line  int
}

var (
	mdLink   = regexp.MustCompile(`\[([^\]]*)\]\([^)]*\)`)
	mdInline = strings.NewReplacer("`", "", "**", "", "*", "", "_", "")
)

// findOutline lists src's ## and ### headings (outside code fences) and the
// lines of rendered they appear on, as the theme renders them: sections in
// capitals, subsections after "› ".
func findOutline(src, rendered string) []outlineItem {
	lines := strings.Split(ansi.Strip(rendered), "\n")
	for i := range lines {
		lines[i] = strings.TrimSpace(lines[i])
	}
	var items []outlineItem
	next, fence := 0, false
	for _, l := range strings.Split(src, "\n") {
		if strings.HasPrefix(strings.TrimSpace(l), "```") {
			fence = !fence
		}
		level := 0
		switch {
		case fence:
		case strings.HasPrefix(l, "## "):
			level = 2
		case strings.HasPrefix(l, "### "):
			level = 3
		}
		if level == 0 {
			continue
		}
		title := strings.TrimSpace(mdInline.Replace(mdLink.ReplaceAllString(l[level+1:], "$1")))
		want := "› " + title
		if level == 2 {
			want = strings.ToUpper(title)
		}
		for i := next; i < len(lines); i++ {
			if lines[i] == want {
				items = append(items, outlineItem{title: title, level: level, line: i})
				next = i + 1
				break
			}
		}
	}
	return items
}

// current is the index of the section shown at the top of the view.
func currentSection(items []outlineItem, top int) int {
	cur := 0
	for i, it := range items {
		if it.line <= top+1 {
			cur = i
		}
	}
	return cur
}

// frame draws the document frame around the visible text lines, with the
// outline column (if the layout has one).
func frame(l docLayout, title string, text []string, items []outlineItem, top int) string {
	edge := func(s string) string { return rule.Render(s) }
	head := " " + bold.Render(ansi.Truncate(title, max(10, l.text-4), "…")) + " "
	textSpan := pagePad + l.text + pagePad
	var b strings.Builder

	// Top edge: ╭─ Title ─────┬─ On this page ─╮
	b.WriteString(edge("╭─") + head)
	if l.outline > 0 {
		b.WriteString(edge(strings.Repeat("─", max(0, textSpan-1-lipgloss.Width(head))) + "┬─"))
		oh := " " + dim.Render("On this page") + " "
		b.WriteString(oh + edge(strings.Repeat("─", max(0, l.outline+1-lipgloss.Width(oh)))+"╮"))
	} else {
		b.WriteString(edge(strings.Repeat("─", max(0, textSpan-1-lipgloss.Width(head))) + "╮"))
	}
	b.WriteString("\n")

	outline := outlineLines(items, l.outline, currentSection(items, top), len(text))
	side, pad := edge("│"), strings.Repeat(" ", pagePad)
	for i, t := range text {
		t = ansi.Truncate(strings.ReplaceAll(t, "\t", "    "), l.text, "")
		b.WriteString(side + pad + t + strings.Repeat(" ", max(0, l.text-lipgloss.Width(t))) + pad)
		if l.outline > 0 {
			o := outline[i]
			b.WriteString(side + " " + o + strings.Repeat(" ", max(0, l.outline-lipgloss.Width(o))) + " ")
		}
		b.WriteString(side + "\n")
	}

	// Bottom edge.
	b.WriteString(edge("╰" + strings.Repeat("─", textSpan)))
	if l.outline > 0 {
		b.WriteString(edge("┴" + strings.Repeat("─", l.outline+2)))
	}
	b.WriteString(edge("╯"))
	return b.String()
}

// outlineLines renders the outline into rows lines of width w, keeping the
// current section in view.
func outlineLines(items []outlineItem, w, cur, rows int) []string {
	out := make([]string, rows)
	if w == 0 || len(items) == 0 {
		return out
	}
	var all []string
	for i, it := range items {
		label := it.title
		indent := ""
		if it.level == 3 {
			indent = "  "
		}
		label = ansi.Truncate(label, w-2-len(indent), "…")
		switch {
		case i == cur:
			all = append(all, keyCap.Render("▸ ")+indent+bold.Render(label))
		case it.level == 2:
			all = append(all, "  "+indent+label)
		default:
			all = append(all, "  "+indent+dim.Render(label))
		}
	}
	first := 0
	if cur >= rows-1 {
		first = cur - rows/2
	}
	first = max(0, min(first, len(all)-rows))
	for i := range rows {
		if first+i < len(all) {
			out[i] = all[first+i]
		}
	}
	return out
}

// page frames a whole document for printing (groundwork guides): the same
// frame, left-aligned, without the outline.
func page(title, body string, w int) string {
	l := layoutDoc(w, false)
	lines := strings.Split("\n"+strings.Trim(body, "\n")+"\n", "\n")
	var wrapped []string
	for _, line := range lines {
		wrapped = append(wrapped, strings.Split(ansi.Hardwrap(strings.TrimRight(line, " "), l.text, true), "\n")...)
	}
	return frame(l, title, wrapped, nil, 0)
}

// sectionRules puts a horizontal rule before every "## " section after the
// first, so long documents read as clearly separated sections. Headings
// inside code fences are left alone.
func sectionRules(md string) string {
	var out []string
	fence, seen := false, false
	for _, line := range strings.Split(md, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			fence = !fence
		}
		if !fence && strings.HasPrefix(line, "## ") {
			if seen {
				out = append(out, "", "---", "")
			}
			seen = true
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}
