// Package tui is the interactive watch mode: it shows the current exercise,
// re-runs its tests whenever a file is saved, and offers instructions, hints
// and a course list without leaving the terminal.
package tui

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"

	"github.com/pjanderson/gopherlings/internal/course"
	"github.com/pjanderson/gopherlings/internal/session"
	"github.com/pjanderson/gopherlings/internal/workspace"
)

type pane int

const (
	paneResult pane = iota
	paneInstructions
	paneHints
	paneList
	paneHelp
)

// Model is the Bubble Tea model for watch mode.
type Model struct {
	s  *session.Session
	ex *course.Exercise

	width, height int
	pane          pane
	body          viewport.Model
	spin          spinner.Model

	watcher  *workspace.Watcher
	watchSeq int

	checking  bool
	checkSeq  int
	cancel    context.CancelFunc
	result    *workspace.Result
	checkErr  error
	completed bool // passed during this visit

	hints      []course.Hint
	todoFiles  []string // files in the exercise that still contain TODO markers
	listLine   int      // line of the list cursor, to keep it scrolled into view
	hintsShown int
	listCursor int
	confirm    string // pending confirmation, e.g. "reset"
	flash      string
	fatal      error

	md      *glamour.TermRenderer
	mdWidth int
}

type (
	checkedMsg struct {
		seq    int
		result *workspace.Result
		err    error
	}
	changedMsg  struct{ seq int }
	editedMsg   struct{ err error }
	clearFlash  struct{ text string }
	openedMsg   struct{}
	openFailMsg struct{ err error }
)

// New creates the model.
func New(s *session.Session) *Model {
	vp := viewport.New(80, 20)
	vp.KeyMap = viewport.KeyMap{
		PageDown:     key.NewBinding(key.WithKeys("pgdown", " ", "f")),
		PageUp:       key.NewBinding(key.WithKeys("pgup", "b")),
		HalfPageDown: key.NewBinding(key.WithKeys("d", "ctrl+d")),
		HalfPageUp:   key.NewBinding(key.WithKeys("u", "ctrl+u")),
		Down:         key.NewBinding(key.WithKeys("down", "j")),
		Up:           key.NewBinding(key.WithKeys("up", "k")),
	}
	sp := spinner.New(spinner.WithSpinner(spinner.MiniDot), spinner.WithStyle(lipgloss.NewStyle().Foreground(accent)))
	return &Model{s: s, width: 100, height: 30, body: vp, spin: sp}
}

// Init opens the current exercise.
func (m *Model) Init() tea.Cmd {
	return tea.Batch(m.open(m.s.Current()), m.spin.Tick)
}

// Err is the error that ended the program, if any.
func (m *Model) Err() error { return m.fatal }

// Close releases the file watcher and stops any running check.
func (m *Model) Close() {
	if m.cancel != nil {
		m.cancel()
	}
	if m.watcher != nil {
		m.watcher.Close()
		m.watcher = nil
	}
}

func (m *Model) open(e *course.Exercise) tea.Cmd {
	m.Close()
	m.ex = e
	m.result, m.checkErr, m.completed = nil, nil, false
	m.hintsShown, m.confirm, m.flash = 0, "", ""
	m.hints, _ = m.s.Course.Hints(e)
	m.pane = paneResult

	err := m.s.Prepare(e)
	if err != nil {
		return func() tea.Msg { return openFailMsg{err} }
	}
	w, err := workspace.Watch(m.s.Dir(e))
	if err != nil {
		return func() tea.Msg { return openFailMsg{err} }
	}
	m.watcher = w
	m.watchSeq++
	return tea.Batch(func() tea.Msg { return openedMsg{} }, m.waitForChange(), m.runCheck())
}

func (m *Model) waitForChange() tea.Cmd {
	seq, changes := m.watchSeq, m.watcher.Changes
	return func() tea.Msg {
		if _, ok := <-changes; !ok {
			return nil
		}
		return changedMsg{seq}
	}
}

