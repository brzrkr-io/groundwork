# Authoring Groundwork exercises

The course is `course.json` (module and exercise order) plus one pack per
exercise in `exercises/<name>/`. Exercise numbers come from their position in
`course.json`, so packs are named without numbers and reordering costs nothing.

The learner is a **beginner** who wants to do DevOps and platform work. Every
exercise should feel doable after reading its README. A sense of progress
matters more than covering every corner of a topic.

## Learning principles

Every pack and guide follows these (see ../LEARNING.md for the research):

1. **Worked example first, then fade.** The README's Learn section fully
   demonstrates each construct with a different example from the task;
   tasks then need more and more independence.
2. **One new idea at a time, quick first win.** Task 1 takes minutes. No
   term or tool the learner hasn't met, judging by the course order.
3. **No hidden requirements.** Everything a test checks is in the task,
   with a literal example of the expected result.
4. **Feedback that teaches.** Failure messages say what's wrong and point
   toward the fix.
5. **Any correct answer passes.** Tests check behaviour and results, not one
   spelling of the solution, and are hermetic and deterministic.
6. **Hints are a ladder:** Nudge (a pointer, no code), Closer (the syntax),
   Answer (the task's code).
7. **Realistic bugs** in fix-the-bug rounds, with the README teaching how
   to diagnose them.
8. **Short pieces** with a bold Goal per task: the learner has ADHD.

## Pack layout

```text
exercises/variables/
├── README.md                 instructions (copied into the learner's folder)
├── hints.md                  progressive hints (shown only in the TUI)
├── exercise.go.tmpl          starter code the learner edits
├── exercise_test.go.tmpl     tests the learner reads but does not edit
├── testdata/…                optional fixtures (non-Go files need no suffix)
└── solution/
    └── exercise.go.tmpl      replaces the starter file once the exercise is done
```

- Go files and `go.mod` carry a `.tmpl` suffix. It keeps the Go toolchain and
  `go:embed` away from them. The suffix is stripped when files are copied.
  Other files (`go.sum`, `.json`, `.txt`, `.golden`) need no suffix.
- `solution/` mirrors the pack layout. Its files are written over the starter
  files. Include only the files that change.
- Every exercise is its own Go module in the learner's folder. Packs without a
  `go.mod.tmpl` get a generated one (`module groundwork.local/<folder>`,
  `go 1.26`). Packs with **third-party dependencies** or **subpackages**
  ship their own `go.mod.tmpl` (module path `groundwork.local/<name>`,
  `go 1.26`) and, with dependencies, a `go.sum`.
- Learner folders are `exercises/<NN>_<module dir>/<NNN>_<topic>`, where the
  topic is the pack name without a `_2`/`_3` suffix. Numbers come from the
  pack's position in `course.json`; each module there has a `dir` and an
  `outcome` (shown in LESSON_PLAN.md). After changing course.json, run
  `go test ./content -run TestLessonPlan -args -update`.
- File and directory names must not start with `.` or `_`, because `go:embed`
  skips them.

## README.md

The learner has ADHD and is new to programming. They lose track in long text,
so a README must be easy to scan, and every task should read like a short
checklist. They also want to understand **why** Go code looks the way it
does: what Go requires, what is convention, and what is their own choice.
Teach that reasoning. Don't give away the answer.

```markdown
# Title

One or two sentences: what this is and why an ops engineer cares.

**New here:** `const`, constant expressions.   (or: **New here:** nothing; practice)

## Learn

The concept, taught from zero, in short sections (### headings), each with a
small ```go example. Explain every piece of syntax or API the task needs that
earlier exercises did not cover. About 200–600 words for concept exercises;
practice exercises recap in a few lines; projects describe the architecture
and include a file map.

> **Why is it written like this?** One short note per new piece of syntax:
> what Go requires (the compiler rejects anything else), what is convention
> (gofmt or the community expects it), and what is the author's choice
> (names, order, line breaks).

## Your task

A one- or two-sentence scenario.

### 1. Imperative task title

**Goal:** one sentence: which function, and what it must return or do.

The exact expected behaviour, followed by a worked example:

    ```go
    ServiceLabel("api", "production")
    // => "production/api"
    ```

**Think it through:**
- 1–3 short questions or steps that point at *what to look at* (which
  parameters, which fixed text, which operator), never the code itself.

### 2. …

## Learn more

- [Official doc title](https://go.dev/…)  (2–4 links: go.dev, pkg.go.dev,
  the spec, the Go blog, gobyexample.com, kubernetes.io, docs.aws.amazon.com)
```

Rules:

- **No hidden requirements.** Every string, format, edge case and error
  condition the tests check is stated in the README. A learner who reads the
  README carefully never has to reverse-engineer the tests.
- **Don't give the answer away.** The README shows the result and teaches the
  syntax with *different* examples, never the line the learner must write.
  Solutions live in the hints' Answer level.
- **Easy to scan.** Short paragraphs of 1–3 sentences. Put each new term in
  *italics* or `code` the first time. Use tables to compare options. Keep the
  details below the **Goal** line, so a learner who only reads the Goal and
  the example knows what to do.
- **Name the parts.** When a task builds something from pieces, say in words
  which pieces come from parameters or constants and which are fixed text.
  When a task changes a value, say which line or name holds it.
- Tasks go from easiest to hardest. Task numbers match the `TODO(n)` markers,
  and task titles match hints.md exactly.
- Write in the second person. Be friendly, direct and concrete. No filler and
  no marketing.
- Refer to earlier exercises by title ("as in *Slices*"), never by number.

## hints.md

Every task has its own hints, in three levels. The learner presses `h` for
the task they're stuck on and sees one level more each time. The format is
fixed, and TestPacks rejects anything else:

````markdown
## Task 1: Name a pod

A pod name is three parameters with a dash between each pair. Which parts are
parameters and which are fixed text?

### Closer

Fixed text goes in quotes (`"-"`); parameters don't. Join every piece with `+`:
`deployment + "-" + ...`.

### Answer

```go
return deployment + "-" + hash + "-" + suffix
```
````

- One `## Task N: Title` per README task (`### N. Title`), in order, with
  **exactly** the README's title text.
- **Nudge** (the text under the task heading): what to look at or think
  about, in plain words. It may name the function or operator, but it gives no
  code that solves the task. Write 1–3 sentences.
- **`### Closer`**: the specific syntax or API, with a partial snippet or the
  pattern. It still leaves the learner a step to take.
- **`### Answer`**: near-complete code for **this task only**, following the
  reference solution's approach, plus at most one sentence on why it works.
- No other `##`/`###` headings. Code blocks may use any content.

## Starter code

- Mark every place the learner works with `// TODO(n): <what to do>`, where
  `n` is the task number. Say what to do, not how.
- Starters compile, unless the lesson is about reading compiler errors (only
  in the first two modules). In that case the README says so.
- Never leave hacks like `var _ = fmt.Errorf` to silence unused imports. If
  the solution needs an import the starter doesn't use, the TODO or README
  tells the learner to add it.
- Every exported identifier has a doc comment. Code is gofmt-clean.
- When the learner adds a new exported declaration, the starter provides its
  doc comment ready, after the TODO and a blank line:

  ```go
  // TODO(1): Write a function named Shout. ...
  // Write your code directly under the "// Shout ..." comment below.

  // Shout returns word followed by an exclamation mark.
  ```

  Otherwise the learner types under the TODO, and the editor flags "comment on
  exported function Shout should be of the form "Shout ..."". If the learner
  renames something instead, the TODO says to rename it in the comment too.
  TestPacks enforces this.
- Every non-test package (starter and solution) has a package comment, such as
  `// Package variables is the Groundwork exercise "Variables".` Otherwise
  editors warn on every file. TestPacks enforces this.
- Use a short package name that fits the topic (`package variables`, not
  `package exercise`).
- Use only concepts from earlier exercises, or concepts the README teaches.

## Tests

The tests are reading material for a beginner. Keep them simple.

- Write one `TestXxx` function per task, in task order, named after the
  behaviour (`TestServiceLabel`). The TUI lists them as a ✓/✗ checklist.
- Modules 1–2 tests use plain sequential `if` checks, one per case (no
  tables, loops or `t.Run`). Table-driven tests with `t.Run(name, …)` start
  in Module 3.
- Failure messages follow `Func(args) = got, want want`, using `%q` for
  strings.
- Give each behaviour at least two cases, so hard-coding an answer fails.
- Tests are deterministic and fast (well under 2s each, including `-race`).
  No internet. Loopback listeners and `httptest` are fine. Do not sleep for
  longer than about 50ms. Inject clocks, or use `testing/synctest`. Never
  depend on map order or goroutine scheduling.
- Use the standard library only, plus a pack's declared dependencies. No
  testify.

### Exercises where the learner writes tests

The **Testing** module grades learner-written tests with *mutation testing*.
The learner edits `exercise_test.go`. A separate `grader_test.go` re-runs the
learner's test inside the same test binary
(`exec.Command(os.Args[0], "-test.run=^TestX$")`). Through an environment
variable it swaps in known-buggy implementations, which the package exposes as
an unexported function variable. The learner's test must pass against the real
code and fail against every mutant. The README explains this in one paragraph.

## Solutions

The solution is shown to the learner after they finish, so it is a teaching
artifact. Make it idiomatic and as simple as possible: the code a good
reviewer would approve. Comment only non-obvious choices.

## Dependencies

Use exactly these versions so every pack shares one module cache:

| Module | Version |
|---|---|
| k8s.io/client-go, k8s.io/api, k8s.io/apimachinery | v0.37.1 |
| sigs.k8s.io/yaml | v1.6.0 |
| github.com/aws/aws-sdk-go-v2 | v1.47.3 |
| github.com/aws/aws-sdk-go-v2/config | v1.33.9 |
| github.com/aws/aws-sdk-go-v2/service/s3 | v1.114.3 |
| github.com/aws/aws-sdk-go-v2/service/ec2 | v1.338.4 |
| github.com/aws/smithy-go | v1.28.5 |
| modernc.org/sqlite | v1.60.1 |

To produce `go.mod.tmpl` and `go.sum`, copy the starter, tests and solution
into a scratch directory and run `go mod tidy`. Then copy `go.mod` back as
`go.mod.tmpl` and copy `go.sum` back as `go.sum`. Tests never touch the
network: use fake clientsets, fake API clients, or `httptest` servers.

## Exercises that aren't Go code

The tracks on the shell, Git, SQL, containers and CI are practised in their
own languages, but every pack is still checked by `go test`: the learner
edits a script, query or config file, and the pack's `exercise_test.go`
runs or reads it. The watcher re-runs on any saved file, and opens the first
file with a `TODO(n)` in it, whatever its type.

- **One file per task**, named for what it does and numbered in task order:
  `1_count_errors.sh`, `2_top_ips.sh`; `1_failed_deploys.sql`. Each test
  function checks one file, so the hints open on the right task.
- **Starter files** hold the task as a comment, any boilerplate (a shebang
  and `set -euo pipefail` for scripts), and a `# TODO(n): ...` (or
  `-- TODO(n): ...` in SQL) where the learner's lines go. The solution
  overlay in `solution/` replaces those files.
- **Fixtures** go in `testdata/`, and the README tells the learner to look
  at them: `testdata/app.log` is what their script reads. The learner can
  run their script by hand from the exercise folder, so tests run it there
  too, against the same fixtures (in a temporary copy if it writes files).
- **Tests are hermetic.** Run scripts with `exec.Command("bash", file)`,
  `LC_ALL=C`, `TZ=UTC`, and for Git `GIT_CONFIG_GLOBAL=/dev/null`,
  `GIT_CONFIG_NOSYSTEM=1` and fixed author and committer names, emails and
  dates. Use only bash, coreutils, grep, sed, awk, find, xargs, sort, uniq,
  cut, tr, wc, head, tail, diff and git: never jq, docker, a network or a
  database server.
- **Failure messages teach**, in the course's got/want style:
  `1_count_errors.sh printed "12", want "3"`; for multi-line output, show
  both in full. If the script exits non-zero, show its stderr: that's where
  bash explains the mistake.
- **Git** packs ship a `setup.sh` that builds the scenario repo in a
  directory given as `$1` (default `sandbox`). The README says to run
  `bash setup.sh` and explore `sandbox/` freely; tests run `setup.sh` in a
  temporary directory, then the learner's script inside the repo, then
  inspect the result with git commands.
- **SQL** packs use SQLite through `modernc.org/sqlite` (pure Go; see
  Dependencies), in memory: load `testdata/schema.sql` and
  `testdata/data.sql`, run the learner's query, and compare rows. Later
  packs call the database from Go with `database/sql`.
- **Config** packs (Dockerfiles, CI workflows) are read and checked by the
  test, not executed: parse the file in Go (`sigs.k8s.io/yaml` for YAML) and
  check each rule the task states. Names starting with `.` can't be
  embedded, so a workflow is `ci.yml`, not `.github/workflows/ci.yml`; the
  README says where it would live in a real repository.
- A pack with only `_test.go` Go files needs no package comment.

## Guides

Exercises teach Go. **Guides** teach the computer science underneath it:
how a program runs, state, abstraction, data structures, concurrency, and so
on. Each guide is one Markdown file, `guides/<name>.md`. A module lists the
guides to read with it under `"guides"` in `course.json`, and the order there
numbers them. The learner reads them in the watcher (`g`), with
`groundwork guides`, or in their course folder's `guides/`.

A guide teaches **ideas**, not syntax. Its Go examples use only what the
course has taught by the end of its module; for anything later, use plain
words or a diagram. It connects every idea to ops work (deploys, logs,
servers, pipelines, incidents).

```markdown
# Concurrency

**In one sentence:** concurrency is structuring a program as several tasks
that make progress independently, so waiting on one doesn't stall the rest.

**Why it matters:** a health checker that probes 500 hosts one by one takes
minutes; probing them concurrently takes about as long as the slowest one.

**You'll learn:** (3–6 bullets, the takeaways)

## (teaching sections: one idea each, short, ### subsections allowed)

## Common mistakes
## Check yourself      (numbered questions, at least three)
## Answers             (numbered, one per question, with the reasoning)
## Key terms           (a table: Term | Meaning)
## Watch               (2–5 videos: `- [Title](url) · Channel · N min: why`)
## Go deeper           (2–5 links, each with one line on why it's worth it)
```

- **Short pieces.** The learner has ADHD: paragraphs of 1–4 sentences, one
  idea per section, a bold lead-in where a skim reader needs one. Tables and
  `text` diagrams beat long prose. 1,500–3,000 words.
- **Concrete first.** Start each idea from an example the learner can picture
  (a deploy queue, a log file, a Kubernetes pod), then name it. Introduce a
  new term in *italics* with a plain definition, and list it in Key terms.
- **Show, then trace.** Code examples are small and complete, and compile.
  After an example, say what happens step by step.
- **Questions stand alone.** Every *Check yourself* question and *Key
  terms* row also becomes a flashcard (`groundwork cards`), read on a
  phone without the guide. So never write "the example above": restate
  what the question needs, briefly.
- **Refer by title.** Name an exercise, module or another guide by its exact
  title in italics: "see the *Abstraction* guide", "as in *Slices*". Never by
  number. TestCrossReferences checks these.
- **Links** must be real pages you have opened: go.dev, pkg.go.dev,
  kubernetes.io, docs.aws.amazon.com, Wikipedia, well-known books and talks.
- **Videos** are the best few on the guide's idea, for watching away from
  the keyboard (the learner reads guides on a phone too). Prefer talks and
  lectures that age well (conference talks, university lectures, well-known
  educators) over listicles. Verify each one exists and copy its exact title
  and channel from
  `https://www.youtube.com/oembed?url=<video url>&format=json`; take the
  length from the video page. Under about 60 minutes unless it's a
  classic; say which part to watch if only part is relevant.

TestGuides enforces the layout.

## Verify

```bash
go test ./content -run 'TestPacks/^variables$' -v
```

This check proves four things:

- The README has a task section, and the hints are present.
- The starter **fails**.
- Starter plus solution **passes** `go test` (with `-race` when the exercise
  sets `"race": true`).
- Starter plus solution passes `go vet` and `gofmt`.

`go test -short ./...` skips packs with their own dependencies.
