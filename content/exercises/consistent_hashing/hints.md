## Task 1: Place nodes on the ring

It's the `Add` from *Hash rings: find a key's node* with one more loop: for
each node, count `i` up to `r.replicas`. Build each label from the node,
`"#"` and the number, hash it, and store the point in both `r.owners` and
`r.points` unless the map already has it.

### Closer

`for i := range r.replicas` counts from 0. `strconv.Itoa(i)` turns the
number into text, and `_, taken := r.owners[p]` tells you whether the point
is already there. After both loops, `slices.Sort(r.points)`.

### Answer

```go
for _, node := range nodes {
	for i := range r.replicas {
		p := hash(node + "#" + strconv.Itoa(i))
		if _, taken := r.owners[p]; taken {
			continue
		}
		r.owners[p] = node
		r.points = append(r.points, p)
	}
}
slices.Sort(r.points)
```

## Task 2: Find a key's node

Return `""` straight away when `r.points` is empty. Otherwise let
`slices.BinarySearch` find the position of `hash(key)` in `r.points`, and
handle the one position that isn't a valid index.

### Closer

`i, _ := slices.BinarySearch(r.points, hash(key))`. When `i` equals
`len(r.points)`, set it to `0`. Then look up `r.points[i]` in `r.owners`.

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

## Task 3: Remove a node's points

Keep only the points whose owner isn't `node`. The point is a number, so ask
`r.owners` who sits there.

### Closer

`r.points = slices.DeleteFunc(r.points, func(p uint32) bool { ... })`, where
the function returns `true` for points to delete: `r.owners[p] == node`.

### Answer

```go
r.points = slices.DeleteFunc(r.points, func(p uint32) bool {
	return r.owners[p] == node
})
```

## Task 4: Forget a node's owners

Now drop every map entry whose value is `node`. This goes after the task 3
code, because that code still needs the map to know who owns each point.

### Closer

`maps.DeleteFunc(r.owners, func(_ uint32, owner string) bool { ... })`. The
point (the key) doesn't matter here, so `_` ignores it.

### Answer

```go
maps.DeleteFunc(r.owners, func(_ uint32, owner string) bool {
	return owner == node
})
```
