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

	"github.com/pjanderson/groundwork/internal/course"
	"github.com/pjanderson/groundwork/internal/session"
	"github.com/pjanderson/groundwork/internal/workspace"
)

type pane int

const (
	paneResult pane = iota
	paneInstructions
	paneHints
	paneList
	paneHelp
	paneSolution
	paneGuides // the list of guides
	paneGuide  // reading one guide
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

	tasks      []course.Task
	hintTask   int         // index into tasks of the task whose hints are shown
	hintLevels map[int]int // per task index: how many levels have been revealed
	todoFiles  []string    // files in the exercise that still contain TODO markers
	listLine   int         // line of the list cursor, to keep it scrolled into view
	listCursor int
	guideLine  int           // line of the guide cursor, like listLine
	guideIdx   int           // index into Course.Guides of the guide cursor
	guide      *course.Guide // the guide open in paneGuide
	confirm    string        // pending confirmation, e.g. "reset"
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
	m.hintTask, m.hintLevels, m.confirm, m.flash = 0, map[int]int{}, "", ""
	m.tasks, _ = m.s.Course.Tasks(e)
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
		switch {
		case k == "x" && pending == "reset":
			return m, m.reset()
		case k == "s" && pending == "reveal":
			m.flash = ""
			if err := m.s.Reveal(m.ex); err != nil {
				m.flash = "Couldn't write the solution: " + err.Error()
			} else {
				m.pane = paneSolution
			}
			m.refresh()
			m.body.GotoTop()
			return m, nil
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

	if m.pane == paneGuides || m.pane == paneGuide {
		if handled, cmd := m.guideKey(msg); handled {
			return m, cmd
		}
	}

	switch k {
	case "esc":
		m.pane = paneResult
	case "g":
		m.openGuides()
	case "i":
		m.toggle(paneInstructions)
		m.body.GotoTop()
	case "h":
		if len(m.tasks) == 0 {
			m.flash = "This exercise has no hints."
			break
		}
		if m.pane != paneHints {
			m.pane = paneHints
			m.hintTask = m.stuckTask()
			m.hintLevels[m.hintTask] = max(m.hintLevels[m.hintTask], 1)
		} else if m.hintLevels[m.hintTask] < len(m.tasks[m.hintTask].Hints) {
			m.hintLevels[m.hintTask]++
		}
		m.refresh()
		m.body.GotoBottom()
		return m, nil
	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		if n := int(k[0] - '0'); m.pane == paneHints && n <= len(m.tasks) {
			m.hintTask = n - 1
			m.hintLevels[m.hintTask] = max(m.hintLevels[m.hintTask], 1)
			m.refresh()
			m.body.GotoBottom()
			return m, nil
		}
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
	case "s":
		switch {
		case m.s.Done(m.ex) || m.s.Revealed(m.ex):
			m.toggle(paneSolution)
			m.body.GotoTop()
		case !m.sawAnswer():
			m.flash = "Try the hints first: h shows them, and the solution unlocks after a task's Answer hint."
		default:
			m.confirm = "reveal"
			m.flash = "Press s again to see the solution. Any other key cancels."
		}
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

// openGuides shows the guide list with the cursor on the next unread guide,
// else the current module's first guide (or the nearest earlier one).
func (m *Model) openGuides() {
	m.pane = paneGuides
	m.guideIdx = 0
	if next := m.s.NextGuide(); next != nil {
		m.guideIdx = next.Number - 1
		return
	}
	for i, g := range m.s.Course.Guides {
		if g.Module.Number <= m.ex.Module.Number {
			m.guideIdx = i
		}
		if g.Module == m.ex.Module {
			m.guideIdx = i
			break
		}
	}
}

// guideKey handles keys in the guide list and the guide reader. It reports
// false for keys the main handler should take (q, l, ?, n, ...).
func (m *Model) guideKey(msg tea.KeyMsg) (bool, tea.Cmd) {
	k, last := msg.String(), len(m.s.Course.Guides)-1
	switch k {
	case "m":
		g := m.cursorGuide()
		read := !m.s.Read(g)
		if err := m.s.SetRead(g, read); err != nil {
			m.flash = "Couldn't save: " + err.Error()
		} else if read {
			m.flash = "Marked " + g.Title + " as read."
		} else {
			m.flash = "Marked " + g.Title + " as unread."
		}
		m.refresh()
		return true, nil
	case "o":
		return true, m.openInEditor(m.s.GuidePath(m.cursorGuide()))
	}
	if m.pane == paneGuide {
		switch k {
		case "esc", "backspace", "left":
			m.pane = paneGuides
		case "g":
			m.pane = paneResult
		case "home":
			m.body.GotoTop()
			return true, nil
		case "end", "G":
			m.body.GotoBottom()
			return true, nil
		default:
			return false, nil
		}
		m.refresh()
		return true, nil
	}
	switch k {
	case "up", "k":
		m.guideIdx = max(0, m.guideIdx-1)
	case "down", "j":
		m.guideIdx = min(last, m.guideIdx+1)
	case "home":
		m.guideIdx = 0
	case "end", "G":
		m.guideIdx = last
	case "enter", "right":
		m.guide, m.pane = m.s.Course.Guides[m.guideIdx], paneGuide
		m.refresh()
		m.body.GotoTop()
		return true, nil
	case "esc", "g", "left":
		m.pane = paneResult
	default:
		return false, nil
	}
	m.refresh()
	return true, nil
}

// cursorGuide is the guide open in the reader, else the one under the cursor.
func (m *Model) cursorGuide() *course.Guide {
	if m.pane == paneGuide {
		return m.guide
	}
	return m.s.Course.Guides[m.guideIdx]
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
	target := m.s.Dir(m.ex)
	if files := m.todoFiles; len(files) > 0 {
		target = filepath.Join(m.s.Root, files[0])
	}
	return m.openInEditor(target)
}

// openInEditor opens target in $VISUAL or $EDITOR, suspending the TUI until
// the editor exits.
func (m *Model) openInEditor(target string) tea.Cmd {
	editor := os.Getenv("VISUAL")
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	if editor == "" {
		m.flash = "Set $EDITOR (for example export EDITOR=vim) to open files in your editor."
		return nil
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
		content = wrapIndent(dim.Render("Same text as "+where+", if you'd rather read it in your editor."), 1, w) + "\n"
		if guides := m.ex.Module.Guides; len(guides) > 0 {
			var titles []string
			for _, g := range guides {
				titles = append(titles, bold.Render(g.Title))
			}
			content += wrapIndent(dim.Render("Background for this module: ")+strings.Join(titles, dim.Render(", "))+dim.Render(". Press ")+keyCap.Render("g")+dim.Render(" for the guides."), 1, w) + "\n"
		}
		content += "\n" + m.markdown(dropTitle(readme), w)
	case paneHints:
		content = m.hintsView(w)
	case paneList:
		content = m.listView(w)
	case paneHelp:
		content = m.helpView(w)
	case paneSolution:
		content = m.solutionView(w)
	case paneGuides:
		content = m.guidesView(w)
	case paneGuide:
		content = m.guideView(w)
	default:
		content = m.resultView(w)
	}
	m.body.SetContent(content)
	switch m.pane {
	case paneList:
		m.scrollTo(m.listLine)
	case paneGuides:
		m.scrollTo(m.guideLine)
	}
}

// scrollTo keeps a cursor line, with a little context, inside the viewport.
func (m *Model) scrollTo(line int) {
	if line < m.body.YOffset+2 {
		m.body.SetYOffset(max(0, line-2))
	} else if line > m.body.YOffset+m.body.Height-4 {
		m.body.SetYOffset(line - m.body.Height + 4)
	}
}

func (m *Model) contentWidth() int { return max(30, min(m.width, 110)) }

// readingWidth caps prose at a comfortable line length, even on wide screens.
const readingWidth = 84

func (m *Model) markdown(src string, width int) string {
	width = min(width, readingWidth)
	if m.md == nil || m.mdWidth != width {
		r, err := newMarkdown(width)
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

// newMarkdown returns a renderer for Markdown prose wrapped at width.
func newMarkdown(width int) (*glamour.TermRenderer, error) {
	return glamour.NewTermRenderer(
		glamour.WithStyles(markdownStyle()),
		glamour.WithChromaFormatter("terminal16"), // code in the terminal's palette
		glamour.WithWordWrap(width-2),
		glamour.WithEmoji(),
	)
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
	if g := m.s.NextGuide(); g != nil && m.result != nil {
		b.WriteString("\n\n" + wrapIndent(dim.Render("Background reading: ")+bold.Render(g.Title)+dim.Render(fmt.Sprintf(" (guide %d), for module %d. Press ", g.Number, g.Module.Number))+keyCap.Render("g")+dim.Render(" to read it."), 0, width))
	}
	return b.String()
}

func (m *Model) successView() string {
	var b strings.Builder
	b.WriteString(okBold.Render("★ Exercise complete!") + "\n\n")
	b.WriteString("  Press " + keyCap.Render("s") + " to compare your code with the reference solution\n  (also saved in " + bold.Render(relative(m.s.SolutionDir(m.ex))) + ").\n\n")
	if next := m.s.NextPending(m.ex); next != nil {
		b.WriteString("  When you're ready, press " + keyCap.Render("n") + " for " + bold.Render(fmt.Sprintf("%03d %s", next.Number, next.Title)) + ".")
	} else {
		b.WriteString(okBold.Render("  You finished every exercise in Groundwork. Brilliant work! 🎉"))
	}
	return b.String()
}

// stuckTask guesses which task the learner is on: the task of the first
// failing test, when each test matches one task (the usual layout), else the
// task whose hints are already open.
func (m *Model) stuckTask() int {
	if m.result == nil {
		return m.hintTask
	}
	var tests []*workspace.Test
	for _, t := range m.result.Tests {
		if !strings.HasSuffix(t.Name, "TestGrader") {
			tests = append(tests, t)
		}
	}
	if len(tests) != len(m.tasks) {
		return m.hintTask
	}
	for i, t := range tests {
		if !t.Passed && !t.Skipped {
			return i
		}
	}
	return m.hintTask
}

// sawAnswer reports whether the learner has read any task's Answer hint.
func (m *Model) sawAnswer() bool {
	for i, t := range m.tasks {
		if m.hintLevels[i] >= len(t.Hints) {
			return true
		}
	}
	return false
}

func (m *Model) hintsView(width int) string {
	if len(m.tasks) == 0 {
		return dim.Render("This exercise has no hints.")
	}
	t := m.tasks[m.hintTask]
	shown := m.hintLevels[m.hintTask]
	var b strings.Builder
	b.WriteString(bold.Render(fmt.Sprintf("Task %d of %d · %s", t.Number, len(m.tasks), t.Title)) + "\n\n")
	for i := range shown {
		h := t.Hints[i]
		b.WriteString(keyCap.Render(h.Title) + dim.Render(fmt.Sprintf("  (%d of %d)", i+1, len(t.Hints))) + "\n")
		b.WriteString(m.markdown(h.Body, width) + "\n\n")
	}
	switch {
	case shown < len(t.Hints)-1:
		b.WriteString(dim.Render("Press ") + keyCap.Render("h") + dim.Render(" for a closer hint."))
	case shown < len(t.Hints):
		b.WriteString(dim.Render("Press ") + keyCap.Render("h") + dim.Render(" for the answer to this task."))
	default:
		b.WriteString(dim.Render("That's the answer for this task. Want the whole solution? Press ") + keyCap.Render("s") + dim.Render("."))
	}
	if len(m.tasks) > 1 {
		var others []string
		for i, o := range m.tasks {
			label := keyCap.Render(fmt.Sprint(o.Number)) + dim.Render(" "+o.Title)
			if i == m.hintTask {
				label = bold.Render(fmt.Sprintf("%d %s", o.Number, o.Title))
			}
			others = append(others, "  "+label)
		}
		b.WriteString("\n\n" + dim.Render("Hints for another task: press its number") + "\n" + strings.Join(others, "\n"))
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
			case m.s.Revealed(e):
				icon = warn.Render("◐")
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

// guidesView lists the guides by module. The row under the cursor shows the
// guide's one-sentence summary.
func (m *Model) guidesView(width int) string {
	lines := []string{
		wrapIndent(dim.Render("The computer science underneath the exercises, one idea per guide. Each module has its own, but any guide is worth reading whenever an idea feels shaky. They're also in "+relative(filepath.Join(m.s.Root, "guides"))+string(filepath.Separator)+", for your editor."), 1, width),
		"",
	}
	var mod *course.Module
	for i, g := range m.s.Course.Guides {
		if g.Module != mod {
			mod = g.Module
			if i > 0 {
				lines = append(lines, "")
			}
			head := bold.Render(fmt.Sprintf("%2d  %s", mod.Number, mod.Title))
			if mod == m.ex.Module {
				head += dim.Render("  · your module")
			}
			lines = append(lines, head)
		}
		label := fmt.Sprintf("%02d  %s", g.Number, g.Title)
		icon := dim.Render("·")
		if m.s.Read(g) {
			icon = okStyle.Render("✓")
		}
		if i != m.guideIdx {
			lines = append(lines, "    "+icon+"  "+label)
			continue
		}
		m.guideLine = len(lines)
		lines = append(lines, keyCap.Render("  ▸ ")+icon+"  "+bold.Render(label))
		if g.Summary != "" {
			lines = append(lines, wrapIndent(dim.Render(g.Summary), 11, width))
		}
	}
	return strings.Join(lines, "\n")
}

// guideView renders the open guide.
func (m *Model) guideView(width int) string {
	g := m.guide
	text, err := m.s.Course.GuideText(g)
	if err != nil {
		return badBold.Render("✗ ") + err.Error()
	}
	where := relative(m.s.GuidePath(g))
	status := dim.Render("Press ") + keyCap.Render("m") + dim.Render(" when you've read it.")
	if m.s.Read(g) {
		status = okStyle.Render("✓ read")
	}
	intro := wrapIndent(dim.Render(fmt.Sprintf("Guide %d of %d, for module %d: %s. Also in %s.", g.Number, len(m.s.Course.Guides), g.Module.Number, g.Module.Title, where)), 1, width) + "\n " + status
	return intro + "\n\n" + m.markdown(text, width)
}

func (m *Model) helpView(width int) string {
	rows := [][2]string{
		{"save a file", "re-run the tests"},
		{"i", "instructions (the exercise README)"},
		{"h", "show a hint (press again for a stronger one)"},
		{"s", "the solution: once you pass, or after every hint"},
		{"l", "list every exercise; enter to open one"},
		{"g", "guides: the computer science behind the exercises"},
		{"  m  o", "in the guides: mark read, open in $EDITOR"},
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
	b.WriteString("\n" + m.markdown(failureGuide, width))
	return b.String()
}

// failureGuide teaches how to read what the Results tab shows.
const failureGuide = `## Reading a failure

**A failing test** looks like this:

    exercise_test.go:19: ServiceLabel("api", "production") = "api", want "production/api"

Read it left to right: the file and line of the check, the call the test
made, what your code returned (after ` + "`=`" + `), and what it should have
returned (after ` + "`want`" + `).

**A compiler error** means the code can't run yet, so no test ran:

    exercise.go:24:9: invalid operation: path + service (mismatched types string and func(...) string)

` + "`exercise.go:24:9`" + ` is the file, line and column. Fix the **first** error
first; the rest are often caused by it. Common ones:

| Message says | Usually means |
|---|---|
| ` + "`undefined: x`" + ` | a typo, or a name from another function |
| ` + "`mismatched types A and B`" + ` | you combined two different kinds of value |
| ` + "`declared and not used`" + ` | a variable you made but never read |
| ` + "`no new variables on left side of :=`" + ` | use ` + "`=`" + ` to change an existing variable |
| ` + "`missing return`" + ` | some path through the function returns nothing |

**Names:** inside a function you can use its parameters (the names in its
` + "`( )`" + `) and the variables you declare in it. A name like ` + "`ServiceLabel`" + `
is a whole function, not a value; call it with ` + "`ServiceLabel(...)`" + `.`

// solutionView shows the files the reference solution adds or changes.
func (m *Model) solutionView(width int) string {
	if !m.s.Done(m.ex) && !m.s.Revealed(m.ex) {
		return dim.Render("The solution unlocks when your tests pass, or after you've read every hint.")
	}
	files, err := m.s.SolutionFiles(m.ex)
	if err != nil {
		return badBold.Render("✗ ") + err.Error()
	}
	var b strings.Builder
	b.WriteString(wrapIndent(dim.Render("Also saved in "+relative(m.s.SolutionDir(m.ex))+". Compare it with yours; there's often more than one good answer. Typing it out yourself sticks better than copying."), 1, width) + "\n\n")
	for _, f := range files {
		lang := "go"
		if !strings.HasSuffix(f.Path, ".go") {
			lang = ""
		}
		b.WriteString(keyCap.Render(" "+f.Path) + "\n")
		b.WriteString(m.markdown("```"+lang+"\n"+strings.TrimRight(f.Content, "\n")+"\n```", width) + "\n\n")
	}
	return strings.TrimRight(b.String(), "\n")
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
	left := brand.Render("groundwork") + dim.Render(fmt.Sprintf("  ·  module %d of %d: %s", m.ex.Module.Number, len(m.s.Course.Modules), m.ex.Module.Title))
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

// tabs shows which view the body holds and the key that opens each one. On a
// narrow terminal the labels get shorter rather than wrapping.
func (m *Model) tabs() string {
	hints := "Hints"
	if len(m.tasks) > 1 {
		hints = fmt.Sprintf("Hints · task %d", m.tasks[m.hintTask].Number)
	}
	row := m.tabRow("Instructions", hints)
	if lipgloss.Width(row) > m.contentWidth() {
		row = m.tabRow("README", "Hints")
	}
	return row
}

func (m *Model) tabRow(instructions, hints string) string {
	views := []struct {
		p     pane
		label string
		key   string
	}{
		{paneResult, "Results", "esc"},
		{paneInstructions, instructions, "i"},
		{paneHints, hints, "h"},
		{paneSolution, "Solution", "s"},
		{paneList, "Exercises", "l"},
		{paneGuides, "Guides", "g"},
		{paneHelp, "Keys", "?"},
	}
	var parts []string
	for _, v := range views {
		if v.p == m.pane || v.p == paneGuides && m.pane == paneGuide {
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
	case m.pane == paneInstructions || m.pane == paneHints || m.pane == paneHelp || m.pane == paneSolution:
		status = dim.Render("↑/↓ scroll · esc back to the results")
	case m.pane == paneList:
		status = dim.Render("↑/↓ move · enter open · esc back to the results")
	case m.pane == paneGuides:
		status = dim.Render("↑/↓ move · enter read · m mark read · o open in editor · esc back")
	case m.pane == paneGuide:
		status = dim.Render("↑/↓ scroll · space next page · m mark read · o open in editor · esc back")
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
