# Project: inventory service

You'll finish a small production-style HTTP service: a JSON API that tracks
your fleet's hosts. It has the things every real service needs: validation,
request IDs, access logs, Prometheus metrics, health probes, idempotent
creates and graceful shutdown. Each piece is something you practised in this
module. Here you wire them together into one program you can run with
`go run` and `curl`.

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

A host looks like this in JSON:

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

The layers stay separate. `internal/hosts` knows nothing about HTTP: it
stores hosts and enforces the rules. `internal/server` turns HTTP into store
calls and store errors into status codes. `cmd/inventoryd` reads flags,
handles signals and starts everything.

### File map

| File | What's in it | Tasks |
|---|---|---|
| `go.mod` | the module `gopherlings.local/inventory_service` | |
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

Do the tasks in order. Later tests send real requests through the whole
stack, so they rely on the earlier tasks working.

### New pieces

A few small things that earlier exercises didn't show:

- **`r.PathValue("id")`** reads the `{id}` part of a route pattern such as
  `"GET /hosts/{id}"`. For `/hosts/h3` it returns `"h3"`.
- **`context.WithValue(ctx, key, value)`** returns a child context carrying
  one value. Middleware uses it to hand data (like a request ID) to later
  handlers: `next.ServeHTTP(w, r.WithContext(ctx))`. `RequestID(ctx)` reads it
  back. The key is a private type, `requestIDKey{}`, so no other package can
  clash with it.
- **`logger.Log(ctx, level, msg, args...)`** is `logger.Info` with the level
  as a parameter, so you can choose `slog.LevelInfo` or `slog.LevelError`
  at run time.
- **`atomic.Bool`** (from `sync/atomic`) is a bool that many goroutines can
  read and write safely without a mutex: `s.ready.Store(true)`,
  `s.ready.Load()`.
- **`slices.SortedFunc(maps.Keys(m), cmp)`** collects a map's keys into a
  sorted slice. Map order is random, so you need this for stable output.
  `cmp.Or(a, b)` returns the first of its arguments that isn't zero, which
  gives you "sort by method, then by code" in one line:

```go
keys := slices.SortedFunc(maps.Keys(m.requests), func(a, b requestKey) int {
	return cmp.Or(cmp.Compare(a.Method, b.Method), cmp.Compare(a.Code, b.Code))
})
```

### Reading the tests

The tests mostly use `httptest.NewRequest` and `httptest.NewRecorder`,
through a helper called `do`. `wg.Go(f)` starts `f` in a goroutine tracked by
the WaitGroup. `TestRun` starts a real server on `127.0.0.1:0` (any free port)
and talks to it over TCP.

## Your task

Your platform team keeps its host list in a spreadsheet. You're replacing it
with a service the deploy tooling can call. The skeleton is wired up; fill in
the seven missing pieces.

### 1. Validate hosts

In `internal/hosts/hosts.go`, make `Validate` check these rules **in this
order** and return an error for the first rule that fails:

| Rule | Message |
|---|---|
| `Name` is not empty | `name is required` |
| `Env` is `"dev"`, `"staging"` or `"prod"` | `env must be dev, staging or prod` |
| `Port` is between 1 and 65535 (inclusive) | `port must be between 1 and 65535` |

Wrap `ErrInvalid` so callers can use `errors.Is`. `Region` is optional.
Return `nil` for a valid host.

```go
Validate(hosts.Host{Env: "prod", Port: 443})
// => error "invalid host: name is required", errors.Is(err, ErrInvalid) == true
Validate(hosts.Host{Name: "web-1", Env: "qa", Port: 0})
// => error "invalid host: env must be dev, staging or prod"
```

### 2. Finish the store

`Create`, `List` and `Len` are done. Use them as a pattern for the rest.
Every method locks `s.mu` first (`defer s.mu.Unlock()`).

- `Get(id)` returns the host, or `ErrNotFound`.
- `Update(id, h)` returns `ErrNotFound` if the ID is unknown, then the
  `Validate` error if `h` is invalid. Otherwise it stores `h` **with its ID
  set to `id`**, whatever `h.ID` said, and returns it.
- `Delete(id)` removes the host, or returns `ErrNotFound`.

```go
s.Update("h1", hosts.Host{ID: "ignored", Name: "web-1", Env: "prod", Port: 8443})
// => Host{ID: "h1", Name: "web-1", Env: "prod", Port: 8443}, nil
s.Delete("h99")
// => ErrNotFound
```

### 3. Finish the handlers

In `internal/server/handlers.go`, `handleList` and `handleGet` show the
pattern. Use `writeJSON`, `writeError` and `writeStoreError`.
`writeStoreError` already maps `ErrNotFound` to 404 and `ErrInvalid` to 400,
with `err.Error()` as the message. Decode request bodies with
`json.NewDecoder(r.Body).Decode(&h)`.

`handleCreate` (`POST /hosts`):

| Situation | Status | Body |
|---|---|---|
| body is not valid JSON | 400 | `{"error":"invalid JSON body"}` |
| host fails validation | 400 | `{"error":"invalid host: name is required"}` (the `Validate` message) |
| created | 201 | the stored host, plus header `Location: /hosts/h1` |

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

### 4. Add request IDs and access logs

In `internal/server/middleware.go`:

