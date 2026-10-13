## Task 1: Log JSON at a minimum level

Two small changes in `NewLogger`: swap the text handler for the JSON one, and
tell the options which minimum level to use.

### Closer

`slog.NewJSONHandler` takes the same arguments as `slog.NewTextHandler`. The
minimum level is the `Level` field of `slog.HandlerOptions`:
`Level: level,`.

### Answer

```go
opts := &slog.HandlerOptions{
	Level: level,
}
return slog.New(slog.NewJSONHandler(w, opts))
```

## Task 2: Tag every line of a deploy

`logger.With` takes the same key/value pairs that `logger.Info` takes after
the message. You need one plain pair and one group.

### Closer

`slog.Group("target", ...)` builds one attribute that holds other key/value
pairs:

```go
return logger.With(
	"deploy_id", d.ID,
	slog.Group("target", ...),
)
```

### Answer

```go
return logger.With(
	"deploy_id", d.ID,
	slog.Group("target", "service", d.Service, "region", d.Region),
)
```

## Task 3: Redact secrets

Look at `a.Key`. For the two secret keys, return a new attribute with the same
key and a fixed value; return everything else unchanged. Then hook `Redact`
into `NewLogger`'s options.

### Closer

`slog.String(key, value)` builds a string attribute. In the options, set the
`ReplaceAttr` field to the function itself: `ReplaceAttr: Redact` (no
parentheses).

### Answer

```go
func Redact(groups []string, a slog.Attr) slog.Attr {
	if a.Key == "password" || a.Key == "token" {
		return slog.String(a.Key, "REDACTED")
	}
	return a
}
```

And in `NewLogger`:

```go
opts := &slog.HandlerOptions{
	Level:       level,
	ReplaceAttr: Redact,
}
```
