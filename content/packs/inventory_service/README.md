# Project: inventory service

You'll finish a small production-style HTTP service: a JSON API that tracks
your fleet's hosts. It has what every real service needs: validation,
request IDs, access logs, Prometheus metrics, health probes, idempotent
creates and graceful shutdown. You practised each piece in this module; here
you wire them into one program you can run with `go run` and `curl`.

**New here:** `slices.SortedFunc` with `cmp.Or`. Everything else is practice, including `r.PathValue` (*Networking and HTTP*), `context.WithValue` (*Observability: trace IDs through logs and calls*), `logger.Log` (*Structured logging: access logs*) and `atomic.Bool` (*Health checks: readiness gates*).

## Learn

### What the service does

| Request | Answer |
|---|---|
| `GET /hosts` | 200, a JSON array of every host |
| `POST /hosts` | 201, the new host (supports `Idempotency-Key`) |
| `GET /hosts/{id}` | 200 the host, or 404 |
| `PUT /hosts/{id}` | 200 the updated host, 400 or 404 |
| `DELETE /hosts/{id}` | 204, or 404 |
| `GET /healthz` | liveness: always 200 `{"status":"ok"}` |
| `GET /readyz` | readiness: 200 `{"status":"ready"}` while serving, else 503 `{"status":"not ready"}` |
| `GET /metrics` | Prometheus text format |

A host in JSON:

```json
{"id":"h1","name":"web-1","env":"prod","region":"eu-west-1","port":443}
```

Errors are always JSON with one field: `{"error":"host not found"}`.

### How a request flows

```text
client → requestID → observe → ServeMux → handler → hosts.Store
         (X-Request-ID) (log + count)      (POST /hosts also goes
                                            through idempotent)
```

`internal/hosts` knows nothing about HTTP: it stores hosts and enforces the
rules. `internal/server` turns HTTP into store calls, and store errors into
status codes. `cmd/inventoryd` reads flags, handles signals and starts it all.

> **Why is it written like this?** Go requires that `internal/` packages are
> only imported from inside this module. Splitting `cmd/` and `internal/` is
> community convention. Which package owns what is the author's choice.

### File map

| File | What's in it | Tasks |
|---|---|---|
| `go.mod` | the module `groundwork.local/inventory_service` | |
| `cmd/inventoryd/main.go` | flags, JSON logger, signal handling, calls `Run` | read only |
| `internal/hosts/hosts.go` | `Host`, `Validate`, the in-memory `Store` | 1, 2 |
| `internal/hosts/hosts_test.go` | `TestValidate`, `TestStore` | |
| `internal/server/server.go` | the `Server` type, `New`, routes in `Handler` | read only |
| `internal/server/handlers.go` | JSON helpers and the `/hosts` handlers | 3 |
| `internal/server/middleware.go` | request IDs and access logging | 4, 5 |
| `internal/server/metrics.go` | the request counter and `GET /metrics` | 5 |
| `internal/idempotency/idempotency.go` | the key `Store`, `Capture` and `Send` | read only |
| `internal/server/idempotency.go` | the `idempotent` middleware | 6 |
| `internal/server/health.go` | `/healthz` and `/readyz` | read only |
| `internal/server/run.go` | `Run`: serve and shut down gracefully | 7 |
| `internal/server/server_test.go` | `TestHostAPI` … `TestRun` | |

Do the tasks in order: later tests send real requests through the whole
stack and rely on earlier tasks.

### New pieces

- **`r.PathValue("id")`** reads the `{id}` part of a route like
  `"GET /hosts/{id}"`. For `/hosts/h3` it returns `"h3"`.
- **`context.WithValue(parent, key, value)`** returns a child context
  carrying one value. Middleware hands it on with `r.WithContext(ctx)`, and
  `RequestID(ctx)` reads it back. The key is a private type,
  `requestIDKey{}`, so no other package can clash with it.
- **`logger.Log(ctx, level, msg, args...)`** is `logger.Info` with the level
  as a parameter: `slog.LevelInfo` or `slog.LevelError`, chosen at run time.
