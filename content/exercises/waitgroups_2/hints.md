## Task 1: Fix the hang

Follow what happens when `patch` returns an error: the goroutine returns early and never calls `wg.Done()`. The counter never gets back to 0, so `wg.Wait()` blocks forever.

### Closer

Delete `wg.Add(len(hosts))` and the `wg.Done()` line, and start each goroutine with `wg.Go(func() { ... })` instead of `go func() { ... }()`. You can store the result directly: `errs[i] = patch(host)`.

### Answer

```go
errs := make([]error, len(hosts))
var wg sync.WaitGroup
for i, host := range hosts {
	wg.Go(func() {
		errs[i] = patch(host)
	})
}
wg.Wait()
return errs
```

`wg.Go` marks the goroutine done however it returns, so no path can skip it.

## Task 2: Checksum files concurrently

Writing to a map from several goroutines is a data race. Collect the checksums into a slice first, one slot per file, and build the map only after `wg.Wait()`.

### Closer

Make `results := make([]string, len(files))`, and in each `wg.Go` set `results[i] = sum(file)`. After `wg.Wait()`, loop over `files` again and set `sums[file] = results[i]`.

### Answer

```go
results := make([]string, len(files))
var wg sync.WaitGroup
for i, file := range files {
	wg.Go(func() {
		results[i] = sum(file)
	})
}
wg.Wait()

sums := make(map[string]string, len(files))
for i, file := range files {
	sums[file] = results[i]
}
return sums
```

Only one goroutine writes to the map, after all the checksums are done.

## Task 3: Roll out in two phases

Both phases do the same thing: deploy a list of hosts at the same time and wait for them all. Do it for `canaries`, then do it again for `rest`.

### Closer

Write a helper `deployAll(hosts []string, deploy func(host string))` that starts each deploy with `wg.Go` and calls `wg.Wait()` before returning. `Rollout` then calls it twice.

### Answer

```go
func Rollout(canaries, rest []string, deploy func(host string)) {
	deployAll(canaries, deploy)
	deployAll(rest, deploy)
}

func deployAll(hosts []string, deploy func(host string)) {
	var wg sync.WaitGroup
	for _, host := range hosts {
		wg.Go(func() {
			deploy(host)
		})
	}
	wg.Wait()
}
```

The helper waits before returning, so phase 2 can't start until every canary is done.
