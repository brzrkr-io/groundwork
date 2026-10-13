# WaitGroups: patching the fleet

More WaitGroup practice on a fleet-patching tool: a hang to fix, a map to
fill safely, and a rollout in two phases.

## Learn

### Recap

`wg.Go(f)` runs `f` in a new goroutine and counts it, and `wg.Wait()` blocks
until every counted goroutine has finished. Results go in a slice made before
the loop, one slot per goroutine.

### Every path must call Done

Older code counts goroutines by hand: `wg.Add(n)` for the goroutines it is
about to start, and `wg.Done()` in each one when it finishes. If a goroutine
can return without calling `wg.Done()`, the counter never reaches 0, and
`wg.Wait()` blocks forever. An early `return` is the usual culprit:

```go
wg.Add(len(hosts))
for _, host := range hosts {
	go func() {
		if err := patch(host); err != nil {
			return // BUG: skips wg.Done(), so Wait never returns
		}
		wg.Done()
	}()
}
```

The best fix is to drop `Add` and `Done` and use `wg.Go`. It calls `Done` for
you however the function ends: normal return, early return or panic. When you
keep the older form, put `defer wg.Done()` on the first line of the goroutine,
which has the same effect.

### Don't write to a map from goroutines

A map must not be written by several goroutines at once. The race detector
reports it, and Go may even crash with `fatal error: concurrent map writes`.
You'll learn to protect a map with a mutex in *Mutexes*. For now, fill a slice
by index in the goroutines, and build the map **after** `wg.Wait()`, back in a
single goroutine.

### Phases

Once `Wait` has returned, you can start another batch of goroutines and wait
again. That gives you phases: everything in phase 1 finishes before anything
in phase 2 starts.

```text
phase 1:  canary-1 ─┐
          canary-2 ─┴─ Wait ─▶ phase 2:  web-1 ─┐
                                         web-2 ─┤
                                         web-3 ─┴─ Wait ─▶ done
```

## Your task

Every patch, checksum or deploy in the tests takes 1 (fake) second.

### 1. Fix the hang

`PatchAll(hosts, patch)` patches every host at the same time and returns one
error per host, in the same order as `hosts` (`nil` for hosts that patched
fine). It works when every patch succeeds, but it hangs forever as soon as one
fails. Fix it, ideally by switching to `wg.Go`.

```go
// patch fails for "web-2" with "web-2: disk full"
PatchAll([]string{"web-1", "web-2"}, patch)
// => []error{nil, errors.New("web-2: disk full")}
```

### 2. Checksum files concurrently

`ChecksumAll(files, sum)` returns a map from each file name to `sum(file)`.
It works, but it handles one file at a time. Make it compute every checksum
at the same time. Don't write to the map from your goroutines: fill a slice by
index, then build the map after `wg.Wait()`.

```go
ChecksumAll([]string{"app.tar", "config.yaml"}, sum)
// => map[string]string{"app.tar": "9f86d0", "config.yaml": "2c26b4"}   (after ~1s)
```

### 3. Roll out in two phases

Write `Rollout(canaries, rest, deploy)`. It deploys all `canaries` at the same
time and waits for them. Only then does it deploy all `rest` hosts at the same
time and wait for them. With 1-second deploys, the whole rollout takes 2
seconds, and no host in `rest` starts before every canary has finished.

```go
Rollout([]string{"canary-1", "canary-2"}, []string{"web-1", "web-2", "web-3"}, deploy)
// canaries finish at 1s, the rest finish at 2s
```

## Learn more

- [sync.WaitGroup](https://pkg.go.dev/sync#WaitGroup)
- [Go spec: Defer statements](https://go.dev/ref/spec#Defer_statements)
- [Go FAQ: Why are map operations not defined to be atomic?](https://go.dev/doc/faq#atomic_maps)
