## Task 1: Validate hosts

Check the three rules in the README's order and return as soon as one fails.
Each error must wrap `ErrInvalid` so `errors.Is` finds it.

### Closer

A `switch` with no value checks its `case` conditions top to bottom, which
gives you "first rule wins" for free. Wrap with
`fmt.Errorf("%w: name is required", ErrInvalid)`.

### Answer

```go
switch {
case h.Name == "":
	return fmt.Errorf("%w: name is required", ErrInvalid)
case h.Env != "dev" && h.Env != "staging" && h.Env != "prod":
	return fmt.Errorf("%w: env must be dev, staging or prod", ErrInvalid)
case h.Port < 1 || h.Port > 65535:
	return fmt.Errorf("%w: port must be between 1 and 65535", ErrInvalid)
}
return nil
```

## Task 2: Finish the store

Copy the shape of `Create`: lock, then work with the `s.hosts` map. Each
method needs to know whether the ID exists before doing anything else.

### Closer

Start with `s.mu.Lock()` and `defer s.mu.Unlock()`, then use the comma-ok
lookup `h, ok := s.hosts[id]` and return `ErrNotFound` when `!ok`. In
`Update`, check `ok`, then `Validate(h)`, then set `h.ID = id`. `Delete` uses
the built-in `delete(s.hosts, id)`.

### Answer

```go
// Get
s.mu.Lock()
defer s.mu.Unlock()
h, ok := s.hosts[id]
if !ok {
	return Host{}, ErrNotFound
}
return h, nil

// Update
s.mu.Lock()
defer s.mu.Unlock()
if _, ok := s.hosts[id]; !ok {
	return Host{}, ErrNotFound
}
if err := Validate(h); err != nil {
	return Host{}, err
}
h.ID = id
s.hosts[id] = h
return h, nil

// Delete
s.mu.Lock()
defer s.mu.Unlock()
if _, ok := s.hosts[id]; !ok {
	return ErrNotFound
}
delete(s.hosts, id)
return nil
```

## Task 3: Finish the handlers

Every handler follows the same steps as `handleGet`: decode the body (if
any), call the store, and turn errors into responses with
`writeStoreError`.

### Closer

Decode with `json.NewDecoder(r.Body).Decode(&h)` and answer a decode error
with `writeError(w, http.StatusBadRequest, "invalid JSON body")`. Set the
`Location` header before `writeJSON`, because it sends the status line. For
a successful delete, call `w.WriteHeader(http.StatusNoContent)` and write
nothing.

### Answer

```go
// handleCreate
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

// handleUpdate: same decode, then
updated, err := s.store.Update(r.PathValue("id"), h)
if err != nil {
	writeStoreError(w, err)
	return
}
writeJSON(w, http.StatusOK, updated)

// handleDelete
if err := s.store.Delete(r.PathValue("id")); err != nil {
	writeStoreError(w, err)
	return
}
w.WriteHeader(http.StatusNoContent)
```

## Task 4: Add request IDs and access logs

`requestID` picks an ID, echoes it in the response, and passes it on through
the context. `observe` must see the status code, so it serves the request
through a `statusRecorder` and logs afterwards.

### Closer

Store the ID with `context.WithValue(r.Context(), requestIDKey{}, id)` and
finish with `next.ServeHTTP(w, r.WithContext(ctx))`. In `observe`, start
with `&statusRecorder{ResponseWriter: w, status: http.StatusOK}`, pick
`slog.LevelError` for `>= 500`, and log with `s.logger.Log(r.Context(),
level, "request", ...)`. Add `"log/slog"` to the imports.

### Answer

```go
// requestID
id := r.Header.Get("X-Request-ID")
if id == "" {
	id = s.newRequestID()
}
w.Header().Set("X-Request-ID", id)
ctx := context.WithValue(r.Context(), requestIDKey{}, id)
next.ServeHTTP(w, r.WithContext(ctx))

// observe
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

## Task 5: Count requests for /metrics

`Inc` adds one to a map entry under the lock. `Write` needs the keys in a
stable order, because map order is random. Don't forget the call in
`observe`.

### Closer

`Inc` is `m.requests[requestKey{Method: method, Code: code}]++` after
locking. In `Write`, lock, sort the keys with the
`slices.SortedFunc(maps.Keys(...), ...)` snippet from the README, and print
each with `fmt.Fprintf`. Add `cmp`, `maps` and `slices` to the imports.

### Answer

```go
// Inc
m.mu.Lock()
defer m.mu.Unlock()
m.requests[requestKey{Method: method, Code: code}]++

// Write, after the two # lines
m.mu.Lock()
defer m.mu.Unlock()
keys := slices.SortedFunc(maps.Keys(m.requests), func(a, b requestKey) int {
	return cmp.Or(cmp.Compare(a.Method, b.Method), cmp.Compare(a.Code, b.Code))
})
for _, k := range keys {
	fmt.Fprintf(w, "inventory_http_requests_total{method=\"%s\",code=\"%d\"} %d\n", k.Method, k.Code, m.requests[k])
}

// observe, after the log line
s.metrics.Inc(r.Method, rec.status)
```

## Task 6: Make POST /hosts idempotent

Follow the README's steps in order: no key, then switch on the state from
`s.idem.Start`, then run `next` into a capture and decide whether to save.

### Closer

`resp, state := s.idem.Start(key)`, then `switch state` with cases
`idempotency.KeyDone` and `idempotency.KeyBusy`, each ending in `return`.
For a new key, `c := idempotency.NewCapture()`, serve into `c`, and read
`c.Response()`. Add the `internal/idempotency` import.

### Answer

```go
key := r.Header.Get("Idempotency-Key")
if key == "" {
	next.ServeHTTP(w, r)
	return
}
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
if resp.Status >= 500 {
	s.idem.Forget(key)
} else {
	s.idem.Finish(key, resp)
}
idempotency.Send(w, resp)
```

Server errors are forgotten, so a retry with the same key runs again.

## Task 7: Run with graceful shutdown

Serve in a goroutine, then wait for whichever happens first: `Serve` fails,
or `ctx` is cancelled. Shutdown needs its own deadline, because `ctx` is
already done.

### Closer

Use `serveErr := make(chan error, 1)` and
`go func() { serveErr <- srv.Serve(ln) }()`, then `select` on `serveErr` and
`ctx.Done()`. For shutdown, `context.WithTimeout(context.Background(),
grace)` with `defer cancel()`, then `srv.Shutdown`; on error, `srv.Close()`.
Add `"fmt"` to the imports.

### Answer

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

s.ready.Store(false)
s.logger.Info("shutting down", "grace", grace.String())
shutdownCtx, cancel := context.WithTimeout(context.Background(), grace)
defer cancel()
if err := srv.Shutdown(shutdownCtx); err != nil {
	srv.Close()
	return fmt.Errorf("graceful shutdown: %w", err)
}
return nil
```
