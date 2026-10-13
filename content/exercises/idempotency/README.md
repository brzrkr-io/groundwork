# Idempotency keys: replay retried requests

Clients retry. When the retried request starts a deploy or creates a VM, the
work can happen twice. *Idempotency keys* let a server spot a repeat and
answer it without doing the work again.

**New here:** idempotency, the `Idempotency-Key` header, capturing a handler's response, `409 Conflict`.

## Learn

### Why retries and POST don't mix

An operation is *idempotent* if doing it twice has the same effect as doing
it once. `GET /deploys/42` and `DELETE /vms/web-1` are idempotent. `POST
/deploys` is not: every call starts another deploy.

A CI job calls `POST /deploys`. The server starts the deploy, but the
response is lost on the network. The client times out and retries. Now two
deploys are running.

### The Idempotency-Key header

The client makes up a unique key for each *operation* and sends the same key
with every retry of it:

```text
POST /deploys
Idempotency-Key: 4f1c2a9e-deploy-api-v42
```

The server keeps a table from key to saved response:

| Request | What the server does |
|---|---|
| No key | Just run the handler. |
| First time with a key | Run the handler, save its response under the key, send it. |
| Repeat of a key | Don't run the handler. Send the saved response, plus the header `Idempotent-Replayed: true`. |

### Duplicates that arrive while the first is still running

The deploy takes 30 seconds; the client retries after 5. Nothing is saved
yet, so the server would run it again. It must also remember which keys are
*in progress*. This pack rejects such a duplicate with `409 Conflict`; the
client retries later and gets the replay.

Checking "in progress?" and marking it must happen under **one** lock. If two
requests could both check before either marks, both would run.

> **Why is it written like this?** Go doesn't force you to lock, but the race
> detector (and real traffic) will catch you if you don't: maps are not safe
> for concurrent use. `mu.Lock()` followed by `defer mu.Unlock()` is the Go
> convention, as in *Mutexes*.

### Capturing a handler's response

`http.ResponseWriter` is an interface with three methods: `Header`,
`WriteHeader` and `Write`. Any type with those methods can stand in for the
real one. The starter has a `recorder` type that keeps everything in memory,
and two helpers:

```go
resp := record(handler, req) // run handler, return what it wrote as a Response
writeResponse(w, resp)       // send a Response (headers, status, body) to w
```

### Setting a header and sending an error

```go
w.Header().Set("Cache-Control", "no-store")             // name, then value
http.Error(w, "service is in maintenance", http.StatusServiceUnavailable)
```

`http.Error` writes the status code and the message (plus a newline) as the
body.

> **Why is it written like this?** Both header name and value are strings, so
> both go in quotes. Using `http.StatusConflict` instead of `409` is
> convention: the name says what it means.

### Reading the tests

`TestMiddlewareConflict` needs two requests at the same time. Its fake
handler closes a `started` channel and then blocks until the test closes
`release`, so the first request is certainly in progress when the duplicate
arrives. `atomic.Int32` is a counter many goroutines can update safely:
`calls.Add(1)`, `calls.Load()`.

## Your task

Your team's deploy API (`POST /deploys`) is called by CI jobs that retry on
timeouts. Add idempotency keys so a retry never starts a second deploy.

### 1. Build the Store

**Goal:** make `Put` save a response under a key and `Get` return it.

Lock `s.mu` in both methods.

- `Put(key, resp)` saves `resp` under `key` in `s.responses`.
- `Get(key)` returns the saved response and `true`, or `Response{}` and
  `false` if there is none.

```go
s := NewStore()
s.Get("deploy-1") // => Response{}, false
s.Put("deploy-1", Response{Status: 202, Body: []byte("deploy d-1 queued")})
s.Get("deploy-1") // => Response{Status: 202, ...}, true
```

**Think it through:**
- A map lookup already gives you a value and an "ok" bool. What does it give
  for a missing key?

### 2. Replay repeated requests

**Goal:** in `Middleware`, send the saved response for a repeated key, and save the response the first time.

Requests without an `Idempotency-Key` header already pass straight through.
For a request with a key:

- If `store.Get(key)` finds a response: set the header
  `Idempotent-Replayed` to `true` on `w`, send the stored response with
  `writeResponse`, and return. Don't call `next`.
- Otherwise the starter already runs `next` with `record` and sends the
  result. Save the result with `store.Put` **before** sending it.

The first response must not have `Idempotent-Replayed`. Status, body and
headers such as `Content-Type` are the same on every replay:

```text
POST /deploys  Idempotency-Key: abc  => 202 {"deploy":"d-1"}
POST /deploys  Idempotency-Key: abc  => 202 {"deploy":"d-1"}  Idempotent-Replayed: true
POST /deploys  Idempotency-Key: xyz  => 202 {"deploy":"d-2"}
POST /deploys  (no key)              => 202 {"deploy":"d-3"}
```

**Think it through:**
- The replay check goes after the "no key" check and before `record`.
- Set the replay header *before* `writeResponse`: headers can't change once
  the status is sent.

### 3. Reject duplicates while the first is running

**Goal:** make `Begin` claim a key once, and make `Middleware` answer `409` for a key that's already claimed.

`Begin(key)`, under the lock:

- If `key` has a stored response in `s.responses`, or is marked in
  `s.inProgress`, return `false`.
- Otherwise mark it in `s.inProgress` (value `true`) and return `true`.

In `Put`, also remove the key from `s.inProgress`: the request has finished.

In `Middleware`, after the `Get` check and before running `next`, call
`store.Begin(key)`. If it returns `false`, respond with `http.Error`, status
`http.StatusConflict` (409), and exactly this message, then return:

```text
request with this Idempotency-Key is still in progress
```

```go
s.Begin("vm-1") // => true   (claimed)
s.Begin("vm-1") // => false  (in progress)
s.Put("vm-1", Response{Status: 201})
s.Begin("vm-1") // => false  (finished: callers replay it instead)
```

**Think it through:**
- `Begin` checks two maps. Which built-in removes the key in `Put`?
- The message is fixed text, so it goes in quotes exactly as shown.

## Learn more

- [IETF draft: The Idempotency-Key HTTP Header Field](https://datatracker.ietf.org/doc/draft-ietf-httpapi-idempotency-key-header/)
- [Stripe: Designing robust and predictable APIs with idempotency](https://stripe.com/blog/idempotency)
- [net/http: ResponseWriter](https://pkg.go.dev/net/http#ResponseWriter)
- [sync/atomic: Int32](https://pkg.go.dev/sync/atomic#Int32)
