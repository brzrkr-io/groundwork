# Worker pools

Starting one goroutine per host is fine for 10 hosts. For 10,000 it can
exhaust file descriptors, or get you rate-limited. A *worker pool* runs a
fixed number of goroutines, the *workers*, that take jobs from a shared
channel. You decide how much work happens at once.

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

1. A goroutine sends every job on the `jobs` channel, then **closes** it.
2. `n` workers each loop `for job := range jobs`. Whichever worker is free
   takes the next job. When `jobs` is closed and empty, the loops end.
3. Each worker sends a result on the `results` channel.
4. When **all** workers have finished, someone closes `results`. That's a
   job for a WaitGroup plus one extra goroutine:

```go
var wg sync.WaitGroup
for range n {
	wg.Go(func() {
		worker(jobs, results)
	})
}
go func() {
	wg.Wait()      // every worker has returned...
	close(results) // ...so nobody will send again
}()
```

5. The collector ranges over `results` until it is closed.

Why the extra goroutine? If you called `wg.Wait()` straight away, nobody
would be receiving from `results`. The workers would block on their sends,
and `Wait` would never return: a deadlock.

### Who closes what

| Channel | Closed by | When |
|---|---|---|
| `jobs` | the goroutine that sends the jobs | after the last job |
| `results` | the goroutine waiting on the WaitGroup | after the last worker returns |

Each channel is closed by whoever knows there will be no more sends, and
closed exactly once.

### Keeping input order

Workers finish in any order. To return results in input order, put the
input's index in the job, copy it into the result, and let the collector
place each result at `out[r.Index]`:

```go
out := make([]Result, n)
for r := range results {
	out[r.Index] = r
}
```

## Your task

Your fleet tool checks hosts with a pool of workers. `CheckAll` and
`sendJobs` are written for you. `CheckAll` wires the three pieces together:

```go
jobs := sendJobs(hosts)                       // provided
results := startWorkers(workers, jobs, check) // task 2
return collect(results, len(hosts))           // task 3
```

### 1. Write the worker

`worker(jobs, results, check)` should receive jobs until `jobs` is closed.
For each job, call `check(job.Host)` and send a `Result` with the job's
`Index` and `Host`, and `Up` set to what `check` returned. Then return.
Don't close `results`, because other workers are still using it.

```go
// jobs carries Job{0, "web-1"}, Job{1, "web-2"}, then is closed
worker(jobs, results, check)
// results gets Result{0, "web-1", true}, Result{1, "web-2", false}
```

### 2. Close results when the workers are done

`startWorkers(n, jobs, check)` makes an unbuffered `results` channel and
starts `n` goroutines that each run `worker`. It returns `results` straight
away, but nothing ever closes it, so `collect` waits forever. Track the
workers with a `sync.WaitGroup` (start them with `wg.Go`), and start one more
goroutine that closes `results` once every worker has returned.

With 3 workers, at most 3 checks run at once, and with 3 hosts all 3 run at
the same time.

### 3. Collect in order

`collect(results, n)` receives every result until `results` is closed, but it
returns them in the order they arrive. Make it return a slice of length `n`
with each result at index `r.Index`.

```go
CheckAll([]string{"web-1", "web-2", "db-1"}, 2, check)
// => []Result{{0, "web-1", true}, {1, "web-2", false}, {2, "db-1", true}}
```

## Learn more

- [Go by Example: Worker pools](https://gobyexample.com/worker-pools)
- [Go blog: Go Concurrency Patterns: Pipelines and cancellation](https://go.dev/blog/pipelines)
- [Effective Go: Channels](https://go.dev/doc/effective_go#channels)
