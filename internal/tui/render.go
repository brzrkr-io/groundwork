package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/pjanderson/groundwork/internal/workspace"
)

// RenderResult formats a check result: a headline, then a ✓/✗ checklist
// with failure messages under each failing test.
func RenderResult(r *workspace.Result, width int) string {
	var b strings.Builder
	passed, failed := r.Counts()
	switch r.Status {
	case workspace.BuildFailed:
		b.WriteString(badBold.Render("✗ Doesn't compile yet") + "\n")
		b.WriteString(dim.Render("  Fix the first error below; later errors are often caused by it.") + "\n\n")
		for _, line := range strings.Split(r.Build, "\n") {
			b.WriteString(wrapIndent(highlightLocation(line), 2, width) + "\n")
		}
	case workspace.Failed:
		switch {
		case failed > 0:
			b.WriteString(badBold.Render(fmt.Sprintf("✗ %d of %d %s failing", failed, passed+failed, plural(passed+failed, "test", "tests"))) + "\n\n")
		default:
			b.WriteString(badBold.Render("✗ Not there yet") + "\n\n")
		}
		writeProblems(&b, r.Problems, width)
		writeTests(&b, r.Tests, width)
	case workspace.Passed:
		b.WriteString(okBold.Render(fmt.Sprintf("✓ All %d %s pass", passed, plural(passed, "test", "tests"))) + "\n\n")
		writeTests(&b, r.Tests, width)
	}
	for _, n := range r.Notes {
		b.WriteString("\n" + wrapIndent(warn.Render("! ")+n, 2, width) + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func writeProblems(b *strings.Builder, problems []string, width int) {
	for _, p := range problems {
		lines := strings.Split(p, "\n")
		b.WriteString(bad.Render("  ✗ ") + bold.Render(lines[0]) + "\n")
		for _, line := range lines[1:] {
			b.WriteString(wrapIndent(highlightLocation(line), 4, width) + "\n")
		}
		b.WriteString("\n")
	}
}

// writeTests lists failing tests first, with their messages, then the
// passing ones on a single line: the thing to fix is always at the top.
func writeTests(b *strings.Builder, tests []*workspace.Test, width int) {
	var passing []string
	for _, t := range tests {
		if t.Passed || t.Skipped {
			passing = append(passing, t.Name)
			continue
		}
		writeTest(b, t, 1, width)
	}
	if len(passing) == 0 {
		return
	}
	if len(passing) < len(tests) {
		b.WriteString("\n")
	}
	b.WriteString(wrapIndent(okStyle.Render("✓ ")+dim.Render(strings.Join(passing, ", ")), 2, width) + "\n")
}

func writeTest(b *strings.Builder, t *workspace.Test, depth, width int) {
	indent := strings.Repeat("  ", depth)
	name := t.Name
	if depth > 1 {
		name = strings.ReplaceAll(name, "_", " ") // go test turns spaces in case names into _
	}
	switch {
	case t.Passed:
		b.WriteString(indent + okStyle.Render("✓ ") + dim.Render(name) + "\n")
		return // hide passing subtests: they add noise
	case t.Skipped:
		b.WriteString(indent + dim.Render("– "+name+" (skipped)") + "\n")
		return
	}
	b.WriteString(indent + bad.Render("✗ ") + name + "\n")
	for _, line := range t.Output {
		b.WriteString(wrapIndent(highlightLocation(line), len(indent)+4, width) + "\n")
	}
	var hidden int
	for _, s := range t.Subtests {
		if s.Passed || s.Skipped {
			hidden++
			continue
		}
		writeTest(b, s, depth+1, width)
	}
	if hidden > 0 && hidden < len(t.Subtests) {
		b.WriteString(indent + "  " + dim.Render(fmt.Sprintf("✓ %d other %s", hidden, plural(hidden, "case passes", "cases pass"))) + "\n")
	}
}

// highlightLocation dims the "file.go:12:" prefix of compiler and test
// messages so the message itself stands out.
func highlightLocation(line string) string {
	trimmed := strings.TrimLeft(line, " \t")
	lead := line[:len(line)-len(trimmed)]
	if i := strings.Index(trimmed, ".go:"); i > 0 && !strings.ContainsAny(trimmed[:i], " \t") {
		rest := trimmed[i+4:]
		end := strings.Index(rest, ": ")
		if end >= 0 && strings.Trim(rest[:end], "0123456789:") == "" {
			return lead + dim.Render(trimmed[:i+4+end+1]) + trimmed[i+4+end+1:]
		}
	}
	return line
}

// wrapIndent indents text and wraps it to width so long assertion messages
// stay readable in narrow terminals.
func wrapIndent(text string, indent, width int) string {
	pad := strings.Repeat(" ", indent)
	if width <= indent+20 {
		return pad + text
	}
	wrapped := lipgloss.NewStyle().PaddingLeft(indent).Width(width).Render(text)
	lines := strings.Split(wrapped, "\n")
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, " ")
	}
	return strings.Join(lines, "\n")
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}
