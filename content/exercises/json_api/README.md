# JSON APIs: build a service catalog

Most internal platform tools talk to each other through small JSON APIs: a
service catalog, a deploy API, an alert manager. Here you finish one that
sends and receives JSON the way clients expect.

**New here:** JSON responses with `json.NewEncoder`, the `any` type, safe
request reading with `http.MaxBytesReader` and `DisallowUnknownFields`.

## Learn

### What makes an API "JSON"

Every response does three things:

1. Sets the header `Content-Type: application/json`, so clients know how to
   read the body.
2. Picks a *status code* that says what happened.
3. Writes the body as JSON, **errors too**. Then a client never has to guess
   the format.

A common error shape is an object with one `error` key:

```json
{"error":"disk is full"}
```

| Code | Constant | Meaning |
|---|---|---|
| 200 | `http.StatusOK` | Here is what you asked for. |
| 201 | `http.StatusCreated` | I created the thing you sent. |
| 400 | `http.StatusBadRequest` | Your request is broken or invalid. |
| 404 | `http.StatusNotFound` | That thing doesn't exist. |
| 409 | `http.StatusConflict` | That clashes with something that exists. |

### Writing JSON

As in *HTTP servers*: set headers, then write the status code, then write the
body. Headers set after `WriteHeader` are ignored.

`json.NewEncoder(w).Encode(v)` turns `v` into JSON and writes it into `w`. It
adds a newline (`"\n"`) at the end.

```go
user := User{Login: "ana"}
json.NewEncoder(os.Stdout).Encode(user)
// prints: {"login":"ana"}\n
```

The JSON field names come from the struct tags, as in *Encoding JSON: a deployment status report*.

> **Why is it written like this?** `json.NewEncoder(w)` makes an encoder and
> `.Encode(v)` uses it straight away: chaining calls on one line is allowed by
> Go, and it's the author's choice here. The `\n` at the end is what `Encode`
> always does; you don't add it.

### The `any` type

A parameter of type `any` accepts a value of **any** type: a struct, a map, a
slice. That's how one helper can send a `Service` or an error object.

```go
func show(v any) { fmt.Println(v) }

show(42)
show(map[string]string{"zone": "eu-west-1"})
```

A `map[string]string` with one entry encodes as a one-key JSON object:
`map[string]string{"zone": "eu-west-1"}` becomes `{"zone":"eu-west-1"}`.

> **Why is it written like this?** `any` is a type built into Go, so you
> don't import it (Go requires that exact name). Inside a map literal, the key
> and value are joined by `:` (required); the key here is fixed text, so it goes
> in quotes.

### Reading JSON safely

A request body comes from anyone, so read it defensively. Here is a reader for
a different struct, `Job`:

```go
r.Body = http.MaxBytesReader(w, r.Body, 512) // read at most 512 bytes
dec := json.NewDecoder(r.Body)
dec.DisallowUnknownFields() // a field Job doesn't have is an error

var job Job
if err := dec.Decode(&job); err != nil {
	// broken JSON, an unknown field, or a body over the limit
}
```

- `http.MaxBytesReader` makes reading past the limit fail, so nobody can fill
  your memory. `1<<20` means 1,048,576 bytes (1 MiB).
- `DisallowUnknownFields` turns a typo like `{"prot":8080}` into an error.
  Without it, the typo is ignored and the port stays `0`.
- `Decode(&job)` needs the `&`: the decoder must change your variable, not a
  copy of it.

> **Why is it written like this?** `if err := ...; err != nil` declares `err`
> and checks it on one line (Go syntax; the community's usual style). The
> names `dec` and `job` are the author's choice.

After you write an error response, `return`. Otherwise the handler keeps
going and writes a second response on top of the first.

### The files

| File | What it is |
|---|---|
| `exercise.go` | The routes and the handlers. You work here. |
| `store.go` | `Service`, `Validate` and the in-memory `Store`. Already done. |

`Store.Get` fails only with `ErrNotFound`, and `Store.Create` fails only with
`ErrExists`. So any error from `Get` means "not found", and any error from
`Create` means "already exists".

## Your task

Your team's service catalog answers `GET /services/{name}` and
`POST /services`, but the handlers only reply "not implemented". The routes
in `Routes` are already wired up.

Every response must have `Content-Type: application/json`, the right status,
and the exact body shown.

### 1. Write the response helpers

**Goal:** finish `writeJSON` (send any value as JSON) and `writeError` (send
`{"error": msg}`).

`writeJSON(w, status, v)` sets `Content-Type` to `application/json`, writes
`status`, then encodes `v` with `json.NewEncoder`. Add the `"encoding/json"`
import.

`writeError(w, status, msg)` calls `writeJSON` with a `map[string]string`
holding one key, `"error"`.

```go
writeJSON(w, http.StatusOK, Service{Name: "api", Port: 8080, Owner: "platform"})
// 200, body: {"name":"api","port":8080,"owner":"platform"}\n

writeError(w, http.StatusNotFound, "service not found: web")
// 404, body: {"error":"service not found: web"}\n
```

**Think it through:**
- Which of the three steps must come first so the header isn't lost?
- In `writeError`, which part of the map is fixed text and which comes from a
  parameter?

### 2. Get one service

**Goal:** `handleGet` answers `GET /services/{name}` with the service, or 404.

Read the name with `r.PathValue("name")` and look it up with `a.store.Get`.

- Found: 200 and the service as JSON.
- Not found: 404 and the error `service not found: <name>`, where `<name>` is
  the name from the path.

```text
GET /services/payments
// 200, body: {"name":"payments","port":8443,"owner":"billing"}\n

GET /services/web
// 404, body: {"error":"service not found: web"}\n
```

**Think it through:**
- `Get` returns two values. Which one tells you it wasn't found?
- The message is fixed text plus the name. Mind the space after the colon.

### 3. Create a service

**Goal:** `handleCreate` reads a `Service` from the body, checks it, saves it
and answers 201.

Do these steps in order. Stop at the first one that fails:

1. Limit the body to 1 MiB and decode it into a `Service`, with unknown
   fields disallowed. Fails: 400, error `invalid request body`.
2. Call `svc.Validate()`. Fails: 400, with the error's own text
   (`err.Error()`) as the message.
3. Save it with `a.store.Create`. Fails: 409, error
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

A body over 1 MiB must also get 400 `invalid request body`.

**Think it through:**
- Reuse the shape from *Reading JSON safely*, but with your limit and type.
- After each error response, what stops the handler from going on?

## Learn more

- [encoding/json: Decoder](https://pkg.go.dev/encoding/json#Decoder)
- [net/http: MaxBytesReader](https://pkg.go.dev/net/http#MaxBytesReader)
- [Go by Example: JSON](https://gobyexample.com/json)
- [net/http: status code constants](https://pkg.go.dev/net/http#pkg-constants)
