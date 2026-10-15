## Task 1: Keep the items that match

The signature is done for you. Inside `Filter`, `T` is just a type name, so the body is the same loop you would write for a `[]string`.

### Closer

Start with an empty `var kept []T`, range over `items`, and `append` each item for which `keep(item)` is `true`.

### Answer

```go
var kept []T
for _, item := range items {
	if keep(item) {
		kept = append(kept, item)
	}
}
return kept
```

If nothing matches, `kept` stays `nil`, which the task allows.

## Task 2: Find the largest item

Handle the empty slice first. Then walk the slice, remembering the largest item so far. Don't start "largest so far" at `0`, or negative numbers break.

### Closer

For the empty case, `var zero T` then `return zero, false`. Otherwise start with `largest := items[0]`. Because `T` is `cmp.Ordered`, you may compare with `>`.

### Answer

```go
if len(items) == 0 {
	var zero T
	return zero, false
}
largest := items[0]
for _, item := range items[1:] {
	if item > largest {
		largest = item
	}
}
return largest, true
```

Starting from a real item means the answer is always one of the values in the slice.

## Task 3: Remember what you've seen

`Set[T]` is really a `map[T]bool`, so adding and checking are ordinary map operations on `s`.

### Closer

`s[v] = true` adds a value and `s[v]` reads one. A missing key reads as `false`, the zero value of `bool`.

### Answer

```go
func (s Set[T]) Add(v T) {
	s[v] = true
}

func (s Set[T]) Has(v T) bool {
	return s[v]
}
```

Maps are shared, so `Add` changes the set even with a value receiver.
