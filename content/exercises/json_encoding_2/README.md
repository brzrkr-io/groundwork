# Encoding JSON: fix the alert stream

Log collectors such as Fluent Bit, Vector and Loki love *JSON Lines*: one
JSON object per line. Here you fix a broken audit logger and alert struct,
then stream alerts as JSON Lines.

**New here:** `io.Writer`, `json.NewEncoder`, debugging struct tags.

## Learn

### Recap

As in *Encoding JSON: a deployment status report*: `json.Marshal(v)` returns
the JSON for `v` as `[]byte`. Only exported (capitalised) fields are encoded.
Tags such as `json:"name"`, `json:"name,omitempty"` and `json:"-"` choose the
key, leave out zero values, or skip a field.

### Silent failures

`encoding/json` never complains about a broken struct. It just produces
different output:

| Mistake | What you get |
|---|---|
| Lowercase field `user string` | Field missing. All fields lowercase → `{}`. |
| Typo in the key `json:"hostnmae"` | Key spelled `"hostnmae"`. |
| Space after the colon `json: "host"` | Tag ignored; key is the field name, e.g. `"Host"`. |

So when JSON looks wrong, read the struct first. `go vet` catches the space
mistake, but not typos or lowercase fields.

> **Why is it written like this?** Go requires a capital first letter for a
> field to be visible outside its package. The type's own name can stay
> lowercase; only the fields matter to `encoding/json`. Renaming a field
> means every place that uses the old name must change too; the compiler
> points at those lines, but the fix starts at the declaration.

### io.Writer

In *Reading input line by line* you met `io.Reader`, something you read
bytes from. Its partner is `io.Writer`, something you write bytes to: a
file, a network connection, or, in tests, a `strings.Builder`.

```go
type Writer interface {
	Write(p []byte) (n int, err error)
}
```

> **Why is it written like this?** Taking an `io.Writer` parameter instead
> of a file is convention: the same function then works with files, sockets
> and test buffers.

### json.NewEncoder

`json.NewEncoder(w)` wraps a writer once. Each call to `enc.Encode(v)`
writes the JSON for `v` to `w`, followed by a newline `"\n"`. That is
exactly one JSON Lines record.

```go
var sb strings.Builder
enc := json.NewEncoder(&sb)
enc.Encode(Host{Name: "web-1"}) // writes {"name":"web-1"}\n
enc.Encode(Host{Name: "web-2"}) // writes {"name":"web-2"}\n
```

`Encode` returns an error if the value can't be encoded or the write fails,
for example when the disk is full. The example above ignores it to stay
short; your code must check it every time.

## Your task

The platform team ships audit entries and alerts to a log collector, but the
JSON coming out is wrong.

### 1. Fix the empty audit entries

**Goal:** make `AuditJSON` return all four fields, with keys `time`, `user`,
`action` and `target`, in that order.

Right now it returns `{}`: every field of `auditRecord` is lowercase. Rename
the fields so they're exported, add tags, and update the struct literal
inside `AuditJSON` to use the new names.

```go
at := time.Date(2026, 10, 11, 14, 5, 0, 0, time.UTC)
AuditJSON("pat", "deploy", "api", at)
// => `{"time":"2026-10-11T14:05:00Z","user":"pat","action":"deploy","target":"api"}`, nil
```

**Think it through:**
- Which letter of each field name decides whether `encoding/json` sees it?
- The field holding the time isn't called `time`. Which part sets the key?
- After renaming, which other line still uses the old lowercase names?

### 2. Fix the alert tags

**Goal:** fix the three broken tags on `Alert` so the keys are `service`,
`severity`, `region`, `message` and `silenced`, and `silenced` is left out
when it's `false`.

```go
json.Marshal(Alert{Service: "api", Severity: "critical", Region: "eu-west-1",
	Message: "error rate above 5%"})
// => {"service":"api","severity":"critical","region":"eu-west-1","message":"error rate above 5%"}

json.Marshal(Alert{Service: "db", Severity: "warning", Region: "us-east-1",
	Message: "disk 85% full", Silenced: true})
// => {"service":"db","severity":"warning","region":"us-east-1","message":"disk 85% full","silenced":true}
```

**Think it through:**
- Compare each tag letter by letter with the key list above.
- Check the *Silent failures* table: which mistake involves a space?
- Which tag option drops a `false` bool?

### 3. Write JSON Lines

**Goal:** `WriteJSONLines(w, alerts)` writes each alert to `w` as one line
of JSON, in order, each ending in `"\n"`.

Create one `json.Encoder` for `w`. With the two alerts above, `w` receives:

```text
{"service":"api","severity":"critical","region":"eu-west-1","message":"error rate above 5%"}
{"service":"db","severity":"warning","region":"us-east-1","message":"disk 85% full","silenced":true}
```

With no alerts (`nil`), write nothing and return `nil`. If `Encode` returns
an error, stop and return that error unchanged.

**Think it through:**
- Should the encoder be made once, or once per alert?
- What do you do with `Encode`'s error inside the loop, and what do you return when the loop finishes?

## Learn more

- [encoding/json: Encoder](https://pkg.go.dev/encoding/json#Encoder)
- [io.Writer](https://pkg.go.dev/io#Writer)
- [JSON Lines format](https://jsonlines.org/)
