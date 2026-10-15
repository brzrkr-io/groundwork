# Subtests and helpers: files and t.TempDir

Real tools read their hosts from files, so their tests need files too. This
time you write the helper yourself: one that creates a throwaway file for
each subtest.

**New here:** `os.ReadFile`, `os.WriteFile`, `filepath.Join`, `t.TempDir`, `t.Cleanup`, `t.Fatal`.

## Learn

### Recap

As in *Subtests and helpers*:

- `t.Run(name, func(t *testing.T) { ... })` runs a named subtest.
- A helper takes `t`, calls `t.Helper()` first, and does the repeated work.
- `slices.Equal(a, b)` compares two slices.
- `TestGrader` runs your test against mutants; it must fail on each one.

### Reading and writing a whole file

Files get their own exercises in *CLI tools and the OS*. For now you need
two functions from the `os` package. `os.WriteFile` creates (or replaces) a
file with the content you give it:

```go
err := os.WriteFile("notes.txt", []byte("hello\n"), 0o644)
```

| Piece | Meaning |
|---|---|
| `"notes.txt"` | the path to create or replace |
| `[]byte("hello\n")` | the content; `[]byte(...)` turns a `string` into bytes |
| `0o644` | the *permission*: you can read and write, others only read |

`os.ReadFile` is the opposite: it reads a whole file and returns its
content as a `[]byte` (a slice of bytes), plus an error if the file can't
be read. `string(data)` turns the bytes back into text:

```go
data, err := os.ReadFile("notes.txt")
if err != nil {
	return err // for example: the file doesn't exist
}
fmt.Print(string(data)) // hello
```

`LoadHosts` in `exercise.go` uses `os.ReadFile` this way; your test uses
`os.WriteFile` to give it a file to read.

> **Why is it written like this?** Files hold bytes, not text, so Go
> requires the `[]byte(...)` and `string(...)` conversions. `0o644` is the
> usual value for a normal file (convention); the `0o` prefix means the
> number is written in octal.

### Building a path

Use `filepath.Join` from `path/filepath` to build paths. It puts the right
separator between the parts:

```go
filepath.Join("/var/log", "app.log") // => "/var/log/app.log"
```

### Temporary directories and cleanup

Tests shouldn't leave files behind or share them with other tests.
`t.TempDir()` creates a fresh, empty directory and returns its path. Each
call gives a new directory, and Go deletes it when the test or subtest that
created it finishes.

Under the hood, `t.TempDir` uses `t.Cleanup(func() { ... })`, which
registers a function to run when the test ends, pass or fail. You can use it
for any teardown:

```go
srv := startFakeServer()
t.Cleanup(func() { srv.Stop() })
```

### t.Fatal

`t.Fatal(err)` is `t.Fatalf` without a format string: it prints its
arguments and stops the test. In a helper it's the usual way to give up when
setup (like writing a file) fails.

> **Why is it written like this?** The helper could also return an error,
> but then every caller would have to check it. Failing the test inside the
> helper keeps each subtest short. That's a common Go convention for setup
> helpers.

## Your task

The deploy tool now loads its targets with `LoadHosts(path)`. The code in
`exercise.go` is correct. It reads the file and applies the same rules as
before:

- trim spaces around each line,
- skip blank lines and lines starting with `#`,
- return the hosts in order,
- return an error if the file can't be read.

```go
// hosts.txt contains "# web tier\nweb-1\n\n  web-2\n"
LoadHosts("hosts.txt") // => []string{"web-1", "web-2"}, nil
```

`exercise_test.go` is almost empty. You write the helper and all three
subtests in `TestLoadHosts`. You'll need the imports `"os"`,
`"path/filepath"` and `"slices"` next to `"testing"`.

### 1. Write the helper and read a file

**Goal:** write the helper `writeFile(t *testing.T, content string) string`, and the subtest `"reads hosts"` that uses it.

The helper:

- calls `t.Helper()`,
- builds a path to `hosts.txt` inside `t.TempDir()`,
- writes `content` there with `os.WriteFile`, calling `t.Fatal(err)` if that
  fails,
- returns the path.

The subtest `"reads hosts"`:

- writes `"# web tier\nweb-1\n\n  web-2\n"` with `writeFile`,
- calls `LoadHosts` on the returned path, failing with `t.Fatalf` on an error,
- checks the result is `[]string{"web-1", "web-2"}` with `slices.Equal`.

**Think it through:**
- `writeFile` gets text, but `os.WriteFile` wants bytes. What converts it?
- Which two parts go into `filepath.Join`?
- In the subtest, which value do you pass to `LoadHosts`: a file name, or what `writeFile` returned?

### 2. Check a file without a final newline

**Goal:** add the subtest `"no trailing newline"`, checking that a file containing `"web-1\ndb-1"` loads as `[]string{"web-1", "db-1"}` with no error.

Many editors save files without a newline at the end. A classic bug is to
assume the last piece after splitting on `"\n"` is always empty and throw it
away, which loses the last host.

**Think it through:**
- This subtest has the same shape as `"reads hosts"`. Which three things change?

### 3. Check a missing file

**Goal:** add the subtest `"missing file"`, checking that `LoadHosts` returns an error for a file that doesn't exist.

Build a path to `missing.txt` inside `t.TempDir()`, but don't create the
file. A tool that treats a missing hosts file as "no hosts" would silently
deploy to nothing.

**Think it through:**
- Should you call `writeFile` here?
- Only the error matters. What does `err` look like when the bug happens?

The exercise is done when `TestLoadHosts` passes and `TestGrader` reports
that every mutant was caught.

## Learn more

- [testing.T.TempDir](https://pkg.go.dev/testing#T.TempDir)
- [testing.T.Cleanup](https://pkg.go.dev/testing#T.Cleanup)
- [os.WriteFile](https://pkg.go.dev/os#WriteFile)
- [Go by Example: Temporary Files and Directories](https://gobyexample.com/temporary-files-and-directories)