- **`atomic.Bool`** (from `sync/atomic`) is a bool many goroutines can use
  safely: `flag.Store(true)`, `flag.Load()`.
- **`slices.SortedFunc(maps.Keys(m), cmp)`** gives a map's keys as a sorted
  slice (map order is random). `cmp.Or(a, b)` returns its first non-zero
  argument, so it sorts by one field, then by another on a tie:

```go
people := slices.SortedFunc(maps.Keys(ages), func(a, b person) int {
	return cmp.Or(cmp.Compare(a.Last, b.Last), cmp.Compare(a.First, b.First))
})
```

> **Why is it written like this?** The compare function must return an `int`
> (negative, zero, positive): Go's `slices` package requires that signature.
> `cmp.Compare` returns exactly that. Using `cmp.Or` to chain fields is
> convention.

### Reading the tests

The tests mostly use `httptest.NewRequest` and `httptest.NewRecorder`,
through a helper called `do`. `wg.Go(f)` starts `f` in a goroutine tracked by
the WaitGroup. `TestRun` starts a real server on `127.0.0.1:0` (any free
port) and talks to it over TCP.

## Your task

Your platform team keeps its host list in a spreadsheet. You're replacing it
with a service the deploy tooling can call. The skeleton is wired up; fill in
the seven missing pieces.

### 1. Validate hosts

**Goal:** make `Validate` in `internal/hosts/hosts.go` return an error for the first broken rule, or `nil`.

Check these rules **in this order**:

| Rule | Message |
|---|---|
| `Name` is not empty | `name is required` |
| `Env` is `"dev"`, `"staging"` or `"prod"` | `env must be dev, staging or prod` |
| `Port` is between 1 and 65535 (inclusive) | `port must be between 1 and 65535` |

Wrap `ErrInvalid` (with `%w`) so callers can use `errors.Is`; the full
message is `invalid host: ` followed by the rule's message. `Region` is optional.

```go
Validate(hosts.Host{Env: "prod", Port: 443})
// => error "invalid host: name is required", errors.Is(err, ErrInvalid) == true
Validate(hosts.Host{Name: "web-1", Env: "qa", Port: 0})
// => error "invalid host: env must be dev, staging or prod"
```

**Think it through:**
- What does `ErrInvalid`'s own text say? That decides what fixed text you add.
- One `if` per rule, top to bottom, each returning straight away.

### 2. Finish the store

**Goal:** write `Get`, `Update` and `Delete` on `hosts.Store`.

`Create`, `List` and `Len` are done: use them as the pattern. Every method
locks `s.mu` first and defers the unlock.

- `Get(id)` returns the host, or `ErrNotFound`.
- `Update(id, h)` returns `ErrNotFound` if the ID is unknown, then the
  `Validate` error if `h` is invalid. Otherwise it stores `h` **with its ID
  set to `id`** (whatever `h.ID` said) and returns it.
- `Delete(id)` removes the host, or returns `ErrNotFound`.

```go
s.Update("h1", hosts.Host{ID: "ignored", Name: "web-1", Env: "prod", Port: 8443})
// => Host{ID: "h1", Name: "web-1", Env: "prod", Port: 8443}, nil
s.Delete("h99")
// => ErrNotFound
```

**Think it through:**
- In `Update`, the "unknown ID" check comes before validation.
- Which value goes into `h.ID` before you store it: the parameter or the field?

### 3. Finish the handlers

**Goal:** write `handleCreate`, `handleUpdate` and `handleDelete` in `internal/server/handlers.go`.

`handleList` and `handleGet` show the pattern. Use `writeJSON`, `writeError`
and `writeStoreError` (which maps `ErrNotFound` to 404 and `ErrInvalid` to
400, with `err.Error()` as the message). Decode bodies with a
`json.NewDecoder(r.Body)` and its `Decode` method, passing a pointer.

`handleCreate` (`POST /hosts`):

| Situation | Status | Body |
|---|---|---|
| body is not valid JSON | 400 | `{"error":"invalid JSON body"}` |
| host fails validation | 400 | `{"error":"invalid host: name is required"}` (the `Validate` message) |
| created | 201 | the stored host, plus header `Location: /hosts/h1` (its real ID) |

