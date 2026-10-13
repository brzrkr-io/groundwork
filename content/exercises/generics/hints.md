## Task 1: an ordinary loop

The signature is done for you. Inside `Filter`, `T` is just a type name, so
the body is the same loop you would write for `[]string`: start with an empty
`[]T`, range over `items`, and `append` each item that `keep` accepts.

## Task 2: start from the first item

Handle the empty slice first with `var zero T` and `return zero, false`. Then
start `largest` at `items[0]`, not at `0`, so negative numbers work. Because
`T` is `cmp.Ordered`, you may compare with `>`.

## Task 3: a Set is a map

`Set[T]` is a `map[T]bool`, so `s[v] = true` adds a value and `s[v]` reads
one. A missing key reads as `false`, the zero value of `bool`.

## Almost the answer

```go
var kept []T
for _, item := range items {
	if keep(item) {
		kept = append(kept, item)
	}
}
return kept
```

```go
largest := items[0]
for _, item := range items[1:] {
	if item > largest {
		largest = item
	}
}
return largest, true
```

```go
func (s Set[T]) Add(v T) { s[v] = true }

func (s Set[T]) Has(v T) bool { return s[v] }
```
