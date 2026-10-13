# Subtests and helpers: files and t.TempDir

Real tools read their hosts from files, so their tests need files too. This
time you write the helper yourself: one that creates a throwaway file for
each subtest.

## Learn

### Recap

`t.Run(name, func(t *testing.T) { ... })` runs a named subtest, and
`go test -run 'TestLoadHosts/missing_file'` runs just one of them. A helper
takes `t`, calls `t.Helper()` first so failures point at the caller's line,
and does the repeated work. `slices.Equal` compares two slices. As before,
`TestGrader` runs `TestLoadHosts` against broken copies of `LoadHosts`, and
your test must fail against each one.

### Reading and writing files

File I/O gets its own exercises later. For now you need two functions from
the `os` package:

```go
data, err := os.ReadFile("hosts.txt")                     // whole file as []byte
err = os.WriteFile("hosts.txt", []byte("web-1\n"), 0o644) // create or replace
```

`0o644` is the file's *permission*: you can read and write it, everyone else
can only read it. It's the usual value for a normal file.

To build a path, use `filepath.Join` from `path/filepath`. It puts the right
separator between the parts: `filepath.Join("/tmp/x", "hosts.txt")` gives
`"/tmp/x/hosts.txt"`.

### Temporary directories and cleanup

Tests shouldn't leave files behind or share them with other tests.
`t.TempDir()` creates a fresh, empty directory and returns its path. Each call
gives a new directory, and Go deletes it automatically when the test or
subtest that created it finishes.

`t.TempDir` does this with `t.Cleanup(func() { ... })`, which registers a
function to run when the test ends, pass or fail. You can use it yourself for
any teardown, such as stopping a fake server:

```go
srv := startFakeServer()
t.Cleanup(func() { srv.Stop() })
```

`t.Fatal(err)` is `t.Fatalf` without a format string: it prints its arguments
and stops the test.

## Your task

The deploy tool now loads its targets with `LoadHosts(path)`. It reads the
file and applies the same rules as before: trim spaces around each line, skip
blank lines and lines starting with `#`, and return the hosts in order. If
the file can't be read, it returns an error. The code in `exercise.go` is
correct. Write `TestLoadHosts` in `exercise_test.go`.

```go
// hosts.txt contains "# web tier\nweb-1\n\n  web-2\n"
LoadHosts("hosts.txt") // => []string{"web-1", "web-2"}, nil
```

### 1. Write the helper and read a file

Write `writeFile(t *testing.T, content string) string`. It:

- calls `t.Helper()`,
- builds a path to `hosts.txt` inside `t.TempDir()`,
- writes `content` there with `os.WriteFile`, calling `t.Fatal(err)` if that
  fails,
- returns the path.

Then add the subtest `"reads hosts"`. Write
`"# web tier\nweb-1\n\n  web-2\n"` with `writeFile`, call `LoadHosts` on
the path, fail with `t.Fatalf` on an error, and check the result is
`[]string{"web-1", "web-2"}` with `slices.Equal`. You'll need the imports
`"os"`, `"path/filepath"` and `"slices"`.

### 2. Check a file without a final newline

Many editors save files without a newline at the end. A classic bug is to
assume the last piece after splitting on `"\n"` is always empty, and throw it
away, which loses the last host.

Add the subtest `"no trailing newline"`. Write `"web-1\ndb-1"` and check
that `LoadHosts` returns `[]string{"web-1", "db-1"}` and no error.

### 3. Check a missing file

Add the subtest `"missing file"`. Build a path to `missing.txt` inside
`t.TempDir()` without creating the file, and check that `LoadHosts` returns
an error. A tool that treats a missing hosts file as "no hosts" would
silently deploy to nothing.

The exercise is done when `TestLoadHosts` passes and `TestGrader` reports
that every mutant was caught.

## Learn more

- [testing.T.TempDir](https://pkg.go.dev/testing#T.TempDir)
- [testing.T.Cleanup](https://pkg.go.dev/testing#T.Cleanup)
- [os.ReadFile](https://pkg.go.dev/os#ReadFile) and [os.WriteFile](https://pkg.go.dev/os#WriteFile)
- [Go by Example: Temporary Files and Directories](https://gobyexample.com/temporary-files-and-directories)
