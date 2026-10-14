# Groundwork

A personal computer-science and software-engineering course in your
terminal, built for someone starting from zero and heading into DevOps and
platform engineering. Go is the spine: you learn it by fixing small programs
until their tests pass, from your first `return "hello"` to a Kubernetes
controller. Around it are tracks on the tools of the trade: **the Linux
shell, Git, SQL databases, software engineering practice, containers and
CI/CD, and system design**. Every exercise, whether a Go function, a bash
script, a SQL query, a Dockerfile or a CI workflow, is checked the same way:
save the file and the tests re-run.

Groundwork takes two ideas from [rustlings](https://github.com/rust-lang/rustlings):
exercises that are nearly finished, and a watcher that re-runs the tests every
time you save. From [Exercism's Go track](https://exercism.org/tracks/go/concepts)
it takes concept-first lessons, a README with numbered tasks and worked
examples, and tests you can read. Each concept gets **several exercises**: an
introduction, a "fix the bug" round, and a small realistic task. Every module
ends with a practice exercise, and every phase ends with a multi-file project.

Alongside the exercises, **29 guides** teach the ideas underneath them: how
a program runs, state, functions, debugging, data structures, Big-O, memory,
abstraction, errors, testing, the shell, Git, databases, code quality and
review, containers, CI/CD, the operating system, networking, concurrency,
distributed systems, system design and scaling, and more. Each ends with
questions to check yourself, key terms, and vetted videos to watch. Each
module points to its guides. Read them in the watcher (`g`), with `groundwork guides`, or in
your editor ([content/guides/](content/guides/), or `guides/` in your course
folder).

See [LESSON_PLAN.md](LESSON_PLAN.md) for the full syllabus and a suggested
schedule.

## Install

You need Go 1.26 or newer ([go.dev/dl](https://go.dev/dl/)).

```bash
cd groundwork
go install .          # installs the groundwork binary into $(go env GOBIN) or ~/go/bin
```

## Start

Everything lives in this one folder: the tool, the course and your work.

```bash
cd groundwork
groundwork init      # first time only: sets up exercises/ and your progress here
groundwork           # opens your current exercise and watches for saves
```

Open this folder in your editor in a second window (VS Code, GoLand,
Neovim…). Groundwork shows which file to edit. Failing tests come first,
each with its exact failure message. Save the file and the checks run again.
When every test passes, press **s** to compare your code with the reference
solution, and **n** to move to the next exercise. Truly stuck? After the last
hint, **s** shows the solution early (you confirm first).

```text
groundwork  ·  module 1 of 20: Basics                     ━━━━━━━━━━━━━━━━━━━━  1/362 done

002  Variables                                                         ✗ 2 of 3 failing
edit  exercises/01_basics/002_variables/exercise.go

 Results   Instructions i   Hints h   Solution s   Exercises l   Guides g   Keys ?
─────────────────────────────────────────────────────────────────────────────────────────
✗ 2 of 3 tests failing

  ✗ TestImageRef
    ✗ github registry
        exercise_test.go:40: ImageRef("ghcr.io/acme", "api", "v1.4.2") = "", want
        "ghcr.io/acme/api:v1.4.2"
  ✗ TestDeployPath
    ✗ api in production
        exercise_test.go:60: DeployPath("production", "api") = "/v1/environments/production",
        want "/v1/environments/production/services/api"

  ✓ TestServiceLabel

Not sure what to do? Press i for the instructions or h for a hint.
─────────────────────────────────────────────────────────────────────────────────────────
Checked in 0.3s. Save any file to re-run.
n next unfinished  ·  ←/→ prev/next  ·  r re-run  ·  e edit  ·  x reset  ·  q quit
```

The tab row shows which view you're looking at. The results always come first.

### Keys

| Key | Action |
|---|---|
| save a file | re-run the tests |
| `i` | instructions (the exercise README, rendered) |
| `h` | hint; press again for a stronger one |
| `s` | the solution: once you pass, or after the last hint |
| `l` | all exercises; `enter` opens one |
| `g` | guides: the computer science behind the exercises; `enter` reads one, `m` marks it read, `o` opens it in `$EDITOR` |
| `?` | every key, plus a guide to reading failures and compiler errors |
| `esc` | back to the results (from a guide: back to the guide list) |
| `n` | next unfinished exercise |
| `←` / `→` | previous / next exercise |
| `↑` `↓` `pgup` `pgdn` | scroll |
| `e` | open the exercise in `$EDITOR` |
| `r` | re-run the tests now |
| `x` `x` | reset the exercise to its starter code |
| `q` | quit; progress is saved |

### Commands

```text
groundwork               open the current exercise and watch for saves
groundwork init [DIR]    start a new course in DIR (default: this directory)
groundwork list [--all]  show your progress through the course
groundwork check [EX]    run one exercise's tests once (handy outside a terminal UI)
groundwork reset EX      restore an exercise's starter code
groundwork guides [G]    list the guides, or read one (groundwork guides concurrency)
groundwork cards [--all] flashcards from the guides, for Anki on your phone or desktop
```

`EX` is usually just the exercise number (`12`). A folder name such as
`012_variables` works too. A guide is its number (`17`) or name
(`concurrency`). Commands work from anywhere inside the course folder;
`groundwork guides` works anywhere.

## Away from the keyboard

Reading and reviewing work on a phone; exercises are for the computer.

- **Read the guides** on GitHub: open [content/guides/](content/guides/) in
  the repository. GitHub renders the tables, diagrams and code, and every
  guide's *Watch* section links to vetted videos.
- **Review with flashcards.** `groundwork cards` writes
  `groundwork-cards.txt`: every *Check yourself* question and *Key terms*
  row from the guides up to your current module (`--all` for every guide),
  one Anki deck per guide. On an iPhone, AirDrop or save the file to Files,
  open it with AnkiMobile, and it imports. In desktop Anki use File > Import
  and sync. Export again as you reach new modules: re-importing updates the
  cards and keeps your review history.
- **Back at the computer**, press `g` in the watcher and `m` on the guides
  you've read; the results screen suggests the next unread one.

## What's in this folder

```text
groundwork/
├── main.go, internal/, content/  the tool and the course content (you don't edit these)
├── LESSON_PLAN.md                the syllabus
├── .groundwork/progress.json    what you've finished
├── guides/                       the guides as Markdown, numbered (rewritten on updates;
│                                 keep your notes in your own files)
├── exercises/
│   ├── 01_basics/                one folder per module
│   │   ├── 001_hello_go/
│   │   ├── 002_variables/        a concept's introduction...
│   │   ├── 003_variables/        ...then its practice rounds, same topic name
│   │   ├── 004_variables/
│   │   │   ├── README.md         instructions
│   │   │   ├── exercise.go       you edit this
│   │   │   ├── exercise_test.go  the tests (read them!)
│   │   │   └── go.mod            every exercise is its own little Go module
│   │   └── …
│   ├── 02_control_flow/
│   └── …                         20 modules, 362 exercises, numbered in order
└── solutions/                    reference solutions, same layout, unlocked when you pass
```

Each exercise is its own folder and its own Go module, with its own tests.
Your editor only ever sees one exercise at a time, so the deliberate bugs in
exercises you haven't reached never show up in a workspace-wide problems
list. The number gives the order.

Every exercise folder exists from the start. When the course is updated, any
file you haven't edited is refreshed automatically. Files you've changed are
never touched. If a course update reorders exercises, your folders move with
them. `groundwork reset EX` restores the starter code.

Groundwork records the format of the course folder in
`.groundwork/progress.json`. An older `groundwork` binary refuses to touch
a folder set up by a newer one and tells you to reinstall, so it can't undo
changes. If the Go toolchain is older than 1.26, it says so up front.

Checks run `go test` (with `-race` in the concurrency modules), then
`go vet`. Unformatted code earns a gentle reminder, not a failure. The
Kubernetes and AWS modules use the real `client-go` and `aws-sdk-go-v2`, so
the first exercise there downloads those modules once. No test ever talks to a
real cluster or cloud account.

## Writing exercises

The course is `content/course.json` (its order) plus one pack per exercise in
`content/exercises/<name>/`. See [content/AUTHORING.md](content/AUTHORING.md)
for the pack format and style rules. To check that a pack starts failing and
that its solution passes:

```bash
go test ./content -run 'TestPacks/^variables$' -v
```

## Development

```bash
go test -short ./...    # engine tests, plus every stdlib pack (starter fails, solution passes)
go test ./...           # also builds the client-go and AWS packs (slower; downloads modules once)
```

| Path | What it does |
|---|---|
| `main.go` | commands |
| `internal/course` | loads course.json, READMEs and hints |
| `internal/session` | the learner's course folder: create exercises, check, record progress, write solutions |
| `internal/workspace` | copies packs, runs `go test -json` and `go vet`, watches files |
| `internal/progress` | `.groundwork/progress.json` |
| `internal/tui` | the Bubble Tea interface and plain-text reports |
