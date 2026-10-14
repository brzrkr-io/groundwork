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
	"github.com/charmbracelet/x/ansi"

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

	// Every pane is drawn in a fixed frame; only the text inside scrolls.
	title     string // in the frame's top edge
	sideTitle string // above the sidebar, when the layout has one
	layout    docLayout
	outline   []outlineItem // headings of the document shown, if any
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
		if msg.err == nil {
			if err := m.s.RecordCheck(); err != nil {
				m.checkErr = err
			}
		}
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
		case k == "v" && pending == "review":
			return m, m.startReview()
		case k == "s" && pending == "reviewsolution":
			m.flash = ""
			if err := m.s.Lapse(m.ex); err != nil {
				m.flash = "Couldn't save: " + err.Error()
			}
			m.pane = paneSolution
			m.refresh()
			m.body.GotoTop()
			return m, nil
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
		if m.sawAnswer() {
			if err := m.s.Lapse(m.ex); err != nil {
				m.flash = "Couldn't save: " + err.Error()
			}
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
		case m.pane != paneSolution && m.inReview() && !m.reviewLapsed():
			m.confirm = "reviewsolution"
			m.flash = "This is a review: try it from memory first. Press s again to see the solution anyway (the review then starts over)."
		case m.s.Done(m.ex) || m.s.Revealed(m.ex):
			m.toggle(paneSolution)
			m.body.GotoTop()
		case !m.sawAnswer():
			m.flash = "Try the hints first: h shows them, and the solution unlocks after a task's Answer hint."
		default:
			m.confirm = "reveal"
			m.flash = "Press s again to see the solution. Any other key cancels."
		}
	case "v":
		due := m.s.DueReviews()
		if len(due) == 0 {
			m.flash = "No reviews due. Practice exercises come back for a review 2, 7 and 30 days after you pass them."
			break
		}
		m.confirm = "review"
		m.flash = fmt.Sprintf("Press v again to review %03d %s: it goes back to its starter code (your code is saved first).", due[0].Number, due[0].Title)
	case "x":
		m.confirm = "reset"
		m.flash = "Press x again to reset this exercise to its starter code. Any other key cancels."
	case "[", "]":
		if len(m.outline) > 0 {
			cur := currentSection(m.outline, m.body.YOffset)
			if k == "]" {
				cur = min(cur+1, len(m.outline)-1)
			} else if m.outline[cur].line >= m.body.YOffset {
				cur = max(cur-1, 0)
			}
			m.body.SetYOffset(max(0, m.outline[cur].line-1))
		}
		return m, nil
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

// startReview resets the first due review's exercise and opens it.
func (m *Model) startReview() tea.Cmd {
	due := m.s.DueReviews()
	if len(due) == 0 {
		return nil
	}
	e := due[0]
	backup, err := m.s.StartReview(e)
	if err != nil {
		m.flash = "Couldn't start the review: " + err.Error()
		return nil
	}
	m.flash = "Review: solve it from scratch. Your previous code is in " + relative(backup) + "."
	text := m.flash
	return tea.Batch(m.open(e), tea.Tick(8*time.Second, func(time.Time) tea.Msg { return clearFlash{text} }))
}

// inReview reports whether the open exercise is being solved again as a
// spaced review.
func (m *Model) inReview() bool {
	r, ok := m.s.Review(m.ex)
	return ok && r.Active
}

