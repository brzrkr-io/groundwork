# Benchmarks

Tests tell you whether code is right. *Benchmarks* tell you how fast it is and
how much memory it uses. When a function runs for every request or every log
line, that is worth knowing before you try to speed it up.

**New here:** `testing.B`, `b.Loop`, `b.ReportAllocs`, allocations, `Grow` on a `strings.Builder`.

## Learn

### Writing a benchmark

A benchmark lives in a `_test.go` file, next to your tests:

```go
func BenchmarkDouble(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		Double(21)
	}
}
```

- `for b.Loop() { ... }` runs the body again and again until Go has a
  stable timing. Code before the loop is setup and isn't timed.
- `b.ReportAllocs()` adds memory numbers to the result.
- `b` also has `Errorf` and `Fatalf`, just like `t`.

> **Why is it written like this?** Go requires the name to start with
> `Benchmark` and the only parameter to be `*testing.B`, or `go test` won't
> find it. Naming the parameter `b` is convention. `b.Loop` arrived in Go
> 1.24; older code writes `for i := 0; i < b.N; i++`, and you'll still see
> that in the wild. `b.Loop` also stops the compiler from deleting a call
> whose result you ignore.

### Running benchmarks

`go test` skips benchmarks unless you ask for them with `-bench`, which takes
a pattern like `-run` does:

```text
go test -bench=FormatTags              run tests, then matching benchmarks
go test -run='^$' -bench=FormatTags    run only the benchmarks
```

`-run='^$'` selects tests with an empty name. There are none, so only the
benchmark runs.

### Reading the output

```text
BenchmarkFormatTags-16    1232925    195.7 ns/op    216 B/op    7 allocs/op
```

| Column | Meaning |
|---|---|
| `BenchmarkFormatTags-16` | The name, plus the number of CPUs Go used. |
| `1232925` | How many times the loop body ran. |
| `195.7 ns/op` | Nanoseconds per call. |
| `216 B/op` | Bytes of memory allocated per call. |
| `7 allocs/op` | *Allocations* per call: how often Go asked for new memory. |

Your numbers depend on the machine. Compare runs on the same machine, before
and after a change.

### Why allocations cost

Each allocation is work, and later the garbage collector has to free it
again. Fewer allocations usually means faster code.

Go strings can't be changed. So `s += "x"` doesn't extend `s`: it allocates a
new string, copies the old text and `"x"` into it, and throws the old one
away. In a loop, that adds up.

### `strings.Builder` and `Grow`

As in *For loops: backoff and progress bars*, a `strings.Builder` collects
text in one growing buffer:

```go
var b strings.Builder
b.Grow(12)              // reserve room for 12 bytes up front
b.WriteString("web")
b.WriteString("-")
b.WriteString("01")
s := b.String()         // "web-01", no extra copy
```

Without `Grow`, the builder enlarges its buffer a few times as text arrives.
If you tell `Grow` the total number of bytes first, it allocates once.
`len(s)` gives a string's length in bytes.

### How Groundwork grades your benchmark

- `TestGrader` runs `BenchmarkFormatTags` for three loop passes. It must pass
  and print `allocs/op`. Then it runs it with a *mutant* `FormatTags` that
  panics on its third call; your benchmark must crash, which proves it calls
  `FormatTags` inside the loop.
- `TestFormatTagsAllocs` uses `testing.AllocsPerRun`, which calls a function
  many times and returns the average allocations per call.
- `TestFormatTags` checks that `FormatTags` still gives the right answer.

## Your task

The deploy tool tags every cloud resource it creates with `FormatTags`:

```go
FormatTags([]Tag{
	{Key: "env", Value: "prod"},
	{Key: "team", Value: "sre"},
	{Key: "region", Value: "eu-west-1"},
})
// => "env=prod,team=sre,region=eu-west-1"
```

Each tag becomes `key=value`, tags are joined with commas in the order
given, and an empty list gives `""`. It works, but builds the text with `+=`.
Measure it, then make it allocate less.

### 1. Write BenchmarkFormatTags

**Goal:** make `BenchmarkFormatTags` in `exercise_test.go` report
allocations and call `FormatTags(sampleTags)` inside a `b.Loop` loop.

The starter already has the call; it just isn't in a loop yet. When done,
run `go test -run='^$' -bench=FormatTags` and note the `allocs/op` number.

**Think it through:**
- Which line must come before the loop, and which line goes inside it?
- Compare with `BenchmarkDouble` above: what's different, apart from the
  function being called?

### 2. Allocate once

**Goal:** rewrite `realFormatTags` in `exercise.go` with a `strings.Builder`
so that `FormatTags` makes **at most 1** allocation per call.

`TestFormatTagsAllocs` checks this with the four tags in `sampleTags`, and
`TestFormatTags` must keep passing. Add `"strings"` to the imports. Steps:

1. Loop over the tags once and add up the result's length. Each tag needs
   its key, one `=`, its value and one comma. That counts one comma too
   many in total, which is fine.
2. Reserve that total with `Grow`.
3. Loop again and write each tag, with a `,` before every tag except the
   first.

Run the benchmark again and compare.

**Think it through:**
- Which pieces of each tag come from its fields, and which are fixed text?
- How does the loop know it is on the first tag?

The exercise is done when `TestGrader`, `TestFormatTags` and
`TestFormatTagsAllocs` all pass.

## Learn more

- [Package testing: Benchmarks](https://pkg.go.dev/testing#hdr-Benchmarks)
- [More predictable benchmarking with testing.B.Loop (Go blog)](https://go.dev/blog/testing-b-loop)
- [strings.Builder](https://pkg.go.dev/strings#Builder)
- [Go by Example: Testing and Benchmarking](https://gobyexample.com/testing-and-benchmarking)
