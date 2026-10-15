package tui

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// Every view sits in the same frame: a rounded border with the title in its
// top edge. The frame stays put and only the text inside scrolls, like a pane
// in lazygit or k9s. On a wide window a sidebar to the right of the text
// holds context for the view: the outline of a document, the exercise's
// tasks next to the test results, stats next to the course list.

const (
	pagePad       = 2  // spaces between the frame and the text
	outlineMin    = 24 // narrowest useful outline column
	outlineMax    = 44
	frameRowsUsed = 2 // the frame's top and bottom edges
)

// docLayout is how a frame divides the window.
type docLayout struct {
	width   int // whole frame
	text    int // text column
	outline int // sidebar column, 0 when there's no room or no sidebar
}

// layoutWide splits a frame w columns wide for lists and test output: the
// text takes most of it, the sidebar about a third, between 30 and 44
// columns, when there's room for both.
func layoutWide(w int) docLayout {
	inner := w - 2 - 2*pagePad
	if inner < 60+3+30 {
		return docLayout{width: w, text: max(20, inner)}
	}
	side := min(44, max(30, inner/3))
	return docLayout{width: w, text: inner - 3 - side, outline: side}
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
// line of rendered each appears on. It matches headings as the theme renders
// them: sections in capitals, subsections after "› ".
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

// frame draws the frame around the visible text lines, with the sidebar
// lines (if the layout has a sidebar) under sideTitle.
func frame(l docLayout, title, sideTitle string, text, side []string) string {
	edge := func(s string) string { return rule.Render(s) }
	head := " " + bold.Render(ansi.Truncate(title, max(10, l.text-4), "…")) + " "
	textSpan := pagePad + l.text + pagePad
	var b strings.Builder

	// Top edge: ╭─ Title ─────┬─ On this page ─╮
	b.WriteString(edge("╭─") + head)
	if l.outline > 0 {
		b.WriteString(edge(strings.Repeat("─", max(0, textSpan-1-lipgloss.Width(head))) + "┬─"))
		oh := " " + dim.Render(sideTitle) + " "
		b.WriteString(oh + edge(strings.Repeat("─", max(0, l.outline+1-lipgloss.Width(oh)))+"╮"))
	} else {
		b.WriteString(edge(strings.Repeat("─", max(0, textSpan-1-lipgloss.Width(head))) + "╮"))
	}
	b.WriteString("\n")

	bar, pad := edge("│"), strings.Repeat(" ", pagePad)
	for i, t := range text {
		t = ansi.Truncate(strings.ReplaceAll(t, "\t", "    "), l.text, "")
		b.WriteString(bar + pad + t + strings.Repeat(" ", max(0, l.text-lipgloss.Width(t))) + pad)
		if l.outline > 0 {
			o := ""
			if i < len(side) {
				o = ansi.Truncate(side[i], l.outline, "…")
			}
			b.WriteString(bar + " " + o + strings.Repeat(" ", max(0, l.outline-lipgloss.Width(o))) + " ")
		}
		b.WriteString(bar + "\n")
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
	return frame(l, title, "", wrapped, nil)
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

// framed frames ready-made lines for the plain CLI (groundwork list, check,
// guides), w columns wide.
func framed(title, body string, w int) string {
	l := docLayout{width: w, text: w - 2 - 2*pagePad}
	lines := strings.Split("\n"+strings.Trim(body, "\n")+"\n", "\n")
	var wrapped []string
	for _, line := range lines {
		wrapped = append(wrapped, strings.Split(ansi.Hardwrap(strings.TrimRight(line, " "), l.text, true), "\n")...)
	}
	return frame(l, title, "", wrapped, nil) + "\n"
}

// sideSection renders a sidebar block: a dim heading, then its lines.
func sideSection(title string, lines ...string) []string {
	return append([]string{dim.Render(strings.ToUpper(title))}, lines...)
}

// wrapSide wraps text for a sidebar w columns wide.
func wrapSide(text string, w int) []string {
	return strings.Split(ansi.Wrap(text, w, " "), "\n")
}

// Links render as their text, underlined and clickable (an OSC 8
// hyperlink, which Ghostty, kitty, iTerm2, WezTerm and recent tmux
// understand), instead of the text followed by its URL: a long URL wraps
// across lines, and a wrapped URL can't be clicked.
const (
	linkOpen  = '\ue000' // marks where a link's text starts, before rendering
	linkClose = '\ue001'
)

// markLinks replaces [text](url) outside code with marked text and returns
// the URLs in order. Glamour then wraps the text like any other words.
func markLinks(md string) (string, []string) {
	var urls []string
	var out strings.Builder
	fence := false
	for i, line := range strings.Split(md, "\n") {
		if i > 0 {
			out.WriteByte('\n')
		}
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			fence = !fence
		}
		if fence {
			out.WriteString(line)
			continue
		}
		code := false
		for j := 0; j < len(line); j++ {
			c := line[j]
			if c == '`' {
				code = !code
			}
			if c != '[' || code {
				out.WriteByte(c)
				continue
			}
			text, url, n := parseLink(line[j:])
			if n == 0 {
				out.WriteByte(c)
				continue
			}
			urls = append(urls, url)
			out.WriteRune(linkOpen)
			out.WriteString(text)
			out.WriteRune(linkClose)
			j += n - 1
		}
	}
	return out.String(), urls
}

// parseLink reads "[text](url)" at the start of s, allowing balanced
// parentheses in the URL (Wikipedia titles). n is 0 when s isn't a link.
func parseLink(s string) (text, url string, n int) {
	end := strings.Index(s, "](")
	if end < 1 || strings.ContainsAny(s[1:end], "[]") {
		return "", "", 0
	}
	depth := 0
	for k := end + 2; k < len(s); k++ {
		switch s[k] {
		case '(':
			depth++
		case ')':
			if depth == 0 {
				url = s[end+2 : k]
				if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
					return "", "", 0
				}
				return s[1:end], url, k + 1
			}
			depth--
		case ' ':
			return "", "", 0
		}
	}
	return "", "", 0
}

// applyLinks turns the markers in rendered text into hyperlinks. Each line
// is complete on its own (a link wrapped onto two lines is two links to the
// same URL), because the frame and the scrolling cut text into lines.
func applyLinks(rendered string, urls []string) string {
	if len(urls) == 0 {
		return rendered
	}
	const style, unstyle = "\x1b[4;96m", "\x1b[24;39m"
	open := func(u string) string { return "\x1b]8;;" + u + "\x1b\\" + style }
	closeLink := unstyle + "\x1b]8;;\x1b\\"
	var out strings.Builder
	next, current := 0, ""
	for i, line := range strings.Split(rendered, "\n") {
		if i > 0 {
			out.WriteByte('\n')
		}
		if current != "" {
			out.WriteString(open(current))
		}
		for _, r := range line {
			switch r {
			case linkOpen:
				if next < len(urls) {
					current = urls[next]
					next++
				}
				out.WriteString(open(current))
			case linkClose:
				out.WriteString(closeLink)
				current = ""
			default:
				out.WriteRune(r)
			}
		}
		if current != "" {
			out.WriteString(closeLink)
		}
	}
	return out.String()
}

// renderMarkdown renders src with r, links as hyperlinks.
func renderMarkdown(r interface{ Render(string) (string, error) }, src string) (string, error) {
	marked, urls := markLinks(src)
	out, err := r.Render(marked)
	if err != nil {
		return "", err
	}
	return applyLinks(out, urls), nil
}
