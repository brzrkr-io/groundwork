## Task 1: two small swaps

`slog.NewJSONHandler` takes the same arguments as `slog.NewTextHandler`. The
minimum level is the `Level` field of `slog.HandlerOptions`.

```go
opts := &slog.HandlerOptions{
	Level: level,
}
return slog.New(slog.NewJSONHandler(w, opts))
```

## Task 2: With takes the same pairs as Info

`logger.With` accepts key/value pairs and `slog.Attr` values, just like
`logger.Info` does after the message. `slog.Group("target", ...)` builds one
attribute that holds other key/value pairs.

## Task 3: compare the key

Check `a.Key`. When it is `"password"` or `"token"`, return a new attribute
with the same key and the value `"REDACTED"`, built with `slog.String`.
Otherwise return `a` unchanged.

## Near-solution

```go
// DeployLogger
return logger.With(
	"deploy_id", d.ID,
	slog.Group("target", "service", d.Service, "region", d.Region),
)

// Redact
if a.Key == "password" || a.Key == "token" {
	return slog.String(a.Key, "REDACTED")
}
return a
```

Don't forget `ReplaceAttr: Redact` in `NewLogger`'s options.
