# Test fixtures and golden files

Ops tools print a lot of text: inventory tables, generated configs, reports.
Checking that output with a giant string inside your test gets painful fast.
*Golden files* keep the expected output in a file of its own, where you can
read it and review changes like any other file.

**New here:** the `testdata` directory, comparing `[]byte` with `string`, an `-update` flag with `flag.Bool`.

## Learn

### The testdata directory

The go tool ignores any directory named `testdata`; nothing in it gets
compiled. That makes it the standard place for files your tests need: sample
inputs, expected outputs, broken configs. This exercise already has one:

```text
golden_files/
├── exercise.go
├── exercise_test.go
└── testdata/
    └── inventory.golden
```

`go test` runs your tests with the package folder as the *working
directory*, so a test reaches that file with the relative path
`testdata/inventory.golden`. `filepath.Join("testdata", "inventory.golden")`
builds it with the right separator for the operating system.

> **Why is it written like this?** The name `testdata` is a Go rule: the
> tool skips it. The `.golden` file ending is only a convention; Go doesn't
> care what you call the file.

### Golden files

A golden file holds the exact output a function should produce. The test:

1. calls the function (`got`),
2. reads the golden file (`want`),
3. fails if the two differ.

You see the expected table exactly as an operator would, without `\n`
escapes and quotes in the way.

### Reading a file

As in *Subtests and helpers: files and t.TempDir*, `os.ReadFile` reads a
whole file and returns its contents as a `[]byte` plus an error:

```go
data, err := os.ReadFile(filepath.Join("testdata", "config.yaml"))
if err != nil {
	t.Fatalf("reading config: %v", err)
}
```

### Comparing bytes with a string

A `string` and a `[]byte` can't be compared with `!=` directly. Convert the
bytes first with `string(...)`:

```go
if text != string(data) {
	t.Errorf("config differs\ngot:\n%s\nwant:\n%s", text, data)
}
```

`%s` prints a `[]byte` as text, so no conversion is needed in the message.
Printing each value after its own `\n` puts every table on its own lines, so
the reader can spot the difference.

> **Why is it written like this?** Go never converts between types for you,
> so `string(data)` is required. Showing both values in the message is
> convention; the exact wording is your choice.

### Updating golden files with -update

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

- `flag.Bool` returns a pointer to a `bool`. It is `false` unless you pass
  `-update`; `*update` reads its value. (The `flag` package gets its own
  exercises in *CLI tools and the OS*.)
- `os.WriteFile` replaces the file's content. `0o644` is the usual file
  permission.

Run it like this:

```text
go test -run TestRenderInventory -update
```

Then comes the important part: **read the new golden file** and check the
change is the one you wanted. (In a real project, Git shows you exactly
what changed; you'll use it in *Git and version control*.)
`-update` makes the test agree with the code, whatever the code does, so the
review is your real check.

### How Groundwork grades your test

As in *Your first test*, `TestGrader` runs your test against the real code
(it must pass) and against *mutants*, copies with one bug each (it must fail
on every one). It names any mutant that slips through. The re-runs never
pass `-update`, so they can't touch your golden file.

## Your task

The fleet tool prints the host inventory with `RenderInventory`. The code in
`exercise.go` is correct. `TestRenderInventory` in `exercise_test.go`
already builds the hosts, calls `RenderInventory` (the result is in `got`),
stores the golden file's path in `golden`, and has the `-update` branch. You
add the comparison.

### 1. Compare with the golden file

**Goal:** below the `if *update` block, read the golden file and fail the test if its contents differ from `got`.

- Read the file at `golden` with `os.ReadFile`. Stop with `t.Fatalf` if
  that fails.
- Compare it with `got`. If they differ, fail with `t.Errorf` and a message
  that prints both `got` and `want`.

This catches mutants that leave out the header line, leave out the last
host, or stop lining up the columns.

**Think it through:**
- The path is already built. Which variable holds it?
- `got` is a `string`; what type does `os.ReadFile` give you?

### 2. Update the outdated golden file

**Goal:** regenerate `testdata/inventory.golden` with `-update` and check that it now has a REGION column.

After task 1 your test **fails**, and `TestGrader` says "your tests must
pass on the real code first". That's expected. The team added a REGION
column on purpose, but nobody updated the golden file; it still holds the
old two-column table.

You don't change any code here. Open a terminal in the exercise folder and
run:

```text
go test -run TestRenderInventory -update
```

Then open `testdata/inventory.golden` and review it. It should read:

```text
NAME         IP              REGION
web-1        10.0.1.15       eu-west-1
web-2        10.0.1.16       eu-west-1
db-primary   10.0.2.5        us-east-1
```

Save `exercise_test.go` (or run `go test` again) and everything passes. The
updated golden file catches the last mutant, which still prints the old
layout without REGION.

**Think it through:**
- In the failure message from task 1, which side has REGION: got or want?
- Which one is out of date: the code or the golden file?

The exercise is done when `TestRenderInventory` passes and `TestGrader`
reports that every mutant was caught.

## Learn more

- [Go by Example: Reading Files](https://gobyexample.com/reading-files)
- [Package os: ReadFile and WriteFile](https://pkg.go.dev/os#ReadFile)
- [go test and the testdata directory (go command docs)](https://pkg.go.dev/cmd/go#hdr-Test_packages)
