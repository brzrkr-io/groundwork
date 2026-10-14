# Decoding JSON: load a service config

Service configs, API responses and alert webhooks all arrive as JSON. Here
you turn JSON bytes into Go structs, catch typos in config files, and dig
into JSON whose shape you don't know in advance.

**New here:** `json.Unmarshal`, `json.NewDecoder`, `DisallowUnknownFields`, `map[string]any`, `*json.SyntaxError`.

## Learn

### json.Unmarshal fills in a value you own

*Decoding* (or *unmarshalling*) turns JSON text into Go values:

```go
type Target struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

var t Target
err := json.Unmarshal([]byte(`{"host":"db-1","port":5432}`), &t)
// t.Host == "db-1", t.Port == 5432
```

> **Why is it written like this?** Go requires the `&` here in practice:
> `Unmarshal` must change your variable, so it needs a *pointer* to it (see
> *Pointers*). Without `&` you get an error instead of data. Declaring `var t`
> first, then filling it, is how the API is designed. The name `t` is your
> choice.

### How keys find fields

The struct tags you wrote in *Encoding JSON: a deployment status report* work
the other way round when decoding:

| JSON | What happens |
|---|---|
| key with a matching tag | fills that field |
| key with no tag match | fills a field with the same name, ignoring case (`"HOST"` → `Host`) |
| key that matches nothing | ignored, silently |
| field whose key is missing | left alone: keeps its zero value |

That last row is useful: a missing number stays `0`, so after decoding you
can spot the zero and apply a default. Unexported (lowercase) fields are
never filled.

### Nested objects and arrays

A JSON object inside an object maps to a struct field. A JSON array maps to
a slice. An object with arbitrary keys can map to a map:

```go
type Deploy struct {
	Image  string            `json:"image"`
	Ports  []int             `json:"ports"`  // [80, 443]
	Limits Limits            `json:"limits"` // {"cpu": "500m"}
	Labels map[string]string `json:"labels"` // {"team": "core"}
}
```

### Strict decoding with json.Decoder

`json.NewDecoder(r)` reads JSON from any `io.Reader` (see *Reading input line
by line*). A decoder has an option `Unmarshal` lacks:
`DisallowUnknownFields()`. Once switched on, a key that matches no field is an
**error** instead of being ignored.

```go
d := json.NewDecoder(os.Stdin)
d.DisallowUnknownFields()
var t Target
err := d.Decode(&t)
// input {"host":"db-1","prot":5432}
// err: json: unknown field "prot"
```

That catches typos that would otherwise leave a field silently at zero.

> **Why is it written like this?** `DisallowUnknownFields` is a method call
> with no return value: it changes the decoder's settings, so it must come
> before `Decode`. That order is required by how the decoder works, not by
> the compiler.

### JSON of unknown shape

When you only need one value from a big or varying payload, decode into
`map[string]any`. The json package picks Go types for you:

| JSON value | Go type inside `any` |
|---|---|
| object `{...}` | `map[string]any` |
| array `[...]` | `[]any` |
| number `3`, `2.5` | `float64` (always, even whole numbers) |
| string `"x"` | `string` |
| `true` / `false` | `bool` |
| `null` | `nil` |

To use a value, you need a type assertion. Use the comma-ok form (see *Type
switches*), so a wrong type gives `ok == false` instead of a panic:

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

Broken JSON gives a `*json.SyntaxError`; its `Offset` field is the byte
where parsing failed. `errors.AsType` finds it, even through a `%w` wrap:

```go
if se, ok := errors.AsType[*json.SyntaxError](err); ok {
	fmt.Println("bad JSON near byte", se.Offset)
}
```

A value of the wrong type, like `"port": "8080"` for an `int` field, gives a
different error (`*json.UnmarshalTypeError`). Either way, `err != nil`.

## Your task

You're writing the config loader for a small deploy tool, plus a helper for
the alert webhook. The starter imports `errors` and `io`; add
`"encoding/json"` and `"fmt"` when you need them. `Config` and `HealthCheck`
already have their tags. A config file looks like this:

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

**Goal:** `ParseConfig(data)` decodes `data` into a `Config` with
`json.Unmarshal`, and sets `Replicas` to `1` when it is `0`.

- `Replicas` is `0` when the key is missing or written as `0`.
- If decoding fails (broken JSON or a wrong type), return `Config{}` and the
  error wrapped with the prefix `parse config: `. Wrap with `%w`, not `%v`,
  so `errors.AsType` can still find the `*json.SyntaxError`.

```go
ParseConfig([]byte(`{"service":"checkout","port":8080}`))
// => Config{Service: "checkout", Port: 8080, Replicas: 1}, nil

ParseConfig([]byte(`{"service":"checkout","replicas":3}`))
// => Config{Service: "checkout", Replicas: 3}, nil

ParseConfig([]byte(`{"service": checkout}`))
// => Config{}, error "parse config: invalid character 'c' looking for beginning of value"
```

**Think it through:**
- `cfg` is declared for you. What does `Unmarshal` need to be able to change it?
- When is it safe to check `Replicas`: before or after the error check?

### 2. Load a config strictly

**Goal:** `LoadStrict(r)` decodes a `Config` from the reader `r` and rejects
any key that matches no field.

- Use a `json.Decoder` with unknown fields disallowed.
- On error, return `Config{}` and the error wrapped with the prefix
  `load config: `. The error must still contain the json package's text,
  e.g. `unknown field "replica"`.
- No replicas default here.

```go
LoadStrict(strings.NewReader(`{"service":"api","replicas":2}`))
// => Config{Service: "api", Replicas: 2}, nil

LoadStrict(strings.NewReader(`{"service":"api","replica":2}`))
// => Config{}, error "load config: json: unknown field \"replica\""
```

**Think it through:**
- Which parameter does the decoder read from?
- Which call must happen before `Decode`?

### 3. Find an alert's severity

**Goal:** `Severity(data)` returns the string at `labels` → `severity` in an
alert payload of unknown shape.

```json
{"status": "firing", "labels": {"alertname": "DiskFull", "severity": "critical"}}
```

- Decode `data` into a `map[string]any`. If that fails, return `""` and the
  error from `json.Unmarshal` (wrapping is optional).
- If `labels` is missing or not an object, or `severity` is missing, `null`
  or not a string, return `""` and the sentinel error `ErrNoSeverity`.

```go
Severity([]byte(`{"labels":{"severity":"critical"}}`)) // => "critical", nil
Severity([]byte(`{"status":"firing"}`))                // => "", ErrNoSeverity
Severity([]byte(`{"labels":"severity=critical"}`))     // => "", ErrNoSeverity
Severity([]byte(`{"labels":{"severity":3}}`))          // => "", ErrNoSeverity
Severity([]byte(`not json`))                           // => "", a *json.SyntaxError
```

**Think it through:**
- Two steps down: what Go type does an inner JSON object become inside `any`?
- Each step needs one comma-ok assertion. What do you return when `ok` is false?

## Learn more

- [JSON and Go (Go blog)](https://go.dev/blog/json)
- [encoding/json package](https://pkg.go.dev/encoding/json)
- [Go by Example: JSON](https://gobyexample.com/json)
