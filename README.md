# Groundwork

Learn Go and the tools around it by fixing small programs until their tests
pass. Groundwork is a course that runs in your terminal. It starts at your
first line of code and ends with the Go that DevOps and platform engineers
write every day: CLIs, log and config processing, HTTP services, concurrency,
Kubernetes controllers and AWS automation.

**400 exercises · 24 modules · 33 guides · about 186 hours**

## Quick start

Needs [Go 1.26 or newer](https://go.dev/dl/), git and bash.

```bash
go install github.com/brzrkr-io/groundwork@latest
groundwork init ~/groundwork    # create your course folder (once)
groundwork                      # start; works from any folder
```

Open `~/groundwork/exercises` in your editor in a second window.

To update, run the `go install` line again. Your course folder picks up the
changes the next time `groundwork` starts.

## How it works

Every exercise is a small program or file that's nearly finished. You fix the
parts marked `TODO` in your own editor and save. Groundwork re-runs the tests
on every save and shows exactly what's still wrong: what your code returned
and what was wanted.

```text
groundwork  1 · Basics                                         3-day streak  ·  ━━━━━━━━━━━  2/400

003  Variables                                                                  ✗ 2 of 3 failing
 Results   Instructions i   Hints h   Solution s   Progress l   Guides g   Keys ?
╭─ Results ───────────────────────────────────────────────────┬─ This exercise ────────────────╮
│  ✗ 2 of 3 tests failing                                     │ TASKS                          │
│                                                             │   ✓ 1 Label a service          │
│    ✗ TestImageRef                                           │ ▸ ✗ 2 Build an image reference │
│        exercise_test.go:40: ImageRef("ghcr.io/acme", "api", │   ✗ 3 Build a deploy path      │
│        "v1.4.2") = "", want "ghcr.io/acme/api:v1.4.2"       │                                │
│    ✗ TestDeployPath                                         │ EDIT                           │
│        exercise_test.go:60: DeployPath("production", "api") │ 003_variables/exercise.go      │
│        = "/v1/environments/production", want                │ press e to open it             │
│        "/v1/environments/production/services/api"           │                                │
│    ✓ TestServiceLabel                                       │ HINTS                          │
│                                                             │ 0 of 9 seen · h                │
╰─────────────────────────────────────────────────────────────┴────────────────────────────────╯
Checked in 0.3s · save any file to re-run
n next  ·  ←/→ prev/next  ·  r re-run  ·  e edit  ·  x reset  ·  q quit
```

1. **Results** shows the current exercise, the file to edit and the failing
   tests.
2. `i` opens the instructions: a worked example, then your task.
3. Edit and save. The tests re-run by themselves.
4. Stuck? `h` gives a hint for the task you're on.
5. When everything passes, explain to yourself why it works, press `s` to
   compare with the reference solution, then `n` for the next exercise.

## What it does

- **Hints in three steps.** Nudge, Closer, then Answer, one task at a time,
  so you only take as much help as you need.
- **Reference solutions** unlock once you pass, or after a task's Answer hint.
- **Spaced reviews.** Practice exercises, and any exercise you passed after
  seeing the solution, come back 2, 7 and 30 days later to solve from scratch
  (`v`). Your earlier code is saved first.
- **Guides.** 33 short guides on the computer science underneath, each with
  *Check yourself* questions, key terms and vetted videos. Read them in the app
  (`g`), in your editor (`guides/` in your course folder) or on GitHub in
  [content/guides/](content/guides/).
- **Flashcards.** `groundwork cards` turns the guides' questions and key terms
  into an Anki import file, one deck per guide. Export again as you progress:
  the cards update and keep their review history.
- **Progress.** Streak, days this week, a 12-week activity grid, and ✓ marks
  on exercises and guides.
- **More than Go.** In the shell, Git, SQL, container, CI and Kubernetes
  modules you edit a `.sh` script, a `.sql` query, a Dockerfile, a GitHub
  Actions workflow or a YAML manifest, and Go tests check it the same way.
- **Nothing to set up.** SQL runs on SQLite through a pure-Go driver,
  Kubernetes exercises use client-go's fake clientset and AWS exercises use
  fakes, so the tests need no database server, Docker, cluster or AWS account.
  Optional *Run it for real* sections show how to try the same thing on
  Docker, a local [kind](https://kind.sigs.k8s.io/) cluster or a real AWS
  account.
- **Your edits are safe.** Course updates refresh files you haven't touched
  and never overwrite ones you've edited. If exercises are renumbered, your
  folders move with them. Each exercise is its own Go module, so your editor
  only shows problems from the one that's open.

## What's in the course

| Phase | Module | Exercises | What it covers |
|---|---|---:|---|
| Go foundations | Basics | 22 | Variables, constants, numbers, strings, fmt, reading Go code and compiler errors |
| | Control flow and functions | 25 | Booleans, if, switch, loops, functions, multiple returns, closures |
| | Collections | 31 | Arrays, slices, maps, sorting, runes and UTF-8 |
| | Types and methods | 19 | Structs, pointers, methods, custom types, stringers |
| | Interfaces and generics | 13 | Interfaces, embedding, type switches, generics |
| | Errors | 19 | Creating, wrapping and inspecting errors, defer, panics |
| | Testing | 15 | Table tests, subtests, golden files, fakes, benchmarks, fuzzing |
| Everyday tools | Linux and the shell | 17 | Bash, files, redirection, environment, text tools, processes and permissions |
| | Git and version control | 14 | Commits, staging, branches, merges, rebase, undo, remotes, bisect |
| Practical Go | Text and data | 16 | Scanning, regular expressions, time, JSON, CSV, templates |
| | Databases and SQL | 14 | Queries, joins, NULL, schemas, indexes, transactions, database/sql |
| | Software engineering practice | 11 | Naming, refactoring, guard clauses, code review, design docs |
| | CLI tools and the OS | 19 | Arguments, flags, subcommands, streams, files, config |
| | Packages and project layout | 12 | Packages, exported names, internal/, embed, build info |
| Building services | Networking and HTTP | 21 | IPs, TCP, DNS, HTTP clients and servers, JSON APIs, middleware, TLS |
| | Concurrency | 25 | Goroutines, channels, select, mutexes, context, subprocesses, signals, pipelines |
| | Production services | 21 | Logging, metrics, tracing, health checks, graceful shutdown, retries, rate limits |
| | Containers and CI/CD | 13 | Dockerfiles, multi-stage builds, image security, GitHub Actions, releases |
| | System design | 14 | Capacity, SLOs, load balancing, sharding, consistent hashing, caches, queues |
| Platform engineering | Kubernetes fundamentals | 12 | Pods, Deployments, Services, config, probes, resources, rollouts, debugging |
| | Kubernetes with client-go | 19 | Typed objects, clientsets, watches, informers, workqueues, a controller |
| | Cloud fundamentals | 11 | ARNs, IAM policies, assume role, VPCs, security groups, CloudTrail, costs |
| | Cloud automation with the AWS SDK | 15 | aws-sdk-go-v2, paginators, API errors, S3, tagging, drift detection |
| Capstones | Capstones | 2 | **go links**: a service built, containerised, deployed to Kubernetes and shipped by CI. **kubepolicy**: a manifest policy checker |

Each concept usually gets three exercises: an introduction, a "fix the bug"
round and a small realistic job. Most modules end with a *practice* exercise
that combines their ideas, or a bigger multi-file *project*.

[LESSON_PLAN.md](LESSON_PLAN.md) lists every exercise with its time estimate.
[LEARNING.md](LEARNING.md) covers how to study with Groundwork and the
research behind it.

## Keys

The bottom line of the screen always lists the keys that work right now. On a
wide terminal (about 100+ columns) a sidebar shows the tasks, a document
outline or your stats, depending on the view.

| Key | What it does |
|---|---|
| *save a file* | re-run the tests |
| `i` | instructions |
| `h` | hint for the current task; press again for a stronger one; `1`–`9` picks another task |
| `s` | reference solution (after passing, or after a task's Answer hint) |
| `l` | progress: stats, activity grid and every exercise (`enter` opens one) |
| `g` | guides (`enter` reads, `m` marks read, `o` opens in your editor) |
| `?` | every key, a session plan, and how to read test failures and compiler errors |
| `esc` | back to the results |
| `n` | next unfinished exercise |
| `←` `→` | previous / next exercise |
| `↑` `↓` `pgup` `pgdn` | scroll |
| `[` `]` | previous / next section of a document |
| `e` | open the file to edit in `$EDITOR` |
| `r` | re-run the tests now |
| `x` `x` | reset the exercise to its starter code |
| `v` `v` | start a due review |
| `q` | quit (progress is saved) |

## Commands

```text
groundwork               open the current exercise and re-run its tests on save
groundwork init [DIR]    create a course folder (once)
groundwork list [--all]  progress through the course (--all: every exercise)
groundwork stats         streak, this week, today, and the last 12 weeks
groundwork check [EX]    run one exercise's tests once (default: current)
groundwork reset EX      put an exercise back to its starter code
groundwork guides [G]    list the guides, or read one
groundwork cards [--all] flashcards for Anki (--all: every guide, not just up to your module)
groundwork version
```

`EX` is an exercise number (`7`) or folder name (`007_constants`). `G` is a
guide number (`1`) or name (`concurrency`).

## Your course folder

```text
~/groundwork/
├── exercises/            every exercise, numbered in order: 01_basics/001_hello_go/ ...
├── solutions/            reference solutions, written as they unlock
├── guides/               the guides as Markdown, for your editor
├── groundwork-cards.txt  flashcards, after groundwork cards
└── .groundwork/          your progress and saved review attempts
```

Commands find this folder from anywhere: the folder you're in or one of its
parents, then `$GROUNDWORK_COURSE`, then wherever you last ran
`groundwork init`. Nothing backs it up; put it in git if you want to keep
your work.

## Troubleshooting

| Problem | Fix |
|---|---|
| `groundwork: command not found` | Add `$(go env GOPATH)/bin` (or `$GOBIN`, if set) to your `PATH` |
| "there's no Groundwork course in this directory or its parents" | Run `groundwork init DIR`, or set `GROUNDWORK_COURSE` to your course folder |
| "the exercises need Go 1.26 or newer" | Update Go from [go.dev/dl](https://go.dev/dl/) |
| "this course folder was set up by a newer groundwork" | Update groundwork (see [Quick start](#quick-start)) |
| A file you edited didn't update with the course | That's deliberate. `groundwork reset EX` restores the starter code and overwrites your version |
| The first Kubernetes or AWS exercise is slow | It downloads client-go or the AWS SDK once; later runs are cached |
| Links aren't clickable inside tmux | Add `set -as terminal-features ",xterm-ghostty:hyperlinks"` to `tmux.conf`, with your terminal's `$TERM` in place of `xterm-ghostty`, and restart tmux |

## Contributing

Work from a clone:

```bash
git clone https://github.com/brzrkr-io/groundwork
cd groundwork
go install .        # again after every pull
groundwork init     # inside the repository this creates course/, which git ignores
```

The course is `content/course.json` (the order of modules and exercises), one
pack per exercise in `content/packs/<name>/` (starter code, tests, README,
hints, solution) and the guides in `content/guides/`, all embedded in the
binary. [content/AUTHORING.md](content/AUTHORING.md) covers the formats and
the learning principles every pack follows.

```bash
go test ./content -run 'TestPacks/^variables$' -v   # one pack: starter fails, solution passes
go test -short ./...                                  # the program and every standard-library pack
go test ./...                                         # everything, incl. Kubernetes and AWS packs
go test ./content -run TestLessonPlan -args -update   # regenerate LESSON_PLAN.md after reordering
```

| Path | What it does |
|---|---|
| `main.go` | commands |
| `internal/course` | loads course.json, packs, hints and guides |
| `internal/session` | the course folder: exercises, checks, progress, reviews, stats |
| `internal/workspace` | copies packs, runs `go test` and `go vet`, watches files |
| `internal/progress` | `.groundwork/progress.json` |
| `internal/cards` | Anki flashcards from the guides |
| `internal/tui` | the terminal interface and printed reports |
