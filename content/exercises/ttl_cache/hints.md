## Task 1: Store and look up entries

`Set` builds an `entry` and puts it in `c.items`. `Get` looks the key up and
has to treat "missing" and "expired" the same way. "Fresh" means
`c.now().Before(e.expiresAt)`.

### Closer

`Set` is one line: `c.items[k] = entry[V]{value: v, expiresAt: ...}`. In
`Get`, use the two-value lookup `e, ok := c.items[k]`, then check `!ok` or
"not fresh". `delete` on a missing key does nothing, so one branch covers
both cases. `var zero V` gives you the zero value to return.

### Answer

```go
// Set
c.items[k] = entry[V]{value: v, expiresAt: c.now().Add(c.ttl)}

// Get
e, ok := c.items[k]
if !ok || !c.now().Before(e.expiresAt) {
	delete(c.items, k)
	var zero V
	return zero, false
}
return e.value, true
```

`!Before` is true when now equals `expiresAt`, so the entry expires exactly
at its TTL.

## Task 2: Delete entries and count them

Both methods are one line each. Go has built-in functions for removing a map
key and counting map entries.

### Closer

`delete(m, key)` removes a key and does nothing if it's missing. `len(m)`
counts the entries.

### Answer

```go
// Delete
delete(c.items, k)

// Len
return len(c.items)
```

## Task 3: Clean up expired entries

Loop over every entry, and for each expired one, delete it and count it. Use
the same "expired" rule as `Get`, with one `now` taken before the loop.

### Closer

Range with `for k, e := range c.items`. Deleting from a map while you range
over it is allowed in Go. Compare with `!now.Before(e.expiresAt)`.

### Answer

```go
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