`handleUpdate` (`PUT /hosts/{id}`):

| Situation | Status | Body |
|---|---|---|
| body is not valid JSON | 400 | `{"error":"invalid JSON body"}` |
| unknown ID | 404 | `{"error":"host not found"}` |
| host fails validation | 400 | the `Validate` message |
| updated | 200 | the stored host |

`handleDelete` (`DELETE /hosts/{id}`): 204 with no body, or 404
`{"error":"host not found"}`. For 204, call `w.WriteHeader` and write nothing.

```console
$ curl -i -X POST localhost:8080/hosts -d '{"name":"web-1","env":"prod","port":443}'
HTTP/1.1 201 Created
Content-Type: application/json
Location: /hosts/h1

{"id":"h1","name":"web-1","env":"prod","region":"","port":443}
```

**Think it through:**
- Set the `Location` header *before* `writeJSON`: headers can't change after the status is sent.
- `Location` is fixed text `/hosts/` plus the stored host's ID.

### 4. Add request IDs and access logs

**Goal:** write the `requestID` and `observe` middleware in `internal/server/middleware.go`.

**`requestID`**: use the request's `X-Request-ID` header if it has one,
otherwise call `s.newRequestID()` (it returns `"req-1"`, `"req-2"`, …). Set
the same ID as the `X-Request-ID` **response** header. Store it in a child of
`r.Context()` under the key `requestIDKey{}`, and call `next` with the
request carrying that context.

**`observe`**: wrap `w` in a `statusRecorder` whose `status` starts at
`http.StatusOK` (handlers that never call `WriteHeader` send 200). Serve the
request through it, then write **one** log line with `s.logger`:

| Field | Value |
|---|---|
| message | `"request"` |
| level | `ERROR` if the status is 500 or more, otherwise `INFO` |
| `request_id` | `RequestID(r.Context())` |
| `method` | `r.Method` |
| `path` | `r.URL.Path` |
| `status` | the recorded status code (a number) |

```json
{"time":"…","level":"INFO","msg":"request","request_id":"req-1","method":"GET","path":"/hosts","status":200}
```

Add `"log/slog"` to the imports in `middleware.go`.

**Think it through:**
- `requestID` touches two different headers: the request's (read) and the response's (write).
- In `observe`, pass the recorder to `next`, not `w`, or you can't see the status.
- Pick the level into a variable first, then make one `Log` call.

### 5. Count requests for /metrics

**Goal:** make `Metrics.Inc` count a request, make `Metrics.Write` print the counts sorted, and call `Inc` from `observe`.

In `internal/server/metrics.go`:

- `Inc(method, code)` locks `m.mu` and adds one to the count stored under
  `requestKey{Method: method, Code: code}` in `m.requests`.
- `Write(w)` already prints the HELP and TYPE lines. After them, print one
  line per key, **sorted by method and then by code** (use the
  `slices.SortedFunc` and `cmp.Or` pattern from *New pieces*), in exactly
  this format. Hold the lock while you read the map. Add `"cmp"`, `"maps"`
  and `"slices"` to the imports in `metrics.go`.

```text
# HELP inventory_http_requests_total HTTP requests handled, by method and status code.
# TYPE inventory_http_requests_total counter
inventory_http_requests_total{method="GET",code="200"} 2
inventory_http_requests_total{method="GET",code="404"} 1
inventory_http_requests_total{method="PATCH",code="405"} 1
```

With nothing counted, print only the two `#` lines.

Then, in `observe`, call `s.metrics.Inc` with the method and recorded status
after the request has been served. (A scrape of `/metrics` doesn't count
itself, because counting happens after its handler finishes.)

**Think it through:**
- The quotes around `GET` and `200` are part of the output, so inside a Go
  string they must be escaped as `\"`.
- Which `requestKey` fields do you compare, and in which order?

### 6. Make POST /hosts idempotent

**Goal:** write `idempotent` in `internal/server/idempotency.go` so a retried `POST /hosts` with the same key never creates a second host.

