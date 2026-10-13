# Gopherlings

Learn Go by fixing small programs until their tests pass, from your first
`return "hello"` to a Kubernetes controller. The course is built for people
heading into DevOps and platform engineering.

Gopherlings takes two ideas from [rustlings](https://github.com/rust-lang/rustlings):
exercises that are nearly finished, and a watcher that re-runs the tests every
time you save. From [Exercism's Go track](https://exercism.org/tracks/go/concepts)
it takes concept-first lessons, a README with numbered tasks and worked
examples, and tests you can read. Each concept gets **several exercises**: an
introduction, a "fix the bug" round, and a small realistic task. Every module
ends with a practice exercise, and every phase ends with a multi-file project.

See [LESSON_PLAN.md](LESSON_PLAN.md) for the full syllabus and a suggested
schedule.

## Install

You need Go 1.26 or newer ([go.dev/dl](https://go.dev/dl/)).

```bash
cd gopherlings
go install .          # installs the gopherlings binary into $(go env GOBIN) or ~/go/bin
```

## Start

Everything lives in this one folder: the tool, the course and your work.

```bash
cd gopherlings
gopherlings init      # first time only: sets up exercises/ and your progress here
gopherlings           # opens your current exercise and watches for saves
```

Open this folder in your editor in a second window (VS Code, GoLand,
Neovim…). Gopherlings shows which file to edit. Failing tests come first,
each with its exact failure message. Save the file and the checks run again.
When every test passes, press **s** to compare your code with the reference
solution, and **n** to move to the next exercise. Truly stuck? After the last
hint, **s** shows the solution early (you confirm first).

```text
gopherlings  ·  module 1 of 14: Basics                     ━━━━━━━━━━━━━━━━━━━━  1/284 done

002  Variables                                                         ✗ 2 of 3 failing
edit  exercises/01_basics/002_variables/exercise.go

 Results   Instructions i   Hints 0/3 h   Solution s   All exercises l   Keys ?
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
| `?` | every key, plus a guide to reading failures and compiler errors |
| `esc` | back to the results |
| `n` | next unfinished exercise |
| `←` / `→` | previous / next exercise |
| `↑` `↓` `pgup` `pgdn` | scroll |
| `e` | open the exercise in `$EDITOR` |
| `r` | re-run the tests now |
| `x` `x` | reset the exercise to its starter code |
| `q` | quit; progress is saved |

### Commands

```text
gopherlings               open the current exercise and watch for saves
gopherlings init [DIR]    start a new course in DIR (default: this directory)
gopherlings list [--all]  show your progress through the course
gopherlings check [EX]    run one exercise's tests once (handy outside a terminal UI)
gopherlings reset EX      restore an exercise's starter code
```

`EX` is usually just the exercise number (`12`). A folder name such as
`012_variables` works too. Commands work from anywhere inside the course
folder.

## What's in this folder

```text
gopherlings/
├── main.go, internal/, content/  the tool and the course content (you don't edit these)
├── LESSON_PLAN.md                the syllabus
├── .gopherlings/progress.json    what you've finished
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
│   └── …                         14 modules, 284 exercises, numbered in order
└── solutions/                    reference solutions, same layout, unlocked when you pass
```

Each exercise is its own folder and its own Go module, with its own tests.
Your editor only ever sees one exercise at a time, so the deliberate bugs in
exercises you haven't reached never show up in a workspace-wide problems
list. The number gives the order.

Every exercise folder exists from the start. When the course is updated, any
file you haven't edited is refreshed automatically. Files you've changed are
never touched. If a course update reorders exercises, your folders move with
them. `gopherlings reset EX` restores the starter code.

Gopherlings records the format of the course folder in
`.gopherlings/progress.json`. An older `gopherlings` binary refuses to touch
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
| `internal/progress` | `.gopherlings/progress.json` |
| `internal/tui` | the Bubble Tea interface and plain-text reports |
