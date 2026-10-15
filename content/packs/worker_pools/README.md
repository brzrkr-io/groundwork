# Worker pools

Starting one goroutine per host is fine for 10 hosts. For 10,000 it can run
out of file descriptors, or get you rate-limited. A *worker pool* runs a
fixed number of goroutines, the *workers*, that take jobs from a shared
channel. You decide how much work happens at once.

**New here:** worker pools, closing a channel after a `WaitGroup`.

## Learn

### The shape

```text
                 ┌──────────┐
            ┌───▶│ worker 1 │───┐
┌───────┐   │    └──────────┘   │    ┌─────────┐   ┌─────────┐
│ jobs  │───┼───▶│ worker 2 │───┼───▶│ results │──▶│ collect │
└───────┘   │    └──────────┘   │    └─────────┘   └─────────┘
            └───▶│ worker 3 │───┘
                 └──────────┘
```

1. One goroutine sends every job on `jobs`, then **closes** it.
2. `n` workers each loop over `jobs`. Whichever is free takes the next job.
   When `jobs` is closed and empty, the loops end.
3. Each worker sends one result per job on `results`.
4. When **all** workers have returned, someone closes `results`.
5. The collector ranges over `results` until it is closed.

### Closing after the last worker

Step 4 needs a `sync.WaitGroup` plus one extra goroutine. Here a pool of
downloaders shares a `files` channel:

```go
var wg sync.WaitGroup
for range 4 {
	wg.Go(func() {
		download(urls, files)
	})
}
go func() {
	wg.Wait()    // every downloader has returned...
	close(files) // ...so nobody will send again
}()
```

> **Why is it written like this?** The extra `go func` is required, not
> style. Calling `wg.Wait()` directly deadlocks: nobody is receiving yet, so
> the workers block on their sends and `Wait` never returns. `for range 4`
> (no loop variable) is fine because every worker is the same.

### Who closes what

| Channel | Closed by | When |
|---|---|---|
| `jobs` | the goroutine that sends the jobs | after the last job |
| `results` | the goroutine waiting on the WaitGroup | after the last worker returns |

Close a channel exactly once, from the side that knows no more sends are
coming. A worker must never close `results`: other workers still send on it.

### Keeping input order

Workers finish in any order. To restore input order, each job carries its
input position, the worker copies it into the result, and the collector
writes each result into that slot of a slice made with full length:

```go
lines := make([]string, total) // length total, not just capacity
lines[3] = "fourth line"       // write straight into slot 3
```

`append` would add to the end instead, in arrival order.

## Your task

Your fleet tool checks hosts with a pool of workers. `CheckAll` and
`sendJobs` are written for you. `CheckAll` wires the pieces together:

```go
jobs := sendJobs(hosts)                       // provided
results := startWorkers(workers, jobs, check) // task 2
return collect(results, len(hosts))           // task 3
```

### 1. Write the worker

**Goal:** make `worker(jobs, results, check)` send one `Result` per job until
`jobs` is closed, then return.

Each `Result` has the job's `Index` and `Host`, and `Up` set to what
`check(job.Host)` returned. Don't close `results`.

```go
// jobs carries Job{0, "web-1"}, Job{1, "web-2"}, then is closed
worker(jobs, results, check)
// results gets Result{0, "web-1", true}, Result{1, "web-2", false}
```

**Think it through:**
- Which loop over a channel stops by itself when the channel is closed?
- Which `Result` fields come from the job, and which from `check`?

### 2. Close results when the workers are done

**Goal:** make `startWorkers(n, jobs, check)` close `results` once all `n`
workers have returned, while still returning `results` straight away.

It already makes an unbuffered `results` channel and starts `n` workers, but
nothing closes `results`, so `collect` waits forever. Track the workers with
a `sync.WaitGroup` (start them with `wg.Go`), and add one more goroutine that
closes `results`. Import `"sync"`.

With 3 workers, at most 3 checks run at once. With 3 workers and 3 hosts,
all 3 checks run at the same time.

**Think it through:**
- What replaces the plain `go` that starts each worker?
- Who waits on the WaitGroup, and why can't it be `startWorkers` itself?

### 3. Collect in order

**Goal:** make `collect(results, n)` return a slice of length `n` with each
result at index `r.Index`.

It receives every result until `results` is closed, but currently keeps
arrival order.

```go
CheckAll([]string{"web-1", "web-2", "db-1"}, 2, check)
// => []Result{{0, "web-1", true}, {1, "web-2", false}, {2, "db-1", true}}
```

**Think it through:**
- `out` is made with length 0. What length does it need?
- Which line adds to the end instead of to a chosen slot?

## Learn more

- [Go by Example: Worker pools](https://gobyexample.com/worker-pools)
- [Go blog: Go Concurrency Patterns: Pipelines and cancellation](https://go.dev/blog/pipelines)
- [Effective Go: Channels](https://go.dev/doc/effective_go#channels)