Use the `s.idem` store (an `*idempotency.Store`) and the
`internal/idempotency` package. Add the import
`"groundwork.local/inventory_service/internal/idempotency"`.

1. No `Idempotency-Key` header: call `next` normally and stop.
2. Call `s.idem.Start(key)`. It returns a saved response and a state:
   - `idempotency.KeyDone`: set the header `Idempotent-Replayed: true`, then
     `idempotency.Send` the saved response to `w`. Don't call `next`.
   - `idempotency.KeyBusy`: another request with this key is still running.
     Use `writeError` with status 409 (`http.StatusConflict`) and the message
     `a request with this Idempotency-Key is still in progress`.
   - `idempotency.KeyNew`: run `next` into a capture from
     `idempotency.NewCapture()` instead of `w`, then take its `Response()`.
3. If the response status is 500 or more, call `s.idem.Forget(key)` so a
   retry runs again. Otherwise call `s.idem.Finish(key, resp)` (4xx answers
   too: the same bad request gets the same answer). Either way, `Send` the
   response to the client.

```console
$ curl -i -X POST localhost:8080/hosts -H 'Idempotency-Key: k1' -d '{"name":"db-1","env":"dev","port":5432}'
HTTP/1.1 201 Created
…
$ curl -i -X POST localhost:8080/hosts -H 'Idempotency-Key: k1' -d '{"name":"db-1","env":"dev","port":5432}'
HTTP/1.1 201 Created
Idempotent-Replayed: true
…        (same body, same Location, no second host)
```

**Think it through:**
- A `switch` on the state fits the three cases; `KeyDone` and `KeyBusy` both return early.
- Step 3 runs only for `KeyNew`.

### 7. Run with graceful shutdown

**Goal:** finish `Run(ctx, ln, grace)` in `internal/server/run.go` so it reports ready, serves, and on `ctx` cancel drains requests within `grace`.

It already builds `srv := s.newHTTPServer()`, but then serves forever.

1. Mark the server ready (`s.ready`), then run `srv.Serve(ln)` in a
   goroutine that sends its error on a buffered channel of size 1.
2. `select` on that channel and `ctx.Done()`:
   - `Serve` returned first: mark not ready and return its error.
   - `ctx` cancelled: mark not ready, log `"shutting down"` at Info level
     with the attribute `"grace"` set to `grace.String()`, and call
     `srv.Shutdown` with a **new** context that times out after `grace`,
     built from `context.Background()`. (`ctx` is already cancelled, so it
     can't be used.)
3. If `Shutdown` returns an error (the grace period ran out), call
   `srv.Close()` and return the error wrapped as `graceful shutdown: <err>`
   (with `%w`). Otherwise return `nil`.

Add `"fmt"` to the imports in `run.go`.

`Shutdown` stops accepting new connections at once and waits for in-flight
requests. In Kubernetes, failing `/readyz` takes the pod out of the
Service's endpoints, so new traffic goes elsewhere while it drains.

```console
$ go run ./cmd/inventoryd
{"time":"…","level":"INFO","msg":"listening","addr":"127.0.0.1:8080"}
^C
{"time":"…","level":"INFO","msg":"shutting down","grace":"10s"}
{"time":"…","level":"INFO","msg":"stopped"}
```

**Think it through:**
- Why buffered? If nobody receives, the goroutine must still be able to send and exit.
- `context.WithTimeout` also returns a cancel function: defer it.

### Going further

- Add `?env=prod` filtering to `GET /hosts`.
- Make idempotency keys expire, as in *Idempotency keys: scope, expire and fingerprint*.
- Add a request-duration histogram, as in *Prometheus-style metrics: histograms*.

## Learn more

- [net/http: ServeMux patterns](https://pkg.go.dev/net/http#hdr-Patterns-ServeMux)
- [net/http: Server.Shutdown](https://pkg.go.dev/net/http#Server.Shutdown)
- [log/slog](https://pkg.go.dev/log/slog)
- [Kubernetes: liveness, readiness and startup probes](https://kubernetes.io/docs/concepts/configuration/liveness-readiness-startup-probes/)