**`requestID`**: use the request's `X-Request-ID` header if it has one,
otherwise call `s.newRequestID()`, which returns `"req-1"`, `"req-2"`, and so
on. Set the same ID as the `X-Request-ID` **response** header. Store it in the
context with `context.WithValue(r.Context(), requestIDKey{}, id)` and call
`next` with `r.WithContext(ctx)`.

**`observe`**: wrap `w` in a `statusRecorder` (start `status` at
`http.StatusOK`, because handlers that never call `WriteHeader` send 200).
Serve the request through it, then write **one** log line with `s.logger`:

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

### 5. Count requests for /metrics

In `internal/server/metrics.go`:

- `Inc(method, code)` locks `m.mu` and adds one to
  `m.requests[requestKey{Method: method, Code: code}]`.
- `Write(w)` already prints the HELP and TYPE lines. After them, print one
  line per key, **sorted by method and then by code**, in exactly this
  format:

```text
# HELP inventory_http_requests_total HTTP requests handled, by method and status code.
# TYPE inventory_http_requests_total counter
inventory_http_requests_total{method="GET",code="200"} 2
inventory_http_requests_total{method="GET",code="404"} 1
inventory_http_requests_total{method="PATCH",code="405"} 1
```

With nothing counted, print only the two `#` lines. Hold the lock while you
read the map.

Finally, in `observe` (`middleware.go`), call `s.metrics.Inc(r.Method,
status)` after the request has been served. `handleMetrics` already appends
the `inventory_hosts` gauge. A scrape doesn't count itself, because `observe`
counts a request after its handler has finished.

### 6. Make POST /hosts idempotent

A deploy script that times out and retries `POST /hosts` must not create the
host twice. In `internal/server/idempotency.go`, write `idempotent` using
the provided `s.idem` store (an `*idempotency.Store`) and the
`internal/idempotency` package. Add its import,
`"gopherlings.local/inventory_service/internal/idempotency"`.

1. No `Idempotency-Key` header: call `next` normally and stop.
2. Call `resp, state := s.idem.Start(key)`:
   - `idempotency.KeyDone`: set the header `Idempotent-Replayed: true`, then
     `idempotency.Send(w, resp)` the saved response. Don't call `next`.
   - `idempotency.KeyBusy`: another request with this key is still running.
     Answer 409 with `writeError(w, http.StatusConflict, "a request with this
     Idempotency-Key is still in progress")`. The client can retry shortly.
   - `idempotency.KeyNew`: run `next` into `c := idempotency.NewCapture()`
     instead of `w`, then take `resp = c.Response()`.
3. If `resp.Status` is 500 or more, call `s.idem.Forget(key)`, so a retry
   runs again. Otherwise call `s.idem.Finish(key, resp)` (this includes 4xx
   answers: the same bad request gets the same answer). Either way,
   `idempotency.Send(w, resp)` to the client.

```console
$ curl -i -X POST localhost:8080/hosts -H 'Idempotency-Key: k1' -d '{"name":"db-1","env":"dev","port":5432}'
HTTP/1.1 201 Created
…
$ curl -i -X POST localhost:8080/hosts -H 'Idempotency-Key: k1' -d '{"name":"db-1","env":"dev","port":5432}'
HTTP/1.1 201 Created
Idempotent-Replayed: true
…        (same body, same Location, no second host)
```

### 7. Run with graceful shutdown

In `internal/server/run.go`, finish `Run(ctx, ln, grace)`. It already builds
the server, but then serves forever: it never reports ready and ignores
`ctx`.

1. `srv := s.newHTTPServer()` is already there (the timeouts are set).
2. Mark the server ready (`s.ready.Store(true)`) and call `srv.Serve(ln)`
   in a goroutine that sends its error on a buffered channel
   (`make(chan error, 1)`).
3. `select` on that channel and `ctx.Done()`:
   - If `Serve` returned first, mark not ready and return its error.
   - If `ctx` was cancelled, mark not ready, log `"shutting down"` at Info
     level with the attribute `"grace", grace.String()`, and call `srv.Shutdown` with a **new** context from
     `context.WithTimeout(context.Background(), grace)`. You can't use
     `ctx` here: it is already cancelled.
4. If `Shutdown` returns an error (the grace period ran out), call
   `srv.Close()` to drop the remaining connections and return
   `fmt.Errorf("graceful shutdown: %w", err)`. Otherwise return `nil`.

`Shutdown` stops accepting new connections at once and waits for in-flight
requests. In Kubernetes, failing `/readyz` takes the pod out of the Service's
endpoints, so new traffic goes to other pods while this one drains.

Try it out:

```console
$ go run ./cmd/inventoryd
{"time":"…","level":"INFO","msg":"listening","addr":"127.0.0.1:8080"}
^C
{"time":"…","level":"INFO","msg":"shutting down","grace":"10s"}
{"time":"…","level":"INFO","msg":"stopped"}
```

### Going further

- Add `?env=prod` filtering to `GET /hosts`.
- Make idempotency keys expire, as in *Idempotency keys: scope, expire and fingerprint*.
- Add a request-duration histogram, as in *Prometheus-style metrics: histograms*.

## Learn more

- [net/http: ServeMux patterns](https://pkg.go.dev/net/http#hdr-Patterns-ServeMux)
- [net/http: Server.Shutdown](https://pkg.go.dev/net/http#Server.Shutdown)
- [log/slog](https://pkg.go.dev/log/slog)
- [Kubernetes: liveness, readiness and startup probes](https://kubernetes.io/docs/concepts/configuration/liveness-readiness-startup-probes/)
