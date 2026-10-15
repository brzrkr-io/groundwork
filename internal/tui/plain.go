package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/brzrkr-io/groundwork/internal/course"
	"github.com/brzrkr-io/groundwork/internal/progress"
	"github.com/brzrkr-io/groundwork/internal/session"
	"github.com/brzrkr-io/groundwork/internal/workspace"
)

const plainWidth = 100

// CheckReport renders a one-shot check for `groundwork check`.
func CheckReport(s *session.Session, e *course.Exercise, r *workspace.Result) string {
	var b strings.Builder
	b.WriteString(dim.Render(fmt.Sprintf("Module %d · %s", e.Module.Number, e.Module.Title)) + "\n\n")
	b.WriteString(RenderResult(r, plainWidth-6) + "\n")
	if r.Status == workspace.Passed {
		b.WriteString("\n" + okBold.Render("★ Exercise complete!") + " Reference solution: " + bold.Render(relative(s.SolutionDir(e))) + "\n")
		if next := s.NextPending(e); next != nil {
			b.WriteString(dim.Render(fmt.Sprintf("Next up: %03d %s", next.Number, next.Title)) + "\n")
		}
	}
	return framed(fmt.Sprintf("%03d  %s", e.Number, e.Title), b.String(), plainWidth)
}

// CourseList renders `groundwork list`: every module with its progress,
// expanding the current module (or all modules with all set).
func CourseList(s *session.Session, all bool) string {
	var b strings.Builder
	done, total := s.DoneCount(), len(s.Course.Exercises)
	b.WriteString(progressBar(done, total, 30) + dim.Render(fmt.Sprintf("  %d/%d exercises done", done, total)) + "\n")
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
	return framed("Progress", b.String(), listWidth)
}

// listWidth is the frame width of groundwork list and groundwork guides.
const listWidth = 80

// GuideList renders `groundwork guides`: every guide by module, ✓ for the
// ones read, with summaries for the guides of current (the learner's module).
// Outside a course, current and read are nil.
func GuideList(c *course.Course, current *course.Module, read func(*course.Guide) bool) string {
	var b strings.Builder
	b.WriteString(dim.Render("The ideas underneath the exercises, one per guide.") + "\n")
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
			b.WriteString(wrapIndent(dim.Render(g.Summary), 12, listWidth-6) + "\n")
		}
	}
	b.WriteString("\n" + keyCap.Render("groundwork guides 10") + dim.Render(" reads guide 10 · in the watcher, press ") + keyCap.Render("g") + "\n")
	return framed("Guides", b.String(), listWidth)
}

// Guide renders one guide for the terminal, framed like the watcher's
// guide view.
func Guide(text string) string {
	w := readingWidth + 2 + 2*pagePad
	title, body := "", text
	if t, rest, ok := strings.Cut(text, "\n"); ok && strings.HasPrefix(t, "# ") {
		title, body = strings.TrimPrefix(t, "# "), rest
	}
	r, err := newMarkdown(layoutDoc(w, false).text)
	if err != nil {
		return text
	}
	out, err := renderMarkdown(r, sectionRules(body))
	if err != nil {
		return text
	}
	return page(title, out, w) + "\n"
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

// courseStats is the learner's stats with enough days for the grid.
func courseStats(s *session.Session) session.Stats {
	offset := (int(time.Now().Weekday()) + 6) % 7 // days since Monday
	return s.Stats(7*(statsWeeks-1) + offset + 1)
}

// statsRows are the summary rows (label, value), each fitting in width
// columns: streak, this week, today, totals.
func statsRows(st session.Stats, width int) []string {
	var rows []string
	row := func(label, value string) {
		rows = append(rows, dim.Render(fmt.Sprintf("%-10s", label))+" "+value)
	}
	streak := bold.Render(plural(st.Streak, "1 day", fmt.Sprintf("%d days", st.Streak)))
	if st.Longest > st.Streak {
		streak += dim.Render(fmt.Sprintf(" · best %d", st.Longest))
	}
	if st.Streak == 0 {
		streak = dim.Render("starts with any practice")
	}
	row("Streak", streak)
	row("This week", fmt.Sprintf("%d of 7 days", st.Week))
	var today []string
	for _, c := range []struct {
		n           int
		one, plural string
	}{
		{st.Today.Checks, "run", "runs"},
		{st.Today.Passed, "passed", "passed"},
		{st.Today.Reviews, "review", "reviews"},
		{st.Today.Guides, "guide", "guides"},
	} {
		if c.n > 0 {
			today = append(today, fmt.Sprintf("%d %s", c.n, plural(c.n, c.one, c.plural)))
		}
	}
	if len(today) == 0 {
		today = []string{dim.Render("nothing yet")}
	}
	row("Today", strings.Join(today, dim.Render(" · ")))
	bar := max(6, min(20, width-11-10))
	row("Exercises", progressBar(st.Passed, st.Exercises, bar)+dim.Render(fmt.Sprintf(" %d/%d", st.Passed, st.Exercises)))
	row("Guides", progressBar(st.GuidesRead, st.Guides, bar)+dim.Render(fmt.Sprintf(" %d/%d", st.GuidesRead, st.Guides)))
	reviews := fmt.Sprintf("%d passed", st.ReviewsPassed)
	if st.ReviewsDue > 0 {
		reviews = warn.Render(fmt.Sprintf("%d due", st.ReviewsDue)) + dim.Render(" · ") + reviews
	}
	row("Reviews", reviews)
	return rows
}

// statsGrid is the activity grid: one column per week, one row per weekday.
func statsGrid(st session.Stats) []string {
	var lines []string
	for weekday, name := range []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"} {
		var b strings.Builder
		b.WriteString(dim.Render(name))
		for week := range statsWeeks {
			i := week*7 + weekday
			if i >= len(st.Days) {
				break
			}
			b.WriteString(" " + cell(st.Days[i]))
		}
		lines = append(lines, b.String())
	}
	return append(lines, dim.Render("·")+dim.Render(" none ")+cell(progress.Day{Checks: 1})+dim.Render(" practised ")+
		cell(progress.Day{Passed: 1})+dim.Render(" passed ")+cell(progress.Day{Passed: 3})+dim.Render(" 3+"))
}

// Stats renders groundwork stats: the summary next to the activity grid.
func Stats(s *session.Session, width int) string {
	st := courseStats(s)
	left := strings.Join(statsRows(st, 46), "\n")
	right := dim.Render(fmt.Sprintf("Last %d weeks", statsWeeks)) + "\n" + strings.Join(statsGrid(st), "\n")
	return framed("Stats", lipgloss.JoinHorizontal(lipgloss.Top, lipgloss.NewStyle().Width(50).Render(left), right), min(width, 88))
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
