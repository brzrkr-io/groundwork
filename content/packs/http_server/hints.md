## Task 1: Answer health checks

Set the header first, then write the body; headers set after the body are
lost. `w` is an `io.Writer`, so the `fmt` printing functions can write to it.

### Closer

Use `w.Header().Set("Content-Type", ...)`, then `fmt.Fprint(w, ...)`. Add
`"fmt"` to the imports.

### Answer

```go
w.Header().Set("Content-Type", "text/plain; charset=utf-8")
fmt.Fprint(w, "ok\n")
```

## Task 2: Look up a service

Read the name from the path, then look it up in the map with the comma-ok
form. Handle the missing case first and return.

### Closer

`name := r.PathValue("name")` and `version, ok := c[name]`. For a missing
name, call `http.Error(w, "unknown service: "+name, http.StatusNotFound)` and
`return`. Otherwise set `X-Service-Version` and write with `fmt.Fprintf`.

### Answer

```go
name := r.PathValue("name")
version, ok := c[name]
if !ok {
	http.Error(w, "unknown service: "+name, http.StatusNotFound)
	return
}
w.Header().Set("X-Service-Version", version)
fmt.Fprintf(w, "%s %s\n", name, version)
```

## Task 3: Wire up the routes

Register one pattern per handler on the mux. The mux produces the 404 and 405
answers on its own.

### Closer

Call `mux.HandleFunc(pattern, handler)` twice. The method goes in front of
the path in the pattern, and a method value like `c.ServeService` works as a
handler.

### Answer

```go
mux.HandleFunc("GET /healthz", Healthz)
mux.HandleFunc("GET /services/{name}", c.ServeService)
```
