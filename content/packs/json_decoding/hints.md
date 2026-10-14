## Task 1: Parse a config with a default

`json.Unmarshal` fills in a value you give it, so it needs a pointer to `cfg`.
After decoding, check whether `Replicas` still needs its default.

### Closer

Add `"encoding/json"` and `"fmt"` to the imports. Call
`json.Unmarshal(data, &cfg)` and wrap a failure with
`fmt.Errorf("parse config: %w", err)`. Then `if cfg.Replicas == 0 { ... }`.

### Answer

```go
var cfg Config
if err := json.Unmarshal(data, &cfg); err != nil {
	return Config{}, fmt.Errorf("parse config: %w", err)
}
if cfg.Replicas == 0 {
	cfg.Replicas = 1
}
return cfg, nil
```

`%w` keeps the original JSON error reachable for `errors.AsType`.

## Task 2: Load a config strictly

A `json.Decoder` reads from an `io.Reader` and has a switch to reject keys the
struct doesn't know. Turn it on before decoding.

### Closer

`dec := json.NewDecoder(r)`, then `dec.DisallowUnknownFields()`, then
`dec.Decode(&cfg)`. You don't write the `unknown field` message yourself; the
decoder does.

### Answer

```go
dec := json.NewDecoder(r)
dec.DisallowUnknownFields()

var cfg Config
if err := dec.Decode(&cfg); err != nil {
	return Config{}, fmt.Errorf("load config: %w", err)
}
return cfg, nil
```

Wrapping with `%w` adds the `load config: ` prefix and keeps the decoder's
message after it.

## Task 3: Find an alert's severity

When the shape varies, decode into a `map[string]any` and dig one step at a
time. Each value you pull out is an `any`, so you must check its real type.

### Closer

Use the two-value type assertion:
`labels, ok := alert["labels"].(map[string]any)`. If `!ok`, return
`ErrNoSeverity`. Do the same for `labels["severity"]` with `.(string)`.

### Answer

```go
var alert map[string]any
if err := json.Unmarshal(data, &alert); err != nil {
	return "", err
}
labels, ok := alert["labels"].(map[string]any)
if !ok {
	return "", ErrNoSeverity
}
severity, ok := labels["severity"].(string)
if !ok {
	return "", ErrNoSeverity
}
return severity, nil
```

A missing key gives `nil`, and asserting `nil` fails, so one check per step
covers both "missing" and "wrong type".
