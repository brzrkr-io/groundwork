# JSON APIs: fix the alerts API

The alerts API your on-call dashboard reads from has several small bugs. Each
one breaks a client in a way that is easy to miss. Fix them, then finish the
list endpoint.

**New here:** why `http.Error` is wrong for JSON, nil slices encoding as
`null`.

## Learn

### Recap

A JSON API sets `Content-Type: application/json`, picks a status code, and
writes every body as JSON, errors included (`{"error":"..."}`).

- Headers must be set **before** `w.WriteHeader`. `WriteHeader` sends them;
  anything set later is silently dropped.
- `dec.DisallowUnknownFields()` turns a misspelled field into a decode error.

### `http.Error` sends plain text

`http.Error(w, msg, code)` always sends `Content-Type: text/plain;
charset=utf-8` and the body `msg` plus a newline. A JSON client can't parse
that. In a JSON API, send errors as JSON.

### Nil slices encode as `null`

`json` encodes a *nil slice* (declared but never filled) as `null`, and an
empty, non-nil slice as `[]`:

```go
var ports []int          // nil
json.Marshal(ports)      // null
json.Marshal([]int{})    // []
```

Many clients crash on `null` where they expect a list. So check for `nil` and
swap in an empty slice before encoding.

> **Why is it written like this?** `[]int{}` is a *composite literal*: the
> type, then `{}` with the items (none here). Go requires the braces. `nil` is
> Go's built-in "nothing" value for slices, maps and pointers.

### Sorting

As in *Sorting*, `slices.SortFunc` sorts a slice in place with a compare
function, and `cmp.Compare` does the comparing:

```go
slices.SortFunc(hosts, func(a, b Host) int {
	return cmp.Compare(a.Zone, b.Zone)
})
```

> **Why is it written like this?** Go requires the compare function to take
> two items and return an `int` (negative, zero or positive). The names `a`
> and `b` are your choice; pick ones that don't clash with other variables.

### The files

| File | What it is |
|---|---|
| `exercise.go` | The routes, helpers and handlers. You work here. |
| `store.go` | `Alert`, `Validate` and the in-memory `Store`. Already done. |

The tests read responses with `rec.Result()`, which shows exactly what a
client receives. Headers set too late are missing there.

## Your task

The API answers `POST /alerts` and `GET /alerts`. Every response is checked
for its status code, `Content-Type: application/json`, and its exact body.

### 1. Fix the response helpers

**Goal:** make `writeJSON` send its `Content-Type`, and make `writeError` send
a JSON body.

Right now `writeJSON` sets `Content-Type` too late, and `writeError` uses
`http.Error`. After your fix:

```go
writeJSON(w, http.StatusOK, Alert{Name: "DiskFull", Severity: "critical", Summary: "disk 95% full"})
// 200, Content-Type: application/json
// body: {"name":"DiskFull","severity":"critical","summary":"disk 95% full"}\n

writeError(w, http.StatusConflict, "alert already exists: DiskFull")
// 409, Content-Type: application/json
// body: {"error":"alert already exists: DiskFull"}\n
```

**Think it through:**
- Look at the order of the lines in `writeJSON`. Which line sends the headers?
- Which helper already knows how to send JSON? `writeError` can use it.

### 2. Fix creating an alert

**Goal:** in `handleCreate`, reject unknown fields and answer a new alert
with 201.

The two bugs:

- A misspelled field such as `"sumary"` is accepted and the summary is lost.
  It must get 400 and the error `invalid request body`.
- A new alert gets 200. It must get 201 Created.

These already work; keep them: broken JSON gives 400 `invalid request body`,
an invalid alert gives 400 with the message from `Validate`, and a taken name
gives 409 `alert already exists: <name>`.

```text
POST /alerts  {"name":"DiskFull","severity":"critical","summary":"disk 95% full"}
// 201, body: {"name":"DiskFull","severity":"critical","summary":"disk 95% full"}\n

POST /alerts  {"name":"DiskFull","severity":"critical","sumary":"disk 95% full"}
// 400, body: {"error":"invalid request body"}\n
```

**Think it through:**
- What must the decoder be told before it decodes?
- Which status constant means "created"?

### 3. Finish listing alerts

**Goal:** `handleList` answers 200 with every alert, sorted by name, and `[]`
when there are none.

`store.All()` returns the alerts in random order, and `nil` when the store is
empty. Sort by `Name`, A to Z. Add the `cmp` and `slices` imports.

```text
GET /alerts   (store is empty)
// 200, body: []\n

GET /alerts   (store holds PodCrashLooping, DiskFull, HighLatency, CertExpiring)
// 200, body: [{"name":"CertExpiring",...},{"name":"DiskFull",...},
//             {"name":"HighLatency",...},{"name":"PodCrashLooping",...}]\n
```

**Think it through:**
- What does `alerts` hold when the store is empty, and how does that encode?
- Which field do you compare when sorting?

## Learn more

- [net/http: Error](https://pkg.go.dev/net/http#Error)
- [encoding/json: Marshal (how slices encode)](https://pkg.go.dev/encoding/json#Marshal)
- [slices: SortFunc](https://pkg.go.dev/slices#SortFunc)
