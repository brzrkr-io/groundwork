# Health checks: readiness gates

A pod shouldn't get traffic while it is still warming up, or while it is
shutting down. In this exercise you fix a readiness endpoint that lies, speed
up its checks, and add a gate for startup and shutdown.

## Learn

### Recap

As in *Health checks: liveness and readiness*, `/livez` only proves the process
answers HTTP. `/readyz` runs named dependency checks, each with its own
`context.WithTimeout`, and answers 200 when all pass or 503 when any fails. A
failing readiness probe takes the pod out of the Service's endpoints; it does
not restart it.

### The status code goes out with the first Write

An HTTP response starts with the status line and headers, then the body. The
`ResponseWriter` sends the status and headers the moment you first call
`Write`. If you haven't called `WriteHeader` by then, it sends `200 OK` for
you. A `WriteHeader` call after that is too late and is ignored (the server
logs "superfluous response.WriteHeader call").

```go
w.Header().Set("Content-Type", "application/json") // 1. headers
w.WriteHeader(http.StatusServiceUnavailable)        // 2. status
json.NewEncoder(w).Encode(report)                   // 3. body (calls Write)
```

### Running checks concurrently

Five checks with a 1-second timeout each can take 5 seconds one after the
other, longer than the probe's timeout. Run them in goroutines instead, and the
slowest check sets the total time. `wg.Go(f)` (Go 1.25+) does `wg.Add(1)`,
starts `f` in a goroutine, and calls `wg.Done()` when `f` returns:

```go
results := make([]Result, len(checks)) // one slot per check
var wg sync.WaitGroup
for i, c := range checks {
	wg.Go(func() {
		results[i] = runCheck(ctx, c, timeout)
	})
}
wg.Wait()
```

No mutex is needed. Each element of a slice is its own variable, and each
goroutine writes only to its own `results[i]`. A data race needs two goroutines
touching the *same* memory. `wg.Wait()` also makes all the writes visible to
the code that runs after it. (Appending to a shared slice from goroutines
*would* race, because `append` touches the shared length.)

### Testing with a fake clock

The tests use *testing/synctest*. Inside `synctest.Test(t, func(t *testing.T)
{...})` the clock is fake. When every goroutine in the test is waiting on a
timer, the fake clock jumps straight to the next timer. A 3-second check
finishes instantly, but `time.Since` still reports exactly `3s`. That lets the
test prove your checks ran at the same time, with no real waiting.

### Startup and draining gates

Readiness is also about the pod itself. While it starts up (loading config,
warming caches) it should answer 503. When it gets SIGTERM, it should answer
503 again (*draining*), so Kubernetes stops routing to it before it exits.
Liveness stays 200 the whole time, because nothing is broken.

The gate is read by every probe request and written by startup or signal
code, all on different goroutines. A plain `bool` would be a data race.
`sync/atomic` has types that are safe to use from many goroutines without a
mutex:

```go
var ready atomic.Bool // the zero value is false
ready.Store(true)     // set it
if ready.Load() {     // read it
	// ...
}
```

## Your task

The `payments-api` readiness endpoint always reports 200, its checks are slow,
and it takes traffic before startup finishes. Fix all three.

### 1. Fix the status code

`WriteReport` writes the JSON body before calling `WriteHeader`, so clients
always see 200. Make it send `code` (and `Content-Type: application/json`)
before the body.

```go
WriteReport(w, 503, Report{Status: "fail"})
// => 503, Content-Type: application/json, body {"status":"fail"}
```

### 2. Run the checks concurrently

Rewrite `RunChecks` so all checks run at the same time, using a
`sync.WaitGroup` (add `"sync"` to the imports). Results stay in the same order
as `checks`.

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

A check slower than the timeout still fails with `context deadline exceeded`,
exactly as before, because `runCheck` is unchanged.

If the test panics with `deadlock: main bubble goroutine has exited but
blocked goroutines remain`, `RunChecks` returned while its goroutines were
still running. Call `wg.Wait()` before `return`.

### 3. Gate readiness on startup and draining

Finish `Gate` using its two `atomic.Bool` fields:

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

## Learn more

- [pkg.go.dev: http.ResponseWriter](https://pkg.go.dev/net/http#ResponseWriter)
- [pkg.go.dev: sync/atomic](https://pkg.go.dev/sync/atomic#Bool)
- [Go blog: Testing concurrent code with testing/synctest](https://go.dev/blog/synctest)
- [Kubernetes: Pod termination](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-termination)
