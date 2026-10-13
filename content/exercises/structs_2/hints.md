## Task 1: Mark every host unhealthy

In `for _, h := range hosts`, `h` is a copy of each host, so changing it never touches `hosts`. You need to change the element inside the slice instead.

### Closer

Loop with `for i := range hosts` to get only the index, then change the element in place with `hosts[i].Healthy = ...`.

### Answer

```go
for i := range hosts {
	hosts[i].Healthy = false
}
```

## Task 2: Record a failed check

`h := hosts[i]` copies the host, so `h.Failures++` changes the copy and the change is lost. Make the changes reach `hosts[i]` itself.

### Closer

Drop `h` and write `hosts[i].Failures++` and `hosts[i].Healthy = false` directly. (Keeping the copy and writing it back with `hosts[i] = h` before `return` also works.)

### Answer

```go
if hosts[i].Name == name {
	hosts[i].Failures++
	if hosts[i].Failures >= 3 {
		hosts[i].Healthy = false
	}
	return
}
```

## Task 3: Reset every host

`reset` gets a copy of the host, so its changes vanish when it returns. Have it hand back the changed copy, and store that back in the slice.

### Closer

Give `reset` a return type, `func reset(h Host) Host`, and end it with `return h`. In `ResetAll`, assign the result: `hosts[i] = reset(...)`.

### Answer

```go
func ResetAll(hosts []Host) {
	for i := range hosts {
		hosts[i] = reset(hosts[i])
	}
}

func reset(h Host) Host {
	h.Failures = 0
	h.Healthy = true
	return h
}
```
