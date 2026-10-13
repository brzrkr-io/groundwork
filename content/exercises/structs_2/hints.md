## Find the copy

In each function, look for a variable that holds a `Host`: `h` in the
`range` loop, `h := hosts[i]`, and the `h` parameter of `reset`. Each one is
a copy. Changing it never touches `hosts`.

## Tasks 1 and 2: go through the index

Loop with `for i := range hosts` and change the element in place:

```go
hosts[i].Healthy = false
```

In task 2, either change `hosts[i].Failures` directly, or keep the copy and
write it back with `hosts[i] = h` before `return`.

## Task 3: return the copy

Give `reset` a return type and return the changed copy:

```go
func reset(h Host) Host {
	h.Failures = 0
	h.Healthy = true
	return h
}
```

Then store the result in `ResetAll`: `hosts[i] = reset(hosts[i])`.
