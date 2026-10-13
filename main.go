// Command gopherlings is a hands-on Go course in your terminal: fix small
// programs until their tests pass, from "Hello, Go" to Kubernetes controllers.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime/debug"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-isatty"

	"github.com/pjanderson/gopherlings/content"
	"github.com/pjanderson/gopherlings/internal/course"
	"github.com/pjanderson/gopherlings/internal/session"
	"github.com/pjanderson/gopherlings/internal/tui"
	"github.com/pjanderson/gopherlings/internal/workspace"
)

const usage = `gopherlings: learn Go by fixing small programs until their tests pass.

Usage:
  gopherlings               open the current exercise and watch for saves
  gopherlings init [DIR]    start a new course in DIR (default: this directory)
  gopherlings list [--all]  show your progress through the course
  gopherlings check [EX]    run one exercise's tests once (default: current)
  gopherlings reset EX      restore an exercise's starter code
  gopherlings version       print the version

EX is an exercise number (12) or folder name (012_variables).
Inside the watcher, press ? for keys.
`

// errSilent means the failure has already been reported.
var errSilent = errors.New("")

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		if err != errSilent {
			fmt.Fprintln(os.Stderr, tui.Notice("error", err.Error()))
		}
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	cmd, rest := "", []string(nil)
	if len(args) > 0 {
		cmd, rest = args[0], args[1:]
	}
	switch cmd {
	case "help", "-h", "--help":
		fmt.Fprint(stdout, usage)
		return nil
	case "version", "--version":
		fmt.Fprintln(stdout, "gopherlings", version())
		return nil
	}

	c, err := course.Load(content.Files)
	if err != nil {
		return fmt.Errorf("the embedded course is broken: %w", err)
	}
	if cmd == "init" {
		return initCourse(c, rest, stdout)
	}
	s, err := openSession(c)
	if err != nil {
		return err
	}
	switch cmd {
	case "", "watch":
		if !isatty.IsTerminal(os.Stdout.Fd()) {
			return check(s, s.Current(), stdout)
		}
		return watch(s, stdout)
	case "list":
		fs := flag.NewFlagSet("list", flag.ContinueOnError)
		fs.SetOutput(stderr)
		all := fs.Bool("all", false, "show every exercise")
		if err := fs.Parse(rest); err != nil {
			return errSilent
		}
		fmt.Fprint(stdout, tui.CourseList(s, *all))
		return nil
	case "check":
		e := s.Current()
		if len(rest) > 0 {
			if e, err = find(c, rest[0]); err != nil {
				return err
			}
		}
		return check(s, e, stdout)
	case "reset":
		if len(rest) != 1 {
			return errors.New("which exercise? For example: gopherlings reset 12")
		}
		e, err := find(c, rest[0])
		if err != nil {
			return err
		}
		files, err := s.Reset(e)
		if err != nil {
			return err
		}
		fmt.Fprintln(stdout, tui.Notice("ok", fmt.Sprintf("Restored %s in %s", strings.Join(files, ", "), e.Dir())))
		return nil
	default:
		fmt.Fprint(stderr, usage)
		return fmt.Errorf("unknown command %q", cmd)
	}
}

func openSession(c *course.Course) (*session.Session, error) {
	root, err := session.FindRoot(".")
	if errors.Is(err, session.ErrNoCourse) {
		return nil, errors.New("there's no Gopherlings course in this directory or its parents.\n\n" +
			"  Start one here:          " + tui.Key("gopherlings init") + "\n" +
			"  Or in a new folder:      " + tui.Key("gopherlings init learn-go") + "\n")
	}
	if err != nil {
		return nil, err
	}
	s, err := session.Open(root, c)
	if err != nil {
		return nil, err
	}
	return s, s.Sync()
}

func find(c *course.Course, query string) (*course.Exercise, error) {
	e, ok := c.Find(query)
	if !ok {
		return nil, fmt.Errorf("no exercise %q; gopherlings list --all shows them all", query)
	}
	return e, nil
}

func initCourse(c *course.Course, args []string, stdout io.Writer) error {
	if len(args) > 1 {
		return errors.New("usage: gopherlings init [DIR]")
	}
	dir := "."
	if len(args) == 1 {
		dir = args[0]
	}
	if _, err := exec.LookPath("go"); err != nil {
		return errors.New("Gopherlings needs the Go toolchain. Install it from https://go.dev/dl/ and try again")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	abs, _ := filepath.Abs(dir)
	if err := session.Init(abs); err != nil {
		return err
	}
	s, err := session.Open(abs, c)
	if err != nil {
		return err
	}
	if err := s.Sync(); err != nil {
		return err
	}
	fmt.Fprintln(stdout, tui.Notice("ok", fmt.Sprintf("Created a Gopherlings course in %s", abs)))
	fmt.Fprintf(stdout, "\n  %d exercises in %d modules, from Hello, Go to Kubernetes controllers.\n\n", len(c.Exercises), len(c.Modules))
	if dir != "." {
		fmt.Fprintf(stdout, "  %s\n", tui.Key("cd "+dir))
	}
	fmt.Fprintf(stdout, "  %s   %s\n\n", tui.Key("gopherlings"), tui.Dim("start the first exercise"))
	fmt.Fprintln(stdout, tui.Dim("  Tip: open this folder in your editor in a second window. Every exercise is in exercises/, numbered in order."))
	return nil
}

func watch(s *session.Session, stdout io.Writer) error {
	m := tui.New(s)
	_, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	m.Close()
	if err != nil {
		return err
	}
	if m.Err() != nil {
		return m.Err()
	}
	done, total := s.DoneCount(), len(s.Course.Exercises)
	fmt.Fprintf(stdout, "%s %s\n", tui.Notice("ok", fmt.Sprintf("Progress saved: %d of %d exercises done.", done, total)), tui.Dim("Run gopherlings to pick up where you left off."))
	return nil
}

func check(s *session.Session, e *course.Exercise, stdout io.Writer) error {
	if err := s.Prepare(e); err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	r, err := s.Check(ctx, e)
	if err != nil {
		return err
	}
	if r.Status == workspace.Passed {
		if err := s.Complete(e); err != nil {
			return err
		}
	}
	fmt.Fprint(stdout, tui.CheckReport(s, e, r))
	if r.Status != workspace.Passed {
		return errSilent
	}
	return nil
}

func version() string {
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" {
		return info.Main.Version
	}
	return "dev"
}
