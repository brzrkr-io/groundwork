## Task 1: Reset a counter

`count` is an address, not a number. You need to change the number that lives at that address.

### Closer

Put `*` in front of a pointer to reach the value it points to, on the left of `=` too: `*count = ...`.

### Answer

```go
*count = 0
```

## Task 2: Scale a deployment

Check for a `nil` pointer first and stop if there is one. Then fix up a negative `replicas` before you store it.

### Closer

Start with `if d == nil { return }`. After that, `d.Replicas = ...` sets the field on the caller's deployment, because Go follows the pointer for you.

### Answer

```go
if d == nil {
	return
}
if replicas < 0 {
	replicas = 0
}
d.Replicas = replicas
```

## Task 3: Find a deployment

You must return the address of the real element inside `ds`, not of a copy. Think about which kind of loop lets you reach the element itself.

### Closer

Loop with `for i := range ds` and take the address with `&ds[i]`. A pointer to the `range` variable (`&d`) points at a copy, and the test will notice.

### Answer

```go
for i := range ds {
	if ds[i].Name == name {
		return &ds[i]
	}
}
return nil
```
