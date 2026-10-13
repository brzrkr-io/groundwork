# Worker pools: fix the image puller

A node agent pre-pulls container images with a small worker pool, so it
doesn't hammer the registry. The pool has the two most common worker-pool
bugs. Find and fix them.

## Learn

### Recap

A worker pool is a `jobs` channel, `n` worker goroutines that `range` over
it, and a `results` channel:

```text
feeder ──▶ jobs ──▶ [worker × n] ──▶ results ──▶ collector
 closes jobs                   closer goroutine
 after the last job            closes results after wg.Wait()
```

- The feeder closes `jobs` when every job is sent, so the workers' loops end.
- A separate goroutine waits for the workers' WaitGroup, then closes
  `results`, so the collector's loop ends.
- Results arrive in **finish** order. Carry an index with each job and
  place results by index to get input order back.

### Symptom → cause

| Symptom | Usual cause |
|---|---|
| hangs after the last result | nobody closes `results`, so the collector's `range` waits forever |
| deadlock right away | `wg.Wait()` is called before anyone receives results |
| hangs with idle workers | nobody closes `jobs` |
| results in a random order | appended in finish order instead of placed by index |

## Your task

`PullAll(images, workers, pull)` pulls every image with `workers`
goroutines and returns one `Result` per image: the image name and the error
from `pull` (`nil` on success).

### 1. Fix the hang

`PullAll` pulls every image, then hangs forever. Close `results` once every
worker has finished, without blocking the collector. Use one extra goroutine,
as in the recap.

```go
// pull fails for "redis:7" with "manifest unknown"
PullAll([]string{"nginx:1.27", "redis:7"}, 2, pull)
// => two Results (task 2 fixes the order)
```

### 2. Keep the input order

Results come back in the order the pulls finish, so a slow first image ends
up last. Store each result at its index, so that `PullAll` returns results in
the same order as `images`.

```go
PullAll([]string{"postgres:16", "nginx:1.27", "redis:7"}, 2, pull)
// => []Result{
//        {Image: "postgres:16", Err: nil},
//        {Image: "nginx:1.27", Err: nil},
//        {Image: "redis:7", Err: errors.New("manifest unknown")},
//    }
// even when postgres:16 finishes last
```

## Learn more

- [Go by Example: Worker pools](https://gobyexample.com/worker-pools)
- [Go blog: Go Concurrency Patterns: Pipelines and cancellation](https://go.dev/blog/pipelines)
