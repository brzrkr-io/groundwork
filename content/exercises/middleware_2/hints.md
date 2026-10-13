## Task 1: loop backwards

Each `h = mw(h)` puts a new layer **outside** the previous ones, so the last
one wrapped ends up outermost. Range over `slices.Backward(mws)` instead of
`mws`, and add the `slices` import:

```go
for _, mw := range slices.Backward(mws) {
	h = mw(h)
}
```

## Task 2: a two-line method

Use a pointer receiver, so saving the code changes the recorder the
middleware reads afterwards:

```go
func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}
```

## Task 3: build what you expect, then compare

```go
want := "Bearer " + token
got := r.Header.Get("Authorization")
if subtle.ConstantTimeCompare([]byte(got), []byte(want)) != 1 {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	fmt.Fprintln(w, `{"error":"unauthorized"}`)
	return
}
next.ServeHTTP(w, r)
```
