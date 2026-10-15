# Worker pools: fix the image puller

A node agent pre-pulls container images with a small worker pool, so it
doesn't hammer the registry. The pool has the two most common worker-pool
bugs. You find and fix them.

**New here:** nothing; practice. You fix a pool that hangs, then one that
returns results in the wrong order.

## Learn

### Recap: the shape of a pool

```text
feeder ──▶ jobs ──▶ [worker × n] ──▶ results ──▶ collector
 closes jobs                   closer goroutine
 after the last job            closes results after wg.Wait()
```

- The *feeder* closes `jobs` after the last job, so the workers' `range` loops end.
- A *closer* goroutine waits for the workers, then closes `results`, so the
  collector's `range` loop ends.
- Results arrive in **finish** order, not input order.

### Why the closer is its own goroutine

`wg.Wait()` blocks until every worker is done. But a worker is only done
after its last send on `results` is received, and the collector is the one
receiving. So waiting *before* the collector loop means everyone waits for
everyone: a deadlock. Waiting *in a separate goroutine* lets the collector
run at the same time.

```go
// A different example: close done once both downloads finish.
go func() {
	downloads.Wait()
	close(done)
}()
```

> **Why is it written like this?** `go func() { ... }()` is required syntax:
> `go` needs a function *call*, so the trailing `()` must be there. Closing a
> channel only from the side that sends (or knows sending is over) is
> convention; Go lets anyone call `close`, but closing twice panics.

### Putting results back in order

Each job carries its `index` (its position in the input). To restore order,
make the output slice full-length first, then write into a slot instead of
appending:

```go
names := make([]string, 3) // 3 empty slots: "", "", ""
names[2] = "c"             // fills slot 2, whatever order you do it in
```

> **Why is it written like this?** `make([]T, n)` with a length is required
> if you want to assign `s[i]`; assigning past the length panics. `append`
> always adds at the end, so it can't place things by index.

### Symptom → cause

| Symptom | Usual cause |
|---|---|
| hangs after the last result | nobody closes `results` |
| deadlock right away | `wg.Wait()` runs before anyone receives results |
| hangs with idle workers | nobody closes `jobs` |
| results in a random order | appended in finish order instead of placed by index |

## Your task

`PullAll(images, workers, pull)` pulls every image with `workers`
goroutines and returns one `Result` per image: the image name (`Image`) and
the error from `pull` (`Err`, `nil` on success). The tests fail if
`PullAll` hasn't returned after 1 second.

### 1. Fix the hang

**Goal:** make `PullAll` return by closing `results` once every worker has finished.

Right now every image is pulled, then `PullAll` hangs forever. Add the
close where `TODO(1)` is, without blocking the collector loop below it.

```go
// pull fails for "redis:7" with "manifest unknown"
PullAll([]string{"nginx:1.27", "redis:7"}, 2, pull)
// => two Results, one per image (task 2 fixes the order)
```

**Think it through:**
- Which variable knows when all the workers are done?
- If you wait for it right here, who is left to receive from `results`?

### 2. Keep the input order

**Goal:** `PullAll` returns the results in the same order as `images`.

Results come back in the order the pulls finish, so a slow first image
ends up last. Each `pulled` value `p` has `p.index` (the image's position in
`images`) and `p.result`. Store each result at its index.

```go
PullAll([]string{"postgres:16", "nginx:1.27", "redis:7"}, 2, pull)
// => []Result{
//        {Image: "postgres:16", Err: nil},
//        {Image: "nginx:1.27", Err: nil},
//        {Image: "redis:7", Err: errors.New("manifest unknown")},
//    }
// even when postgres:16 finishes last
```

**Think it through:**
- How many results will there be? Use that to size `out` before the loop.
- Which field of `p` says where its result belongs?

## Learn more

- [Go by Example: Worker pools](https://gobyexample.com/worker-pools)
- [Go blog: Go Concurrency Patterns: Pipelines and cancellation](https://go.dev/blog/pipelines)
