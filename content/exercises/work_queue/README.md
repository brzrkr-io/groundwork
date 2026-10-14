# Work queues: bounded queues and backpressure

A *work queue* sits between the part of a system that accepts work and the
part that does it. You'll build a bounded one that says "not now" when it's
full, instead of quietly growing until the server runs out of memory.

**New here:** queues and backpressure, a buffered channel as a queue,
non-blocking send with `select` and `default`, `len` of a channel, `202
Accepted`, `503 Service Unavailable` with `Retry-After`.

## Learn

### Why put a queue in the middle?

A build takes two minutes, but a CI job that asks for one wants an answer in
milliseconds. So the API only *accepts* the request and puts it in a queue;
a fixed pool of *workers* takes jobs out and does the slow part.

```text
POST /builds ──> API ──> [ job | job | job ] ──> worker 1
                          the queue           ──> worker 2
```

That buys you two things:

- **Bursts are absorbed.** 50 requests in one second land in the queue, and
  the workers catch up over the next minute.
- **The API is decoupled from the slow work.** It stays fast even while every
  worker is busy.

### Unbounded queues hide overload

If work arrives faster than workers finish it, the queue grows forever:

| Arrive | Finish | After 1 hour |
|---|---|---|
| 40/s | 40/s | queue stays small |
| 50/s | 40/s | 36,000 jobs waiting |

Nothing looks wrong at first: every request still gets a quick "accepted".
But each new job waits longer and longer (36,000 jobs at 40/s is 15 minutes),
memory grows, and eventually the process is OOM-killed and loses everything
in the queue.

### Bounded queues and backpressure

A *bounded* queue has a fixed capacity. When it's full, it pushes back on
whoever is adding work. That push is called *backpressure*. There are two
common ways to do it:

| Choice | What the caller sees | Good for |
|---|---|---|
| Reject fast | An error at once (over HTTP: `503` with `Retry-After`) | Clients that can retry later |
| Block with a deadline | Waits for room, gives up when its context is done | Producers inside your own program |

Either way, the overload is visible right away, and memory stays bounded.

### A buffered channel is a queue

As in *Channels*, `make(chan T, n)` makes a channel that holds up to `n`
values. Sends go in at the back, receives come out at the front. Two built-in
functions describe it:

```go
alerts := make(chan string, 3)
alerts <- "disk 91%"
len(alerts) // => 1 (values waiting)
cap(alerts) // => 3 (room in total)
```

A send on a full channel blocks until someone receives.

### Sending without waiting

A `select` with a `default` case never blocks: if no other case is ready
right now, `default` runs instead.

```go
select {
case alerts <- msg:
	// there was room
default:
	// full: msg was not sent
}
```

### Waiting, but not forever

Without `default`, `select` waits for whichever case is ready first. Add a
case for `ctx.Done()`, as in *Context and cancellation*, and the
wait ends when the context is cancelled or its deadline passes:

```go
select {
case results <- r:
	return nil
case <-ctx.Done():
	return ctx.Err() // context.Canceled or context.DeadlineExceeded
}
```

If both cases are ready at once, Go picks one at random. That's fine here.

> **Why is it written like this?** `select`, `case` and `default` are Go
> syntax: the compiler requires each `case` to be a send or a receive. Using
> `default` to make a send non-blocking is the standard Go idiom.

### Workers, Close and draining

A worker ranges over the channel. The loop ends only when the channel is
**closed and empty**, so closing it means "no more jobs, finish what's
there":

```go
for e := range emails {
	send(e)
}
```

`Close` closes the channel, then waits on a `sync.WaitGroup` until every
worker's loop has ended. `wg.Go(f)` runs `f` in a new goroutine and tracks
it, as in *WaitGroups*.

Only the sending side closes a channel, and sending on a closed channel
panics. So in a graceful shutdown, stop the HTTP server first, then call
`Close`, as in *Graceful shutdown: HTTP server and worker*.

> **Why is it written like this?** "Only the sender closes" is a Go rule of
> thumb that keeps you away from that panic. Ranging over a channel is Go
> syntax for "receive until closed".

### Telling HTTP clients what happened

| Status | Meaning here |
|---|---|
| `202 Accepted` | Queued; the work happens later |
| `503 Service Unavailable` | Full; `Retry-After: 1` says "try again in 1 second" |
| `400 Bad Request` | The body isn't a valid job |

