## Task 1: let wg.Go count

Delete `wg.Add(len(hosts))` and the `wg.Done()` at the bottom, and start each
goroutine with `wg.Go(func() { errs[i] = patch(host) })`. `wg.Go` marks the
goroutine done however it returns, so nothing can skip it.

## Task 2: slice first, map later

```go
results := make([]string, len(files))
var wg sync.WaitGroup
for i, file := range files {
	wg.Go(func() {
		results[i] = sum(file)
	})
}
wg.Wait()
```

Then loop over `files` again and set `sums[file] = results[i]`.

## Task 3: the same thing twice

Write a small helper `deployAll(hosts, deploy)` that starts every deploy with
`wg.Go` and then calls `wg.Wait()`. `Rollout` calls it for `canaries`, then
for `rest`. Because the helper waits before returning, phase 2 can't start
early.
