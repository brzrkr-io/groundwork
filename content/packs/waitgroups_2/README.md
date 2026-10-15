# WaitGroups: patching the fleet

More WaitGroup practice on a fleet-patching tool: a hang to fix, a map to
fill safely, and a rollout in two phases.

**New here:** a few new patterns, no new syntax.

## Learn

### Recap

`wg.Go(f)` runs `f` in a new goroutine and counts it. `wg.Wait()` blocks until
every counted goroutine has finished. Results go in a slice made before the
loop, one slot per goroutine (`results[i]`).

### Every path must call Done

In the older style you count by hand: `wg.Add(n)` before starting `n`
goroutines, `wg.Done()` in each as it finishes. If any goroutine can return
**without** calling `wg.Done()`, the counter never reaches 0 and `wg.Wait()`
blocks forever. An early `return` is the usual culprit:

```go
wg.Add(len(jobs))
for _, job := range jobs {
	go func() {
		if job.Skip {
			return // BUG: skips wg.Done(), so Wait never returns
		}
		run(job)
		wg.Done()
	}()
}
```

Two fixes:

| Fix | How |
|---|---|
| `wg.Go` (best) | drop `Add` and `Done`; `wg.Go` calls `Done` however the function ends |
| `defer wg.Done()` | first line of the goroutine; runs on every `return` |

### Don't write to a map from goroutines

A map must not be written by several goroutines at once. The race detector
reports it, and Go may crash with `fatal error: concurrent map writes`.
(*Mutexes* shows how to protect a map.) For now: fill a slice by index in the
goroutines, then build the map **after** `wg.Wait()`, when only one goroutine
is left.

> **Why is it written like this?** Go *requires* that only one goroutine
> touches a map while it is being written; breaking that is a bug, not a
> style issue. Filling a slice first and building the map later is *your*
> choice; it is the simplest safe way.

### Phases

After `Wait` returns, you can start a new batch and wait again. Everything in
phase 1 finishes before anything in phase 2 starts:

```text
phase 1:  canary-1 ─┐
          canary-2 ─┴─ Wait ─▶ phase 2:  web-1 ─┐
                                         web-2 ─┤
                                         web-3 ─┴─ Wait ─▶ done
```

When two phases do the same job on different lists, a small helper function
saves you writing the loop twice. You choose its name and parameters.

## Your task

Every checksum or deploy in the tests takes 1 (fake) second. Patches return
at once; a hung `PatchAll` fails its test after 2 seconds.

### 1. Fix the hang

**Goal:** make `PatchAll(hosts, patch)` return without hanging, even when some
patches fail.

It patches every host at the same time and returns one `error` per host, in
the same order as `hosts`: the error from `patch(host)`, or `nil` if it
succeeded. Right now it hangs as soon as one patch fails.

```go
// patch fails for "web-2" with "web-2: disk full"
PatchAll([]string{"web-1", "web-2"}, patch)
// => []error{nil, errors.New("web-2: disk full")}
```

**Think it through:**
- Follow a failing host through the goroutine. Which line does it never reach?
- Which fix in the table above makes that impossible?

### 2. Checksum files concurrently

**Goal:** make `ChecksumAll(files, sum)` compute every checksum at the same
time and return a map from each file name to `sum(file)`.

It already returns the right map, but one file at a time. Don't write to the
map from your goroutines.

```go
ChecksumAll([]string{"app.tar", "config.yaml"}, sum)
// => map[string]string{"app.tar": "9f86d0", "config.yaml": "2c26b4"}, after 1s
```

**Think it through:**
- What can each goroutine safely write to instead of the map?
- After waiting, how do you pair each file name with its result?

### 3. Roll out in two phases

**Goal:** write `Rollout(canaries, rest, deploy)`: deploy all `canaries` at
once and wait, then deploy all `rest` hosts at once and wait.

No host in `rest` may start before every canary has finished. With 1-second
deploys, every canary finishes at 1s, every `rest` host at 2s, and `Rollout`
returns at 2s.

```go
Rollout([]string{"canary-1", "canary-2"}, []string{"web-1", "web-2", "web-3"}, deploy)
// canaries finish at 1s, the rest finish at 2s
```

**Think it through:**
- Phase 1 and phase 2 do the same job with a different list. What would a
  helper for "deploy these hosts and wait" take as parameters?
- Where must `Wait` happen so phase 2 can't start early?

## Learn more

- [sync.WaitGroup](https://pkg.go.dev/sync#WaitGroup)
- [Go spec: Defer statements](https://go.dev/ref/spec#Defer_statements)
- [Go FAQ: Why are map operations not defined to be atomic?](https://go.dev/doc/faq#atomic_maps)