func (m *Model) runCheck() tea.Cmd {
	if m.cancel != nil {
		m.cancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel
	m.checkSeq++
	m.checking = true
	seq, e := m.checkSeq, m.ex
	return func() tea.Msg {
		r, err := m.s.Check(ctx, e)
		return checkedMsg{seq, r, err}
	}
}

// Update handles input, file changes and check results.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.refresh()
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spin, cmd = m.spin.Update(msg)
		return m, cmd

	case openedMsg:
		m.refresh()
		m.body.GotoTop()
		return m, nil

	case openFailMsg:
		m.fatal = msg.err
		return m, tea.Quit

	case changedMsg:
		if msg.seq != m.watchSeq {
			return m, nil
		}
		m.flash = ""
		if m.pane == paneInstructions || m.pane == paneHints {
			m.pane = paneResult // show the result of the save
		}
		cmd := m.runCheck()
		m.refresh()
		return m, tea.Batch(cmd, m.waitForChange())

	case checkedMsg:
		if msg.seq != m.checkSeq {
			return m, nil // a newer check is running
		}
		m.checking = false
		m.result, m.checkErr = msg.result, msg.err
		if msg.err == nil && msg.result.Status == workspace.Passed && !m.completed {
			m.completed = true
			if err := m.s.Complete(m.ex); err != nil {
				m.checkErr = err
			}
			if m.pane != paneList {
				m.pane = paneResult
			}
		}
		m.refresh()
		if m.pane == paneResult {
			m.body.GotoTop()
		}
		return m, nil

	case editedMsg:
		if msg.err != nil {
			m.flash = "Editor failed: " + msg.err.Error()
		}
		m.refresh()
		return m, nil

	case clearFlash:
		if m.flash == msg.text {
			m.flash = ""
		}
		return m, nil

	case tea.KeyMsg:
		return m.key(msg)
	}
	return m, nil
}

func (m *Model) key(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	k := msg.String()
	if k == "ctrl+c" || k == "q" {
		return m, tea.Quit
	}
	if m.confirm != "" {
		pending := m.confirm
		m.confirm = ""
		if k == "x" && pending == "reset" {
			return m, m.reset()
		}
		m.flash = ""
	}

	if m.pane == paneList {
		switch k {
		case "up", "k":
			m.listCursor = max(0, m.listCursor-1)
		case "down", "j":
			m.listCursor = min(len(m.s.Course.Exercises)-1, m.listCursor+1)
		case "pgup", "b":
			m.listCursor = max(0, m.listCursor-m.body.Height)
		case "pgdown", " ", "f":
			m.listCursor = min(len(m.s.Course.Exercises)-1, m.listCursor+m.body.Height)
		case "home", "g":
			m.listCursor = 0
		case "end", "G":
			m.listCursor = len(m.s.Course.Exercises) - 1
		case "enter":
			return m, m.open(m.s.Course.Exercises[m.listCursor])
		case "esc", "l":
			m.pane = paneResult
		default:
			return m, nil
		}
		m.refresh()
		return m, nil
	}

	switch k {
	case "esc":
		m.pane = paneResult
	case "i":
		m.toggle(paneInstructions)
		m.body.GotoTop()
	case "h":
		if m.pane != paneHints {
			m.pane = paneHints
			m.hintsShown = max(m.hintsShown, 1)
		} else if m.hintsShown < len(m.hints) {
			m.hintsShown++
		}
		m.refresh()
		m.body.GotoBottom()
		return m, nil
	case "l":
		m.pane = paneList
		m.listCursor = m.ex.Number - 1
	case "?":
		m.toggle(paneHelp)
	case "r":
		m.pane = paneResult
		cmd := m.runCheck()
		m.refresh()
		return m, cmd
	case "n":
		if next := m.s.NextPending(m.ex); next != nil {
			return m, m.open(next)
		}
		m.flash = "Every exercise is done. 🎉"
	case "right", "tab":
		if m.ex.Number < len(m.s.Course.Exercises) {
			return m, m.open(m.s.Course.Exercises[m.ex.Number])
		}
	case "left", "shift+tab":
		if m.ex.Number > 1 {
			return m, m.open(m.s.Course.Exercises[m.ex.Number-2])
		}
	case "e":
		return m, m.edit()
	case "x":
		m.confirm = "reset"
		m.flash = "Press x again to reset this exercise to its starter code. Any other key cancels."
	default:
		var cmd tea.Cmd
		m.body, cmd = m.body.Update(msg)
		return m, cmd
	}
	m.refresh()
	return m, nil
}

func (m *Model) toggle(p pane) {
	if m.pane == p {
		m.pane = paneResult
	} else {
		m.pane = p
	}
}

func (m *Model) reset() tea.Cmd {
	files, err := m.s.Reset(m.ex)
	if err != nil {
		m.flash = "Reset failed: " + err.Error()
	} else {
		m.flash = "Reset " + strings.Join(files, ", ") + " to the starter code."
	}
	m.pane = paneResult
	text := m.flash
	return tea.Batch(m.runCheck(), tea.Tick(4*time.Second, func(time.Time) tea.Msg { return clearFlash{text} }))
}

