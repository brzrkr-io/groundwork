# Idempotency keys: replay retried requests

Clients retry. When a retried request starts a deploy or creates a VM, a
retry can do the work twice. *Idempotency keys* let a server spot a repeat
and answer it without doing the work again.

## Learn

### Why retries and POST don't mix

An operation is *idempotent* if doing it twice has the same effect as doing it
once. `GET /deploys/42` and `DELETE /vms/web-1` are idempotent. `POST
/deploys` is not: every call starts another deploy.

Picture a CI job calling `POST /deploys`. The server starts the deploy, but
the response is lost on the network. The client sees a timeout and retries.
Now two deploys are running. With `POST /vms` you would get two VMs and two
bills.

### The Idempotency-Key header

The fix, made popular by Stripe and now an IETF draft, is a header. The client
makes up a unique key (often a UUID) for each *operation* and sends the same
key with every retry of it:

```text
POST /deploys
Idempotency-Key: 4f1c2a9e-deploy-api-v42
```

The server keeps a table from key to response:

1. **First request** with a key: run the handler, save its status, headers
   and body under the key, and send the response.
2. **Repeat** with the same key: don't run the handler. Send the saved
   response again. This pack also adds the header `Idempotent-Replayed: true`,
   so clients and logs can tell it was a replay.
3. **No key**: the client didn't ask for protection, so just run the handler.

### Duplicates that arrive while the first is still running

Suppose the deploy takes 30 seconds and the client retries after 5. Nothing is
saved yet, so step 1 would start a second deploy. The server must remember
which keys are *in progress* too. There are two common choices:

| Choice | What the duplicate gets |
|---|---|
| Reject it (this pack) | `409 Conflict` straight away. The client retries later and gets the replay. |
| Wait | The duplicate blocks until the first finishes, then gets the replay. Harder to build. |

Checking and marking must happen under one lock. If two requests could both
check "not in progress" before either marks it, both would run.

### Capturing a handler's response

To save a response, the middleware must see what the handler writes.
`http.ResponseWriter` is an interface with three methods: `Header`,
`WriteHeader` and `Write`. Any type with those methods can stand in for the
real one. The starter includes a `recorder` type that keeps everything in
memory, plus two helpers you can call:

```go
resp := record(next, r) // run next, capture its Response
writeResponse(w, resp)  // send a Response (headers, status, body) to w
```

`httptest.ResponseRecorder`, which you've used in handler tests, works the
same way, but it belongs in tests, so production code uses its own small
type.

### Reading the tests

`TestMiddlewareConflict` must run two requests at the same time without
guessing about scheduling. Its fake handler closes a `started` channel and
then blocks until the test closes `release`. The test sends the duplicate
only after `started` is closed, so the first request is certainly in
progress. `atomic.Int32` (from `sync/atomic`) is a counter that many
goroutines can update safely: `calls.Add(1)` adds one and returns the new
value, and `calls.Load()` reads it.

## Your task

Your team's deploy API (`POST /deploys`) is called by CI jobs that retry on
timeouts. Add idempotency keys so a retry never starts a second deploy.

### 1. Build the Store

Fill in `Get` and `Put` on `Store`. Lock `s.mu` in both, because many
requests use the store at once.

- `Put(key, resp)` saves `resp` under `key` in `s.responses`.
- `Get(key)` returns the saved response and `true`, or `Response{}` and
  `false` if there is none.

```go
s := NewStore()
s.Get("deploy-1") // => Response{}, false
s.Put("deploy-1", Response{Status: 202, Body: []byte("deploy d-1 queued")})
s.Get("deploy-1") // => Response{Status: 202, ...}, true
```

### 2. Replay repeated requests

Finish `Middleware`. Requests without an `Idempotency-Key` header already pass
straight through. For a request with a key:

- If `store.Get(key)` finds a response, set the header
  `Idempotent-Replayed: true` on `w`, send the stored response with
  `writeResponse`, and return. Don't call `next`.
- Otherwise the starter already runs `next` with `record` and sends the
  result with `writeResponse`. Save the result with `store.Put` before
  sending it.

The first response must not have the `Idempotent-Replayed` header. Status,
body and headers such as `Content-Type` are the same on every replay:

```text
POST /deploys  Idempotency-Key: abc  => 202 {"deploy":"d-1"}
POST /deploys  Idempotency-Key: abc  => 202 {"deploy":"d-1"}  Idempotent-Replayed: true
POST /deploys  Idempotency-Key: xyz  => 202 {"deploy":"d-2"}
POST /deploys  (no key)              => 202 {"deploy":"d-3"}
```

### 3. Reject duplicates while the first is running

Fill in `Begin(key)`. Under the lock:

- If `key` has a stored response in `s.responses`, or is marked in
  `s.inProgress`, return `false`.
- Otherwise set `s.inProgress[key] = true` and return `true`.

In `Put`, also remove the key from `s.inProgress` with `delete`, because the
request has finished.

In `Middleware`, after the `Get` check and before running `next`, call
`store.Begin(key)`. If it returns `false`, respond with status `409 Conflict`
and this exact message, then return:

```go
http.Error(w, "request with this Idempotency-Key is still in progress", http.StatusConflict)
```

(`http.Error` adds a newline at the end. The test ignores it.)

```go
s.Begin("vm-1") // => true   (claimed)
s.Begin("vm-1") // => false  (in progress)
s.Put("vm-1", Response{Status: 201})
s.Begin("vm-1") // => false  (finished: callers replay it instead)
```

## Learn more

- [IETF draft: The Idempotency-Key HTTP Header Field](https://datatracker.ietf.org/doc/draft-ietf-httpapi-idempotency-key-header/)
- [Stripe: Designing robust and predictable APIs with idempotency](https://stripe.com/blog/idempotency)
- [net/http: ResponseWriter](https://pkg.go.dev/net/http#ResponseWriter)
- [sync/atomic: Int32](https://pkg.go.dev/sync/atomic#Int32)
