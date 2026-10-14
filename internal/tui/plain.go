package tui

import (
	"fmt"
	"strings"

	"github.com/pjanderson/groundwork/internal/course"
	"github.com/pjanderson/groundwork/internal/session"
	"github.com/pjanderson/groundwork/internal/workspace"
)

const plainWidth = 100

// CheckReport renders a one-shot check for `groundwork check`.
func CheckReport(s *session.Session, e *course.Exercise, r *workspace.Result) string {
	var b strings.Builder
	b.WriteString(bold.Render(fmt.Sprintf("%03d  %s", e.Number, e.Title)) + dim.Render("  ·  "+e.Module.Title) + "\n\n")
	b.WriteString(RenderResult(r, plainWidth) + "\n")
	if r.Status == workspace.Passed {
		b.WriteString("\n" + okBold.Render("★ Exercise complete!") + " Reference solution: " + bold.Render(relative(s.SolutionDir(e))) + "\n")
		if next := s.NextPending(e); next != nil {
			b.WriteString(dim.Render(fmt.Sprintf("Next up: %03d %s", next.Number, next.Title)) + "\n")
		}
	}
	return b.String()
}

// CourseList renders `groundwork list`: every module with its progress,
// expanding the current module (or all modules with all set).
func CourseList(s *session.Session, all bool) string {
	var b strings.Builder
	done, total := s.DoneCount(), len(s.Course.Exercises)
	b.WriteString(brand.Render("groundwork") + "  " + progressBar(done, total, 30) + dim.Render(fmt.Sprintf("  %d/%d exercises done", done, total)) + "\n")
	if due := s.DueReviews(); len(due) > 0 {
		b.WriteString(warn.Render(fmt.Sprintf("↻ %d %s due", len(due), plural(len(due), "review", "reviews"))) + dim.Render(": run groundwork and press ") + keyCap.Render("v") + dim.Render(fmt.Sprintf(" (first: %03d %s)", due[0].Number, due[0].Title)) + "\n")
	}
	current := s.Current()
	phase := ""
	for _, mod := range s.Course.Modules {
		if mod.Phase != phase {
			phase = mod.Phase
			b.WriteString("\n" + dim.Render(strings.ToUpper(phase)) + "\n")
		}
		modDone := 0
		for _, e := range mod.Exercises {
			if s.Done(e) {
				modDone++
			}
		}
		marker := "  "
		if modDone == len(mod.Exercises) {
			marker = okStyle.Render("✓ ")
		} else if current.Module == mod {
			marker = keyCap.Render("● ")
		}
		head := fmt.Sprintf("%s%2d  %-36s", marker, mod.Number, mod.Title)
		b.WriteString(head + " " + progressBar(modDone, len(mod.Exercises), 12) + dim.Render(fmt.Sprintf(" %2d/%-2d", modDone, len(mod.Exercises))) + "\n")
		if !all && current.Module != mod {
			continue
		}
		for _, e := range mod.Exercises {
			icon := dim.Render("·")
			switch {
			case s.Done(e):
				icon = okStyle.Render("✓")
			case s.Revealed(e):
				icon = warn.Render("◐")
			case e == current:
				icon = keyCap.Render("▸")
			}
			label := fmt.Sprintf("%03d  %s", e.Number, e.Title)
			if e == current {
				label = bold.Render(label)
			}
			b.WriteString("        " + icon + "  " + label + "\n")
		}
	}
	b.WriteString("\n" + dim.Render("✓ done   ◐ solution seen before passing   ▸ current") + "\n")
	if !all {
		b.WriteString(dim.Render("groundwork list --all shows every exercise.") + "\n")
	}
	return b.String()
}

// GuideList renders `groundwork guides`: every guide by module, ✓ for the
// ones read, with summaries for the guides of current (the learner's module).
// Outside a course, current and read are nil.
func GuideList(c *course.Course, current *course.Module, read func(*course.Guide) bool) string {
	var b strings.Builder
	b.WriteString(brand.Render("groundwork guides") + dim.Render("  ·  the computer science underneath the exercises") + "\n")
	var mod *course.Module
	for _, g := range c.Guides {
		if g.Module != mod {
			mod = g.Module
			head := "  " + dim.Render(fmt.Sprintf("%2d  %s", mod.Number, mod.Title))
			if mod == current {
				head = keyCap.Render("● ") + bold.Render(fmt.Sprintf("%2d  %s", mod.Number, mod.Title)) + dim.Render("  · your module")
			}
			b.WriteString("\n" + head + "\n")
		}
		icon := " "
		if read != nil && read(g) {
			icon = okStyle.Render("✓")
		}
		label := fmt.Sprintf("%2d  %s", g.Number, g.Title)
		if mod != current {
			b.WriteString("     " + icon + "  " + label + "\n")
			continue
		}
		b.WriteString("     " + icon + "  " + bold.Render(label) + "\n")
		if g.Summary != "" {
			b.WriteString(wrapIndent(dim.Render(g.Summary), 12, plainWidth) + "\n")
		}
	}
	b.WriteString("\n" + keyCap.Render("groundwork guides 10") + dim.Render(" reads guide 10 · in the watcher, press ") + keyCap.Render("g") + "\n")
	return b.String()
}

// Guide renders one guide's Markdown for the terminal.
func Guide(text string) string {
	r, err := newMarkdown(readingWidth)
	if err != nil {
		return text
	}
	out, err := r.Render(text)
	if err != nil {
		return text
	}
	return out
}

// Notice styles a short informational line for the CLI.
func Notice(icon, text string) string {
	switch icon {
	case "ok":
		return okBold.Render("✓ ") + text
	case "error":
		return badBold.Render("✗ ") + text
	}
	return keyCap.Render("› ") + text
}

// Key styles a command or key name.
func Key(s string) string { return keyCap.Render(s) }

// Dim styles secondary text.
func Dim(s string) string { return dim.Render(s) }
