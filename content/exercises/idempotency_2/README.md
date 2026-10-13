# Idempotency keys: scope, expire and fingerprint

A shared API has many clients, keys pile up forever, and sometimes a client
reuses a key by mistake. Here you harden idempotency keys for real traffic.

**New here:** scoped store keys, `time.Time.Sub`, SHA-256 fingerprints with `crypto/sha256` and `encoding/hex`, re-reading a request body.

## Learn

### Recap

As in *Idempotency keys: replay retried requests*, a client sends the same
`Idempotency-Key` header with every retry. The first request runs the
handler and the server saves the response. Repeats get the saved response
with `Idempotent-Replayed: true`. This pack leaves out the in-progress check.

### Scope keys per client

Clients make up keys, so two clients can pick the same one. With the raw key,
client B would get client A's response. So the store key is built from **who**
is asking, **what** they call, and the key, joined by `|`:

```text
<client ID>|<method> <path>|<key>
```

The client ID comes from the `X-Client-ID` header. Note the single space
between method and path.

### Expire old keys

Retries happen within minutes, not weeks. Stores keep entries for a *TTL*,
often 24 hours. `t.Sub(u)` returns the `time.Duration` from `u` to `t`, and
durations compare with `<`, `>=` and friends:

```go
age := now.Sub(lastBackup)
if age > 6*time.Hour { /* backup is overdue */ }
```

### Fingerprint the request body

What if a client reuses key `abc` for a *different* body? Replaying the old
response would be a lie. The server stores a *fingerprint* (a hash) of the
body and compares it on every repeat. A mismatch gets
`422 Unprocessable Entity`.

`sha256.Sum256` turns bytes into a `[32]byte` **array**. `hex.EncodeToString`
wants a **slice**, and `arr[:]` makes a slice of a whole array:

```go
digits := [3]int{1, 2, 3}
s := digits[:] // []int{1, 2, 3}, sharing the array's memory
```

> **Why is it written like this?** Go requires the conversion: an array and
> a slice are different types, and `[:]` is how you get one from the other.
> Lowercase hex is what `encoding/hex` produces by convention.

### Reading a body twice

`io.ReadAll(r.Body)` reads the whole body into a `[]byte`. A body can only be
read once, so `next` would then see nothing. Put a new reader over the same
bytes back on `r.Body`:

| Piece | What it does |
|---|---|
| `bytes.NewReader(b)` | an `io.Reader` that reads from the slice `b` |
| `io.NopCloser(rd)` | wraps a reader with the do-nothing `Close` method `r.Body` needs |

> **Why is it written like this?** `r.Body` has type `io.ReadCloser`, so Go
> requires both `Read` and `Close`; `NopCloser` supplies `Close`.

## Your task

A shared VM provisioning API (`POST /vms`) is called by CI pipelines from
several teams. Make its idempotency keys safe to share.

### 1. Scope keys per client

**Goal:** make `storeKey` return client ID, method, path and key joined in the exact format below.

The pieces, in order: the `X-Client-ID` request header, `|`, the method, a
space, the URL path, `|`, the `key` parameter.

```go
// POST /vms with X-Client-ID: ci-payments
storeKey(r, "abc")  // => "ci-payments|POST /vms|abc"

// DELETE /vms/web-1 with X-Client-ID: ci-search
storeKey(r, "k-17") // => "ci-search|DELETE /vms/web-1|k-17"
```

**Think it through:**
- Which pieces come from `r` (header, method, path) and which is the parameter?
- The fixed text is `"|"` and `" "`. Did you include the space?

### 2. Expire old entries

**Goal:** make `Store.Get` treat an entry as missing once `s.ttl` or more has passed since it was saved.

Return `Entry{}` and `false` for an expired entry, just like a missing one.
With a 24-hour TTL:

| Time since Put | `Get` returns ok |
|---|---|
| 23h59m | `true` |
| 24h | `false` |
| 72h | `false` |

After expiry the middleware runs the handler again and saves the new response.

**Think it through:**
- `e.Saved` holds when it was saved; `s.now()` gives the current time. Which one goes first in `Sub`?
- Exactly 24h counts as expired: `>` or `>=`?

### 3. Reject a key reused with a different body

**Goal:** write `Fingerprint`, then make `Middleware` answer `422` when a key comes back with a different body.

`Fingerprint(body)` returns the SHA-256 of `body` as 64 lowercase hex characters:

```go
Fingerprint([]byte(""))
// => "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
Fingerprint([]byte(`{"name":"web-1"}`))
// => "c170d260c5801ebdc0dc96b0bce790fa3fb88bc151f3d6efc215e57fd008eff0"
```

Then, in `Middleware`, for requests that have a key:

- Read the body with `io.ReadAll`. If that fails, use `http.Error` with
  status `http.StatusBadRequest` and the message `cannot read request body`,
  then return.
- Put the body back on `r.Body`, and fingerprint it.
- If `store.Get` finds an entry whose `Fingerprint` differs, use `http.Error`
  with status `http.StatusUnprocessableEntity` and the message
  `Idempotency-Key reused with a different request body`, then return
  without replaying.
- Save the fingerprint in the `Entry` you `Put`.

Add `bytes`, `crypto/sha256`, `encoding/hex` and `io` to the imports if they aren't there.

```text
key abc  body {"name":"web-1"}  => 201 vm-1 created from {"name":"web-1"}
key abc  body {"name":"web-1"}  => 201 vm-1 created from {"name":"web-1"}  (replay)
key abc  body {"name":"web-2"}  => 422 Idempotency-Key reused with a different request body
key def  body {"name":"web-2"}  => 201 vm-2 created from {"name":"web-2"}
```

**Think it through:**
- The body must be read *before* `store.Get`, because you compare fingerprints there.
- Where in `Middleware` does the replay happen? The mismatch check goes just before it.

## Learn more

- [IETF draft: The Idempotency-Key HTTP Header Field](https://datatracker.ietf.org/doc/draft-ietf-httpapi-idempotency-key-header/)
- [crypto/sha256](https://pkg.go.dev/crypto/sha256)
- [io.NopCloser](https://pkg.go.dev/io#NopCloser)
- [time.Time.Sub](https://pkg.go.dev/time#Time.Sub)
