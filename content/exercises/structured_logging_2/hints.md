## Task 1: one missing line

`WriteHeader` receives the code as `code`. Save it in `r.status` before
passing it on: `r.status = code`.

## Task 2: pick the level first

Start with `level := slog.LevelInfo`, and switch to `slog.LevelError` when
`rec.status >= 500`. Then call `logger.Log(r.Context(), level, "request", ...)`.
Read the header with `r.Header.Get("X-Request-ID")`.

## Task 3: pass the LevelVar

`HandlerOptions.Level` accepts a `*slog.LevelVar` as well as a fixed level.
Replace `slog.LevelInfo` with `level`.

## Near-solution for task 2

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