func (m *Model) reviewLapsed() bool {
	r, _ := m.s.Review(m.ex)
	return r.Lapsed
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
	m.body.Height = max(3, m.height-lipgloss.Height(m.header())-lipgloss.Height(m.footer())-frameRowsUsed)
	m.outline, m.sideTitle = nil, ""
	var content string
	switch m.pane {
	case paneInstructions:
		inner := m.startDoc(m.ex.Title, "On this page", true)
		readme, _ := m.s.Course.README(m.ex)
		where := relative(filepath.Join(m.s.Dir(m.ex), "README.md"))
		meta := dim.Render("Also in " + where + ".")
		if guides := m.ex.Module.Guides; len(guides) > 0 {
			var titles []string
			for _, g := range guides {
				titles = append(titles, g.Title)
			}
			meta += dim.Render(" Background reading: "+strings.Join(titles, ", ")+" (") + keyCap.Render("g") + dim.Render(").")
		}
		src := dropTitle(readme)
		content = m.docText(meta, m.markdown(src, inner), src)
	case paneHints:
		content = m.docText("", m.hintsView(m.startDoc("Hints", "Tasks", true)), "")
	case paneList:
		m.startWide("Progress", "Stats")
		content = m.listView(m.layout.text)
	case paneHelp:
		content = m.docText("", m.helpView(m.startDoc("Keys", "", false)), "")
	case paneSolution:
		content = m.docText("", m.solutionView(m.startDoc("Reference solution", "", false)), "")
	case paneGuides:
		m.startWide("Guides", "About this guide")
		content = m.guidesView(m.layout.text)
	case paneGuide:
		content = m.guideView(w)
	default:
		m.startWide("Results", "This exercise")
		content = m.resultView(m.layout.text)
	}
	m.body.Width = m.layout.text
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

func (m *Model) contentWidth() int { return max(30, min(m.width, 140)) }

// readingWidth caps prose at a comfortable line length, even on wide screens.
const readingWidth = 84

// widestText is the widest a framed document's text may get when an outline
// takes the rest of a wide window.
const widestText = 90

func (m *Model) markdown(src string, width int) string {
	width = min(width, widestText)
	src = sectionRules(src)
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
		glamour.WithStyles(markdownStyle(width)),
		glamour.WithChromaFormatter("terminal16"), // code in the terminal's palette
		glamour.WithWordWrap(width),
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
	wide := m.layout.outline > 0
	if !wide { // no sidebar: say where to edit at the top
		b.WriteString(dim.Render("edit ") + m.editPath() + "\n\n")
	}
	switch {
	case m.checkErr != nil:
		b.WriteString(badBold.Render("✗ Couldn't run the checks") + "\n\n  " + m.checkErr.Error())
	case m.result == nil:
		b.WriteString(dim.Render("Running the tests for the first time…"))
	default:
		b.WriteString(RenderResult(m.result, width))
	}
	switch {
	case m.result != nil && m.result.Status == workspace.Passed && m.checkErr == nil:
		b.WriteString("\n\n" + m.successView(width))
	case m.result != nil && m.s.DoneCount() == 0:
		b.WriteString("\n\n" + bold.Render("How this works") + "\n" +
			wrapIndent("1. Open "+bold.Render(m.editPath())+" in your editor (keep this window open next to it).", 2, width) + "\n" +
			wrapIndent("2. Press "+keyCap.Render("i")+" to read the instructions, then change the code where it says TODO.", 2, width) + "\n" +
			wrapIndent("3. Save the file. The tests above re-run by themselves. Stuck? Press "+keyCap.Render("h")+" for a hint.", 2, width))
	case m.result != nil && !wide:
		b.WriteString("\n\n" + dim.Render("Not sure what to do? Press ") + keyCap.Render("i") + dim.Render(" for the instructions or ") + keyCap.Render("h") + dim.Render(" for a hint."))
	}
	if !wide && m.result != nil {
		if due := m.s.DueReviews(); len(due) > 0 && !m.inReview() {
			b.WriteString("\n\n" + wrapIndent(warn.Render("↻ Review due: ")+bold.Render(fmt.Sprintf("%03d %s", due[0].Number, due[0].Title))+dim.Render(". Solving it again from scratch, days later, is what makes it stick. Press ")+keyCap.Render("v")+dim.Render(" to start."), 0, width))
		}
		if g := m.s.NextGuide(); g != nil {
			b.WriteString("\n\n" + wrapIndent(dim.Render("Background reading: ")+bold.Render(g.Title)+dim.Render(". Press ")+keyCap.Render("g")+dim.Render(" to read it."), 0, width))
		}
	}
	return b.String()
}

// editPath is the file to edit: the first with a TODO, else the folder.
func (m *Model) editPath() string {
	where := relative(m.s.Dir(m.ex)) + string(filepath.Separator)
	if files := m.todoFiles; len(files) > 0 {
		where = relative(filepath.Join(m.s.Root, files[0]))
		if len(files) > 1 {
			where += dim.Render(fmt.Sprintf(" (+%d more)", len(files)-1))
		}
	}
	return where
}

// taskMarks returns each task's state for the sidebar: passed, failing or
// unknown (when tests don't map one-to-one onto tasks).
func (m *Model) taskMarks() []string {
	marks := make([]string, len(m.tasks))
	for i := range marks {
		marks[i] = dim.Render("·")
	}
	if m.result == nil {
		return marks
	}
	if m.result.Status == workspace.Passed {
		for i := range marks {
			marks[i] = okStyle.Render("✓")
		}
		return marks
	}
	var tests []*workspace.Test
	for _, t := range m.result.Tests {
		if !strings.HasSuffix(t.Name, "TestGrader") {
			tests = append(tests, t)
		}
	}
	if len(tests) != len(m.tasks) {
		return marks
	}
	for i, t := range tests {
		switch {
		case t.Passed:
			marks[i] = okStyle.Render("✓")
		case !t.Skipped:
			marks[i] = bad.Render("✗")
		}
	}
	return marks
}

// resultSide is the results sidebar: the tasks, where to edit, hints used,
// and what's waiting (a due review, the next guide).
func (m *Model) resultSide(w int) []string {
	var out []string
	if len(m.tasks) > 0 {
		marks, cur := m.taskMarks(), m.stuckTask()
		var rows []string
		for i, t := range m.tasks {
			label := ansi.Truncate(fmt.Sprintf("%d %s", t.Number, t.Title), w-4, "…")
			if i == cur && m.result != nil && m.result.Status != workspace.Passed {
				rows = append(rows, keyCap.Render("▸ ")+marks[i]+" "+bold.Render(label))
			} else {
				rows = append(rows, "  "+marks[i]+" "+label)
			}
		}
		out = append(out, sideSection("Tasks", rows...)...)
		out = append(out, "")
	}
	dir := filepath.Base(m.s.Dir(m.ex))
	edit := dir + "/"
	if files := m.todoFiles; len(files) > 0 {
		edit = dir + "/" + filepath.Base(files[0])
		if len(files) > 1 {
			edit += dim.Render(fmt.Sprintf(" +%d", len(files)-1))
		}
	}
	out = append(out, sideSection("Edit", edit, dim.Render("press ")+keyCap.Render("e")+dim.Render(" to open it"))...)
	if len(m.tasks) > 0 {
		seen, total := 0, 0
		for i, t := range m.tasks {
			seen += m.hintLevels[i]
			total += len(t.Hints)
		}
		out = append(out, "")
		out = append(out, sideSection("Hints", fmt.Sprintf("%d of %d seen", seen, total)+dim.Render(" · ")+keyCap.Render("h"))...)
	}
	if due := m.s.DueReviews(); len(due) > 0 && !m.inReview() {
		out = append(out, "")
		out = append(out, sideSection("Review due", append(wrapSide(fmt.Sprintf("%03d %s", due[0].Number, due[0].Title), w),
			dim.Render("press ")+keyCap.Render("v")+dim.Render(" to start"))...)...)
	}
	if g := m.s.NextGuide(); g != nil {
		out = append(out, "")
		out = append(out, sideSection("Read next", append(wrapSide(g.Title, w),
			dim.Render("press ")+keyCap.Render("g"))...)...)
	}
	return out
}

func (m *Model) successView(width int) string {
	var b strings.Builder
	line := func(s string) { b.WriteString(wrapIndent(s, 2, width) + "\n\n") }
	b.WriteString(okBold.Render("★ Exercise complete!") + "\n\n")
	line(bold.Render("Explain it back.") + " Before moving on, say in a sentence or two why your code works. If you can't, that's the part to reread.")
	if r, ok := m.s.Review(m.ex); ok && !r.Active {
		days := max(1, int(time.Until(r.Due).Hours()/24+0.5))
		line(dim.Render(fmt.Sprintf("This one comes back for a review in %d days, to solve again from memory.", days)))
	}
	line("Press " + keyCap.Render("s") + " to compare your code with the reference solution (also saved in " + bold.Render(relative(m.s.SolutionDir(m.ex))) + ").")
	if next := m.s.NextPending(m.ex); next != nil {
		b.WriteString(wrapIndent("When you're ready, press "+keyCap.Render("n")+" for "+bold.Render(fmt.Sprintf("%03d %s", next.Number, next.Title))+".", 2, width))
	} else {
		b.WriteString(okBold.Render("  You finished every exercise in Groundwork. Brilliant work! 🎉"))
	}
	return strings.TrimRight(b.String(), "\n")
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
	if len(m.tasks) > 1 && m.layout.outline == 0 { // the sidebar lists them otherwise
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
	if m.layout.outline == 0 { // no sidebar: stats on top
		lines = append(statsRows(courseStats(m.s), width), "", rule.Render(strings.Repeat("─", width)), "")
	}
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
		if mod.Number > 1 {
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
		wrapIndent(dim.Render("The ideas underneath the exercises, one per guide. Also in "+relative(filepath.Join(m.s.Root, "guides"))+string(filepath.Separator)+" for your editor."), 0, width),
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
		if g.Summary != "" && m.layout.outline == 0 {
			lines = append(lines, wrapIndent(dim.Render(g.Summary), 11, width))
		}
	}
	return strings.Join(lines, "\n")
}

// guideView renders the open guide.
func (m *Model) guideView(int) string {
	g := m.guide
	text, err := m.s.Course.GuideText(g)
	if err != nil {
		return badBold.Render("✗ ") + err.Error()
	}
	inner := m.startDoc(g.Title, "On this page", true)
	status := dim.Render("press ") + keyCap.Render("m") + dim.Render(" when you've read it")
	if m.s.Read(g) {
		status = okStyle.Render("✓ read")
	}
	meta := dim.Render(fmt.Sprintf("Guide %d of %d · module %d: %s · also in %s · ", g.Number, len(m.s.Course.Guides), g.Module.Number, g.Module.Title, relative(m.s.GuidePath(g)))) + status
	src := dropTitle(text)
	return m.docText(meta, m.markdown(src, inner), src)
}

// startDoc frames a document titled title (with a sidebar titled
// sideTitle when withSide and there's room) and returns the text width to
// render it at.
func (m *Model) startDoc(title, sideTitle string, withSide bool) int {
	m.title, m.sideTitle = title, sideTitle
	m.layout = layoutDoc(m.contentWidth(), withSide)
	return m.layout.text
}

// startWide frames a list or test output, with a sidebar when there's room.
func (m *Model) startWide(title, sideTitle string) {
	m.title, m.sideTitle = title, sideTitle
	m.layout = layoutWide(m.contentWidth())
}

// docText puts an optional meta line above rendered text and, when src is
// given, finds the outline of its headings.
func (m *Model) docText(meta, rendered, src string) string {
	text := strings.Trim(rendered, "\n")
	if meta != "" {
		var wrapped []string
		for l := range strings.SplitSeq(meta, "\n") {
			wrapped = append(wrapped, ansi.Wrap(l, m.layout.text, " "))
		}
		text = strings.Join(wrapped, "\n") + "\n" + rule.Render(strings.Repeat("┄", m.layout.text)) + "\n\n" + text
	}
	if src != "" && m.layout.outline > 0 {
		m.outline = findOutline(src, text)
	}
	if src != "" && len(m.outline) == 0 && m.layout.outline > 0 {
		m.layout = layoutDoc(m.contentWidth(), false) // nothing to outline: hug the text
		return m.docText(meta, m.markdown(src, m.layout.text), "")
	}
	return text
}

func (m *Model) helpView(width int) string {
	rows := [][2]string{
		{"save a file", "re-run the tests"},
		{"i", "instructions (the exercise README)"},
		{"h", "show a hint (press again for a stronger one)"},
		{"s", "the solution: once you pass, or after a task's Answer hint"},
		{"l", "progress: your streak, stats and every exercise; enter opens one"},
		{"g", "guides: the computer science behind the exercises"},
		{"  m  o", "in the guides: mark read, open in $EDITOR"},
		{"n", "next unfinished exercise"},
		{"← →", "previous / next exercise"},
		{"↑ ↓ pgup pgdn", "scroll"},
		{"[ ]", "previous / next section of a long document"},
		{"e", "open the exercise in $EDITOR"},
		{"r", "re-run the tests now"},
		{"x", "reset the exercise to its starter code"},
		{"v", "start a due review: solve a past exercise again from scratch"},
		{"esc", "back to the test results"},
		{"q", "quit (progress is saved)"},
	}
	var b strings.Builder
	for _, r := range rows {
		b.WriteString(fmt.Sprintf("  %s  %s\n", keyCap.Render(fmt.Sprintf("%-14s", r[0])), r[1]))
	}
	b.WriteString("\n" + m.markdown(sessionGuide+"\n\n"+failureGuide, width))
	return b.String()
}

// sessionGuide is the short version of LEARNING.md: what a good session
// looks like and why.
const sessionGuide = `## A good session (25–45 minutes)

1. **Flashcards** (5 minutes): ` + "`groundwork cards`" + ` makes an Anki deck from the guides.
2. **A due review** if the results say *↻ Review due*: press ` + "`v`" + ` and solve it from memory.
3. **One guide section** (` + "`g`" + `), answering its *Check yourself* questions before the answers.
4. **New exercises.** Try each task before a hint; after a pass, explain to yourself why it works.
5. **Stop on a pass**, so the next session starts with a win.

Pulling answers from memory, spaced over days, is what makes things stick;
rereading mostly feels like learning. LEARNING.md has the details.`

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
		return dim.Render("The solution unlocks when your tests pass, or after you've read a task's Answer hint.")
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
	lines := strings.Split(m.body.View(), "\n")
	for len(lines) < m.body.Height {
		lines = append(lines, "")
	}
	body := frame(m.layout, m.title, m.sideTitle, lines[:m.body.Height], m.sidebar(m.body.Height))
	return lipgloss.JoinVertical(lipgloss.Left, m.header(), body, m.footer())
}

// sidebar is the current pane's sidebar, rows lines tall.
func (m *Model) sidebar(rows int) []string {
	w := m.layout.outline
	if w == 0 {
		return nil
	}
	switch m.pane {
	case paneInstructions, paneGuide:
		return outlineLines(m.outline, w, currentSection(m.outline, m.body.YOffset), rows)
	case paneHints:
		var out []string
		for i, t := range m.tasks {
			dots := strings.Repeat("●", m.hintLevels[i]) + strings.Repeat("○", len(t.Hints)-m.hintLevels[i])
			label := ansi.Truncate(fmt.Sprintf("%d %s", t.Number, t.Title), w-len(t.Hints)-3, "…")
			if i == m.hintTask {
				out = append(out, keyCap.Render("▸ ")+bold.Render(label)+" "+keyCap.Render(dots))
			} else {
				out = append(out, "  "+label+" "+dim.Render(dots))
			}
		}
		return append(out, "", dim.Render("press a task's number"), dim.Render("for its hints"))
	case paneList:
		st := courseStats(m.s)
		out := sideSection("This course", statsRows(st, w)...)
		out = append(out, "")
		return append(out, sideSection(fmt.Sprintf("Last %d weeks", statsWeeks), statsGrid(st)...)...)
	case paneGuides:
		g := m.s.Course.Guides[m.guideIdx]
		out := wrapSide(bold.Render(g.Title), w)
		out = append(out, dim.Render(fmt.Sprintf("Module %d · %s", g.Module.Number, g.Module.Title)))
		if m.s.Read(g) {
			out = append(out, okStyle.Render("✓ read"))
		} else {
			out = append(out, dim.Render("not read yet"))
		}
		if g.Summary != "" {
			out = append(out, "")
			out = append(out, wrapSide(g.Summary, w)...)
		}
		return out
	case paneResult:
		return m.resultSide(w)
	}
	return nil
}

func (m *Model) header() string {
	w := m.contentWidth()
	done, total := m.s.DoneCount(), len(m.s.Course.Exercises)
	left := brand.Render("groundwork") + dim.Render(fmt.Sprintf("  %d · %s", m.ex.Module.Number, m.ex.Module.Title))
	right := progressBar(done, total, min(20, w/6)) + dim.Render(fmt.Sprintf("  %d/%d", done, total))
	if streak := m.s.Stats(1).Streak; streak > 1 && lipgloss.Width(left)+lipgloss.Width(right) < w-16 {
		right = okStyle.Render(fmt.Sprintf("%d-day streak", streak)) + dim.Render("  ·  ") + right
	}

	title := bold.Render(fmt.Sprintf("%03d  %s", m.ex.Number, m.ex.Title))
	if m.ex.Kind != course.Concept {
		title += "  " + dim.Render(string(m.ex.Kind))
	}
	if m.inReview() {
		title += "  " + warn.Render("↻ review")
	}

	return strings.Join([]string{
		spread(left, right, w),
		"",
		spread(title, m.statusBadge(), w),
		m.tabs(),
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
		{paneList, "Progress", "l"},
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
	case m.pane == paneResult && m.result != nil && m.result.Duration > 0:
		status = dim.Render(fmt.Sprintf("Checked in %.1fs · save any file to re-run", m.result.Duration.Seconds()))
	}
	if m.body.TotalLineCount() > m.body.Height {
		status = spread(status, dim.Render(fmt.Sprintf("%.0f%%", m.body.ScrollPercent()*100)), w)
	}

	var keys []string
	switch m.pane {
	case paneResult:
		keys = []string{"n next", "←/→ prev/next", "r re-run", "e edit", "x reset"}
		if len(m.s.DueReviews()) > 0 {
			keys = append(keys, "v review")
		}
	case paneInstructions, paneGuide:
		keys = []string{"↑/↓ scroll"}
		if len(m.outline) > 0 {
			keys = append(keys, "[ ] sections")
		}
		if m.pane == paneGuide {
			keys = append(keys, "m mark read", "o open in editor")
		} else {
			keys = append(keys, "e edit")
		}
		keys = append(keys, "esc back")
	case paneHints:
		keys = []string{"h stronger hint", "1-9 another task", "s solution", "esc back"}
	case paneList:
		keys = []string{"↑/↓ move", "enter open", "esc back"}
	case paneGuides:
		keys = []string{"↑/↓ move", "enter read", "m mark read", "o open in editor", "esc back"}
	default:
		keys = []string{"↑/↓ scroll", "esc back"}
	}
	keys = append(keys, "q quit")
	var parts []string
	for _, k := range keys {
		name, desc, _ := strings.Cut(k, " ")
		parts = append(parts, keyCap.Render(name)+" "+dim.Render(desc))
	}
	help := ansi.Truncate(strings.Join(parts, dim.Render("  ·  ")), w, "…")
	return status + "\n" + help
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
