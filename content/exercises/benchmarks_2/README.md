# Benchmarks: sub-benchmarks and preallocation

Code that is fast for 10 services can crawl for 1000. Here you benchmark the
same function at two sizes and fix a slow pattern that shows up all over Go
code: growing a slice one `append` at a time.

## Learn

### Recap

As in *Benchmarks*, a benchmark is a
`func BenchmarkXxx(b *testing.B)` in a `_test.go` file. It calls the code
inside `for b.Loop() { ... }`, and `b.ReportAllocs()` adds the `B/op` and
`allocs/op` columns. Run it with `go test -run='^$' -bench=Ports`.

### Sub-benchmarks with b.Run

`b.Run(name, func(b *testing.B) { ... })` runs a *sub-benchmark*, just like
`t.Run` runs a subtest. Each one gets its own result line, named
`Parent/name`. A loop over sizes is the usual way to see how the cost grows:

```go
func BenchmarkSum(b *testing.B) {
	for _, n := range []int{10, 1000} {
		b.Run(strconv.Itoa(n), func(b *testing.B) {
			nums := make([]int, n) // setup, not timed
			b.ReportAllocs()
			for b.Loop() {
				Sum(nums)
			}
		})
	}
}
```

```text
BenchmarkSum/10-16       ...
BenchmarkSum/1000-16     ...
```

Call `b.ReportAllocs()` inside each sub-benchmark. A sub-benchmark doesn't
inherit it from its parent. You can pick one size with a `/` in the pattern:
`go test -run='^$' -bench=Sum/1000`.

### Why append allocates

A slice has a length and a *capacity* (the room in its backing array). When
`append` runs out of capacity, it allocates a bigger array, roughly double the
size, and copies everything over. Starting from a nil slice, 1000 appends
trigger about ten of those copies. If you know the final length, make the
slice with that capacity, and `append` never has to grow it:

```go
out := make([]string, 0, len(hosts)) // length 0, room for len(hosts)
```

`TestGrader` re-runs `BenchmarkPorts` once and checks both sub-benchmarks
run, print `allocs/op`, and crash when `Ports` is swapped for a mutant that
panics.

## Your task

The load balancer config is generated from `Ports`, which lists the port of
every service in the deploy manifest, in order. The big manifests have
thousands of services.

```go
Ports([]Service{{Name: "api", Port: 8080}, {Name: "web", Port: 443}})
// => [8080 443]
```

### 1. Benchmark two sizes

Write `BenchmarkPorts` in `exercise_test.go`. For `n` = 10 and `n` = 1000, run
a sub-benchmark named after `n` with `b.Run`, so the results are named
exactly `BenchmarkPorts/10` and `BenchmarkPorts/1000`. In each one:

- build a slice of `n` services (any names and ports will do, for example
  port `8000 + i`),
- call `b.ReportAllocs()`,
- call `Ports` with your slice inside a `for b.Loop()` loop.

Add `"strconv"` to the imports and use `strconv.Itoa(n)` for the name. Run
the benchmark and note how `allocs/op` differs between the two sizes.

### 2. Preallocate

Change `realPorts` in `exercise.go` to create `out` with
`make([]int, 0, len(svcs))` instead of starting from a nil slice.

`TestPortsAllocs` passes when `Ports` makes **at most 1** allocation per call
for 100 services. `TestPorts` checks the result stays right: `nil` services
give an empty list, and three services with ports 80, 443 and 8080 give
`[80 443 8080]`. Run your benchmark again: both sizes should show
`1 allocs/op`.

The exercise is done when `TestGrader`, `TestPorts` and `TestPortsAllocs`
all pass.

## Learn more

- [Package testing: Subtests and Sub-benchmarks](https://pkg.go.dev/testing#hdr-Subtests_and_Sub_benchmarks)
- [Go Slices: usage and internals (Go blog)](https://go.dev/blog/slices-intro)
- [Go by Example: Slices](https://gobyexample.com/slices)
