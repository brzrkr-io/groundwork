## Task 1: Unmarshal needs a pointer

Add `"encoding/json"` and `"fmt"` to the imports. Then pass `&cfg`, not `cfg`:

```go
if err := json.Unmarshal(data, &cfg); err != nil {
	return Config{}, fmt.Errorf("parse config: %w", err)
}
```

After that, an `if` that checks `cfg.Replicas == 0` sets the default.

## Task 2: a strict decoder

`json.NewDecoder(r)` gives you a `*json.Decoder`. Call
`dec.DisallowUnknownFields()` before decoding, then `dec.Decode(&cfg)`. You
don't write the `unknown field` message yourself; the decoder does.

## Task 3: one assertion per step

Decode into `var alert map[string]any`. `alert["labels"]` is an `any`. Assert
it with `labels, ok := alert["labels"].(map[string]any)`. If `!ok`, return
`ErrNoSeverity`. Do the same for `labels["severity"]` with `.(string)`.

## Almost the answer

```go
func LoadStrict(r io.Reader) (Config, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	var cfg Config
	if err := dec.Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("load config: %w", err)
	}
	return cfg, nil
}

func Severity(data []byte) (string, error) {
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
}
```
