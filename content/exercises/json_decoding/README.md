# Decoding JSON: load a service config

Service configs, API responses and alert webhooks all arrive as JSON. In this
exercise you turn JSON bytes into Go structs, catch typos in config files, and
dig into JSON whose shape you don't know in advance.

## Learn

### json.Unmarshal fills in a value you own

*Decoding* (or *unmarshalling*) turns JSON text into Go values. The
`encoding/json` package does it with `json.Unmarshal`:

```go
type Target struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

var t Target
err := json.Unmarshal([]byte(`{"host":"db-1","port":5432}`), &t)
// t.Host == "db-1", t.Port == 5432
```

Notice the `&t`. `Unmarshal` must change your variable, so it needs a
**pointer** to it, just like any function that modifies its argument (see
*Pointers*). Pass `t` without the `&` and you get an error instead of data.

### How keys find fields

The text in backticks after a field is a *struct tag*. `json:"host"` says
"this field holds the JSON key `host`". You saw the same tags in *Encoding
JSON*; decoding reads them the other way round.

| JSON | Go struct field it fills |
|---|---|
| key with a matching tag | that field |
| key with no tag match | a field with the same name, ignoring case (`"HOST"` → `Host`) |
| key that matches nothing | ignored, silently |
| field whose key is missing | left alone: keeps its zero value |

That last row is useful. If `replicas` is missing, your `Replicas` field stays
`0`, so after decoding you can spot the zero and apply a default.

Unexported (lowercase) fields are never filled. The json package can't see
them.

### Nested objects and arrays

A JSON object inside an object maps to a struct field of struct type. A JSON
array maps to a slice. A JSON object with arbitrary keys can map to a map:

```go
type Deploy struct {
	Image   string            `json:"image"`
	Ports   []int             `json:"ports"`  // [80, 443]
	Limits  Limits            `json:"limits"` // {"cpu": "500m"}
	Labels  map[string]string `json:"labels"` // {"team": "core"}
}
```

### Strict decoding with json.Decoder

`json.NewDecoder(r)` reads JSON from any `io.Reader` (see *Reading input line by line*). A
decoder has an option that `Unmarshal` lacks: `DisallowUnknownFields`. With it
switched on, a key that matches no field is an **error** instead of being
ignored. That catches typos such as `"replica": 3` in a config file, which
would otherwise silently leave `Replicas` at zero.

```go
dec := json.NewDecoder(r)
dec.DisallowUnknownFields()
var cfg Config
if err := dec.Decode(&cfg); err != nil {
	return Config{}, err
}
```

The error text comes from `encoding/json` itself and looks like
`json: unknown field "replica"`.

### JSON of unknown shape

Sometimes you only need one value from a big payload, or the shape varies.
Decode into `map[string]any`. The json package then picks Go types for you:

| JSON value | Go type inside `any` |
|---|---|
| object `{...}` | `map[string]any` |
| array `[...]` | `[]any` |
| number `3`, `2.5` | `float64` (always, even for whole numbers) |
| string `"x"` | `string` |
| `true` / `false` | `bool` |
| `null` | `nil` |

To use a value you need a type assertion. Use the comma-ok form (see *Type
switches*), so a wrong type gives you `ok == false` instead of a panic:

```go
var doc map[string]any
json.Unmarshal(data, &doc)

host, ok := doc["host"].(string)
if !ok {
	// missing, or not a string
}
```

A missing key gives `nil`, and asserting `nil` to any type also gives
`ok == false`. One check covers both cases.

### Syntax errors

Broken JSON gives a `*json.SyntaxError`. Its `Offset` field is the byte
position where parsing failed. Use `errors.AsType` to get at it, even through
a `%w` wrap:

```go
if se, ok := errors.AsType[*json.SyntaxError](err); ok {
	fmt.Println("bad JSON near byte", se.Offset)
}
```

A value of the wrong type, like `"port": "8080"` for an `int` field, gives a
different error (`*json.UnmarshalTypeError`). Either way, `err != nil`.

## Your task

You're writing the config loader for a small deploy tool, plus a helper for
the alert webhook. The starter imports `errors` and `io`. Add `encoding/json`
and `fmt` to the import list when you need them.

The `Config` and `HealthCheck` structs are already written, with their tags.
A config file looks like this:

```json
{
  "service": "checkout",
  "port": 8080,
  "replicas": 3,
  "regions": ["eu-west-1", "us-east-1"],
  "healthcheck": {"path": "/healthz", "interval": "10s"},
  "labels": {"team": "payments", "tier": "backend"}
}
```

### 1. Parse a config with a default

Write `ParseConfig(data)`:

- Decode `data` into a `Config` with `json.Unmarshal`.
- If decoding fails, return `Config{}` and the error wrapped as
  `parse config: %w`. The original error must stay reachable with
  `errors.AsType`, so use `%w`, not `%v`.
- If `Replicas` is `0` after decoding (missing or written as `0`), set it to
  `1`.

```go
ParseConfig([]byte(`{"service":"checkout","port":8080}`))
// => Config{Service: "checkout", Port: 8080, Replicas: 1}, nil

ParseConfig([]byte(`{"service":"checkout","replicas":3}`))
// => Config{Service: "checkout", Replicas: 3}, nil

ParseConfig([]byte(`{"service": checkout}`))
// => Config{}, error "parse config: invalid character 'c' looking for beginning of value"
```

### 2. Load a config strictly

Write `LoadStrict(r)`:

- Create a decoder with `json.NewDecoder(r)` and call
  `DisallowUnknownFields()` on it.
- Decode into a `Config`. On error, return `Config{}` and the error wrapped
  as `load config: %w`.
- `LoadStrict` does not apply the replicas default.

```go
LoadStrict(strings.NewReader(`{"service":"api","replicas":2}`))
// => Config{Service: "api", Replicas: 2}, nil

LoadStrict(strings.NewReader(`{"service":"api","replica":2}`))
// => Config{}, error "load config: json: unknown field \"replica\""
```

The tests check that the error starts with `load config: ` and contains
`unknown field "replica"` (with the real key name in quotes).

### 3. Find an alert's severity

Alert webhooks carry lots of fields that vary between senders. You only need
one: `labels` → `severity`.

```json
{"status": "firing", "labels": {"alertname": "DiskFull", "severity": "critical"}}
```

Write `Severity(data)`:

- Decode `data` into a `map[string]any`. If that fails, return `""` and the
  error from `json.Unmarshal` (wrapping it is optional).
- Get `labels` and assert it to `map[string]any`.
- Get `severity` from that map and assert it to `string`.
- If `labels` is missing or not an object, or `severity` is missing or not a
  string, return `""` and the sentinel error `ErrNoSeverity`.

```go
Severity([]byte(`{"labels":{"severity":"critical"}}`))  // => "critical", nil
Severity([]byte(`{"status":"firing"}`))                 // => "", ErrNoSeverity
Severity([]byte(`{"labels":"severity=critical"}`))      // => "", ErrNoSeverity
Severity([]byte(`{"labels":{"severity":3}}`))           // => "", ErrNoSeverity
Severity([]byte(`not json`))                            // => "", a *json.SyntaxError
```

## Learn more

- [JSON and Go (Go blog)](https://go.dev/blog/json)
- [encoding/json package](https://pkg.go.dev/encoding/json)
- [Go by Example: JSON](https://gobyexample.com/json)
