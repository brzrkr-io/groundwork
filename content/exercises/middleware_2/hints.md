## Task 1: Fix the chain order

Each `h = mw(h)` puts a new layer **outside** the previous ones, so whatever
you wrap last ends up outermost. Which end of `mws` should you start from?

### Closer

Range over `slices.Backward(mws)` instead of `mws`, and add the `slices`
import. The loop body stays the same.

### Answer

```go
for _, mw := range slices.Backward(mws) {
	h = mw(h)
}
return h
```

## Task 2: Log the real status code

Declare your own `WriteHeader` method on the recorder. It replaces the promoted
one, so it must save the code and then pass it on to the embedded writer.

### Closer

Use a pointer receiver, `func (s *statusRecorder) WriteHeader(code int)`, so
saving into `s.status` changes the recorder `AccessLog` reads afterwards. The
real call is `s.ResponseWriter.WriteHeader(code)`.

### Answer

```go
func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}
```

## Task 3: Require a token

Build the header value you expect (`"Bearer "` plus the token), compare it with
the one the request sent, and stop before `next` when they differ.

### Closer

`subtle.ConstantTimeCompare([]byte(got), []byte(want))` returns `1` when equal.
On a mismatch, set `Content-Type`, call `w.WriteHeader(http.StatusUnauthorized)`,
write the body with `fmt.Fprintln`, then `return`. Add the `crypto/subtle`
import.

### Answer

```go
want := []byte("Bearer " + token)
got := []byte(r.Header.Get("Authorization"))
if subtle.ConstantTimeCompare(got, want) != 1 {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	fmt.Fprintln(w, `{"error":"unauthorized"}`)
	return
}
next.ServeHTTP(w, r)
```
