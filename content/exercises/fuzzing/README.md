# Fuzzing

A *fuzz test* throws thousands of generated inputs at a function to find the
ones that crash it. For anything that reads text from the outside world,
such as config files, API requests or command-line flags, that is how you
find the input nobody thought of before it takes a service down.

## Learn

### A fuzz target

So far you picked every test input yourself. You can only test the cases
you think of, and bugs hide in the ones you don't. Fuzzing lets Go invent
inputs for you.

A *fuzz target* lives in a `_test.go` file like a test, but its name starts
with `Fuzz` and it takes `f *testing.F`:

```go
func FuzzParsePort(f *testing.F) {
	f.Add("8080") // a seed: an example input
	f.Add("443")
	f.Fuzz(func(t *testing.T, s string) {
		ParsePort(s) // must never panic
	})
}
```

- `f.Add(...)` adds a *seed*: a sample input to start from. Give it typical,
  valid inputs.
- `f.Fuzz(func(t *testing.T, ...))` hands Go the function to run on every
  input. After `t`, its parameters say what kind of input it wants, here one
  `string`. The values you pass to `f.Add` must have the same types.

Inside the function you can't write a want: you don't know the input in
advance. Instead you check a *property*, something that holds for every
input. The simplest property is "it doesn't panic". A panic in a parser is a
crashed program, so it's worth checking on its own. Calling the function is
the whole check, because a panic fails the fuzz target.

### The seed corpus

The *seed corpus* is the list of inputs a plain `go test` feeds to your fuzz
target. It holds your `f.Add` seeds plus every file in
`testdata/fuzz/<FuzzTargetName>/`. Each file holds one input:

```text
go test fuzz v1
string("[::1")
```

The first line is always `go test fuzz v1`. Then comes one line per `f.Fuzz`
parameter, written like a Go conversion: `string("")`, `int(2)`,
`bool(true)`.

A plain `go test` runs every corpus entry as a subtest named after it, such
as `FuzzParseTarget/seed#0` or `FuzzParseTarget/empty_target` for a file
called `empty_target`. Nothing random happens, so it is as fast and
repeatable as any test. Gopherlings grades with plain `go test` too.

### Real fuzzing (going further)

To let Go generate new inputs, run this in the exercise folder:

```text
go test -run='^$' -fuzz=FuzzParseTarget -fuzztime=10s
```

`-fuzz` names one fuzz target to fuzz, `-fuzztime` says when to stop
(without it, fuzzing runs until it finds a failure or you press Ctrl+C), and
`-run='^$'` skips the normal tests. Go mutates your seeds, flipping and
adding bytes, and keeps inputs that reach new code. When an input fails, Go
shrinks it to a small example and writes it to
`testdata/fuzz/FuzzParseTarget/<hash>`, a file with a long hex name. From
then on every plain `go test` replays it. Commit that file, and the bug can
never come back unnoticed.

The two files in this exercise's `testdata/fuzz/FuzzParseTarget/` are
exactly that: crashes a fuzzer found, renamed so you can tell them apart.

### How Gopherlings grades your fuzz target

`TestGrader` in `grader_test.go` re-runs `FuzzParseTarget` with plain
`go test` against the real `ParseTarget`, where it must pass, and against
*mutants*: copies of `ParseTarget` with one deliberate bug each. Your fuzz
target must fail against every mutant. Here each mutant puts back one of the
original crashes. `TestParseTarget`, also in `grader_test.go`, checks that
your fixed `ParseTarget` still parses targets correctly.

## Your task

The monitoring system reads scrape targets like `web-1:8080` from config
with `ParseTarget`. Someone deployed a config with a typo and the whole
monitoring service crashed. A fuzzer already found two inputs that make
`ParseTarget` panic, and saved them in `testdata/fuzz/FuzzParseTarget/`:

| File | Input |
|---|---|
| `empty_target` | `""` |
| `missing_bracket` | `"[::1"` (no closing `]`) |

`ParseTarget` returns the host and port. IPv6 addresses go in brackets
because they contain colons themselves:

```go
ParseTarget("web-1:8080") // => "web-1", 8080, nil
ParseTarget("[::1]:443")  // => "::1", 443, nil
ParseTarget("web-1")      // => "", 0, error (no port)
```

### 1. Write the fuzz target

In `FuzzParseTarget` in `exercise_test.go`:

- add the seeds `"web-1:8080"` and `"[::1]:443"` with `f.Add`,
- inside the function passed to `f.Fuzz`, call `ParseTarget(s)`. Ignore its
  results; the property is only "never panics".

Save, and your fuzz target fails: `go test` replays `empty_target`, and the
output shows `--- FAIL: FuzzParseTarget/empty_target`, the panic, and the
line in `exercise.go` where it happened. That's your fuzz target doing its
job. A panic stops the whole run, so you see one crash at a time.

### 2. Fix the crashes

Change `realParseTarget` in `exercise.go` so both inputs return an error
instead of panicking:

```go
ParseTarget("")     // => "", 0, error
ParseTarget("[::1") // => "", 0, error (no closing ])
```

- `s[0]` panics when `s` is empty. Check for the empty string first, or use
  `strings.HasPrefix(s, "[")`, which is simply false for `""`.
- `strings.Index` returns `-1` when there is no `]`, and `s[1:-1]` panics.
  Return an error when `end` is negative.

Every other target must keep working as before. `TestParseTarget` checks the
examples above, `"db.internal:5432"`, `"[fe80::2]:9100"`, and errors for
`":8080"` (no host), `"web-1:http"`, `"web-1:70000"` and `"[::1]"`.

The exercise is done when `FuzzParseTarget` and `TestParseTarget` pass and
`TestGrader` reports that every mutant was caught.

## Learn more

- [Go Fuzzing](https://go.dev/doc/security/fuzz/)
- [Tutorial: Getting started with fuzzing](https://go.dev/doc/tutorial/fuzz)
- [Package testing: Fuzzing](https://pkg.go.dev/testing#hdr-Fuzzing)
