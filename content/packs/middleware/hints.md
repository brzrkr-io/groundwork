## Task 1: Log every request

Your code goes in the innermost function, before `next.ServeHTTP(w, r)`. The
method and path are both on the request `r`.

### Closer

`fmt.Fprintf(out, ...)` writes formatted text to any `io.Writer`. Use
`r.Method` and `r.URL.Path` with a `"%s %s\n"` format. Add the `"fmt"` import.

### Answer

```go
fmt.Fprintf(out, "%s %s\n", r.Method, r.URL.Path)
next.ServeHTTP(w, r)
```

## Task 2: Tag every request with an ID

Read the incoming header first. Only when it's empty do you make a new ID and
put it on the request; either way the response gets the same ID.

### Closer

`r.Header.Get("X-Request-ID")` returns `""` when the header is missing. Set it
on the request with `r.Header.Set(...)` and on the response with
`w.Header().Set(...)`, both before `next.ServeHTTP`.

### Answer

```go
id := r.Header.Get("X-Request-ID")
if id == "" {
	id = newID()
	r.Header.Set("X-Request-ID", id)
}
w.Header().Set("X-Request-ID", id)
next.ServeHTTP(w, r)
```

## Task 3: Recover from panics

A deferred function runs even when the code after it panics. Set it up
**before** `next.ServeHTTP` runs, and call `recover()` inside it.

### Closer

`defer func() { if v := recover(); v != nil { ... } }()`. Inside, log with
`fmt.Fprintf(out, "panic: %v\n", v)` and reply with `http.Error` and
`http.StatusInternalServerError`.

### Answer

```go
defer func() {
	if v := recover(); v != nil {
		fmt.Fprintf(out, "panic: %v\n", v)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}()
next.ServeHTTP(w, r)
```
