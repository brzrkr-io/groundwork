# Test fixtures and golden files

Ops tools print a lot of text: inventory tables, generated configs, reports.
Checking that output with a giant string inside your test gets painful fast.
*Golden files* keep the expected output in a file of its own, where you can
read it and review changes to it like any other file.

## Learn

### The `testdata` directory

The go tool ignores any directory named `testdata`. Nothing in it gets
compiled, so it is the standard place for files your tests need: sample
inputs, expected outputs, broken configs. Gopherlings already put one in this
exercise:

```text
golden_files/
├── exercise.go
├── exercise_test.go
└── testdata/
    └── inventory.golden
```

`go test` runs your tests with the package folder as the *working directory*,
so a test opens that file with the relative path `testdata/inventory.golden`.
`filepath.Join("testdata", "inventory.golden")` builds that path with the
right separator for the operating system (`/` on Linux and macOS, `\` on
Windows).

### Golden files

A golden file holds the exact output a function should produce. The test
calls the function, reads the golden file, and fails if the two differ. The
expected output is now a plain file: you see the table exactly as an operator
would, without `\n` escapes and quotes in the way.

### Reading a file

`os.ReadFile` reads a whole file and returns its contents as a `[]byte` (a
slice of bytes) plus an error:

```go
want, err := os.ReadFile(filepath.Join("testdata", "inventory.golden"))
if err != nil {
	t.Fatalf("reading golden file: %v", err)
}
```

A `string` and a `[]byte` can't be compared directly, so convert with
`string(want)`. When the outputs differ, print both on their own lines so
the reader can spot the difference:

```go
if got != string(want) {
	t.Errorf("output differs from golden file\ngot:\n%s\nwant:\n%s", got, want)
}
```

`%s` prints a `[]byte` as text, so `want` needs no conversion there.

### Updating golden files with `-update`

When you *mean* to change the output, editing the golden file by hand is
tedious and error-prone. The usual trick is an `-update` flag that makes the
test write the golden file for you:

```go
var update = flag.Bool("update", false, "rewrite golden files")

// inside the test, before reading the golden file:
if *update {
	os.WriteFile(golden, []byte(got), 0o644)
}
```

The `flag` package reads command-line options (a later exercise covers it
properly). `flag.Bool` returns a pointer to a `bool` that starts as `false`
and becomes `true` when you pass `-update`, so `*update` reads its value.
`os.WriteFile` writes the bytes to the file, replacing what was there. `0o644`
is the usual file permission: you can read and write it, everyone else can
read it.

Running the test with the flag rewrites the golden file:

```text
go test -run TestRenderInventory -update
```

`go test` passes `-update` on to your test, and `-run` keeps the run to the
one test you care about. Then comes the important part: **read the new golden
file** (in a real project, `git diff testdata/`) and check that the change is
the one you wanted before you commit it. `-update` makes the test agree with
the code, whatever the code does, so the review is your actual check.

### How Gopherlings grades your test

`TestGrader` in `grader_test.go` re-runs `TestRenderInventory` against the
real `RenderInventory`, where it must pass, and then against *mutants*:
copies of `RenderInventory` with one deliberate bug each. Your test must fail
on every mutant. If one slips through, `TestGrader` names the bug, for
example "your tests didn't notice when RenderInventory left out the last
host". The re-runs never pass `-update`, so they can't touch your golden file.

## Your task

The fleet tool prints the host inventory with `RenderInventory`. The code in
`exercise.go` is correct. Finish `TestRenderInventory` in `exercise_test.go`
so it checks the output against `testdata/inventory.golden`. The hosts,
the call and the `-update` branch are already written.

### 1. Compare with the golden file

Below the `if *update` block, read `testdata/inventory.golden` with
`os.ReadFile` (stop with `t.Fatalf` if that fails), then compare it with
`got`. If they differ, fail with `t.Errorf` and a message that prints both
`got` and `want`.

This catches mutants that leave out the header line, leave out the last
host, or stop lining up the columns.

### 2. Update the outdated golden file

Once task 1 is done, your test **fails**, and `TestGrader` says "your tests
must pass on the real code first". That's expected: the team recently added a
REGION column on purpose, but nobody updated the golden file. It still holds
the old two-column table.

Open a terminal in the exercise folder and run:

```text
go test -run TestRenderInventory -update
```

Then open `testdata/inventory.golden` and review it. It should now read:

```text
NAME         IP              REGION
web-1        10.0.1.15       eu-west-1
web-2        10.0.1.16       eu-west-1
db-primary   10.0.2.5        us-east-1
```

Save `exercise_test.go` (or run `go test` again) and everything passes. The
updated golden file is what catches the last mutant, which still prints the
old layout without REGION.

The exercise is done when `TestRenderInventory` passes and `TestGrader`
reports that every mutant was caught.

## Learn more

- [Go by Example: Reading Files](https://gobyexample.com/reading-files)
- [Package os: ReadFile and WriteFile](https://pkg.go.dev/os#ReadFile)
- [go test and the testdata directory (go command docs)](https://pkg.go.dev/cmd/go#hdr-Test_packages)
