## Task 1: one line

`storeKey` only needs the client in front. Read it with
`r.Header.Get("X-Client-ID")` and join the parts with `+`.

## Task 2: compare durations

`s.now().Sub(e.Saved)` is how old the entry is. If it is `>= s.ttl`, the entry
has expired. Combine it with the "missing" case:

```go
if !ok || s.now().Sub(e.Saved) >= s.ttl {
	return Entry{}, false
}
return e, true
```

## Task 3: Fingerprint

```go
sum := sha256.Sum256(body)
return hex.EncodeToString(sum[:])
```

## Task 3: the middleware

```go
body, err := io.ReadAll(r.Body)
if err != nil {
	http.Error(w, "cannot read request body", http.StatusBadRequest)
	return
}
r.Body = io.NopCloser(bytes.NewReader(body))
fingerprint := Fingerprint(body)
```

Inside the `store.Get` branch, compare `e.Fingerprint != fingerprint` before
replaying, and use `Entry{Response: resp, Fingerprint: fingerprint}` in `Put`.
If the handler's reply says `created from ` with nothing after it, you forgot
to put the body back.
