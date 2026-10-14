## Task 1: Break ties in list order

Replace the loop over the `lc.active` map with a loop over `lc.addrs`, which
keeps the list order. Start with the first backend as the best so far, and
only switch when another one has strictly fewer requests in flight.

### Closer

`best := lc.addrs[0]`, then `for _, a := range lc.addrs[1:]`. Look up
counts with `lc.active[a]` and compare with `<`, not `<=`.

### Answer

```go
best := lc.addrs[0]
for _, a := range lc.addrs[1:] {
	if lc.active[a] < lc.active[best] {
		best = a
	}
}
lc.active[best]++
```

On a tie, `<` is false, so the earlier backend stays the best.

## Task 2: Release only once

Declare a `sync.Once` inside `Acquire`, right before the closure, so each
request gets its own. Inside `release`, put the decrement in `once.Do`.

### Closer

```go
var once sync.Once
release = func() {
	once.Do(func() {
		// the decrement goes here
	})
}
```

### Answer

```go
var once sync.Once
release = func() {
	once.Do(func() {
		lc.active[best]--
	})
}
```

## Task 3: Fix the data race in release

The racing write is `lc.active[best]--` inside the closure. `Acquire`'s
deferred unlock ran when `Acquire` returned, long before `release` is
called, so the closure has to lock `lc.mu` itself.

### Closer

Inside the function you pass to `once.Do`, lock `lc.mu` and defer the
unlock before the decrement.

### Answer

```go
var once sync.Once
release = func() {
	once.Do(func() {
		lc.mu.Lock()
		defer lc.mu.Unlock()
		lc.active[best]--
	})
}
```
