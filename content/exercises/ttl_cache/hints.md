## Task 1: build an entry

`Set` is one line: `c.items[k] = entry[V]{value: v, expiresAt: ...}`. In
`Get`, use the two-value lookup `e, ok := c.items[k]`, then check `ok` and
`c.now().Before(e.expiresAt)`.

## Task 2: built-ins do the work

`delete(c.items, k)` removes a key and does nothing if it's missing.
`len(c.items)` counts the entries.

## Task 3: loop and delete

Range over `c.items` with `for k, e := range c.items`. When an entry is not
fresh, delete it and add one to a counter.

## Near-solution

```go
// Get
e, ok := c.items[k]
if !ok || !c.now().Before(e.expiresAt) {
	delete(c.items, k)
	var zero V
	return zero, false
}
return e.value, true

// Cleanup
now := c.now()
removed := 0
for k, e := range c.items {
	if !now.Before(e.expiresAt) {
		delete(c.items, k)
		removed++
	}
}
return removed
```
