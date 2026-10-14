## Task 1: Fix the status recorder

`WriteHeader` already receives the code and passes it on. It only forgets to
remember it in the recorder.

### Closer

Assign the parameter `code` to the field `r.status` before calling the wrapped
`WriteHeader`.

### Answer

```go
func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}
```

## Task 2: Log one structured line per request

Pick the level first, based on `rec.status`, then log once with a fixed
message and the values as attributes. Because the level depends on data, use
`logger.Log` instead of `logger.Info`.

### Closer

Start with `level := slog.LevelInfo` and switch to `slog.LevelError` when
`rec.status >= 500`. Then:

```go
logger.Log(r.Context(), level, "request",
	"method", r.Method,
	// path, status, request_id ...
)
```

Read the header with `r.Header.Get("X-Request-ID")`.

### Answer

```go
level := slog.LevelInfo
if rec.status >= 500 {
	level = slog.LevelError
}
logger.Log(r.Context(), level, "request",
	"method", r.Method,
	"path", r.URL.Path,
	"status", rec.status,
	"request_id", r.Header.Get("X-Request-ID"),
)
```

Delete the old `fmt.Sprintf` line and the `"fmt"` import.

## Task 3: Make the level switchable

The handler is given a fixed level, so changing the `LevelVar` does nothing.
Give the handler the `LevelVar` instead.

### Closer

`HandlerOptions.Level` accepts a `*slog.LevelVar` as well as a fixed level.
Replace `slog.LevelInfo` in the options with the `level` variable.

### Answer

```go
opts := &slog.HandlerOptions{Level: level}
```

The handler reads the `LevelVar` on every record, so `level.Set` takes effect
straight away.
