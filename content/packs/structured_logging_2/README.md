# Structured logging: access logs

An *access log* records one line per HTTP request. It answers "how many 500s
did the API return in the last hour, and which request IDs?"

**New here:** `logger.Log`, `slog.LevelVar`, `r.Header.Get`.

## Learn

### Recap

As in *Structured logging with slog*: a `*slog.Logger` writes through a
handler, `slog.NewJSONHandler` writes JSON lines, attributes are key/value
pairs after the message, and `HandlerOptions.Level` sets the minimum level.
*Middleware* wraps the next `http.Handler` and does extra work around it.

### Messages are names, attributes are data

Don't build the message with `fmt.Sprintf`:

```go
logger.Info(fmt.Sprintf("job %s took %dms", job, ms))
// {"msg":"job resize took 812ms"}
```

Every line now has a different message, and the number is buried in text.
Keep the message a fixed name; put changing values in attributes:

```go
logger.Info("job finished", "job", job, "ms", ms)
// {"msg":"job finished","job":"resize","ms":812}
```

Now `ms > 500` is a simple query.

> **Why is it written like this?** Go doesn't require this: both compile.
> It is a logging convention, because log platforms filter on fields.

### Choosing the level at runtime: logger.Log

`logger.Info` and `logger.Error` have the level built into their name. When
the level depends on data, use `logger.Log`. It takes a context and the level
before the message:

```go
level := slog.LevelWarn
if retries > 3 {
	level = slog.LevelError
}
logger.Log(ctx, level, "retrying", "retries", retries)
```

Inside an HTTP handler, the context to pass is the request's: `r.Context()`.

> **Why is it written like this?** The argument order (context, level,
> message, attributes) is required by `Log`'s signature. Starting with a
> default and overriding it in an `if` is a common Go pattern, not a rule.

### Reading a request header

`r.Header.Get("Name")` returns a header's value, or `""` when it is missing:

```go
agent := r.Header.Get("User-Agent")
```

### Changing the minimum level live: slog.LevelVar

During an incident you want debug logs *now*, without a restart. A
`*slog.LevelVar` is a level you can change at any time, from any goroutine.
It starts at `slog.LevelInfo`.

If you give the handler a `LevelVar` as its `Level`, it checks the current
value on every record. If you give it a fixed level, `Set` changes nothing.

```go
lv := new(slog.LevelVar)
lv.Set(slog.LevelWarn)
fmt.Println(lv.Level()) // WARN
```

> **Why is it written like this?** `new(T)` returns a pointer to a zero `T`;
> that's Go syntax. `HandlerOptions.Level` accepts both `slog.LevelInfo` and
> a `*slog.LevelVar` because both satisfy the `slog.Leveler` interface.

## Your task

The deploy API's access logs say every request returned 200, the messages
can't be queried, and debug logging can't be switched on. Fix
`exercise.go`.

### 1. Fix the status recorder

**Goal:** make `statusRecorder.WriteHeader` remember `code` in `r.status`.

A plain `http.ResponseWriter` doesn't tell you which status was sent, so
`statusRecorder` wraps it. `WriteHeader` already passes the code on; it just
forgets to store it. (Here `r` is the recorder, not a request.)

```go
rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
rec.WriteHeader(http.StatusNotFound)
// rec.status => 404
```

`status` starts at 200 because a handler that never calls `WriteHeader`
sends 200.

**Think it through:**
- Which parameter holds the new code, and which field should receive it?

### 2. Log one structured line per request

**Goal:** make `AccessLog` log exactly one line with message `request` after
the request is served.

| Key | Value |
|---|---|
| `method` | the request's method |
| `path` | the request's URL path |
| `status` | the recorded status, as a number |
| `request_id` | the `X-Request-ID` request header (`""` if missing) |

Level: `slog.LevelError` when the status is 500 or higher, otherwise
`slog.LevelInfo`. Replace the `fmt.Sprintf` line, and remove the `fmt`
import (Go rejects unused imports).

```text
GET /deploys with X-Request-ID: r-42, handler sends 503
{"time":"...","level":"ERROR","msg":"request","method":"GET","path":"/deploys","status":503,"request_id":"r-42"}
```

**Think it through:**
- The status lives on `rec`, not on `r`. Which comparison picks the level?
- Which logger method takes the level as an argument?

### 3. Make the level switchable

**Goal:** make the logger from `NewLeveledLogger` obey its `LevelVar`.

The function already creates `level` and returns it, but the handler options
use a fixed `slog.LevelInfo`, so `level.Set` has no effect.

```go
logger, level := NewLeveledLogger(os.Stdout)
logger.Debug("hidden")         // not written
level.Set(slog.LevelDebug)
logger.Debug("cache lookup")   // written
```

**Think it through:**
- Which line sets the handler's `Level`, and which variable should it use?

## Learn more

- [slog.Logger.Log](https://pkg.go.dev/log/slog#Logger.Log)
- [slog.LevelVar](https://pkg.go.dev/log/slog#LevelVar)
- [Structured Logging with slog (Go blog)](https://go.dev/blog/slog)
