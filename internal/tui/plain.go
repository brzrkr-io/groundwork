package tui

import (
	"fmt"
	"strings"

	"github.com/pjanderson/gopherlings/internal/course"
	"github.com/pjanderson/gopherlings/internal/session"
	"github.com/pjanderson/gopherlings/internal/workspace"
)

const plainWidth = 100

// CheckReport renders a one-shot check for `gopherlings check`.
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

// CourseList renders `gopherlings list`: every module with its progress,
// expanding the current module (or all modules with all set).
func CourseList(s *session.Session, all bool) string {
	var b strings.Builder
	done, total := s.DoneCount(), len(s.Course.Exercises)
	b.WriteString(brand.Render("gopherlings") + "  " + progressBar(done, total, 30) + dim.Render(fmt.Sprintf("  %d/%d exercises done", done, total)) + "\n")
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
		b.WriteString(dim.Render("gopherlings list --all shows every exercise.") + "\n")
	}
	return b.String()
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
