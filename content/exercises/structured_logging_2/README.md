# Structured logging: access logs

An *access log* records one line per HTTP request. It is how you answer "how
many 500s did the API return in the last hour, and which request IDs?" In this
exercise you fix a half-finished access-log middleware and add a log level
you can change while the service runs.

## Learn

### Recap

As in *Structured logging with slog*, a `*slog.Logger` writes records through
a handler; `slog.NewJSONHandler` writes one JSON object per line. Attributes
come as key/value pairs after the message, and `HandlerOptions.Level` sets the
minimum level. Middleware takes the next `http.Handler` and returns a new one
that does extra work around it.

### Messages are names, attributes are data

It is tempting to build the message with `fmt.Sprintf`:

```go
logger.Info(fmt.Sprintf("%s %s -> %d", r.Method, r.URL.Path, status))
// {"level":"INFO","msg":"GET /deploys -> 500"}
```

Now every request has a different message, and the status is buried in text.
To find all 500s, your log platform needs a regular expression. Keep the
message a fixed name and put the changing values in attributes:

```go
logger.Info("request", "method", r.Method, "path", r.URL.Path, "status", status)
// {"level":"INFO","msg":"request","method":"GET","path":"/deploys","status":500}
```

Now `status >= 500` is a simple query.

### Choosing the level at runtime: logger.Log

`logger.Info` and `logger.Error` hard-code the level. When the level depends on
data, use `logger.Log`, which takes a `context.Context` and the level first:

```go
level := slog.LevelInfo
if failed {
	level = slog.LevelError
}
logger.Log(r.Context(), level, "request", "path", r.URL.Path)
```

Inside a handler, pass the request's context, `r.Context()`.

### Changing the minimum level live: slog.LevelVar

During an incident you want debug logs *now*, without a restart. A
`*slog.LevelVar` is a level that you can change at any time, safely from any
goroutine. Its zero value is `slog.LevelInfo`. Pass it as the handler's
`Level`, and the handler checks its current value on every record:

```go
level := new(slog.LevelVar) // starts at Info
logger := slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: level}))

logger.Debug("cache lookup") // dropped
level.Set(slog.LevelDebug)
logger.Debug("cache lookup") // written
```

Real services often wire `level.Set` to an admin endpoint or a signal.

## Your task

The deploy API's access logs say every request returned 200, the messages
can't be queried, and debug logging can't be switched on.

### 1. Fix the status recorder

A plain `http.ResponseWriter` doesn't tell you which status the handler sent,
so `statusRecorder` wraps it. Its `WriteHeader` method passes the code on but
forgets to store it in `r.status`. Fix it so `r.status` holds the code.

```go
rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
rec.WriteHeader(http.StatusNotFound)
// rec.status => 404
```

`status` starts at 200 because a handler that never calls `WriteHeader` sends
200.

### 2. Log one structured line per request

Finish `AccessLog`. After the request is served, log exactly one line with the
message `request` and these attributes:

| Key | Value |
|---|---|
| `method` | `r.Method` |
| `path` | `r.URL.Path` |
| `status` | the recorded status code, as a number |
| `request_id` | the `X-Request-ID` request header (empty string if missing) |

Log at `slog.LevelError` when the status is 500 or higher, and at
`slog.LevelInfo` otherwise. Remove the `fmt.Sprintf` line and the `fmt`
import.

```text
GET /deploys with X-Request-ID: r-42, handler sends 503
{"time":"...","level":"ERROR","msg":"request","method":"GET","path":"/deploys","status":503,"request_id":"r-42"}
```

### 3. Make the level switchable

`NewLeveledLogger` returns a JSON logger and the `*slog.LevelVar` that should
control it, but the handler is built with a fixed `slog.LevelInfo`, so
`level.Set` has no effect. Pass the `LevelVar` as the handler's `Level`.

```go
logger, level := NewLeveledLogger(os.Stdout)
logger.Debug("hidden")         // not written
level.Set(slog.LevelDebug)
logger.Debug("cache lookup")   // written
```

## Learn more

- [slog.Logger.Log](https://pkg.go.dev/log/slog#Logger.Log)
- [slog.LevelVar](https://pkg.go.dev/log/slog#LevelVar)
- [Structured Logging with slog (Go blog)](https://go.dev/blog/slog)
