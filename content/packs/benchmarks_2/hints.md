## Task 1: Benchmark two sizes

Loop over the two sizes and run a separate sub-benchmark for each one. Inside each sub-benchmark, the body is a normal benchmark like the one in *Benchmarks*.

### Closer

Use `for _, n := range []int{10, 1000}` and `b.Run(strconv.Itoa(n), func(b *testing.B) { ... })`. Inside, build `svcs := make([]Service, n)`, fill it in a loop, then call `b.ReportAllocs()` and `for b.Loop()`.

### Answer

```go
func BenchmarkPorts(b *testing.B) {
	for _, n := range []int{10, 1000} {
		b.Run(strconv.Itoa(n), func(b *testing.B) {
			svcs := make([]Service, n)
			for i := range svcs {
				svcs[i] = Service{Name: "svc-" + strconv.Itoa(i), Port: 8000 + i}
			}
			b.ReportAllocs()
			for b.Loop() {
				Ports(svcs)
			}
		})
	}
}
```

## Task 2: Preallocate

You already know how many ports there will be: one per service. If the slice starts with room for all of them, `append` never has to grow it.

### Closer

Replace `var out []int` with a `make` call that has length `0` and capacity `len(svcs)`. The `append` loop stays as it is.

### Answer

```go
out := make([]int, 0, len(svcs))
for _, s := range svcs {
	out = append(out, s.Port)
}
return out
```
