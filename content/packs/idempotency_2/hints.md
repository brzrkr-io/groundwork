## Task 1: Scope keys per client

`storeKey` already joins method, path and key. It only needs the client ID
in front, read from a request header.

### Closer

Read the client with `r.Header.Get("X-Client-ID")` and join the parts with
`+`, using `"|"` between the three sections.

### Answer

```go
return r.Header.Get("X-Client-ID") + "|" + r.Method + " " + r.URL.Path + "|" + key
```

## Task 2: Expire old entries

Work out how old the entry is and compare that with the TTL. An expired
entry should look exactly like a missing one.

### Closer

`s.now().Sub(e.Saved)` is a `time.Duration`: the entry's age. If it is
`>= s.ttl`, the entry has expired. Combine that with `!ok` in one `if`.

### Answer

```go
e, ok := s.entries[key]
if !ok || s.now().Sub(e.Saved) >= s.ttl {
	return Entry{}, false
}
return e, true
```

## Task 3: Reject a key reused with a different body

Write `Fingerprint` first. Then the middleware has to read the body, put it
back for `next`, and compare fingerprints before replaying.

### Closer

`sha256.Sum256(body)` returns a `[32]byte`; pass `sum[:]` to
`hex.EncodeToString`. Read the body with `io.ReadAll(r.Body)` and restore it
with `r.Body = io.NopCloser(bytes.NewReader(body))`. Inside the `store.Get`
branch, check `e.Fingerprint != fingerprint` before replaying.

### Answer

```go
// Fingerprint
sum := sha256.Sum256(body)
return hex.EncodeToString(sum[:])

// Middleware, after the "no key" check
body, err := io.ReadAll(r.Body)
if err != nil {
	http.Error(w, "cannot read request body", http.StatusBadRequest)
	return
}
r.Body = io.NopCloser(bytes.NewReader(body))
fingerprint := Fingerprint(body)

sk := storeKey(r, key)
if e, ok := store.Get(sk); ok {
	if e.Fingerprint != fingerprint {
		http.Error(w, "Idempotency-Key reused with a different request body", http.StatusUnprocessableEntity)
		return
	}
	w.Header().Set("Idempotent-Replayed", "true")
	writeResponse(w, e.Response)
	return
}
resp := record(next, r)
store.Put(sk, Entry{Response: resp, Fingerprint: fingerprint})
writeResponse(w, resp)
```

If the handler's reply says `created from ` with nothing after it, the body
wasn't put back on `r.Body`.
