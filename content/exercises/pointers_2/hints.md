## Task 1: Fix SwapPorts

`live, standby = standby, live` swaps the addresses held by two local variables, which the caller never sees. You want to swap the ints they point at.

### Closer

Put `*` on every name, on both sides of `=`, so the swap works on the pointed-to values.

### Answer

```go
*live, *standby = *standby, *live
```

## Task 2: Fix CloneHost

Returning `h` hands back the same host, so changes to the "clone" change the original. You need a new `Host` value with the same fields.

### Closer

`*h` is the `Host` value that `h` points to. Storing it in a new variable (`c := *h`) makes a copy; then return that variable's address.

### Answer

```go
c := *h
return &c
```

## Task 3: Fix DrainHost

`markDrained` takes a `Host`, so it changes its own copy. It needs a pointer to the host in the slice instead.

### Closer

Change the parameter from `h Host` to `h *Host`; the body can stay the same, because `h.Drained` follows the pointer. Then pass an address in `DrainHost`: `&hosts[i]`.

### Answer

```go
func markDrained(h *Host) {
	h.Drained = true
	h.Sessions = 0
}

// In DrainHost:
markDrained(&hosts[i])
```
