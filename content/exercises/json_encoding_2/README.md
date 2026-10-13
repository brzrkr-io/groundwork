# Encoding JSON: fix the alert stream

Log collectors such as Fluent Bit, Vector and Loki love *JSON Lines*: one
JSON object per line. Here you fix a broken audit logger and alert struct,
then stream alerts as JSON Lines.

## Learn

### Recap

`json.Marshal(v)` returns the JSON for `v` as a `[]byte`. Only exported
(capitalised) fields are encoded. Struct tags such as `json:"name"`,
`json:"name,omitempty"` and `json:"-"` control the key names, leave out
zero values, and skip fields.

### Silent failures

`encoding/json` never complains about a broken struct. It just produces
different output:

| Mistake | What you get |
|---|---|
| Lowercase field `user string` | Field missing. If every field is lowercase, you get `{}`. |
| Typo in the key `json:"sevrity"` | Key spelled `"sevrity"`. |
| Space after the colon `json: "region"` | Tag ignored, key is the field name `"Region"`. |

So when JSON looks wrong, read the struct first. `go vet` catches the
space-after-colon mistake, but not typos or lowercase fields.

### io.Writer

In *Reading input line by line* you met `io.Reader`, something you can read bytes from.
Its partner is `io.Writer`, something you can write bytes to: a file, a
network connection, or, in tests, a `strings.Builder`. The interface has one
method:

```go
type Writer interface {
	Write(p []byte) (n int, err error)
}
```

### json.NewEncoder

`json.NewEncoder(w)` wraps a writer. Each call to `Encode(v)` writes the JSON
for `v` to `w`, followed by a newline `"\n"`. That is exactly one JSON Lines
record:

```go
var sb strings.Builder
enc := json.NewEncoder(&sb)
enc.Encode(Host{Name: "web-1"}) // writes {"name":"web-1"}\n
enc.Encode(Host{Name: "web-2"}) // writes {"name":"web-2"}\n
```

`Encode` returns an error if the value can't be encoded or the write fails,
for example when the disk is full. Check it every time.

Side note: a `time.Duration` field encodes as a plain number of
nanoseconds (`5s` becomes `5000000000`), which surprises most dashboards.

## Your task

The platform team ships audit entries and alerts to a log collector, but the
JSON coming out is wrong.

### 1. Fix the empty audit entries

`AuditJSON` always returns `{}` because every field of `auditRecord` is
lowercase. Rename the fields so they are exported (you will need to update
the struct literal in `AuditJSON` too) and add tags so the keys are `time`,
`user`, `action` and `target`, in that order:

```go
at := time.Date(2026, 10, 11, 14, 5, 0, 0, time.UTC)
AuditJSON("pat", "deploy", "api", at)
// => `{"time":"2026-10-11T14:05:00Z","user":"pat","action":"deploy","target":"api"}`, nil
```

The type `auditRecord` itself can stay lowercase: it's the fields that count.

### 2. Fix the alert tags

Three of the tags on `Alert` are wrong. Fix them so the keys are `service`,
`severity`, `region`, `message` and `silenced`, and `silenced` is left out
when it is `false`:

```go
json.Marshal(Alert{Service: "api", Severity: "critical", Region: "eu-west-1",
	Message: "error rate above 5%"})
// => {"service":"api","severity":"critical","region":"eu-west-1","message":"error rate above 5%"}

json.Marshal(Alert{Service: "db", Severity: "warning", Region: "us-east-1",
	Message: "disk 85% full", Silenced: true})
// => {"service":"db","severity":"warning","region":"us-east-1","message":"disk 85% full","silenced":true}
```

### 3. Write JSON Lines

Implement `WriteJSONLines(w io.Writer, alerts []Alert) error`. Create one
`json.Encoder` for `w` and encode each alert in order, so each alert becomes
one line ending in `"\n"`. With the two alerts above, `w` receives:

```text
{"service":"api","severity":"critical","region":"eu-west-1","message":"error rate above 5%"}
{"service":"db","severity":"warning","region":"us-east-1","message":"disk 85% full","silenced":true}
```

With no alerts (`nil`), write nothing and return `nil`. If `Encode` returns an
error, stop and return that error unchanged.

## Learn more

- [encoding/json: Encoder](https://pkg.go.dev/encoding/json#Encoder)
- [io.Writer](https://pkg.go.dev/io#Writer)
- [JSON Lines format](https://jsonlines.org/)
