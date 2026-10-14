package tui

import (
	"github.com/charmbracelet/glamour/ansi"
	"github.com/charmbracelet/glamour/styles"
	"github.com/charmbracelet/lipgloss"
)

// Every colour is one of the terminal's 16 palette slots, never a fixed hex
// value, so Groundwork takes on whatever theme the terminal uses (gruvbox,
// tokyo night, rose pine, light or dark) and follows it when the theme
// changes.
const (
	ansiBlack   = "0"
	ansiYellow  = "3"
	ansiGray    = "8"
	ansiRed     = "9"
	ansiGreen   = "10"
	ansiAccent  = "11" // bright yellow
	ansiBlue    = "12"
	ansiMagenta = "13"
	ansiCyan    = "14"
)

var (
	accent  = lipgloss.Color(ansiAccent)
	muted   = lipgloss.Color(ansiGray)
	brand   = lipgloss.NewStyle().Bold(true).Foreground(accent)
	bold    = lipgloss.NewStyle().Bold(true)
	dim     = lipgloss.NewStyle().Foreground(muted)
	rule    = lipgloss.NewStyle().Foreground(muted)
	okStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(ansiGreen))
	okBold  = okStyle.Bold(true)
	bad     = lipgloss.NewStyle().Foreground(lipgloss.Color(ansiRed))
	badBold = bad.Bold(true)
	warn    = lipgloss.NewStyle().Foreground(lipgloss.Color(ansiYellow))
	keyCap  = lipgloss.NewStyle().Foreground(accent).Bold(true)

	tabActive   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(ansiBlack)).Background(accent).Padding(0, 1)
	tabInactive = lipgloss.NewStyle().Foreground(muted).Padding(0, 1)
)

// Chroma maps these to the 16 palette slots ("terminal16" formatter), so
// code blocks are highlighted in the terminal theme's colours too.
const (
	codeComment = "#555555" // bright black
	codeRed     = "#ff0000"
	codeGreen   = "#00ff00"
	codeYellow  = "#ffff00"
	codeBlue    = "#0000ff"
	codeMagenta = "#ff00ff"
	codeCyan    = "#00ffff"
)

// markdownStyle renders READMEs and hints with the palette above.
func markdownStyle() ansi.StyleConfig {
	s := styles.DarkStyleConfig
	yes, margin := true, uint(1)
	color := func(c string) *string { return &c }

	s.Document.Color = nil // the terminal's own foreground
	s.Document.Margin = &margin
	s.Heading.Color = color(ansiAccent)
	s.H1.Color, s.H1.BackgroundColor = color(ansiAccent), nil
	// Section headings (LEARN, YOUR TASK) in accent capitals; task headings
	// in bold. No literal "## " markers.
	s.H2.Prefix, s.H2.Upper, s.H2.Bold = "", &yes, &yes
	s.H3.Prefix, s.H3.Color = "", nil
	s.H4.Prefix, s.H4.Color = "", nil
	s.H6.Color = color(ansiGray)
	s.HorizontalRule.Color = color(ansiGray)
	s.Link.Color = color(ansiBlue)
	s.LinkText.Color = color(ansiCyan)
	s.Image.Color = color(ansiMagenta)
	s.ImageText.Color = color(ansiGray)
	// Inline code: coloured, no background, so no padding spaces either
	// (they'd read as extra spaces around `+` and `"-"`).
	s.Code.Color, s.Code.BackgroundColor = color(ansiCyan), nil
	s.Code.Prefix, s.Code.Suffix = "", ""

	s.CodeBlock.Color = nil
	c := *s.CodeBlock.Chroma
	fg := func(hex string) ansi.StylePrimitive { return ansi.StylePrimitive{Color: color(hex)} }
	plain := ansi.StylePrimitive{} // the terminal's own foreground
	c.Text, c.Name, c.NameOther, c.Operator, c.Punctuation = plain, plain, plain, plain, plain
	c.Error = fg(codeRed)
	c.Comment, c.CommentPreproc = fg(codeComment), fg(codeComment)
	c.Keyword, c.KeywordReserved, c.KeywordNamespace = fg(codeRed), fg(codeRed), fg(codeRed)
	c.KeywordType, c.NameClass = fg(codeYellow), fg(codeYellow)
	c.NameBuiltin, c.NameAttribute, c.NameDecorator = fg(codeCyan), fg(codeCyan), fg(codeCyan)
	c.NameTag, c.NameException = fg(codeRed), fg(codeRed)
	c.NameFunction = fg(codeBlue)
	c.NameConstant, c.Literal, c.LiteralNumber, c.LiteralDate = fg(codeMagenta), fg(codeMagenta), fg(codeMagenta), fg(codeMagenta)
	c.LiteralString, c.GenericInserted = fg(codeGreen), fg(codeGreen)
	c.LiteralStringEscape = fg(codeCyan)
	c.GenericDeleted = fg(codeRed)
	c.Background = ansi.StylePrimitive{}
	s.CodeBlock.Chroma = &c
	return s
}
