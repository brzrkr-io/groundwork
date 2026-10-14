package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/pjanderson/groundwork/internal/course"
	"github.com/pjanderson/groundwork/internal/progress"
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
	if st := s.Stats(7); st.Streak > 0 || st.Week > 0 {
		b.WriteString(dim.Render(fmt.Sprintf("streak %d · %d of 7 days this week · ", st.Streak, st.Week)) + keyCap.Render("groundwork stats") + dim.Render(" for more") + "\n")
	}
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

// Brand styles the tool's name in a heading.
func Brand(s string) string { return brand.Render(s) }

// Key styles a command or key name.
func Key(s string) string { return keyCap.Render(s) }

// Dim styles secondary text.
func Dim(s string) string { return dim.Render(s) }

// statsWeeks is how many weeks the activity grid shows.
const statsWeeks = 12

// Stats renders the progress summary: streak, this week, today, totals and
// a grid of the last weeks, one column per week and one row per weekday.
func Stats(s *session.Session, width int) string {
	today := time.Now()
	offset := (int(today.Weekday()) + 6) % 7 // days since Monday
	st := s.Stats(7*(statsWeeks-1) + offset + 1)

	var b strings.Builder
	row := func(label, value string) {
		b.WriteString("  " + dim.Render(fmt.Sprintf("%-11s", label)) + " " + value + "\n")
	}
	streak := bold.Render(plural(st.Streak, "1 day", fmt.Sprintf("%d days", st.Streak)))
	if st.Longest > st.Streak {
		streak += dim.Render(fmt.Sprintf("  (longest %d)", st.Longest))
	}
	if st.Streak == 0 {
		streak = dim.Render("none yet: any practice today starts one")
	}
	row("Streak", streak)
	row("This week", fmt.Sprintf("%d of 7 days", st.Week))
	var todayParts []string
	for _, c := range []struct {
		n           int
		one, plural string
	}{
		{st.Today.Checks, "test run", "test runs"},
		{st.Today.Passed, "exercise passed", "exercises passed"},
		{st.Today.Reviews, "review", "reviews"},
		{st.Today.Guides, "guide read", "guides read"},
	} {
		if c.n > 0 {
			todayParts = append(todayParts, fmt.Sprintf("%d %s", c.n, plural(c.n, c.one, c.plural)))
		}
	}
	if len(todayParts) == 0 {
		todayParts = []string{dim.Render("nothing yet")}
	}
	row("Today", strings.Join(todayParts, dim.Render(" · ")))
	bar := min(24, max(8, width/5))
	row("Exercises", progressBar(st.Passed, st.Exercises, bar)+dim.Render(fmt.Sprintf("  %d/%d", st.Passed, st.Exercises)))
	row("Guides", progressBar(st.GuidesRead, st.Guides, bar)+dim.Render(fmt.Sprintf("  %d/%d read", st.GuidesRead, st.Guides)))
	reviews := fmt.Sprintf("%d passed", st.ReviewsPassed)
	if st.ReviewsDue > 0 {
		reviews = warn.Render(fmt.Sprintf("%d due", st.ReviewsDue)) + dim.Render(" · ") + reviews
	}
	row("Reviews", reviews)

	var grid strings.Builder
	grid.WriteString(dim.Render(fmt.Sprintf("Last %d weeks", statsWeeks)) + "\n")
	for weekday, name := range []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"} {
		grid.WriteString(dim.Render(name) + " ")
		for week := range statsWeeks {
			i := week*7 + weekday
			if i >= len(st.Days) {
				break
			}
			grid.WriteString(" " + cell(st.Days[i]))
		}
		grid.WriteString("\n")
	}
	grid.WriteString(dim.Render("·") + dim.Render(" none ") + cell(progress.Day{Checks: 1}) + dim.Render(" practised ") +
		cell(progress.Day{Passed: 1}) + dim.Render(" passed ") + cell(progress.Day{Passed: 3}) + dim.Render(" 3+"))

	summary := strings.TrimRight(b.String(), "\n")
	activity := strings.TrimRight(grid.String(), "\n")
	if width >= 90 { // side by side
		joined := lipgloss.JoinHorizontal(lipgloss.Top, lipgloss.NewStyle().Width(52).Render(summary), activity)
		lines := strings.Split(joined, "\n")
		for i, l := range lines {
			lines[i] = strings.TrimRight(l, " ")
		}
		return strings.Join(lines, "\n") + "\n"
	}
	return summary + "\n\n" + lipgloss.NewStyle().PaddingLeft(2).Render(activity) + "\n"
}

// cell is one day in the activity grid.
func cell(d progress.Day) string {
	switch wins := d.Passed + d.Reviews; {
	case wins >= 3:
		return okBold.Render("■")
	case wins > 0:
		return okStyle.Render("■")
	case d.Checks+d.Guides > 0:
		return keyCap.Render("▪")
	}
	return dim.Render("·")
}
