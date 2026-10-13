## Task 1: Write BenchmarkFormatTags

A benchmark repeats the same call many times and measures it. You need two things: ask it to count allocations, and put the call inside the benchmark loop.

### Closer

Call `b.ReportAllocs()` first. Then use `for b.Loop() { ... }` with `FormatTags(sampleTags)` inside. You can ignore the result.

### Answer

```go
func BenchmarkFormatTags(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		FormatTags(sampleTags)
	}
}
```

## Task 2: Allocate once

Every `+=` on a string makes a brand new string. If you work out the final size first, a `strings.Builder` can reserve that space once and fill it in.

### Closer

First add up the size: `n += len(t.Key) + 1 + len(t.Value) + 1` for each tag. Then `var b strings.Builder`, `b.Grow(n)`, and write each piece with `b.WriteString`. Write `","` only when `i > 0`.

### Answer

```go
func realFormatTags(tags []Tag) string {
	n := 0
	for _, t := range tags {
		n += len(t.Key) + 1 + len(t.Value) + 1
	}

	var b strings.Builder
	b.Grow(n)
	for i, t := range tags {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(t.Key)
		b.WriteString("=")
		b.WriteString(t.Value)
	}
	return b.String()
}
```

Remember to add `"strings"` to the imports.
