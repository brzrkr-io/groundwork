# Encoding JSON: a deployment status report

Dashboards, APIs, log collectors and `kubectl -o json` all speak JSON. Here you
turn Go structs into JSON with the `encoding/json` package, so your tools can
report deployment status to anything that listens.

**New here:** `encoding/json`, `json.Marshal`, `json.MarshalIndent`, struct tags.

## Learn

### json.Marshal

`json.Marshal(v)` takes any Go value and returns two things: its JSON form as
a `[]byte` (a slice of bytes) and an `error`. Turn the bytes into a string
with `string(b)`.

```go
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

> **Why is it written like this?** Go requires the `"encoding/json"` import
> before you can call `json.Marshal`. Returning `[]byte` instead of `string`
> is the package's choice: bytes are what files and networks want. Checking
> `err` right away is convention, as in *Errors*.

### Only exported fields are encoded

`encoding/json` is another package, so it can only see *exported* fields:
the ones that start with a capital letter. Lowercase fields are skipped
silently. No error, they just don't appear.

```go
type Host struct {
	Name string
	port int // lowercase: never encoded
}
// {"Name":"web-1"}
```

### Struct tags

By default the JSON key is the Go field name (`"Name"`). To choose a
different key, add a *struct tag*: a string in backquotes after the field's
type.

```go
type User struct {
	Login    string `json:"login"`
	Nickname string `json:"nickname,omitempty"`
	Password string `json:"-"`
}
```

| Tag | Effect |
|---|---|
| `json:"login"` | Use `login` as the key. |
| `json:"nickname,omitempty"` | Use `nickname`, but leave the field out when it holds its zero value (`""`, `0`, `false`, `nil`). |
| `json:"-"` | Never encode this field. Good for secrets. |

> **Why is it written like this?** Go requires the backquotes; the tag is
> just text that the compiler stores. The `json:"…"` format is what the
> `encoding/json` package expects, and it's strict: no spaces anywhere.
> Lining the tags up in a column is gofmt's job, not yours. The key names
> are your (or your API's) choice.

### How Go values become JSON

| Go value | JSON |
|---|---|
| `string` | `"text"` |
| `int`, `float64` | `42`, `0.5` |
| `bool` | `true` / `false` |
| struct | object `{...}`, keys in field order |
| `map[string]int` | object `{...}`, keys **sorted** alphabetically |
| slice with items | array `[...]` |
| empty slice, e.g. `[]string{}` | `[]` |
| nil slice, e.g. `var s []string` | `null` |
| `time.Time` | RFC 3339 string: `"2026-10-11T09:30:00Z"` |

That nil-slice row catches people out. Many dashboards choke on `null`
where they expect a list. An empty slice (made with `{}`) encodes as `[]`.

### Pretty output with MarshalIndent

`json.MarshalIndent(v, prefix, indent)` works like `Marshal`, but puts each
field on its own line. `prefix` goes at the start of every line; `indent` is
added once per nesting level.

```go
b, err := json.MarshalIndent(Host{Name: "web-1", Port: 8080}, "", "\t")
```

```json
{
	"Name": "web-1",
	"Port": 8080
}
```

There is no newline after the final `}`.

## Your task

The deploy dashboard wants a JSON status report for every environment. The
`Deployment` and `Report` types are in `exercise.go`. `Report` already has
its tags, so use it as a model.

### 1. Tag the Deployment struct

**Goal:** add a struct tag to every field of `Deployment` so `json.Marshal`
produces the keys below.

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

**Think it through:**
- Look at how `Report` writes its tags: where does the tag go on the line?
- `"replicas":0` and `"healthy":false` still appear. Which single field gets `omitempty`?
- Which row of the tag table hides a field completely?

### 2. Count healthy deployments

**Goal:** `HealthCounts(deps)` returns a map with the keys `"healthy"` and
`"unhealthy"`.

`"healthy"` counts deployments whose `Healthy` is `true`; `"unhealthy"`
counts all the others. Both keys are always present, even when a count is 0.

```go
HealthCounts([]Deployment{{Healthy: true}, {Healthy: true}, {Healthy: false}})
// => map[healthy:2 unhealthy:1]

HealthCounts(nil)
// => map[healthy:0 unhealthy:0]
```

**Think it through:**
- With `nil` the loop never runs. How do both keys still end up in the map?
- Each deployment adds 1 to exactly one of the two keys. Which field decides?

### 3. Encode the report

**Goal:** `ReportJSON(env, generatedAt, deps)` returns the `Report` encoded
as indented JSON, as a string.

Build a `Report` from the pieces:

| Report field | Comes from |
|---|---|
| `Environment` | parameter `env` |
| `GeneratedAt` | parameter `generatedAt` |
| `Counts` | your `HealthCounts`, called on `deps` |
| `Deployments` | parameter `deps`, but an empty slice when `deps` is nil |

Encode it with `json.MarshalIndent`, using an empty prefix and **two spaces**
as the indent. If encoding fails, return `""` and the error. Add
`"encoding/json"` to the imports. This task uses your tags from task 1.

With the two deployments from task 1, `env` = `"production"` and
`generatedAt` = `time.Date(2026, 10, 11, 10, 0, 0, 0, time.UTC)`:

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

With `env` = `"staging"` and `deps` = `nil` (an empty `[]Deployment{}` gives
the same `"deployments": []`):

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

**Think it through:**
- Where does `null` come from, and what can you check before building the `Report`?
- `MarshalIndent` returns bytes and an error. What must you do with each before returning?

## Learn more

- [encoding/json: Marshal](https://pkg.go.dev/encoding/json#Marshal)
- [Go by Example: JSON](https://gobyexample.com/json)
- [Go blog: JSON and Go](https://go.dev/blog/json)
