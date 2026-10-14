## Task 1: Deploy every region at once

Right now each `deploy(region)` runs before the next one starts. Run each one in its own goroutine with a `sync.WaitGroup`, then wait for all of them after the loop.

### Closer

Declare `var wg sync.WaitGroup` before the loop. Inside it, wrap the call: `wg.Go(func() { ... })`. After the loop, call `wg.Wait()`.

### Answer

```go
var wg sync.WaitGroup
for _, region := range regions {
	wg.Go(func() {
		deploy(region)
	})
}
wg.Wait()
```

`wg.Wait()` blocks until every goroutine started with `wg.Go` has returned.

## Task 2: Resolve hostnames in order

The `addrs` slice already has one slot per host. If each goroutine writes only to its own slot `addrs[i]`, the order stays right however the lookups finish.

### Closer

Keep `addrs[i] = resolve(host)`, but move it inside `wg.Go(func() { ... })`. Call `wg.Wait()` after the loop and before `return addrs`.

### Answer

```go
addrs := make([]string, len(hosts))
var wg sync.WaitGroup
for i, host := range hosts {
	wg.Go(func() {
		addrs[i] = resolve(host)
	})
}
wg.Wait()
return addrs
```

Each goroutine writes a different index, so they never touch the same slot.

## Task 3: Start a backup from a helper

`startBackup` already receives `wg` as a pointer, so it can start goroutines that `BackupAll`'s `wg.Wait()` will wait for. The three lines that are there now just run the backup straight away.

### Closer

Replace `wg.Add(1)`, `backup(db)` and `wg.Done()` with a single `wg.Go` call. `wg` is a pointer, but you still call its methods as `wg.Go(...)`.

### Answer

```go
wg.Go(func() {
	backup(db)
})
```

`wg.Go` counts the goroutine in the caller's WaitGroup and marks it done when `backup` returns.
