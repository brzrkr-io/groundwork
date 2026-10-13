# Health checks: liveness and readiness

Kubernetes asks every pod two questions, many times a minute: "are you
alive?" and "are you ready for traffic?" If you answer them well, deploys and
outages go smoothly. If you answer them badly, Kubernetes can restart every pod
you have at the worst possible moment.

## Learn

### Liveness and readiness

A *probe* is an HTTP request that Kubernetes (the kubelet on each node) sends
to your container on a schedule. Two probes matter most:

| Probe | The question | Usual path | When it fails, Kubernetes… |
|---|---|---|---|
| *Liveness* | Is the process stuck beyond repair? | `/livez` | kills and restarts the container |
| *Readiness* | Can this pod serve requests right now? | `/readyz` | stops sending it traffic (removes it from the Service's endpoints) until it passes again |

Any status from 200 to 399 counts as a pass. Services answer `200 OK` for
"yes" and `503 Service Unavailable` for "no".

### Keep liveness dumb

Imagine `/livez` checked the database and the database went down. Every pod
would fail liveness and Kubernetes would restart all of them. The restarts
can't fix the database, so the pods keep crashing in a loop. A database outage
has become a full outage.

So liveness only proves that the process can still answer HTTP. Dependency
checks belong in readiness. When the database is down, pods stop getting
traffic. When it recovers, they pass readiness again and traffic comes back,
with no restarts.

### Named checks

A readiness endpoint runs a list of *checks*, one per dependency:

```go
type Check struct {
	Name string
	Run  func(ctx context.Context) error
}

postgres := Check{Name: "postgres", Run: func(ctx context.Context) error {
	return db.PingContext(ctx) // db is a *sql.DB connection pool
}}
```

`Run` returns `nil` when the dependency is healthy, and an error when it isn't.

### A timeout for each check

Probes have their own timeout (`timeoutSeconds`, 1 second by default). One
check that hangs would make the whole probe hang, so each check gets its own
deadline from `context.WithTimeout`:

```go
ctx, cancel := context.WithTimeout(parent, 500*time.Millisecond)
defer cancel() // always release the context's timer
err := check.Run(ctx)
```

A well-behaved check watches `ctx.Done()`. When the deadline passes, it gives
up and returns `ctx.Err()`, which is `context.DeadlineExceeded`. That error's
message is:

```text
context deadline exceeded
```

### The readiness report

People debugging an outage want to see *which* dependency failed, so the
readiness endpoint sends a JSON report:

```json
{"status":"fail","checks":[{"name":"postgres","status":"ok"},{"name":"redis","status":"fail","error":"connection refused"}]}
```

The `Result` struct uses `json:"error,omitempty"`, so the `error` field is left
out when it's empty. You only have to fill in the structs. `encoding/json`
does the rest.

The status code and headers go out before the body, so set them first:

```go
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusServiceUnavailable)
json.NewEncoder(w).Encode(report) // the body comes last
```

## Your task

The `orders-api` service needs probe endpoints before it can run on
Kubernetes. The types are written for you. Fill in the logic.

### 1. Answer the liveness probe

Make `Livez` reply with status 200 and exactly this body: `ok` followed by a
newline (`"ok\n"`). It must not run any checks.

`fmt.Fprint(w, "ok\n")` writes text to the response. Add `"fmt"` to the
imports. A handler that never calls `WriteHeader` sends 200 automatically.

```text
GET /livez
=> 200, body "ok\n"
```

### 2. Run one check with a timeout

`RunChecks` already loops over the checks in order and calls `RunCheck` for
each one. Write `RunCheck`:

- Create a context from `ctx` that expires after `timeout`, and pass it to
  `c.Run`. Remember `defer cancel()`.
- If `Run` returns `nil`, return `Result{Name: c.Name, Status: "ok"}`.
- Otherwise return `Status: "fail"` and `Error: err.Error()`.

`healthy`, `refused` and `hangs` are small checks defined in
`exercise_test.go`:

```go
RunChecks(ctx, []Check{{"postgres", healthy}, {"redis", refused}}, time.Second)
// => []Result{
//        {Name: "postgres", Status: "ok"},
//        {Name: "redis", Status: "fail", Error: "connection refused"},
//    }

// hangs blocks until its context expires:
RunChecks(ctx, []Check{{"vault", hangs}}, 5*time.Millisecond)
// => []Result{{Name: "vault", Status: "fail", Error: "context deadline exceeded"}}
```

### 3. Serve the readiness report

Finish the handler returned by `Readyz`. It already runs the checks and builds
a `Report` with `Status: "ok"`. Before the `Encode` line:

- If any result's `Status` is `"fail"`, set `report.Status` to `"fail"` and
  reply `503`. Otherwise reply `200`.
- Set the `Content-Type` header to `application/json`.

```text
GET /readyz   (postgres ok, redis down)
=> 503
   Content-Type: application/json
   {"status":"fail","checks":[{"name":"postgres","status":"ok"},{"name":"redis","status":"fail","error":"connection refused"}]}
```

## Learn more

- [Kubernetes: Liveness, Readiness and Startup Probes](https://kubernetes.io/docs/concepts/configuration/liveness-readiness-startup-probes/)
- [Kubernetes: Configure liveness and readiness probes](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/)
- [pkg.go.dev: context.WithTimeout](https://pkg.go.dev/context#WithTimeout)
- [pkg.go.dev: http.ResponseWriter](https://pkg.go.dev/net/http#ResponseWriter)
