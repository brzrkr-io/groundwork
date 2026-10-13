## Task 1: swap the values, not the pointers

`live, standby = standby, live` swaps the addresses held by the two local
variables. You want to swap what they point *at*, so put `*` on every name:
`*live, *standby = *standby, *live`.

## Task 2: copy the struct

`*h` is the `Host` value that `h` points to. Store it in a new variable, which
makes a copy, and return the address of that new variable.

## Task 3: two edits

First change the parameter of `markDrained` from `h Host` to `h *Host`. The
body can stay the same, because `h.Drained` follows the pointer. Then the
call in `DrainHost` no longer compiles until you pass an address:
`&hosts[i]`.

## Almost the answer

```go
func CloneHost(h *Host) *Host {
	c := *h
	return &c
}

func markDrained(h *Host) {
	h.Drained = true
	h.Sessions = 0
}

// In DrainHost:
markDrained(&hosts[i])
```
