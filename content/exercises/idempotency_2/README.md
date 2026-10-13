# Idempotency keys: scope, expire and fingerprint

A shared API has many clients, keys pile up forever, and sometimes a client
reuses a key by mistake. Here you harden idempotency keys for real traffic.

## Learn

### Recap

As in *Idempotency keys: replay retried requests*, a client sends the same `Idempotency-Key` header
with every retry of one operation. The first request runs the handler and the
server saves the response. Repeats get the saved response with
`Idempotent-Replayed: true` and the handler doesn't run again. This pack
leaves out the in-progress check so you can focus on three new problems.

### Scope keys per client

Keys are made up by clients, so two clients can pick the same one. If the
store used the raw key, client B would get client A's response. Real servers
build the store key from who is asking and what they are calling:

```text
<client ID> | <method> <path> | <key>
ci-payments|POST /vms|abc
```

Here the client ID comes from the `X-Client-ID` header. (A real server would
take it from the authenticated user instead.)

### Expire old keys

A retry happens seconds or minutes after the first attempt, not weeks later.
Stores keep entries for a *TTL* (time to live), often 24 hours, and then
forget them, which keeps memory bounded. The `Store` takes a `now func()
time.Time` so tests can move a fake clock. `t.Sub(u)` returns the
`time.Duration` between two times.

### Fingerprint the request body

What if a client reuses key `abc` for a *different* request, say a VM called
`web-2` instead of `web-1`? Replaying the `web-1` response would be a lie. So
the server stores a *fingerprint* of the request body and compares it on
every repeat. A mismatch gets `422 Unprocessable Entity`, the status the IETF
draft suggests.

A good fingerprint is a hash. `crypto/sha256` turns any bytes into a 32-byte
array, and `encoding/hex` turns that into 64 hex characters:

```go
sum := sha256.Sum256([]byte("hello")) // sum is a [32]byte array
hex.EncodeToString(sum[:])            // sum[:] makes a slice of the whole array
// => "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
```

To hash the body, the middleware reads it with `io.ReadAll(r.Body)`. A body
can only be read once, so `next` would then see an empty body. Put a fresh
reader over the same bytes back on the request before calling `next`:

```go
body, err := io.ReadAll(r.Body)
// ...handle err...
r.Body = io.NopCloser(bytes.NewReader(body))
```

`bytes.NewReader` reads from a byte slice. `io.NopCloser` adds the `Close`
method that `r.Body` needs.

## Your task

A shared VM provisioning API (`POST /vms`) is called by CI pipelines from
several teams. Make its idempotency keys safe to share.

### 1. Scope keys per client

`storeKey` ignores the client, so two teams sending the same
`Idempotency-Key` get each other's VM. Fix it to return exactly the
`X-Client-ID` header, `|`, the method, a space, the URL path, `|`, and the key:

```go
// POST /vms with X-Client-ID: ci-payments
storeKey(r, "abc")  // => "ci-payments|POST /vms|abc"

// DELETE /vms/web-1 with X-Client-ID: ci-search
storeKey(r, "k-17") // => "ci-search|DELETE /vms/web-1|k-17"
```

### 2. Expire old entries

Finish `Store.Get`. An entry has expired once `s.ttl` or more has passed
since `e.Saved`. Return `Entry{}` and `false` for an expired entry, just like
a missing one. With a 24-hour TTL:

| Time since Put | `Get` returns ok |
|---|---|
| 23h59m | `true` |
| 24h | `false` |
| 72h | `false` |

After expiry the middleware runs the handler again and saves the new
response.

### 3. Reject a key reused with a different body

First write `Fingerprint`: the SHA-256 of `body` as lowercase hex.

```go
Fingerprint([]byte(""))
// => "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
Fingerprint([]byte(`{"name":"web-1"}`))
// => "c170d260c5801ebdc0dc96b0bce790fa3fb88bc151f3d6efc215e57fd008eff0"
```

Then update `Middleware` for requests that have a key:

- Read the body with `io.ReadAll`. If that fails, respond with
  `http.Error(w, "cannot read request body", http.StatusBadRequest)` and
  return. Put the body back on `r.Body` and compute its fingerprint.
- When `store.Get` finds an entry whose `Fingerprint` differs, respond with
  this and return, without replaying:

  ```go
  http.Error(w, "Idempotency-Key reused with a different request body", http.StatusUnprocessableEntity)
  ```

- Save the fingerprint in the `Entry` you `Put`.

Add `crypto/sha256`, `encoding/hex` and `io` to the imports.

```text
key abc  body {"name":"web-1"}  => 201 vm-1 created from {"name":"web-1"}
key abc  body {"name":"web-1"}  => 201 vm-1 created from {"name":"web-1"}  (replay)
key abc  body {"name":"web-2"}  => 422 Idempotency-Key reused with a different request body
key def  body {"name":"web-2"}  => 201 vm-2 created from {"name":"web-2"}
```

## Learn more

- [IETF draft: The Idempotency-Key HTTP Header Field](https://datatracker.ietf.org/doc/draft-ietf-httpapi-idempotency-key-header/)
- [crypto/sha256](https://pkg.go.dev/crypto/sha256)
- [io.NopCloser](https://pkg.go.dev/io#NopCloser)
- [time.Time.Sub](https://pkg.go.dev/time#Time.Sub)