// edit opens the first file with a TODO in $VISUAL or $EDITOR, suspending
// the TUI until the editor exits.
func (m *Model) edit() tea.Cmd {
	editor := os.Getenv("VISUAL")
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	if editor == "" {
		m.flash = "Set $EDITOR (for example export EDITOR=vim) to open files with e."
		return nil
	}
	target := m.s.Dir(m.ex)
	if files := m.todoFiles; len(files) > 0 {
		target = filepath.Join(m.s.Root, files[0])
	}
	fields := strings.Fields(editor)
	cmd := exec.Command(fields[0], append(fields[1:], target)...)
	return tea.ExecProcess(cmd, func(err error) tea.Msg { return editedMsg{err} })
}

// refresh lays out the scrollable body for the current pane and size.
func (m *Model) refresh() {
	if m.ex == nil {
		return
	}
	m.todoFiles = m.s.TodoFiles(m.ex)
	w := m.contentWidth()
	m.body.Width = w
	m.body.Height = max(3, m.height-lipgloss.Height(m.header())-lipgloss.Height(m.footer()))
	var content string
	switch m.pane {
	case paneInstructions:
		readme, _ := m.s.Course.README(m.ex)
		where := relative(filepath.Join(m.s.Dir(m.ex), "README.md"))
		content = dim.Render(" Same text as "+where+", if you'd rather read it in your editor.") + "\n\n" + m.markdown(dropTitle(readme), w)
	case paneHints:
		content = m.hintsView(w)
	case paneList:
		content = m.listView(w)
	case paneHelp:
		content = helpView()
	default:
		content = m.resultView(w)
	}
	m.body.SetContent(content)
	if m.pane == paneList {
		if m.listLine < m.body.YOffset+2 {
			m.body.SetYOffset(max(0, m.listLine-2))
		} else if m.listLine > m.body.YOffset+m.body.Height-3 {
			m.body.SetYOffset(m.listLine - m.body.Height + 3)
		}
	}
}

func (m *Model) contentWidth() int { return max(30, min(m.width, 110)) }

// readingWidth caps prose at a comfortable line length, even on wide screens.
const readingWidth = 84

func (m *Model) markdown(src string, width int) string {
	width = min(width, readingWidth)
	if m.md == nil || m.mdWidth != width {
		r, err := glamour.NewTermRenderer(
			glamour.WithStyles(markdownStyle()),
			glamour.WithChromaFormatter("terminal16"), // code in the terminal's palette
			glamour.WithWordWrap(width-2),
			glamour.WithEmoji(),
		)
		if err != nil {
			return src
		}
		m.md, m.mdWidth = r, width
	}
	out, err := m.md.Render(src)
	if err != nil {
		return src
	}
	return strings.Trim(out, "\n")
}

// dropTitle removes the README's "# Title" line; the header already shows it.
func dropTitle(readme string) string {
	if strings.HasPrefix(readme, "# ") {
		if _, rest, ok := strings.Cut(readme, "\n"); ok {
			return rest
		}
	}
	return readme
}

func (m *Model) resultView(width int) string {
	var b strings.Builder
	switch {
	case m.checkErr != nil:
		b.WriteString(badBold.Render("✗ Couldn't run the checks") + "\n\n  " + m.checkErr.Error())
	case m.result == nil:
		b.WriteString(dim.Render("Running the tests for the first time…"))
	default:
		b.WriteString(RenderResult(m.result, width))
	}
	if m.result != nil && m.result.Status == workspace.Passed && m.checkErr == nil {
		b.WriteString("\n\n" + m.successView())
	} else if m.result != nil && m.s.DoneCount() == 0 {
		file := relative(m.s.Dir(m.ex))
		if len(m.todoFiles) > 0 {
			file = relative(filepath.Join(m.s.Root, m.todoFiles[0]))
		}
		b.WriteString("\n\n" + bold.Render("How this works") + "\n" +
			"  1. Open " + bold.Render(file) + " in your editor (keep this window open next to it).\n" +
			"  2. Press " + keyCap.Render("i") + " to read the instructions, then change the code where it says TODO.\n" +
			"  3. Save the file. The tests above re-run by themselves. Stuck? Press " + keyCap.Render("h") + " for a hint.")
	} else if m.result != nil {
		b.WriteString("\n\n" + dim.Render("Not sure what to do? Press ") + keyCap.Render("i") + dim.Render(" for the instructions or ") + keyCap.Render("h") + dim.Render(" for a hint."))
	}
	return b.String()
}

