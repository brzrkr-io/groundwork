# Structured logging with slog

When a service misbehaves at 3am, its logs are the first thing you search.
*Structured logs* turn every log line into data you can filter, like "all
errors for service `checkout` in `eu-west-1`", instead of text you squint at.

## Learn

### Why structured logs

A plain log line is a sentence for humans:

```text
2026/10/11 09:15:02 deploy d-101 of checkout to eu-west-1 failed after 3 tries
```

To find every failed `checkout` deploy you would need a fragile regular
expression. A structured log line is a JSON object, one per line:

```json
{"time":"2026-10-11T09:15:02Z","level":"ERROR","msg":"deploy failed","deploy_id":"d-101","service":"checkout","tries":3}
```

Log platforms such as Loki, Elasticsearch and CloudWatch Logs parse each line
into fields, so a query like `service="checkout" AND level="ERROR"` just works.

### log/slog

Go's standard library ships a structured logger in the `log/slog` package. A
`*slog.Logger` sends each *record* (one log call) to a *handler*, and the
handler decides how to write it:

```go
handler := slog.NewJSONHandler(os.Stdout, nil) // nil = default options
logger := slog.New(handler)

logger.Info("service started", "service", "api", "port", 8080)
// {"time":"...","level":"INFO","msg":"service started","service":"api","port":8080}
```

After the message come *attributes* as key/value pairs: `"service", "api"`
means key `service`, value `"api"`. You can also build an attribute
explicitly, which reads well when there are many:

```go
logger.Warn("disk almost full", slog.String("host", "web-1"), slog.Int("percent", 93))
```

`slog.NewTextHandler` writes `key=value` text instead of JSON. It is handy on
your laptop, but machines prefer JSON.

### Levels

Each record has a *level*. The handler drops records below its minimum level.

| Method | Level | Use it for |
|---|---|---|
| `logger.Debug` | `slog.LevelDebug` | Noisy detail for troubleshooting |
| `logger.Info` | `slog.LevelInfo` | Normal events: started, deployed |
| `logger.Warn` | `slog.LevelWarn` | Something odd, but still working |
| `logger.Error` | `slog.LevelError` | Something failed |

You set the minimum level with `slog.HandlerOptions`:

```go
opts := &slog.HandlerOptions{Level: slog.LevelWarn}
logger := slog.New(slog.NewJSONHandler(w, opts))
logger.Info("ignored")  // below Warn: not written
logger.Warn("written")
```

In JSON output the level appears as `"level":"DEBUG"`, `"INFO"`, `"WARN"` or
`"ERROR"`.

### With and Group

`logger.With(...)` returns a **new** logger that adds the given attributes to
every line. The original logger is unchanged. `slog.Group` nests attributes
under one key:

```go
reqLogger := logger.With("request_id", "r-7", slog.Group("client", "ip", "10.0.0.5", "country", "NL"))
reqLogger.Info("cache miss")
// {"time":"...","level":"INFO","msg":"cache miss","request_id":"r-7","client":{"ip":"10.0.0.5","country":"NL"}}
```

### ReplaceAttr

`HandlerOptions.ReplaceAttr` is a function the handler calls on every
attribute just before writing it. Whatever it returns is written instead:

```go
func(groups []string, a slog.Attr) slog.Attr
```

`a.Key` is the attribute's key and `groups` lists the groups it sits inside
(`[]string{"client"}` for `ip` above). It is the perfect place to stop secrets
from ever reaching your logs.

## Your task

Your deploy tool logs as plain text, logs everything, and has leaked a
database password into the logs. Fix all three.

### 1. Log JSON at a minimum level

Change `NewLogger` so it returns a logger that writes JSON lines to `w` and
drops records below `level`. Use `slog.NewJSONHandler` and set `Level` in the
options.

```go
logger := NewLogger(os.Stdout, slog.LevelInfo)
logger.Debug("cache miss")  // not written
logger.Info("service started", "service", "api", "port", 8080)
// {"time":"...","level":"INFO","msg":"service started","service":"api","port":8080}
```

### 2. Tag every line of a deploy

Change `DeployLogger` so it returns `logger.With(...)` with these attributes:

- `deploy_id`: the deploy's `ID`
- a group called `target` holding `service` (the `Service`) and `region`
  (the `Region`)

```go
d := Deploy{ID: "d-101", Service: "checkout", Region: "eu-west-1"}
DeployLogger(logger, d).Info("rollout started", "replicas", 3)
// {"time":"...","level":"INFO","msg":"rollout started","deploy_id":"d-101","target":{"service":"checkout","region":"eu-west-1"},"replicas":3}
```

### 3. Redact secrets

Change `Redact` so that any attribute whose key is exactly `password` or
`token` gets the string value `REDACTED` (keep the key). Every other attribute
comes back unchanged. The key counts at any depth, so a `token` inside a group
is redacted too. You don't need `groups` for this.

```go
Redact(nil, slog.String("password", "hunter2"))         // => password=REDACTED
Redact([]string{"db"}, slog.String("token", "s3cr3t"))  // => token=REDACTED
Redact(nil, slog.String("user", "alice"))               // => user=alice
```

Then set `ReplaceAttr: Redact` in the options in `NewLogger`, so every logger
it builds redacts secrets:

```go
logger.Info("login", "user", "alice", "password", "hunter2")
// {"time":"...","level":"INFO","msg":"login","user":"alice","password":"REDACTED"}
```

## Learn more

- [Package log/slog](https://pkg.go.dev/log/slog)
- [Structured Logging with slog (Go blog)](https://go.dev/blog/slog)
- [Go by Example: Logging](https://gobyexample.com/logging)
