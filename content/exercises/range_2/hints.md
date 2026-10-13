## Task 1: Fix the downtime total

With only one loop variable, `range` gives you the index, not the element.
For `[]int{30, 5, 12}` the loop adds 0 + 1 + 2.

### Closer

Put `_, ` in front of `minutes` so the index is thrown away and `minutes` is
the element: `for _, minutes := range ...`.

### Answer

```go
for _, minutes := range outages {
	total += minutes
}
```

## Task 2: Fix the latency cap

`v` is a copy of the element, so `v = limit` changes only the copy. To change
the slice, you need the index.

### Closer

Keep the index: `for i, v := range ms`. Then assign through it:
`ms[i] = limit`.

### Answer

```go
for i, v := range ms {
	if v > limit {
		ms[i] = limit
	}
}
```

Writing to `ms[i]` changes the element the caller sees too.
