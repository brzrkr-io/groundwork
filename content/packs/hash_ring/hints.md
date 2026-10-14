## Task 1: Hash a name onto the ring

`hash/crc32` has a function that takes bytes and returns a `uint32` straight
away. `s` is a `string`, so convert it first.

### Closer

`crc32.ChecksumIEEE(...)` takes a `[]byte`; `[]byte(s)` converts the string.

### Answer

```go
return crc32.ChecksumIEEE([]byte(s))
```

## Task 2: Find the next point

`slices.BinarySearch` already gives you "the first element at or after".
Only one position it returns isn't a valid index: handle that one.

### Closer

`i, _ := slices.BinarySearch(points, h)`. When `i` equals `len(points)`,
return `0`; otherwise return `i`.

### Answer

```go
i, _ := slices.BinarySearch(points, h)
if i == len(points) {
	return 0 // past the last point: wrap around to the first
}
return i
```

## Task 3: Place nodes on the ring

Loop over `nodes`. For each one, compute its point, skip it if the map
already has that point, and otherwise store it in both `r.owners` and
`r.points`. Sort after the loop.

### Closer

`_, taken := r.owners[p]` tells you whether the point is already there, and
`continue` skips to the next node. After the loop, `slices.Sort(r.points)`.

### Answer

```go
for _, node := range nodes {
	p := Point(node)
	if _, taken := r.owners[p]; taken {
		continue
	}
	r.owners[p] = node
	r.points = append(r.points, p)
}
slices.Sort(r.points)
```

## Task 4: Find a key's node

Return `""` straight away when `r.points` is empty. Otherwise it's a chain:
hash the key, ask `Next` for an index, read the point at that index, then
look up that point's owner.

### Closer

`i := Next(r.points, Point(key))` is an index. `r.points[i]` is the point,
and `r.owners[...]` of the point is the node.

### Answer

```go
if len(r.points) == 0 {
	return ""
}
return r.owners[r.points[Next(r.points, Point(key))]]
```
