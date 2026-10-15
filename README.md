# Groundwork

My path from zero to writing real software: Go first, then the tools and
systems around it (Linux, Git, SQL, containers, CI/CD, Kubernetes, the
cloud), all in one terminal app.

Every exercise is a small program or file that's nearly finished. I fix it,
save, and the tests re-run straight away and tell me exactly what's still
wrong. Alongside the exercises, guides explain the computer science
underneath, with questions, key terms and vetted videos.

**400 exercises · 24 modules · 33 guides · about 186 hours**

- [Daily use](#daily-use)
- [What's in the course](#whats-in-the-course)
- [The screen and keys](#the-screen-and-keys)
- [Commands](#commands)
- [How it helps me learn](#how-it-helps-me-learn)
- [Away from the keyboard](#away-from-the-keyboard)
- [Where things are](#where-things-are)
- [Setting up a new machine](#setting-up-a-new-machine)
- [Troubleshooting](#troubleshooting)
- [Changing the course](#changing-the-course)

## Daily use

```bash
groundwork                 # works from any folder
```

Next to it, open the exercises in the editor:

```bash
nvim ~/code/golang/groundwork/course/exercises
```

1. The **Results** screen shows the exercise I'm on, which file to edit, and
   the failing tests.
2. `i` opens the instructions. I change the code where it says `TODO`.
3. **Save.** The tests re-run by themselves.
4. When everything passes: explain to myself why it works, press `s` to
   compare with the reference solution, then `n` for the next exercise.

Stuck? `h` gives a hint for the task I'm on, and each press gives a stronger
one (Nudge → Closer → Answer).

**A good session (25–45 minutes):**

1. Flashcards on the phone (5 minutes).
2. A due review, if the results screen says *Review due*: press `v`.
3. One section of a guide (`g`), answering its *Check yourself* questions
   before reading the answers.
4. New exercises. Try each task before taking a hint.
5. Stop on a pass, so the next session starts with a win.

[LEARNING.md](LEARNING.md) has the longer version and the research behind it.

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
round, and a small realistic job. Most modules end with a *practice*
exercise that combines their ideas, or a bigger multi-file *project*.

Not everything is Go. In the shell, Git, SQL, container and Kubernetes
modules I edit a `.sh` script, a `.sql` query, a Dockerfile, a CI workflow or
a YAML manifest instead, and Go tests check it the same way. Some exercises
have an optional *Run it for real* section for trying it on Docker or a
local Kubernetes cluster.

[LESSON_PLAN.md](LESSON_PLAN.md) lists every exercise with its time estimate.

## The screen and keys

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

On a wide window (about 100+ columns) the sidebar shows context for the
view: the tasks next to the results, an outline next to a document, stats
next to the course list. The bottom line always lists the keys that work
right now.

| Key | What it does |
|---|---|
| *save a file* | re-run the tests |
| `i` | instructions |
| `h` | a hint for the task I'm stuck on; press again for a stronger one; `1`–`9` picks another task |
| `s` | the reference solution: after passing, or after a task's Answer hint |
| `l` | progress: stats, a 12-week activity grid, and every exercise (`enter` opens one) |
| `g` | guides (`enter` reads, `m` marks read, `o` opens in the editor) |
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

Links in the instructions and guides are clickable titles (in Ghostty, and
in tmux with the hyperlinks feature on).

## Commands

```text
groundwork               the watcher: current exercise, re-run on save
groundwork list [--all]  progress through the course (--all: every exercise)
groundwork stats         streak, this week, today, and the last 12 weeks
groundwork check [EX]    run one exercise's tests once
groundwork reset EX      put an exercise back to its starter code
groundwork guides [G]    list the guides, or read one: groundwork guides concurrency
groundwork cards [--all] flashcards from the guides, for Anki
groundwork init [DIR]    create the course folder (first time only)
groundwork version
```

`EX` is an exercise number (`12`) or folder name (`012_constants`). `G` is a
guide number (`17`) or name (`concurrency`). Every command finds the course
from any folder; `$GROUNDWORK_COURSE` points it somewhere else.

## How it helps me learn

| Technique | Where it shows up |
|---|---|
| **Worked examples, then less help** | Each instructions page fully works through an example that's different from the task; later exercises in a concept give less help |
| **Instant, specific feedback** | Tests re-run on every save and say what I returned and what was wanted |
| **Hints as a ladder** | Nudge, then Closer, then Answer, one task at a time |
| **Explaining it back** | After every pass: say why the code works before moving on |
| **Spaced reviews** | Practice exercises, and any exercise I passed after seeing the solution, come back 2, 7 and 30 days later to solve from scratch (`v`). My old code is saved in `course/.groundwork/attempts/`. Using the solution during a review restarts the schedule |
| **Flashcards** | Every guide's questions and key terms become an Anki deck |
| **Visible progress** | Streak, days this week, a 12-week grid, and ✓ marks on exercises and guides |

## Away from the keyboard

Reading and flashcards work on the phone; exercises wait for the computer.

- **Guides:** open [content/guides/](content/guides/) on GitHub. It renders
  the tables, code and links, including each guide's *Watch* videos.
- **Flashcards:** `groundwork cards` writes `course/groundwork-cards.txt`
  with the guides up to my current module (`--all` for every guide). AirDrop
  it to the iPhone and open it with AnkiMobile. Each guide gets its own deck
  under *Groundwork*. Exporting again as I progress updates the cards and
  keeps their review history.
- Back at the computer, `m` in the guides marks what I've read, and the
  results screen suggests the next unread guide.

## Where things are

```text
groundwork/                      this repository
├── course/                      MY WORK (not in git)
│   ├── exercises/               every exercise, numbered in order: 01_basics/001_hello_go/ ...
│   ├── solutions/               reference solutions, written as they unlock
│   ├── guides/                  the guides as Markdown, for the editor
│   ├── groundwork-cards.txt     flashcards, after groundwork cards
│   └── .groundwork/             progress.json and review attempts
├── content/                     the course itself
│   ├── course.json              the order of modules and exercises
│   ├── packs/                   one source pack per exercise: starter, tests, README, hints, solution
│   ├── guides/                  the guides
│   └── AUTHORING.md             how exercises and guides are written
├── main.go, internal/           the groundwork program
├── LESSON_PLAN.md               the full syllabus
└── LEARNING.md                  how to study with it
```

Each exercise is its own Go module, so the editor only shows problems from
the exercise that's open. When the course changes, files I haven't edited
are refreshed and files I have edited are never touched; if exercises are
renumbered, my folders move with them.

`course/` isn't in git, so my progress and code aren't backed up with the
repository. To back it up too, remove `/course/` from `.gitignore`.

## Setting up a new machine

Needs Go 1.26 or newer ([go.dev/dl](https://go.dev/dl/)) and git.

```bash
git clone https://github.com/pjanderson/groundwork ~/code/golang/groundwork
cd ~/code/golang/groundwork
go install .            # puts groundwork in $(go env GOBIN) or ~/go/bin
groundwork init         # creates course/ (first time only)
groundwork
```

After pulling changes, run `go install .` again; `groundwork` updates
`course/` the next time it starts.

## Troubleshooting

| Problem | Fix |
|---|---|
| "there's no Groundwork course" | Run `groundwork init` in the repository, or set `GROUNDWORK_COURSE` to the course folder |
| A file I edited didn't update with the course | That's deliberate. `groundwork reset EX` restores the starter code (my old version is overwritten) |
| Links aren't clickable inside tmux | Add `set -as terminal-features ",xterm-ghostty:hyperlinks"` to tmux.conf and restart tmux |
| `kind create cluster` fails waiting for a log line (NixOS) | Set `virtualisation.docker.logDriver = "json-file";` and `sudo nixos-rebuild switch` |
| The first Kubernetes or AWS exercise is slow | It downloads client-go or the AWS SDK once; later runs are cached |
| "set up by a newer groundwork" | Run `go install .` in the repository to update the program |

## Changing the course

The course is `content/course.json` plus one pack per exercise in
`content/packs/<name>/`. [content/AUTHORING.md](content/AUTHORING.md)
covers the formats and the learning principles every pack follows.

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
| `internal/progress` | `course/.groundwork/progress.json` |
| `internal/cards` | Anki flashcards from the guides |
| `internal/tui` | the terminal interface and printed reports |
