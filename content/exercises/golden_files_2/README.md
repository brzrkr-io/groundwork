# Golden files: write your own

This time you write a whole golden-file test yourself, for a function that
writes a file instead of returning a string. Config generators like this one
sit behind many load balancers, and a stray character in their output can take
a service offline.

## Learn

Recap from *Test fixtures and golden files*: files your tests need go in
`testdata/`, which `go test` can open with relative paths. `os.ReadFile` reads
a whole file as `[]byte`, and an `-update` flag (`flag.Bool`) lets the test
rewrite its golden file when the output changes on purpose. Always review the
rewritten file afterwards.

Two things are new here. First, `testdata/` holds an *input fixture* as well:
`backends.txt` lists the backends to feed the function, one per line.
`strings.Fields` turns that text into a `[]string`, because it splits on
every run of spaces and newlines and ignores the final newline:

```go
strings.Fields("10.0.1.15:8080\n10.0.1.16:8080\n")
// => []string{"10.0.1.15:8080", "10.0.1.16:8080"}
```

Second, the function writes a file, so it needs a folder to write into. As in
*Subtests and helpers: files and t.TempDir*, `t.TempDir()` gives each test a fresh, empty
directory and deletes it when the test ends. Your test never leaves junk
behind, and it never writes into `testdata/` unless you pass `-update`.

`TestGrader` re-runs your test against mutants of `WriteUpstreamConfig` (never
with `-update`), and `TestUpdateFlag` checks your `-update` flag in a scratch
copy of `testdata/`, so your real golden file stays untouched.

## Your task

The platform team generates nginx upstream blocks with `WriteUpstreamConfig`.
The code in `exercise.go` is correct. Write `TestWriteUpstreamConfig` in
`exercise_test.go`, from the empty function up. You'll need to import `flag`,
`os`, `path/filepath` and `strings`.

`WriteUpstreamConfig(dir, name, backends)` writes the file
`<dir>/<name>.conf`. For the backends in `testdata/backends.txt`, the file
`api.conf` must match `testdata/api.conf.golden` exactly:

```text
upstream api {
    server 10.0.1.15:8080;
    server 10.0.1.16:8080;
    server 10.0.1.17:8080;
}
```

### 1. Write the config into a temporary folder

Read `testdata/backends.txt` with `os.ReadFile` and split it with
`strings.Fields`. Call `WriteUpstreamConfig(t.TempDir(), "api", backends)`
(keep the folder in a variable, since you need it in task 2). Stop with
`t.Fatalf` if reading the fixture fails or `WriteUpstreamConfig` returns an
error.

This catches a mutant that writes the file and then returns an error anyway.

### 2. Compare with the golden file

Read the written file, `api.conf` inside your temporary folder (stop with
`t.Fatalf` if it's missing). Read `testdata/api.conf.golden` and fail with
`t.Errorf` if the two differ, printing both.

This catches mutants that write the wrong file name (`api.config`), indent
with a tab instead of four spaces, leave out the closing `}` line, or write
the backends in reverse order.

### 3. Add the -update flag

Declare `var update = flag.Bool("update", false, "rewrite golden files")` at
the top of the file, outside the test. When `*update` is true, write the
contents of the generated `api.conf` to `testdata/api.conf.golden` with
`os.WriteFile` (permission `0o644`) **before** you read the golden file, so
the comparison after it passes.

`TestUpdateFlag` checks this: it runs your test with `-update` next to an
outdated golden file and expects your test to pass and leave the golden file
matching the real output. You can try it yourself with
`go test -run TestWriteUpstreamConfig -update`; the golden file is already
correct, so it shouldn't change.

The exercise is done when `TestWriteUpstreamConfig` and `TestUpdateFlag` pass
and `TestGrader` reports that every mutant was caught.

## Learn more

- [Package testing: T.TempDir](https://pkg.go.dev/testing#T.TempDir)
- [Package strings: Fields](https://pkg.go.dev/strings#Fields)
- [Go by Example: Writing Files](https://gobyexample.com/writing-files)
- [nginx upstream module](https://nginx.org/en/docs/http/ngx_http_upstream_module.html)
