# Health checks: readiness gates

A pod shouldn't get traffic while it is still warming up, or while it is
shutting down. Here you fix a readiness endpoint that lies, speed up its
checks, and add a gate for startup and shutdown.

**New here:** the order of status and body, running checks with
`sync.WaitGroup`, `atomic.Bool`.

## Learn

### Recap

As in *Health checks: liveness and readiness*, `/livez` only proves the process
answers HTTP. `/readyz` runs named dependency checks, each with its own
timeout, and answers 200 when all pass or 503 when any fails. Failing
readiness takes the pod out of traffic; it does not restart it.

### The status code goes out with the first Write

A response is: status line, headers, then body. The `ResponseWriter` sends
the status and headers the moment you first write body bytes. If you haven't
called `WriteHeader` by then, it sends `200 OK` for you. A later
`WriteHeader` is ignored (the server logs "superfluous response.WriteHeader
call").

```go
w.Header().Set("Content-Type", "text/plain") // 1. headers
w.WriteHeader(http.StatusTeapot)             // 2. status (418)
fmt.Fprint(w, "short and stout\n")           // 3. body
```

> **Why is it written like this?**
> - **Go requires** nothing here; the compiler accepts any order. The rule
>   comes from HTTP: the status line is the first thing on the wire.
> - **Convention:** headers, then `WriteHeader`, then body, every time.

### Running work concurrently

Five checks of 1 second each take 5 seconds one after another, longer than the
probe's timeout. In goroutines, the slowest one sets the total time.
`wg.Go(f)` (Go 1.25+) starts `f` in a goroutine and counts it, and `wg.Wait()`
blocks until every counted goroutine has returned:

```go
sizes := make([]int, len(files)) // one slot per file, made up front
var wg sync.WaitGroup
for i, name := range files {
	wg.Go(func() {
		sizes[i] = countBytes(name)
	})
}
wg.Wait()
```

No mutex is needed: each goroutine writes only its own `sizes[i]`, and a data
race needs two goroutines touching the *same* memory. `append` from many
goroutines *would* race, because every call changes the shared length.

> **Why is it written like this?**
> - **Go requires** the `"sync"` import to use `sync.WaitGroup`.
> - **Convention:** make the slice with its full length (`len(files)`), not
>   length 0, so every index exists before the goroutines start.
> - **Author's choice:** names like `sizes` and `wg`.

### Testing with a fake clock

The tests use *testing/synctest*. Inside `synctest.Test`, the clock is fake:
when every goroutine is waiting on a timer, it jumps straight ahead. A
3-second check finishes instantly, yet `time.Since` still says `3s`. That lets
the test prove your checks ran at the same time.

### Startup and draining gates

While starting up (loading config, warming caches) a pod should answer 503 on
readiness. After SIGTERM it should answer 503 again (*draining*), so
Kubernetes stops routing to it before it exits. Liveness stays 200 the whole
time, because nothing is broken.

Probe requests read the gate and startup or signal code writes it, on
different goroutines. A plain `bool` would race. `sync/atomic` types are safe
without a mutex:

```go
var paused atomic.Bool // the zero value is false
paused.Store(true)     // set it
if paused.Load() {     // read it
	// ...
}
```

> **Why is it written like this?**
> - **Go requires** method calls (`Load`, `Store`); you can't assign `true`
>   to an `atomic.Bool` with `=`.
> - **Convention:** use the zero value; no constructor is needed.

## Your task

The `payments-api` readiness endpoint always reports 200, its checks are slow,
and it takes traffic before startup finishes. Fix all three.

### 1. Fix the status code

**Goal:** make `WriteReport` send `code` before the JSON body.

Right now it writes the body first, so clients always see 200. The response
must have status `code`, `Content-Type: application/json`, and the report as
JSON.

```go
WriteReport(w, 503, Report{Status: "fail"})
// => 503, Content-Type: application/json, body {"status":"fail"}
```

**Think it through:**
- Which line in `WriteReport` writes the body? Which line sets the status?
- Which of them must come first?

### 2. Run the checks concurrently

**Goal:** make `RunChecks` run all checks at the same time, with results in
the same order as `checks`.

Use a `sync.WaitGroup` (add `"sync"` to the imports). `runCheck` is written
for you and doesn't change.

```go
// postgres takes 3s, redis 1s, and s3 2s before failing with "access denied":
RunChecks(ctx, checks, 5*time.Second)
// => []Result{
//        {Name: "postgres", Status: "ok"},
//        {Name: "redis", Status: "ok"},
//        {Name: "s3", Status: "fail", Error: "access denied"},
//    }
//    after 3s in total, not 1+2+3 = 6s
```

A check slower than the timeout still fails with `context deadline exceeded`.

If the test panics with `deadlock: main bubble goroutine has exited but
blocked goroutines remain`, `RunChecks` returned while its goroutines were
still running.

**Think it through:**
- The current loop uses `append`. Why is that unsafe from goroutines, and
  what does the slice need instead?
- You need the index of each check. Which loop variable gives it?
- What must happen before `return results`?

### 3. Gate readiness on startup and draining

**Goal:** finish `MarkReady`, `MarkDraining` and `State` using the gate's two
`atomic.Bool` fields, `ready` and `draining`.

- `MarkReady` stores `true` in `ready`. `MarkDraining` stores `true` in
  `draining`.
- `State` returns `"draining"` if draining was marked (even before ready),
  otherwise `"starting"` until `MarkReady` is called, otherwise `"ready"`.

`Readyz` (already written) uses `State`. Here is what the probes see:

| `State()` | `GET /readyz` | `GET /livez` |
|---|---|---|
| `"starting"` | 503 `{"status":"starting"}` | 200 `ok` |
| `"ready"` | runs the checks: 200 `{"status":"ok",...}` | 200 `ok` |
| `"draining"` | 503 `{"status":"draining"}` | 200 `ok` |

**Think it through:**
- The methods have receiver `g`. How do you reach a field of `g`?
- In `State`, which question must you ask first so draining always wins?

## Learn more

- [pkg.go.dev: http.ResponseWriter](https://pkg.go.dev/net/http#ResponseWriter)
- [pkg.go.dev: sync/atomic](https://pkg.go.dev/sync/atomic#Bool)
- [Go blog: Testing concurrent code with testing/synctest](https://go.dev/blog/synctest)
- [Kubernetes: Pod termination](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-termination)
