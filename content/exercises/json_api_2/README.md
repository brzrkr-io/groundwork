# JSON APIs: fix the alerts API

An alerts API that your on-call dashboard reads from has several small bugs.
Each one breaks a client in a way that is easy to miss. Find and fix them,
then finish the list endpoint.

## Learn

### Recap

A JSON API sets `Content-Type: application/json`, picks a status code, and
writes every body as JSON, errors included (`{"error":"..."}`). Headers must be
set **before** `w.WriteHeader`, because `WriteHeader` sends them; anything set
afterwards is silently dropped. When reading a body, `http.MaxBytesReader`
caps its size and `dec.DisallowUnknownFields()` turns a misspelled field into
a decode error instead of a silently empty value.

### New: `http.Error` sends plain text

`http.Error(w, msg, code)` is handy, but it always sends
`Content-Type: text/plain; charset=utf-8` and the body `msg` plus a newline.
A JSON client that tries to decode `invalid request body` fails with a
confusing parse error. In a JSON API, send errors as JSON instead.

### New: nil slices encode as `null`

`json` encodes a nil slice as `null`, and an empty, non-nil slice as `[]`:

```go
var none []Alert          // nil
json.Marshal(none)        // null
json.Marshal([]Alert{})   // []
```

Many clients crash on `null` where they expect a list. Turn a nil slice into
an empty one before encoding:

```go
if alerts == nil {
	alerts = []Alert{}
}
```

### Sorting

As in the sorting exercises, `slices.SortFunc` sorts a slice in place using
a compare function, and `cmp.Compare` does the comparing:

```go
slices.SortFunc(hosts, func(a, b Host) int {
	return cmp.Compare(a.Name, b.Name)
})
```

### The files

| File | What it is |
|---|---|
| `exercise.go` | The routes, helpers and handlers. You work here. |
| `store.go` | `Alert`, `Validate` and the in-memory `Store`. Already done. |

The tests read responses with `rec.Result()`, which shows exactly what a
client would receive, so headers set too late are missing there.

## Your task

The API answers `POST /alerts` and `GET /alerts`. Each response is checked for
its status code, a `Content-Type` of `application/json`, and its exact body.

### 1. Fix the response helpers

`writeJSON` sets `Content-Type` too late, so clients never see it. `writeError`
uses `http.Error`, so error bodies are plain text. Fix both so that:

```go
writeJSON(w, http.StatusOK, Alert{Name: "DiskFull", Severity: "critical", Summary: "disk 95% full"})
// 200, Content-Type: application/json
// body: {"name":"DiskFull","severity":"critical","summary":"disk 95% full"}\n

writeError(w, http.StatusConflict, "alert already exists: DiskFull")
// 409, Content-Type: application/json
// body: {"error":"alert already exists: DiskFull"}\n
```

### 2. Fix creating an alert

`handleCreate` has two bugs:

- A misspelled field such as `"sumary"` is accepted, and the summary is lost.
  It must be rejected with 400 and the error `invalid request body`.
- A new alert gets status 200. It must get 201 Created.

The other responses are already right: broken JSON gives 400
`invalid request body`, an invalid alert gives 400 with the message from
`Validate`, and a taken name gives 409 `alert already exists: <name>`.

```text
POST /alerts  {"name":"DiskFull","severity":"critical","summary":"disk 95% full"}
// 201, body: {"name":"DiskFull","severity":"critical","summary":"disk 95% full"}\n

POST /alerts  {"name":"DiskFull","severity":"critical","sumary":"disk 95% full"}
// 400, body: {"error":"invalid request body"}\n
```

### 3. Finish listing alerts

`handleList` serves `GET /alerts`. It always answers 200 with a JSON list of
every alert, **sorted by name** (A to Z). With no alerts, the body must be
`[]`, not `null`. `store.All()` returns the alerts in random order, and `nil`
when the store is empty.

```text
GET /alerts   (store is empty)
// 200, body: []\n

GET /alerts   (store holds PodCrashLooping, DiskFull, HighLatency, CertExpiring)
// 200, body: [{"name":"CertExpiring",...},{"name":"DiskFull",...},
//             {"name":"HighLatency",...},{"name":"PodCrashLooping",...}]\n
```

Add the `cmp` and `slices` imports.

## Learn more

- [net/http: Error](https://pkg.go.dev/net/http#Error)
- [encoding/json: Marshal (how slices encode)](https://pkg.go.dev/encoding/json#Marshal)
- [slices: SortFunc](https://pkg.go.dev/slices#SortFunc)
