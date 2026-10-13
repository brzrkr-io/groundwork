## Only touch the innermost function

In all three functions, your code goes inside
`http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { ... })`,
next to the `next.ServeHTTP(w, r)` line. The outer layers are done.

## Task 1: one Fprintf

Add `"fmt"` to the imports, then before `next.ServeHTTP`:

```go
fmt.Fprintf(out, "%s %s\n", r.Method, r.URL.Path)
```

## Task 2: read, maybe make, then set

```go
id := r.Header.Get("X-Request-ID")
if id == "" {
	id = newID()
	r.Header.Set("X-Request-ID", id)
}
w.Header().Set("X-Request-ID", id)
```

## Task 3: defer before next

The deferred function must be set up **before** `next.ServeHTTP` runs:

```go
defer func() {
	if v := recover(); v != nil {
		fmt.Fprintf(out, "panic: %v\n", v)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}()
next.ServeHTTP(w, r)
```
