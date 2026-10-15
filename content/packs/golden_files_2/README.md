# Golden files: write your own

This time you write a whole golden-file test yourself, for a function that
writes a file instead of returning a string. Config generators like this one
sit behind many load balancers, and one stray character in their output can
take a service offline.

**New here:** input fixtures, `t.TempDir` for output files, writing your own `-update` flag.

## Learn

### Recap: golden files

As in *Test fixtures and golden files*:

- Files your tests need live in `testdata/`. Tests open them with relative
  paths, such as `filepath.Join("testdata", "x.txt")`.
- `os.ReadFile(path)` returns the whole file as `[]byte`, plus an error.
- A *golden file* holds the exact output you expect. The test compares the
  real output with it.

### Input fixtures and `strings.Fields`

`testdata/` can hold *input* too. Here `backends.txt` lists the backends to
feed the function, one per line. As in *The strings package: fix the
bootstrap parser*, `strings.Fields` splits text at every run of spaces and
newlines, so a final newline adds no empty item:

```go
strings.Fields("alpha beta\ngamma\n")
// => []string{"alpha", "beta", "gamma"}
```

`os.ReadFile` gives you `[]byte`, but `strings.Fields` wants a `string`.
Convert with `string(data)`.

### A temporary folder for output

The function writes a file, so it needs a folder. As in *Subtests and
helpers: files and t.TempDir*, `t.TempDir()` returns the path of a fresh,
empty folder and deletes it when the test ends:

```go
dir := t.TempDir()
path := filepath.Join(dir, "report.txt")
```

Your test never leaves junk behind, and never writes into `testdata/` by
accident.

### Package-level flags

A *flag* is a command-line option such as `-update`. `flag.Bool` declares one
and returns a *pointer* to its value, which `go test` fills in before your
tests run. You read the value with `*`:

```go
var verbose = flag.Bool("verbose", false, "print extra output")

if *verbose {
	// ...
}
```

> **Why is it written like this?** Go requires the three arguments: the flag
> name (without the dash), the default value, and a help text. The `var`
> must sit at package level, outside every function, so the flag exists
> before `go test` reads the command line. The variable name and help text
> are your choice; naming the flag `update` is what this exercise checks.

### Grading

`TestGrader` re-runs your test against *mutants* of `WriteUpstreamConfig`,
each with one bug, and never with `-update`. `TestUpdateFlag` tries your
`-update` flag on a scratch copy of `testdata/`, so your real golden file
stays untouched.

## Your task

The platform team generates nginx upstream blocks with `WriteUpstreamConfig`.
The code in `exercise.go` is correct. Write `TestWriteUpstreamConfig` in
`exercise_test.go`, from the empty function up. Tasks 1 and 2 need the
imports `os`, `path/filepath` and `strings`; add `flag` in task 3, because Go
rejects an import you don't use yet.

`WriteUpstreamConfig(dir, name, backends)` writes the file
`<dir>/<name>.conf`. With the name `"api"` and the backends from
`testdata/backends.txt`, it writes `api.conf`, which must match
`testdata/api.conf.golden` exactly:

```text
upstream api {
    server 10.0.1.15:8080;
    server 10.0.1.16:8080;
    server 10.0.1.17:8080;
}
```

### 1. Write the config into a temporary folder

**Goal:** read the backends from the fixture and call `WriteUpstreamConfig`
with a temporary folder, the name `"api"` and those backends.

- Read `testdata/backends.txt` and split it into a `[]string`.
- Get a folder from `t.TempDir()` and keep it in a variable; task 2 needs it.
- Stop with `t.Fatalf` if reading the fixture fails, or if
  `WriteUpstreamConfig` returns an error.

This catches a mutant that writes the file and then returns an error anyway.

**Think it through:**
- Which of the three arguments comes from the fixture, which from
  `t.TempDir()`, and which is fixed text?
- What type does `os.ReadFile` return, and what type does `strings.Fields`
  want?

### 2. Compare with the golden file

**Goal:** read the file the function wrote and fail with `t.Errorf` if it
differs from `testdata/api.conf.golden`, printing both.

- The written file is `api.conf` inside your temporary folder. Stop with
  `t.Fatalf` if it's missing.
- Read the golden file the same way.

This catches mutants that write the wrong file name (`api.config`), indent
with a tab instead of four spaces, leave out the closing `}` line, or write
the backends in reverse order.

**Think it through:**
- Which folder holds the written file, and which holds the golden file?
- Both reads give `[]byte`. How do you compare them as text?

### 3. Add the -update flag

**Goal:** add a flag named `update`; when it is set, your test rewrites
`testdata/api.conf.golden` with the generated output and then passes.

- Add `"flag"` to the imports, and declare the flag with `flag.Bool` at
  package level (above the test function), default `false`.
- When the flag is true, write the contents of the generated `api.conf` to
  the golden file with `os.WriteFile` (permission `0o644`).
- Do this **before** you read the golden file, so the comparison after it
  passes.

`TestUpdateFlag` runs your test with `-update` next to an outdated golden
file. It expects your test to pass and the golden file to match the real
output afterwards. Try it yourself with
`go test -run TestWriteUpstreamConfig -update`; the golden file is already
correct, so it shouldn't change. Always review a rewritten golden file.

**Think it through:**
- Where in the test do you already have the generated bytes?
- Why must the write come before the read of the golden file, not after?

The exercise is done when `TestWriteUpstreamConfig` and `TestUpdateFlag` pass
and `TestGrader` reports that every mutant was caught.

## Learn more

- [Package testing: T.TempDir](https://pkg.go.dev/testing#T.TempDir)
- [Package strings: Fields](https://pkg.go.dev/strings#Fields)
- [Package flag](https://pkg.go.dev/flag)
- [nginx upstream module](https://nginx.org/en/docs/http/ngx_http_upstream_module.html)
