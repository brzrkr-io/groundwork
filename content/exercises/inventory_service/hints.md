## Task 1: one switch, three rules

A `switch` with no value checks its `case` conditions top to bottom, which
gives you "first rule wins" for free:

```go
switch {
case h.Name == "":
	return fmt.Errorf("%w: name is required", ErrInvalid)
case h.Env != "dev" && h.Env != "staging" && h.Env != "prod":
	// ...
}
return nil
```

## Task 2: copy the shape of Create

Every store method starts with `s.mu.Lock()` and `defer s.mu.Unlock()`.
Then use the comma-ok lookup:

```go
h, ok := s.hosts[id]
if !ok {
	return Host{}, ErrNotFound
}
```

In `Update`, check `ok` first, then call `Validate(h)`, then set `h.ID = id`
before storing. `Delete` uses the built-in `delete(s.hosts, id)`.

## Task 3: decode, call the store, answer

Every handler follows the same three steps. For `handleCreate`:

```go
var h hosts.Host
if err := json.NewDecoder(r.Body).Decode(&h); err != nil {
	writeError(w, http.StatusBadRequest, "invalid JSON body")
	return
}
created, err := s.store.Create(h)
if err != nil {
	writeStoreError(w, err)
	return
}
w.Header().Set("Location", "/hosts/"+created.ID)
writeJSON(w, http.StatusCreated, created)
```

Set headers before `writeJSON`, because it sends the status line. For
`handleDelete`, success is just `w.WriteHeader(http.StatusNoContent)`.

## Task 4: the recorder and the log line

```go
rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
next.ServeHTTP(rec, r)

level := slog.LevelInfo
if rec.status >= 500 {
	level = slog.LevelError
}
s.logger.Log(r.Context(), level, "request",
	"request_id", RequestID(r.Context()),
	"method", r.Method,
	"path", r.URL.Path,
	"status", rec.status,
)
```

Add `"log/slog"` to the imports. In `requestID`, finish with
`next.ServeHTTP(w, r.WithContext(ctx))`.

## Task 5: sort the keys, then print

`Inc` is `m.requests[requestKey{Method: method, Code: code}]++` under the
lock. In `Write`, lock, then sort the keys with the
`slices.SortedFunc(maps.Keys(...), ...)` snippet from the README, and print
each one:

```go
fmt.Fprintf(w, "inventory_http_requests_total{method=\"%s\",code=\"%d\"} %d\n",
	k.Method, k.Code, m.requests[k])
```

Add `cmp`, `maps` and `slices` to the imports.

## Task 6: switch on the key state

```go
resp, state := s.idem.Start(key)
switch state {
case idempotency.KeyDone:
	w.Header().Set("Idempotent-Replayed", "true")
	idempotency.Send(w, resp)
	return
case idempotency.KeyBusy:
	writeError(w, http.StatusConflict, "a request with this Idempotency-Key is still in progress")
	return
}
c := idempotency.NewCapture()
next.ServeHTTP(c, r)
resp = c.Response()
```

Then `s.idem.Forget` for 5xx, `s.idem.Finish` otherwise, and
`idempotency.Send(w, resp)`.

## Task 7: serve, wait, shut down

```go
srv := s.newHTTPServer()
serveErr := make(chan error, 1)
s.ready.Store(true)
go func() { serveErr <- srv.Serve(ln) }()

select {
case err := <-serveErr:
	s.ready.Store(false)
	return err
case <-ctx.Done():
}
```

After the `select`: mark not ready, log, create
`context.WithTimeout(context.Background(), grace)` (and `defer cancel()`),
then call `srv.Shutdown`. On error, call `srv.Close()` and return the wrapped
error. Add `"fmt"` to the imports.
