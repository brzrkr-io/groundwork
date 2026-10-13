# Authoring Gopherlings exercises

The course is `course.json` (module and exercise order) plus one pack per
exercise in `exercises/<name>/`. Exercise numbers come from their position in
`course.json`, so packs are named without numbers and reordering costs nothing.

The learner is a **beginner** who wants to do DevOps and platform work. Every
exercise should feel doable after reading its README. A sense of progress
matters more than covering every corner of a topic.

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
  `go.mod.tmpl` get a generated one (`module gopherlings.local/<folder>`,
  `go 1.26`). Packs with **third-party dependencies** or **subpackages**
  ship their own `go.mod.tmpl` (module path `gopherlings.local/<name>`,
  `go 1.26`) and, with dependencies, a `go.sum`.
- Learner folders are `exercises/<NN>_<module dir>/<NNN>_<topic>`, where the
  topic is the pack name without a `_2`/`_3` suffix. Numbers come from the
  pack's position in `course.json`; each module there has a `dir` and an
  `outcome` (shown in LESSON_PLAN.md). After changing course.json, run
  `go test ./content -run TestLessonPlan -args -update`.
- File and directory names must not start with `.` or `_`, because `go:embed`
  skips them.

## README.md

```markdown
# Title

One or two sentences: what this is and why an ops engineer cares.

## Learn

The concept, taught from zero, with short ```go examples. Explain every piece
of syntax or API the task needs that earlier exercises did not cover. Use
tables when comparing options. About 200–600 words for concept exercises.
Practice exercises recap briefly. Projects describe the architecture and
include a file map.

## Your task

A one- or two-sentence scenario.

### 1. Imperative task title

Which function to change and the exact expected behaviour, followed by a worked
example:

    ```go
    ServiceLabel("api", "production")
    // => "production/api"
    ```

### 2. …

## Learn more

- [Official doc title](https://go.dev/…)  (2–4 links: go.dev, pkg.go.dev,
  the spec, the Go blog, gobyexample.com, kubernetes.io, docs.aws.amazon.com)
```

Rules:

- **No hidden requirements.** Every string, format, edge case and error
  condition the tests check is stated in the README. A learner who reads the
  README carefully never has to reverse-engineer the tests.
- Tasks go from easiest to hardest. Task numbers match the `TODO(n)` markers.
- Write in the second person. Be friendly, direct and concrete. No filler and
  no marketing.
- Refer to earlier exercises by title ("as in *Slices*"), never by number.
- Introduce a new term in *italics* the first time it appears.

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
- Every non-test package (starter and solution) has a package comment, such as
  `// Package variables is the Gopherlings exercise "Variables".` Otherwise
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

To produce `go.mod.tmpl` and `go.sum`, copy the starter, tests and solution
into a scratch directory and run `go mod tidy`. Then copy `go.mod` back as
`go.mod.tmpl` and copy `go.sum` back as `go.sum`. Tests never touch the
network: use fake clientsets, fake API clients, or `httptest` servers.

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
