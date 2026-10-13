# Benchmarks

Tests tell you whether code is right. *Benchmarks* tell you how fast it is and
how much memory it uses. When a function runs for every request or every log
line, that is worth knowing before you try to speed it up.

## Learn

### Writing a benchmark

A benchmark lives in a `_test.go` file, next to your tests. Its name starts
with `Benchmark`, and it takes a `*testing.B` instead of a `*testing.T`:

```go
func BenchmarkDouble(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		Double(21)
	}
}
```

- `for b.Loop() { ... }` runs the body again and again until Go has a
  stable timing. Anything before the loop is setup and isn't timed. `b.Loop`
  also stops the compiler from optimizing away a call whose result you
  ignore, which would otherwise make the benchmark measure nothing. It
  arrived in Go 1.24. Older code writes `for i := 0; i < b.N; i++` instead,
  and you'll still see that in the wild.
- `b.ReportAllocs()` adds memory numbers to the result (more on those below).

`b` also has `Errorf` and `Fatalf`, just like `t`.

### Running benchmarks

`go test` skips benchmarks unless you ask for them with `-bench`, which takes
a regular expression, like `-run` does:

```text
go test -bench=FormatTags              run tests, then matching benchmarks
go test -run='^$' -bench=FormatTags    run only the benchmarks
```

`-run='^$'` selects tests whose name is empty. No test has an empty name, so
no tests run and the output shows only the benchmark.

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
| `7 allocs/op` | *Allocations* per call: how often Go had to ask for new memory. |

Your numbers will differ: they depend on the machine. Compare runs on the
same machine, before and after a change.

### Why allocations cost

Each allocation is work for Go's memory allocator. Later, the garbage
collector has to find and free that memory again. Fewer allocations usually
means faster code.

Strings in Go can't be changed. So `s += "x"` doesn't extend `s`. It
allocates a new string, copies the old text and `"x"` into it, and throws
the old one away. In a loop, that adds up.

A `strings.Builder` collects text in one growing buffer instead:

```go
var b strings.Builder
b.Grow(12)              // reserve room for 12 bytes up front
b.WriteString("web")
b.WriteString("-")
b.WriteString("01")
s := b.String()         // "web-01", no extra copy
```

Without `Grow`, the builder still has to enlarge its buffer a few times as
the text grows. If you tell `Grow` how many bytes you'll write, it allocates
once and never again. `len(s)` gives you the length of a string in bytes.

### How Gopherlings grades your benchmark

`TestGrader` in `grader_test.go` re-runs `BenchmarkFormatTags` once (with
`-test.benchtime=1x`). It must run, pass, and print `allocs/op`. Then it runs
it again with a *mutant*, a copy of `FormatTags` that panics. Your benchmark
must crash then, which proves it really calls `FormatTags`. For the speed-up,
`TestFormatTagsAllocs` uses `testing.AllocsPerRun`, which calls a function
many times and returns the average number of allocations per call.
`TestFormatTags` checks that `FormatTags` still gives the right answer.

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

Each tag becomes `key=value`, the tags are joined with commas in the order
given, and no tags give `""`. It works, but it builds the text with `+=`.
Measure it, then make it allocate less.

### 1. Write BenchmarkFormatTags

In `exercise_test.go`, fill in `BenchmarkFormatTags`:

- call `b.ReportAllocs()`,
- move the starter's `FormatTags(sampleTags)` call inside a `for b.Loop()`
  loop.

`TestGrader` passes when the benchmark runs on the real code, prints
`allocs/op`, and crashes when `FormatTags` panics. Run
`go test -run='^$' -bench=FormatTags` and note the `allocs/op` number.

### 2. Allocate once

Rewrite `realFormatTags` in `exercise.go` with a `strings.Builder` (add
`"strings"` to the imports):

1. Loop over the tags once to add up the length of the result. Each tag needs
   `len(t.Key) + 1 + len(t.Value) + 1` bytes: the key, `=`, the value and a
   comma. That is one comma too many in total, which is fine.
2. Call `Grow` with that total.
3. Loop again and write each tag, with a `,` before every tag except the
   first.

`TestFormatTagsAllocs` passes when `FormatTags` makes **at most 1**
allocation per call for the four tags in `sampleTags`. `TestFormatTags` must
keep passing. Run the benchmark again and compare.

The exercise is done when `TestGrader`, `TestFormatTags` and
`TestFormatTagsAllocs` all pass.

## Learn more

- [Package testing: Benchmarks](https://pkg.go.dev/testing#hdr-Benchmarks)
- [More predictable benchmarking with testing.B.Loop (Go blog)](https://go.dev/blog/testing-b-loop)
- [strings.Builder](https://pkg.go.dev/strings#Builder)
- [Go by Example: Testing and Benchmarking](https://gobyexample.com/testing-and-benchmarking)
