# Encoding JSON: a deployment status report

Dashboards, APIs, log collectors and `kubectl -o json` all speak JSON. In this
exercise you turn Go structs into JSON with the `encoding/json` package, so
your tools can report deployment status to anything that listens.

## Learn

### json.Marshal

`json.Marshal` takes any Go value and returns its JSON form as a `[]byte`
(a slice of bytes), plus an error. Convert the bytes to a string with
`string(b)`:

```go
import "encoding/json"

type Host struct {
	Name string
	Port int
}

b, err := json.Marshal(Host{Name: "web-1", Port: 8080})
if err != nil {
	return "", err
}
fmt.Println(string(b))
// {"Name":"web-1","Port":8080}
```

### Only exported fields are encoded

`encoding/json` lives in another package, so it can only see *exported*
fields: the ones that start with a capital letter. Lowercase fields are
silently skipped. There is no error, they just don't appear:

```go
type Host struct {
	Name string
	port int // lowercase: never encoded
}
// {"Name":"web-1"}
```

### Struct tags

By default the JSON key is the Go field name (`"Name"`). Most JSON uses
lowercase keys, so you attach a *struct tag* to the field. A tag is a string
in backquotes after the field type:

```go
type Host struct {
	Name     string `json:"name"`
	Port     int    `json:"port"`
	Comment  string `json:"comment,omitempty"`
	Password string `json:"-"`
}
```

| Tag | Effect |
|---|---|
| `json:"name"` | Use `name` as the key. |
| `json:"comment,omitempty"` | Use `comment` as the key, but leave the field out when it holds its zero value (`""`, `0`, `false`, `nil`). |
| `json:"-"` | Never encode this field. Good for secrets. |

The format is strict: `json:"name"` with no spaces anywhere.

### How Go values become JSON

| Go value | JSON |
|---|---|
| `string` | `"text"` |
| `int`, `float64` | `42`, `0.5` |
| `bool` | `true` / `false` |
| struct | object `{...}`, keys in field order |
| `map[string]int` | object `{...}`, keys **sorted** alphabetically |
| slice with items | array `[...]` |
| empty slice `[]T{}` | `[]` |
| nil slice | `null` |
| `time.Time` | a string in RFC 3339 format: `"2026-10-11T09:30:00Z"` |

That nil-slice row catches people out. `var deps []Deployment` is nil and
encodes as `null`, which many dashboards choke on. If you want `[]`, replace
the nil slice with an empty one before encoding:

```go
if deps == nil {
	deps = []Deployment{}
}
```

### Pretty output with MarshalIndent

`json.MarshalIndent(v, prefix, indent)` works like `Marshal`, but puts each
field on its own line. Use an empty prefix and two spaces of indent:

```go
b, err := json.MarshalIndent(Host{Name: "web-1", Port: 8080}, "", "  ")
```

```json
{
  "name": "web-1",
  "port": 8080
}
```

There is no newline after the final `}`.

## Your task

The deploy dashboard wants a JSON status report for every environment. The
`Deployment` and `Report` types are in `exercise.go`. `Report` already has its
tags.

### 1. Tag the Deployment struct

Add struct tags to `Deployment` so that `json.Marshal` produces these keys,
in this order:

| Field | Key | Rule |
|---|---|---|
| `Service` | `service` | always |
| `Version` | `version` | always |
| `Replicas` | `replicas` | always |
| `Healthy` | `healthy` | always |
| `Started` | `started` | always |
| `Error` | `error` | left out when empty |
| `AuthToken` | (none) | never encoded |

```go
started := time.Date(2026, 10, 11, 9, 30, 0, 0, time.UTC)

json.Marshal(Deployment{Service: "api", Version: "1.4.2", Replicas: 3,
	Healthy: true, Started: started, AuthToken: "s3cr3t"})
// => {"service":"api","version":"1.4.2","replicas":3,"healthy":true,"started":"2026-10-11T09:30:00Z"}

json.Marshal(Deployment{Service: "worker", Version: "2.0.0", Replicas: 0,
	Healthy: false, Started: started, Error: "image pull failed"})
// => {"service":"worker","version":"2.0.0","replicas":0,"healthy":false,"started":"2026-10-11T09:30:00Z","error":"image pull failed"}
```

Notice that `"replicas":0` and `"healthy":false` still appear: only `error`
uses `omitempty`.

### 2. Count healthy deployments

Implement `HealthCounts(deps []Deployment) map[string]int`. Return a map with
two keys: `"healthy"` (deployments where `Healthy` is true) and `"unhealthy"`
(all the others). Both keys are always present, even when a count is zero.

```go
HealthCounts([]Deployment{{Healthy: true}, {Healthy: true}, {Healthy: false}})
// => map[healthy:2 unhealthy:1]

HealthCounts(nil)
// => map[healthy:0 unhealthy:0]
```

### 3. Encode the report

Implement `ReportJSON(env string, generatedAt time.Time, deps []Deployment) (string, error)`:

1. Build a `Report` with `Environment: env`, `GeneratedAt: generatedAt`,
   `Counts` from `HealthCounts(deps)`, and `Deployments: deps`.
2. If `deps` is nil, use an empty slice instead, so the report shows
   `"deployments": []` and not `null`.
3. Encode it with `json.MarshalIndent(report, "", "  ")` (two spaces).
4. Return the JSON as a string. If encoding fails, return `""` and the error.

Add `"encoding/json"` to the imports. This task uses your tags from task 1,
so finish that first.

With the two deployments from task 1, `env` = `"production"` and
`generatedAt` = `time.Date(2026, 10, 11, 10, 0, 0, 0, time.UTC)`, the result
is:

```json
{
  "environment": "production",
  "generated_at": "2026-10-11T10:00:00Z",
  "counts": {
    "healthy": 1,
    "unhealthy": 1
  },
  "deployments": [
    {
      "service": "api",
      "version": "1.4.2",
      "replicas": 3,
      "healthy": true,
      "started": "2026-10-11T09:30:00Z"
    },
    {
      "service": "worker",
      "version": "2.0.0",
      "replicas": 0,
      "healthy": false,
      "started": "2026-10-11T09:30:00Z",
      "error": "image pull failed"
    }
  ]
}
```

With `env` = `"staging"` and `deps` = `nil`:

```json
{
  "environment": "staging",
  "generated_at": "2026-10-11T10:00:00Z",
  "counts": {
    "healthy": 0,
    "unhealthy": 0
  },
  "deployments": []
}
```

An empty, non-nil slice (`[]Deployment{}`) gives the same `"deployments": []`.

## Learn more

- [encoding/json: Marshal](https://pkg.go.dev/encoding/json#Marshal)
- [Go by Example: JSON](https://gobyexample.com/json)
- [Go blog: JSON and Go](https://go.dev/blog/json)
