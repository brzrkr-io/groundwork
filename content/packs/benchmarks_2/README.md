# Benchmarks: sub-benchmarks and preallocation

Code that is fast for 10 services can crawl for 1000. Here you benchmark the
same function at two sizes and fix a slow pattern that shows up all over Go
code: growing a slice one `append` at a time.

**New here:** sub-benchmarks with `b.Run`, measuring what preallocating with `make` (from *append and copy: fix the inventory*) saves.

## Learn

### Recap

As in *Benchmarks*: a benchmark is a `func BenchmarkXxx(b *testing.B)` in a
`_test.go` file. It calls the code inside `for b.Loop() { ... }`, and
`b.ReportAllocs()` adds the `B/op` and `allocs/op` columns. Run it with
`go test -run='^$' -bench=Ports`.

### Sub-benchmarks with `b.Run`

`b.Run(name, func(b *testing.B) { ... })` runs a *sub-benchmark*, just like
`t.Run` runs a subtest. Each gets its own result line, named `Parent/name`.
A loop over sizes shows how the cost grows:

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

> **Why is it written like this?** `b.Run` requires a `string` name, so
> `strconv.Itoa` turns the number `n` into text like `"10"`. The inner
> function must take exactly one `*testing.B`; reusing the name `b` for it is
> convention, and it hides the outer `b` on purpose. `b.ReportAllocs()` must
> be called inside each sub-benchmark, because sub-benchmarks don't inherit
> it from their parent.

Pick one size with a `/` in the pattern: `go test -run='^$' -bench=Sum/1000`.

### Why append allocates

A slice has a *length* (items in it) and a *capacity* (room in its backing
array). When `append` runs out of room, it allocates a bigger array, about
double the size, and copies everything over. From a nil slice, 1000 appends
trigger about ten such copies.

If you know the final length, make the slice with that capacity up front, and
`append` never has to grow it:

```go
names := make([]string, 0, len(hosts)) // length 0, room for len(hosts)
```

| `make` argument | Meaning |
|---|---|
| `[]string` | the slice type |
| `0` | starting length: empty, so `append` adds from the start |
| `len(hosts)` | capacity: how many items fit before it must grow |

### Grading

`TestGrader` runs `BenchmarkPorts` once and checks that both sub-benchmarks
run, print `allocs/op`, and crash when `Ports` is swapped for a *mutant* that
panics.

## Your task

The load balancer config is generated from `Ports`, which lists the port of
every service in the deploy manifest, in order. Big manifests have thousands
of services.

```go
Ports([]Service{{Name: "api", Port: 8080}, {Name: "web", Port: 443}})
// => [8080 443]
```

### 1. Benchmark two sizes

**Goal:** write `BenchmarkPorts` in `exercise_test.go` with two
sub-benchmarks, named exactly `BenchmarkPorts/10` and `BenchmarkPorts/1000`.

For `n` = 10 and `n` = 1000, start a sub-benchmark named after `n`. In each:

- build a slice of `n` services (any names and ports will do, for example
  port `8000 + i`),
- call `b.ReportAllocs()`,
- call `Ports` with your slice inside a `for b.Loop()` loop.

Add `"strconv"` to the imports. Run the benchmark and note how `allocs/op`
differs between the two sizes.

**Think it through:**
- Compare with `BenchmarkSum` above: which parts change for `Ports`?
- How do you fill each of the `n` services in a slice made with `make`?

### 2. Preallocate

**Goal:** change `realPorts` in `exercise.go` so `out` starts with room for
every port, and `Ports` makes **at most 1** allocation per call.

`TestPortsAllocs` checks this with 100 services. `TestPorts` checks the
result stays right: `nil` services give an empty list, and three services
with ports 80, 443 and 8080 give `[80 443 8080]`. Run your benchmark again:
both sizes should show `1 allocs/op`.

**Think it through:**
- Which line creates `out` now, and what does it start as?
- What is the element type of `out`, and how many ports will it hold?

The exercise is done when `TestGrader`, `TestPorts` and `TestPortsAllocs`
all pass.

## Learn more

- [Package testing: Subtests and Sub-benchmarks](https://pkg.go.dev/testing#hdr-Subtests_and_Sub_benchmarks)
- [Go Slices: usage and internals (Go blog)](https://go.dev/blog/slices-intro)
- [Go by Example: Slices](https://gobyexample.com/slices)
