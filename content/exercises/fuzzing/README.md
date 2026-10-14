# Fuzzing

A *fuzz test* throws thousands of generated inputs at a function to find the
ones that crash it. For anything that reads text from the outside world, such
as config files, API requests or flags, that's how you find the input nobody
thought of, before it takes a service down.

**New here:** fuzz targets, `testing.F`, seeds, the seed corpus, `go test -fuzz`.

## Learn

### A fuzz target

So far you picked every test input yourself, and bugs hide in the cases you
didn't think of. Fuzzing lets Go invent inputs for you.

A *fuzz target* lives in a `_test.go` file like a test:

```go
func FuzzParsePort(f *testing.F) {
	f.Add("8080") // a seed: an example input
	f.Add("443")
	f.Fuzz(func(t *testing.T, s string) {
		ParsePort(s) // must never panic
	})
}
```

- `f.Add(...)` adds a *seed*: a typical, valid input to start from.
- `f.Fuzz(...)` hands Go the function to run on every input. After `t`, its
  parameters say what kind of input it wants, here one `string`.

> **Why is it written like this?** Go requires the name to start with `Fuzz`
> and the parameter to be `*testing.F`. The function given to `f.Fuzz` must
> take `*testing.T` first; the values passed to `f.Add` must match the types
> of the parameters after it. Naming them `f`, `t` and `s` is convention.

### Checking a property

Inside the function you can't write a "want": you don't know the input in
advance. Instead you check a *property*, something true for every input.

The simplest property is "it doesn't panic". A panic in a parser is a crashed
program. Calling the function is the whole check, because a panic fails the
fuzz target.

### The seed corpus

The *seed corpus* is the list of inputs a plain `go test` feeds your fuzz
target: your `f.Add` seeds plus every file in
`testdata/fuzz/<FuzzTargetName>/`. Each file holds one input:

```text
go test fuzz v1
string("hello")
```

The first line is always `go test fuzz v1`. Then comes one line per input
parameter, written like a Go conversion: `string("")`, `int(2)`,
`bool(true)`.

A plain `go test` runs each entry as a subtest, such as
`FuzzParseTarget/seed#0`, or `FuzzParseTarget/empty_target` for a file named
`empty_target`. Nothing random happens, so it is fast and repeatable.
Groundwork grades with plain `go test` too.

### Real fuzzing (going further)

To let Go generate new inputs, run this in the exercise folder:

```text
go test -run='^$' -fuzz=FuzzParseTarget -fuzztime=10s
```

| Flag | Meaning |
|---|---|
| `-fuzz` | which fuzz target to fuzz (one only) |
| `-fuzztime` | when to stop; without it, it runs until a failure or Ctrl+C |
| `-run='^$'` | skip the normal tests |

Go mutates your seeds and keeps inputs that reach new code. When one fails,
Go shrinks it and saves it to `testdata/fuzz/FuzzParseTarget/<hash>`. From
then on, every plain `go test` replays it. Commit that file, and the bug can
never come back unnoticed.

The two files in this exercise's `testdata/fuzz/FuzzParseTarget/` are exactly
that: crashes a fuzzer found, renamed so you can tell them apart.

### How Groundwork grades your fuzz target

- `TestGrader` runs `FuzzParseTarget` with plain `go test` against the real
  `ParseTarget`, where it must pass, and against *mutants* that each put back
  one of the original crashes, where it must fail.
- `TestParseTarget` checks that your fixed `ParseTarget` still parses targets
  correctly.

## Your task

The monitoring system reads scrape targets like `web-1:8080` from config with
`ParseTarget`. A config with a typo crashed the whole monitoring service. A
fuzzer already found two inputs that make `ParseTarget` panic, saved in
`testdata/fuzz/FuzzParseTarget/`:

| File | Input |
|---|---|
| `empty_target` | `""` |
| `missing_bracket` | `"[::1"` (no closing `]`) |

`ParseTarget` returns the host and port. IPv6 addresses go in brackets,
because they contain colons themselves:

```go
ParseTarget("web-1:8080") // => "web-1", 8080, nil
ParseTarget("[::1]:443")  // => "::1", 443, nil
ParseTarget("web-1")      // => "", 0, error (no port)
```

### 1. Write the fuzz target

**Goal:** in `FuzzParseTarget` in `exercise_test.go`, add the seeds
`"web-1:8080"` and `"[::1]:443"` and call `ParseTarget(s)` on every input.

Ignore the results; the property is only "never panics".

When you save, your fuzz target **fails**. That's it doing its job: `go test`
replays `empty_target`, and the output shows
`--- FAIL: FuzzParseTarget/empty_target`, the panic, and the line in
`exercise.go` where it happened. A panic stops the run, so you see one crash
at a time.

**Think it through:**
- Which `TODO(1)` is outside `f.Fuzz`, and which is inside?
- Where does the input `s` come from inside the function?

### 2. Fix the crashes

**Goal:** change `realParseTarget` in `exercise.go` so both saved inputs
return an error instead of panicking.

```go
ParseTarget("")     // => "", 0, error
ParseTarget("[::1") // => "", 0, error (no closing ])
```

- `s[0]` panics when `s` is empty. Check for the empty string first, or use
  `strings.HasPrefix`, which is simply false for `""`.
- `strings.Index` returns `-1` when there is no `]`, and slicing up to `-1`
  panics. Return an error when `end` is negative.

Everything else must keep working. `TestParseTarget` checks the examples
above, `"db.internal:5432"` and `"[fe80::2]:9100"`, and expects errors for
`":8080"` (no host), `"web-1:http"`, `"web-1:70000"` and `"[::1]"`.

**Think it through:**
- The panic message names a line in `exercise.go`. What is that line doing
  with `s`?
- Which value would make the slice expression unsafe, and where can you check
  it first?

The exercise is done when `FuzzParseTarget` and `TestParseTarget` pass and
`TestGrader` reports that every mutant was caught.

## Learn more

- [Go Fuzzing](https://go.dev/doc/security/fuzz/)
- [Tutorial: Getting started with fuzzing](https://go.dev/doc/tutorial/fuzz)
- [Package testing: Fuzzing](https://pkg.go.dev/testing#hdr-Fuzzing)
