## Task 1: Fix the labels

The label is built from two parameters joined with `+`. Put a piece of fixed
text between them that never appears in a node name.

### Closer

The doc comment says `node + "#" + i`. `i` is an `int`, so it still goes
through `strconv.Itoa`.

### Answer

```go
return node + "#" + strconv.Itoa(i)
```

## Task 2: Fix the wrap-around

When `slices.BinarySearch` returns `len(r.points)`, the key is past the last
point. Instead of giving up, move to the first point. An empty ring is the
only case that returns `""`, and it needs its own check before any indexing.

### Closer

Start with `if len(r.points) == 0 { return "" }`. After the search, turn
`i == len(r.points)` into `i = 0`, then always return
`r.owners[r.points[i]]`.

### Answer

```go
if len(r.points) == 0 {
	return ""
}
i, _ := slices.BinarySearch(r.points, hash(key))
if i == len(r.points) {
	i = 0 // past the last point: wrap around to the first
}
return r.owners[r.points[i]]
```

## Task 3: Fix the sort

Sorting only the new points and appending them leaves the old and new
points out of order. Sort the whole `r.points` once all the new points are
in it.

### Closer

Append each point straight to `r.points` inside the loop (so `added` can
go), and call `slices.Sort(r.points)` at the end of `Add`.

### Answer

```go
for _, node := range nodes {
	for i := range r.replicas {
		p := hash(label(node, i))
		if _, taken := r.owners[p]; taken {
			continue
		}
		r.owners[p] = node
		r.points = append(r.points, p)
	}
}
slices.Sort(r.points) // the whole ring, not just the new points
```
