## Task 1: Add new entries

Two steps: put a new `*entry` on the front of the list, then remember the
element the list gives back, under `key` in the map.

### Closer

`c.order.PushFront(x)` adds `x` at the front and returns its
`*list.Element`. Build `x` with `&entry{key: ..., value: ...}`, then store
the element: `c.items[key] = ...`.

### Answer

```go
c.items[key] = c.order.PushFront(&entry{key: key, value: value})
```

## Task 2: Look up an entry and mark it used

Find the element in `c.items`. A miss returns straight away. A hit moves the
element to the front, then reads the value out of the `*entry` it holds.

### Closer

`el, ok := c.items[key]` tells a hit from a miss. `c.order.MoveToFront(el)`
marks it used, and `el.Value.(*entry)` gives you the entry, so
`el.Value.(*entry).value` is the value.

### Answer

```go
el, ok := c.items[key]
if !ok {
	return "", false
}
c.order.MoveToFront(el)
return el.Value.(*entry).value, true
```

## Task 3: Update existing entries

Check the map first. If the key is there, change the value inside its entry,
move the element to the front, and `return` before the task 1 code adds a
second element.

### Closer

Use `if el, ok := c.items[key]; ok { ... }`. Inside, assign to
`el.Value.(*entry).value`, call `c.order.MoveToFront(el)`, and `return`.

### Answer

```go
if el, ok := c.items[key]; ok {
	el.Value.(*entry).value = value
	c.order.MoveToFront(el)
	return
}
```

The assertion gives you the pointer stored in the list, so the assignment
changes the cached entry itself.

## Task 4: Evict the least recently used entry

After the new entry is in, compare the number of entries with `c.capacity`.
If there are too many, the victim is at the back of the list. Remove it from
both the list and the map, and count it. Count with the fields, not with
`c.Len()`: once task 5 locks `Put`, `c.Len()` would wait for that lock
forever.

### Closer

`oldest := c.order.Back()` is the least recently used element.
`c.order.Remove(oldest)` unlinks it, and `oldest.Value.(*entry).key` is the
key to `delete` from `c.items`. Then `c.evictions++`.

### Answer

```go
if c.order.Len() > c.capacity {
	oldest := c.order.Back()
	c.order.Remove(oldest)
	delete(c.items, oldest.Value.(*entry).key)
	c.evictions++
}
```

## Task 5: Make it safe for concurrent use

Copy how `Len` starts into the top of `Get` and of `Put`. Then check that
neither method calls `c.Len()`, `c.Keys()` or `c.Get()`.

### Closer

`c.mu.Lock()` on one line, `defer c.mu.Unlock()` on the next, as the very
first lines of each method.

### Answer

At the top of both `Get` and `Put`:

```go
c.mu.Lock()
defer c.mu.Unlock()
```
