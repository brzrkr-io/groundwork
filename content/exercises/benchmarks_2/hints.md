## Task 1: a loop around b.Run

Loop over `[]int{10, 1000}`. For each `n`, call
`b.Run(strconv.Itoa(n), func(b *testing.B) { ... })`. Everything else from
*Benchmarks* goes inside that function.

## Task 1, nearly done

```go
for _, n := range []int{10, 1000} {
	b.Run(strconv.Itoa(n), func(b *testing.B) {
		svcs := make([]Service, n)
		for i := range svcs {
			svcs[i] = Service{Name: "svc", Port: 8000 + i}
		}
		b.ReportAllocs()
		for b.Loop() {
			Ports(svcs)
		}
	})
}
```

## Task 2: one line

Replace `var out []int` with a `make` call: length `0`, capacity
`len(svcs)`. The loop with `append` stays as it is.