func (m *Model) successView() string {
	var b strings.Builder
	b.WriteString(okBold.Render("★ Exercise complete!") + "\n\n")
	b.WriteString("  Compare your code with the reference solution in\n  " + bold.Render(relative(m.s.SolutionDir(m.ex))) + "\n\n")
	if next := m.s.NextPending(m.ex); next != nil {
		b.WriteString("  When you're ready, press " + keyCap.Render("n") + " for " + bold.Render(fmt.Sprintf("%03d %s", next.Number, next.Title)) + ".")
	} else {
		b.WriteString(okBold.Render("  You finished every exercise in Gopherlings. Brilliant work! 🎉"))
	}
	return b.String()
}

func (m *Model) hintsView(width int) string {
	if len(m.hints) == 0 {
		return dim.Render("This exercise has no hints.")
	}
	var b strings.Builder
	for i := range m.hintsShown {
		h := m.hints[i]
		b.WriteString(keyCap.Render(fmt.Sprintf("Hint %d of %d", i+1, len(m.hints))) + dim.Render(" · ") + bold.Render(h.Title) + "\n")
		b.WriteString(m.markdown(h.Body, width) + "\n\n")
	}
	if m.hintsShown < len(m.hints) {
		b.WriteString(dim.Render("Press ") + keyCap.Render("h") + dim.Render(" for a stronger hint."))
	} else {
		b.WriteString(dim.Render("That's every hint. The solution unlocks once your tests pass."))
	}
	return b.String()
}

func (m *Model) listView(width int) string {
	var lines []string
	for _, mod := range m.s.Course.Modules {
		done := 0
		for _, e := range mod.Exercises {
			if m.s.Done(e) {
				done++
			}
		}
		head := bold.Render(fmt.Sprintf("%2d  %s", mod.Number, mod.Title))
		count := dim.Render(fmt.Sprintf("%d/%d", done, len(mod.Exercises)))
		gap := max(1, width-lipgloss.Width(head)-lipgloss.Width(count)-2)
		if len(lines) > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, head+strings.Repeat(" ", gap)+count)
		for _, e := range mod.Exercises {
			icon := dim.Render("·")
			switch {
			case m.s.Done(e):
				icon = okStyle.Render("✓")
			case e == m.ex:
				icon = keyCap.Render("●")
			}
			label := fmt.Sprintf("%03d  %s", e.Number, e.Title)
			if e.Kind != course.Concept {
				label += dim.Render("  " + string(e.Kind))
			}
			row := "  " + icon + "  " + label
			if e.Number-1 == m.listCursor {
				row = keyCap.Render("▸ ") + icon + "  " + bold.Render(label)
				m.listLine = len(lines)
			}
			lines = append(lines, row)
		}
	}
	return strings.Join(lines, "\n")
}

func helpView() string {
	rows := [][2]string{
		{"save a file", "re-run the tests"},
		{"i", "instructions (the exercise README)"},
		{"h", "show a hint (press again for a stronger one)"},
		{"l", "list every exercise; enter to open one"},
		{"n", "next unfinished exercise"},
		{"← →", "previous / next exercise"},
		{"↑ ↓ pgup pgdn", "scroll"},
		{"e", "open the exercise in $EDITOR"},
		{"r", "re-run the tests now"},
		{"x", "reset the exercise to its starter code"},
		{"esc", "back to the test results"},
		{"q", "quit (progress is saved)"},
	}
	var b strings.Builder
	b.WriteString(bold.Render("Keys") + "\n\n")
	for _, r := range rows {
		b.WriteString(fmt.Sprintf("  %s  %s\n", keyCap.Render(fmt.Sprintf("%-14s", r[0])), r[1]))
	}
	return b.String()
}

// View renders header, scrollable body and footer.
func (m *Model) View() string {
	if m.ex == nil {
		return ""
	}
	return lipgloss.JoinVertical(lipgloss.Left, m.header(), m.body.View(), m.footer())
}

