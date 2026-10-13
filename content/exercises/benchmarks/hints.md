## Task 1: two lines of setup

Inside `BenchmarkFormatTags`, call `b.ReportAllocs()` first. Then write a
`for b.Loop()` loop whose body calls `FormatTags(sampleTags)`. You can ignore
the result.

## Task 1, nearly done

```go
func BenchmarkFormatTags(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		FormatTags(sampleTags)
	}
}
```

## Task 2: count, then grow

First add up the size in one loop:

```go
n := 0
for _, t := range tags {
	n += len(t.Key) + 1 + len(t.Value) + 1
}
```

Then make a `var b strings.Builder`, call `b.Grow(n)`, and write the tags in
a second loop. Don't build pieces with `+` inside the loop, because every `+`
makes a new string.

## Task 2, nearly done

```go
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
```
