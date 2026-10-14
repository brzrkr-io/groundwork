# Health checks: liveness and readiness

Kubernetes asks every pod two questions, many times a minute: "are you
alive?" and "are you ready for traffic?" Answer well and deploys go smoothly.
Answer badly and Kubernetes can restart every pod you have at the worst moment.

**New here:** liveness and readiness probes. (`context.WithTimeout` is from
*DNS: lookups, timeouts and failure types*; status codes and headers from
*Networking and HTTP*.)

## Learn

### Liveness and readiness

A *probe* is an HTTP request that Kubernetes sends to your container on a
schedule. Two probes matter most:

| Probe | The question | Usual path | When it fails, Kubernetes… |
|---|---|---|---|
| *Liveness* | Is the process stuck beyond repair? | `/livez` | kills and restarts the container |
| *Readiness* | Can this pod serve requests right now? | `/readyz` | stops sending it traffic until it passes again |

Any status from 200 to 399 is a pass. Services answer `200 OK` for "yes" and
`503 Service Unavailable` for "no".

### Keep liveness dumb

Imagine `/livez` checked the database, and the database went down. Every pod
would fail liveness, and Kubernetes would restart all of them. Restarts can't
fix the database, so the pods crash in a loop.

So liveness only proves the process can answer HTTP. Dependency checks belong
in readiness: pods stop getting traffic while the database is down, and get it
back when it recovers. No restarts.

### Named checks

Readiness runs a list of *checks*, one per dependency. `Run` returns `nil`
when the dependency is healthy, and an error when it isn't:

```go
postgres := Check{Name: "postgres", Run: func(ctx context.Context) error {
	return db.PingContext(ctx) // db is a *sql.DB connection pool
}}
```

### A timeout for each check

One hanging check would make the whole probe hang. `context.WithTimeout`
gives a piece of work its own deadline:

```go
dl, cancel := context.WithTimeout(parent, 2*time.Second)
defer cancel()
resp, err := fetch(dl, url) // fetch gives up when dl expires
```

A well-behaved check watches its context. When the deadline passes, it returns
`ctx.Err()`, whose message is exactly `context deadline exceeded`.

> **Why is it written like this?**
> - **Go requires** you to use both return values: Go won't compile if a
>   variable like `cancel` is never used.
> - **Convention:** `defer cancel()` right after creating the context, so its
>   timer is always released. `go vet` warns if you forget.
> - **Author's choice:** the names `dl` and `cancel`. The new context is often
>   also called `ctx`, replacing the old one.

### Writing a response

A handler writes to `w http.ResponseWriter`. Headers and the status code go
out *before* the body, so set them first:

```go
w.Header().Set("Content-Type", "text/plain")
w.WriteHeader(http.StatusNotFound) // 404
fmt.Fprint(w, "no such page\n")    // the body comes last
```

If you never call `WriteHeader`, the first body write sends `200` for you.

> **Why is it written like this?**
> - **Go requires** nothing about the order, but HTTP does: once body bytes
>   are written, the status and headers are already sent and can't change.
> - **Convention:** use named constants like `http.StatusOK` instead of `200`.

### The readiness report

People debugging an outage want to see *which* dependency failed:

```json
{"status":"fail","checks":[{"name":"postgres","status":"ok"},{"name":"redis","status":"fail","error":"connection refused"}]}
```

The structs and their `json:"..."` tags are written for you. The tag
`omitempty` leaves `error` out when it's empty. `encoding/json` does the rest.

## Your task

The `orders-api` service needs probe endpoints before it can run on
Kubernetes. The types are written for you. Fill in the logic.

### 1. Answer the liveness probe

**Goal:** make `Livez` reply with status 200 and the body `"ok\n"`.

The body is the two letters `ok` followed by a newline. It must not run any
checks.

```text
GET /livez
=> 200, body "ok\n"
```

**Think it through:**
- Which function from *Formatting with fmt* can write text to an
  `io.Writer` like `w`? Remember to add its package to the imports.
- Do you need to call `WriteHeader` at all for a 200?

### 2. Run one check with a timeout

**Goal:** make `RunCheck` run `c.Run` with its own deadline and return a
`Result`.

`RunChecks` already calls `RunCheck` for each check, in order. In `RunCheck`:

- Make a context from `ctx` that expires after `timeout`, and pass it to `c.Run`.
- If `Run` returns `nil`: `Result{Name: c.Name, Status: "ok"}`.
- Otherwise: `Status: "fail"` and `Error` set to the error's message (a
  `string`, from `err.Error()`).

`healthy`, `refused` and `hangs` are small checks in `exercise_test.go`:

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

**Think it through:**
- Which parameter is the parent context, and which is the duration?
- `c.Run` must get the *new* context, not the original `ctx`.
- `Name` always comes from the check; only `Status` and `Error` change.

### 3. Serve the readiness report

**Goal:** in the handler returned by `Readyz`, reply 200 when all checks
passed and 503 when any failed, with a JSON content type.

The handler already runs the checks and builds `report` with `Status: "ok"`.
Before the `Encode` line:

- If any result's `Status` is `"fail"`, set `report.Status` to `"fail"` and
  reply `503`. Otherwise reply `200`.
- Set the `Content-Type` header to `application/json`.

```text
GET /readyz   (postgres ok, redis down)
=> 503
   Content-Type: application/json
   {"status":"fail","checks":[{"name":"postgres","status":"ok"},{"name":"redis","status":"fail","error":"connection refused"}]}
```

**Think it through:**
- Loop over `results` and remember the status code in a variable.
- Headers, then status code, then body. The `Encode` line is the body.

## Learn more

- [Kubernetes: Liveness, Readiness and Startup Probes](https://kubernetes.io/docs/concepts/configuration/liveness-readiness-startup-probes/)
- [Kubernetes: Configure liveness and readiness probes](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/)
- [pkg.go.dev: context.WithTimeout](https://pkg.go.dev/context#WithTimeout)
- [pkg.go.dev: http.ResponseWriter](https://pkg.go.dev/net/http#ResponseWriter)