func (m *Model) header() string {
	w := m.contentWidth()
	done, total := m.s.DoneCount(), len(m.s.Course.Exercises)
	left := brand.Render("gopherlings") + dim.Render(fmt.Sprintf("  ·  module %d of %d: %s", m.ex.Module.Number, len(m.s.Course.Modules), m.ex.Module.Title))
	right := progressBar(done, total, min(20, w/5)) + dim.Render(fmt.Sprintf("  %d/%d done", done, total))

	title := bold.Render(fmt.Sprintf("%03d  %s", m.ex.Number, m.ex.Title))
	if m.ex.Kind != course.Concept {
		title += "  " + dim.Render("["+string(m.ex.Kind)+"]")
	}

	files := m.todoFiles
	where := relative(m.s.Dir(m.ex)) + string(filepath.Separator)
	if len(files) > 0 {
		where = relative(filepath.Join(m.s.Root, files[0]))
		if len(files) > 1 {
			where += dim.Render(fmt.Sprintf("  (+%d more with TODOs)", len(files)-1))
		}
	}

	return strings.Join([]string{
		spread(left, right, w),
		"",
		spread(title, m.statusBadge(), w),
		dim.Render("edit  ") + where,
		"",
		m.tabs(),
		rule.Render(strings.Repeat("─", w)),
	}, "\n")
}

// statusBadge is the at-a-glance state of the current exercise.
func (m *Model) statusBadge() string {
	switch {
	case m.checking:
		return m.spin.View() + dim.Render(" running tests")
	case m.checkErr != nil:
		return badBold.Render("✗ checks couldn't run")
	case m.result == nil:
		return ""
	case m.result.Status == workspace.Passed:
		return okBold.Render("✓ all tests pass")
	case m.result.Status == workspace.BuildFailed:
		return badBold.Render("✗ doesn't compile")
	}
	passed, failed := m.result.Counts()
	return badBold.Render(fmt.Sprintf("✗ %d of %d failing", failed, passed+failed))
}

// tabs shows which view the body holds and the key that opens each one.
func (m *Model) tabs() string {
	hints := "Hints"
	if len(m.hints) > 0 {
		hints = fmt.Sprintf("Hints %d/%d", m.hintsShown, len(m.hints))
	}
	views := []struct {
		p     pane
		label string
		key   string
	}{
		{paneResult, "Results", "esc"},
		{paneInstructions, "Instructions", "i"},
		{paneHints, hints, "h"},
		{paneList, "All exercises", "l"},
		{paneHelp, "Keys", "?"},
	}
	var parts []string
	for _, v := range views {
		if v.p == m.pane {
			parts = append(parts, tabActive.Render(v.label))
		} else {
			parts = append(parts, tabInactive.Render(v.label+" "+keyCap.Render(v.key)))
		}
	}
	return strings.Join(parts, " ")
}

func (m *Model) footer() string {
	w := m.contentWidth()
	status := ""
	switch {
	case m.flash != "":
		status = warn.Render(m.flash)
	case m.checking && m.result == nil:
		status = dim.Render("Compiling and running the tests. The first run of an exercise takes a moment.")
	case m.pane == paneInstructions || m.pane == paneHints || m.pane == paneHelp:
		status = dim.Render("↑/↓ scroll · esc back to the results")
	case m.pane == paneList:
		status = dim.Render("↑/↓ move · enter open · esc back to the results")
	case m.result != nil && m.result.Duration > 0:
		status = dim.Render(fmt.Sprintf("Checked in %.1fs. Save any file to re-run.", m.result.Duration.Seconds()))
	}
	keys := []string{"n next unfinished", "←/→ prev/next", "r re-run", "e edit", "x reset", "q quit"}
	var parts []string
	for _, k := range keys {
		name, desc, _ := strings.Cut(k, " ")
		parts = append(parts, keyCap.Render(name)+" "+dim.Render(desc))
	}
	help := lipgloss.NewStyle().Width(w).Render(strings.Join(parts, dim.Render("  ·  ")))
	scroll := ""
	if m.body.TotalLineCount() > m.body.Height {
		scroll = dim.Render(fmt.Sprintf("  ·  %.0f%% scrolled", m.body.ScrollPercent()*100))
	}
	return strings.Join([]string{rule.Render(strings.Repeat("─", w)), status + scroll, help}, "\n")
}

// spread puts left and right on one line, pushed to opposite edges.
func spread(left, right string, width int) string {
	return left + strings.Repeat(" ", max(1, width-lipgloss.Width(left)-lipgloss.Width(right))) + right
}

// relative shows paths relative to the working directory when that is shorter.
func relative(path string) string {
	if wd, err := os.Getwd(); err == nil {
		if rel, err := filepath.Rel(wd, path); err == nil && !strings.HasPrefix(rel, "..") {
			return rel
		}
	}
	return path
}

func progressBar(done, total, width int) string {
	filled := 0
	if total > 0 {
		filled = done * width / total
	}
	if done > 0 && filled == 0 {
		filled = 1
	}
	return okStyle.Render(strings.Repeat("━", filled)) + rule.Render(strings.Repeat("━", width-filled))
}