`json.NewDecoder(r.Body).Decode(&v)` reads a JSON body into `v` and returns
an error when the JSON is broken. With Go's router you register the handler
as `mux.Handle("POST /builds", q.Handler())`, so the router already rejects
other methods.

## Your task

Your deploy API accepts build requests faster than the build workers can run
them during a release rush. Put a bounded queue in front of the workers.

### 1. Reject jobs when the queue is full

**Goal:** make `TryEnqueue` add a job without waiting, and `Len` count the
waiting jobs.

- `TryEnqueue(j)` sends `j` on `q.jobs` and returns `nil` if there is room.
  If the queue is full, it returns `ErrFull` straight away and `j` is not
  queued.
- `Len()` returns how many jobs are waiting in `q.jobs`.

```go
q := New(2)
q.TryEnqueue(Job{ID: "b1", Service: "api"}) // => nil
q.TryEnqueue(Job{ID: "b2", Service: "api"}) // => nil
q.TryEnqueue(Job{ID: "b3", Service: "api"}) // => ErrFull
q.Len()                                     // => 2
```

**Think it through:**
- Which `select` case runs when the send can't happen right now?
- Which built-in tells you how many values a channel holds?

### 2. Wait for room until the deadline

**Goal:** make `Enqueue(ctx, j)` wait for room, but give up when `ctx` is
done.

- Room now, or room appears later: send `j` and return `nil`.
- `ctx` is done first: return `ctx.Err()`, and `j` is not queued.

```go
q := New(1)
q.TryEnqueue(Job{ID: "b1", Service: "api"})
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
defer cancel()
q.Enqueue(ctx, Job{ID: "b2", Service: "api"})
// waits 2s, then => context.DeadlineExceeded
// (if a worker takes b1 within the 2s, it returns nil instead)
```

**Think it through:**
- This is task 1 without `default`, plus one more case. Which channel closes
  when a context is done?

### 3. Run the workers and drain on Close

**Goal:** make `Start(workers, handle)` start the workers, and `Close` wait
until every queued job has been handled.

- `Start` starts exactly `workers` goroutines, tracked by `q.wg`. Each takes
  jobs from `q.jobs` and calls `handle(job)` for each, one at a time.
- `Close` closes `q.jobs`, then waits for the workers (the `q.wg.Wait()` is
  already there). When `Close` returns, every job that was queued has been
  handled.

```go
// 6 jobs queued, each takes 1s to handle
q.Start(2, build)
q.Close() // returns after 3s: 2 workers x 3 rounds, all 6 jobs built
```

**Think it through:**
- `Start` doesn't handle jobs itself: it only starts goroutines and returns.
- Which loop over a channel stops by itself once the channel is closed and
  empty?
- `Close` must close before it waits. What would the workers do otherwise?

### 4. Answer HTTP with 202 or 503

**Goal:** make `Handler()` queue the JSON job in each request body and
answer with the right status.

In this order:

1. Decode the body into a `Job`. If that fails, answer `400`.
2. If `ID` or `Service` is empty, answer `400`.
3. Call `q.TryEnqueue`. If it fails, set the header `Retry-After` to `1` and
   answer `503`.
4. Otherwise answer `202`.

Only the `503` response has a `Retry-After` header. The tests check status
codes and headers, not the body text, so `http.Error` with any message is
fine for the errors. Add `"encoding/json"` to the imports.

```text
POST {"id":"b7","service":"billing"}   => 202
POST {"id":"b8","service":"api"}       => 503, Retry-After: 1   (queue full)
POST {"id":"b9",                       => 400                   (broken JSON)
POST {"service":"api"}                 => 400                   (no id)
```

**Think it through:**
- A bad job is a `400` even when the queue is full, because you check it
  before you queue it.
- Headers must be set before the status is written. Which comes first?
- Which `http.Status...` constants hold 202, 400 and 503?

## Learn more

- [Go spec: Select statements](https://go.dev/ref/spec#Select_statements)
- [Go blog: Pipelines and cancellation](https://go.dev/blog/pipelines)
- [Amazon Builders' Library: Avoiding insurmountable queue backlogs](https://aws.amazon.com/builders-library/avoiding-insurmountable-queue-backlogs/)
- [MDN: 503 Service Unavailable](https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Status/503)
