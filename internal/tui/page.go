package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// Documents (instructions, hints, guides, the solution, keys) are shown as a
// page: a rounded frame with the title in its top edge, a comfortable
// reading width, centred in the window. The frame's sides run down every
// line, so text never floats loose as it scrolls.

const pagePad = 2 // spaces between the frame and the text

// pageWidth is the frame's total width for a window of width w.
func pageWidth(w int) int { return min(w, readingWidth+2+2*pagePad) }

// pageInner is the text width inside the frame.
func pageInner(w int) int { return max(20, pageWidth(w)-2-2*pagePad) }

// page frames body (already rendered to pageInner(w) columns) under title,
// with an optional dim meta line, centred in a window w columns wide.
func page(title, meta, body string, w int) string {
	pw, inner := pageWidth(w), pageInner(w)
	edge := func(s string) string { return rule.Render(s) }
	head := " " + bold.Render(ansi.Truncate(title, inner-2, "…")) + " "
	top := edge("╭─") + head + edge(strings.Repeat("─", max(0, pw-3-lipgloss.Width(head)))+"╮")

	var content []string
	if meta != "" {
		content = append(content, strings.Split(ansi.Wrap(meta, inner, " "), "\n")...)
		content = append(content, edge(strings.Repeat("┄", inner)))
	}
	content = append(content, "")
	content = append(content, strings.Split(strings.Trim(body, "\n"), "\n")...)
	content = append(content, "")

	lines := []string{top}
	side, pad := edge("│"), strings.Repeat(" ", pagePad)
	for _, l := range content {
		l = strings.TrimRight(strings.ReplaceAll(l, "\t", "    "), " ")
		for _, part := range strings.Split(ansi.Hardwrap(l, inner, true), "\n") {
			fill := strings.Repeat(" ", max(0, inner-lipgloss.Width(part)))
			lines = append(lines, side+pad+part+fill+pad+side)
		}
	}
	lines = append(lines, edge("╰"+strings.Repeat("─", pw-2)+"╯"))

	if left := (w - pw) / 2; left > 0 {
		indent := strings.Repeat(" ", left)
		for i := range lines {
			lines[i] = indent + lines[i]
		}
	}
	return strings.Join(lines, "\n")
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
