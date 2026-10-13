## Headers first, body last

Every handler follows the same order: set headers with `w.Header().Set`,
optionally set a status with `w.WriteHeader` (or `http.Error`), then write the
body with `fmt.Fprint` or `fmt.Fprintf`. Headers set after the body are lost.

## Task 1: the health check

```go
w.Header().Set("Content-Type", "text/plain; charset=utf-8")
fmt.Fprint(w, "ok\n")
```

Remember to add `"fmt"` to the imports.

## Task 2: the map lookup

Use the comma-ok form, as with any map:

```go
name := r.PathValue("name")
version, ok := c[name]
if !ok {
	http.Error(w, "unknown service: "+name, http.StatusNotFound)
	return
}
```

Then set `X-Service-Version` and write the body with
`fmt.Fprintf(w, "%s %s\n", name, version)`.

## Task 3: register the patterns

```go
mux.HandleFunc("GET /healthz", Healthz)
mux.HandleFunc("GET /services/{name}", c.ServeService)
```

The mux produces the 404 and 405 answers on its own.
