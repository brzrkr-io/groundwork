# Structured logging with slog

*Project: uptime monitor* closed *Concurrency*. This module adds what a
long-running service needs in production: logs, metrics, health checks,
clean shutdown and resilience.

When a service misbehaves at 3am, its logs are the first thing you search.
*Structured logs* turn every log line into data you can filter, like "all
errors for service `checkout` in `eu-west-1`".

**New here:** `log/slog`, handlers, levels, `With`, `slog.Group`, `ReplaceAttr`.

## Learn

### Why structured logs

A plain log line is a sentence for humans:

```text
2026/10/11 09:15:02 deploy d-101 of checkout to eu-west-1 failed after 3 tries
```

A structured log line is one JSON object per line:

```json
{"time":"2026-10-11T09:15:02Z","level":"ERROR","msg":"deploy failed","deploy_id":"d-101","service":"checkout","tries":3}
```

Log platforms (Loki, Elasticsearch, CloudWatch Logs) split it into fields, so
a query like `service="checkout" AND level="ERROR"` just works.

### Logger and handler

The `log/slog` package has two parts:

- a `*slog.Logger`: the thing you call (`logger.Info(...)`).
- a *handler*: decides **how** each log call (a *record*) is written.

```go
handler := slog.NewTextHandler(os.Stdout, nil) // nil = default options
logger := slog.New(handler)

logger.Info("queue drained", "queue", "emails", "jobs", 12)
// time=... level=INFO msg="queue drained" queue=emails jobs=12
```

| Handler constructor | Writes |
|---|---|
| `slog.NewTextHandler(w, opts)` | `key=value` text (nice on a laptop) |
| `slog.NewJSONHandler(w, opts)` | one JSON object per line (what machines want) |

Both take the same two arguments, so swapping one for the other is easy.

> **Why is it written like this?** `slog.New(handler)` is required: a logger
> always needs a handler. Passing `nil` for options is allowed by the API and
> means "defaults". The names `handler` and `logger` are the author's choice.

### Attributes

After the message come *attributes*: alternating key, value, key, value.
`"queue", "emails"` means key `queue`, value `"emails"`. You can also build
one attribute at a time:

```go
logger.Warn("slow query", slog.String("table", "orders"), slog.Int("ms", 950))
```

> **Why is it written like this?** Keys are strings in quotes, because they
> become the JSON field names. Values can be any type. Mixing pairs and
> `slog.String(...)` in one call is allowed; pick whichever reads better.

### Levels and options

Every record has a *level*. The handler drops records below its minimum.

| Method | Level constant | JSON shows |
|---|---|---|
| `logger.Debug` | `slog.LevelDebug` | `"DEBUG"` |
| `logger.Info` | `slog.LevelInfo` | `"INFO"` |
| `logger.Warn` | `slog.LevelWarn` | `"WARN"` |
| `logger.Error` | `slog.LevelError` | `"ERROR"` |

You set the minimum in a `slog.HandlerOptions` struct:

```go
opts := &slog.HandlerOptions{Level: slog.LevelError}
logger := slog.New(slog.NewTextHandler(w, opts))
logger.Warn("ignored")  // below Error: dropped
logger.Error("written")
```

> **Why is it written like this?** `Field: value,` inside `{ }` is Go's
> syntax for setting struct fields by name. The trailing comma is required
> when the closing `}` is on its own line. Fields you leave out keep their
> zero value.

### With and Group

`logger.With(...)` returns a **new** logger that adds attributes to every
line. The original logger is unchanged. `slog.Group(name, ...)` nests
attributes under one key:

```go
reqLogger := logger.With("request_id", "r-7", slog.Group("client", "ip", "10.0.0.5"))
reqLogger.Info("cache miss")
// {...,"msg":"cache miss","request_id":"r-7","client":{"ip":"10.0.0.5"}}
```

### ReplaceAttr

`HandlerOptions.ReplaceAttr` is a function the handler calls on every
attribute just before writing it. It writes whatever the function returns.
The function must have exactly this shape:

```go
func(groups []string, a slog.Attr) slog.Attr
```

`a.Key` is the attribute's key; `a.Value` its value. `groups` lists the
groups it sits inside. Example that upper-cases one field:

```go
func shout(groups []string, a slog.Attr) slog.Attr {
	if a.Key == "env" {
		return slog.String("env", strings.ToUpper(a.Value.String()))
	}
	return a
}
```

> **Why is it written like this?** You assign a function to `ReplaceAttr` by
> its **name**, without `()`: you hand over the function itself, you don't
> call it. Go requires the parameter and result types to match exactly.

## Your task

Your deploy tool logs plain text, logs everything, and leaked a database
password into the logs. Fix all three in `exercise.go`.

### 1. Log JSON at a minimum level

**Goal:** make `NewLogger` return a logger that writes JSON to `w` and drops
records below `level`.

```go
logger := NewLogger(os.Stdout, slog.LevelInfo)
logger.Debug("cache miss")  // not written
logger.Info("service started", "service", "api", "port", 8080)
// {"time":"...","level":"INFO","msg":"service started","service":"api","port":8080}
```

**Think it through:**
- Which options field holds the minimum level, and which parameter has it?
- Which handler constructor writes JSON instead of text?

### 2. Tag every line of a deploy

**Goal:** make `DeployLogger` return `logger.With(...)` carrying the deploy's
details.

The attributes to add:

- `deploy_id`: the deploy's `ID` field
- a group named `target` holding `service` (from `Service`) and `region`
  (from `Region`)

The keys are fixed text; the values come from `d`. Don't change `logger`
itself: lines from the original logger must not get `deploy_id`.

```go
d := Deploy{ID: "d-101", Service: "checkout", Region: "eu-west-1"}
DeployLogger(logger, d).Info("rollout started", "replicas", 3)
// {"time":"...","level":"INFO","msg":"rollout started","deploy_id":"d-101","target":{"service":"checkout","region":"eu-west-1"},"replicas":3}
```

**Think it through:**
- How many key/value pairs go directly in `With`, and how many inside the group?
- How do you read a field from the struct `d`?

### 3. Redact secrets

**Goal:** make `Redact` hide the value of `password` and `token` attributes,
then make `NewLogger` use it.

- Key exactly `password` or `token` → return an attribute with the same key
  and the string value `REDACTED`.
- Any other key → return `a` unchanged.
- Groups don't matter: a `token` inside a group is redacted too. You don't
  need the `groups` parameter.

```go
Redact(nil, slog.String("password", "hunter2"))         // => password=REDACTED
Redact([]string{"db"}, slog.String("token", "s3cr3t"))  // => token=REDACTED
Redact(nil, slog.String("user", "alice"))               // => user=alice
```

Then fill in the second TODO in `NewLogger` so every logger it builds
redacts:

```go
logger.Info("login", "user", "alice", "password", "hunter2")
// {"time":"...","level":"INFO","msg":"login","user":"alice","password":"REDACTED"}
```

**Think it through:**
- Which part of `a` do you compare, and with which operator for "either one"?
- Which options field takes a function, and do you write `()` after the name?

## Learn more

- [Package log/slog](https://pkg.go.dev/log/slog)
- [Structured Logging with slog (Go blog)](https://go.dev/blog/slog)
- [Go by Example: Logging](https://gobyexample.com/logging)
