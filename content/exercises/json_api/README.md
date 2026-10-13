# JSON APIs: build a service catalog

Most internal platform tools talk to each other through small JSON APIs: a
service catalog, a deploy API, an alert manager. Here you finish one that
sends and receives JSON the way clients expect.

## Learn

### What makes an API "JSON"

A JSON API does three things on every response:

1. It sets the header `Content-Type: application/json`, so clients know how
   to read the body.
2. It picks a status code that says what happened.
3. It writes the body as JSON, **including errors**. A client that always gets
   JSON never has to guess the format.

A common error shape is an object with one `error` key:

```json
{"error":"service not found: web"}
```

These are the status codes you need:

| Code | Constant | Meaning |
|---|---|---|
| 200 | `http.StatusOK` | Here is what you asked for. |
| 201 | `http.StatusCreated` | I created the thing you sent. |
| 400 | `http.StatusBadRequest` | Your request is broken or invalid. |
| 404 | `http.StatusNotFound` | That thing doesn't exist. |
| 409 | `http.StatusConflict` | That clashes with something that exists. |

### Writing JSON

In the HTTP servers exercises you learned the order: set headers, then write
the status code, then write the body. Headers set after `WriteHeader` are
ignored. `json.NewEncoder(w).Encode(v)` turns any value into JSON and writes it
straight into the response. It adds a newline (`"\n"`) at the end.

```go
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusOK)
json.NewEncoder(w).Encode(svc)
// body: {"name":"api","port":8080,"owner":"platform"}\n
```

The field names come from the struct tags on `Service`, as in the JSON
encoding exercises.

A parameter of type `any` accepts a value of any type. That's how one helper
can send a `Service`, a map, or anything else. A `map[string]string` with one
entry encodes as a one-key object:

```go
json.NewEncoder(w).Encode(map[string]string{"error": "boom"})
// body: {"error":"boom"}\n
```

### Reading JSON safely

A request body is a stream. Anyone can send you anything, so read it
defensively:

```go
r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // read at most 1 MiB
dec := json.NewDecoder(r.Body)
dec.DisallowUnknownFields()                   // a typo is an error

var svc Service
if err := dec.Decode(&svc); err != nil {
	// broken JSON, an unknown field, or a body over 1 MiB
}
```

- `http.MaxBytesReader` wraps the body so that reading past the limit fails.
  Without it, a client could send gigabytes and fill your memory. `1<<20` is
  1,048,576 bytes, or 1 MiB. (Some APIs answer an oversized body with 413
  Content Too Large. This one keeps it simple and uses 400.)
- `DisallowUnknownFields` makes `Decode` fail when the JSON has a field the
  struct doesn't. Without it, `{"prot":8080}` would be accepted silently, and
  the port would be `0`.

After writing an error response, `return`. Otherwise the handler keeps going
and writes a second response on top of the first.

### The files

| File | What it is |
|---|---|
| `exercise.go` | The routes and the handlers. You work here. |
| `store.go` | `Service`, `Validate` and the in-memory `Store`. Already done. |

Read `store.go` once. `Store.Get` fails only with `ErrNotFound`, and
`Store.Create` fails only with `ErrExists`. So in your handlers, any error from
`Get` means "not found" and any error from `Create` means "already exists".

## Your task

Your team runs a service catalog. It answers `GET /services/{name}` and
`POST /services`, but the handlers only reply "not implemented". The routes in
`Routes` are already wired up.

### 1. Write the response helpers

`writeJSON(w, status, v)` must:

- set `Content-Type` to `application/json`
- write `status`
- encode `v` as the body with `json.NewEncoder`

`writeError(w, status, msg)` sends `{"error": msg}` by calling `writeJSON` with
a `map[string]string`.

```go
writeJSON(w, http.StatusOK, Service{Name: "api", Port: 8080, Owner: "platform"})
// 200, body: {"name":"api","port":8080,"owner":"platform"}\n

writeError(w, http.StatusNotFound, "service not found: web")
// 404, body: {"error":"service not found: web"}\n
```

Every response in tasks 2 and 3 goes through these two helpers, so do this
task first.

### 2. Get one service

`handleGet` serves `GET /services/{name}`. Read the name with
`r.PathValue("name")` and look it up with `a.store.Get`.

- Found: 200 and the service as JSON.
- Not found: 404 and the error `service not found: <name>`.

```text
GET /services/payments
// 200, body: {"name":"payments","port":8443,"owner":"billing"}\n

GET /services/web
// 404, body: {"error":"service not found: web"}\n
```

### 3. Create a service

`handleCreate` serves `POST /services`. Do these steps in order, and stop at
the first one that fails:

1. Limit the body to 1 MiB and decode it into a `Service`, with unknown
   fields disallowed. If decoding fails: 400 and the error
   `invalid request body`.
2. Call `svc.Validate()`. If it returns an error: 400, with the error's text
   (`err.Error()`) as the message.
3. Save it with `a.store.Create`. If that fails: 409 and the error
   `service already exists: <name>`.
4. Otherwise: 201 and the saved service as JSON.

```text
POST /services  {"name":"api","port":8080,"owner":"platform"}
// 201, body: {"name":"api","port":8080,"owner":"platform"}\n

POST /services  {"name":"api","prot":8080}
// 400, body: {"error":"invalid request body"}\n

POST /services  {"name":"api","port":70000}
// 400, body: {"error":"port must be between 1 and 65535"}\n

POST /services  {"name":"payments","port":9000}   (payments already exists)
// 409, body: {"error":"service already exists: payments"}\n
```

## Learn more

- [encoding/json: Encoder and Decoder](https://pkg.go.dev/encoding/json#Decoder)
- [net/http: MaxBytesReader](https://pkg.go.dev/net/http#MaxBytesReader)
- [Go by Example: JSON](https://gobyexample.com/json)
- [net/http: status code constants](https://pkg.go.dev/net/http#pkg-constants)
