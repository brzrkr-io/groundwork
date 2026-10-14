# HTTP servers: handlers and routes

Nearly every service you run speaks HTTP: the load balancer calls its health
check, Prometheus scrapes its metrics, other services call its API. Here you
build a small service catalog server with only Go's standard library.

**New here:** handlers, `http.ResponseWriter`, `http.Error`, `ServeMux` routing, path wildcards.

## Learn

### Handlers

A server receives *requests* and sends back *responses*. The code that
answers a request is a *handler*. The simplest handler is a function with
exactly this shape:

```go
func Version(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "1.0\n")
}
```

- `r *http.Request` is what the client sent: `r.Method` (`"GET"`),
  `r.URL.Path` (`"/version"`), headers, body.
- `w http.ResponseWriter` is where you write the answer. It is an
  `io.Writer`, so `fmt.Fprint(w, ...)` writes the response body.

> **Why is it written like this?** Go requires the parameter *types*
> `http.ResponseWriter` then `*http.Request`, in that order, or the router
> won't accept the function. The names `w` and `r` are convention. The
> function name is your choice.

### Headers, then status, then body

A response has three parts, set in this order:

| Step | Example | If you skip it |
|---|---|---|
| 1. Headers | `w.Header().Set("X-Region", "eu-west-1")` | no extra headers |
| 2. Status code | `w.WriteHeader(http.StatusTeapot)` | 200 OK |
| 3. Body | `fmt.Fprint(w, "...")` | empty body |

Once the status or the first byte of body is written, later header changes
are ignored. Status constants: `http.StatusOK` (200), `http.StatusNotFound`
(404), `http.StatusMethodNotAllowed` (405).

For an error reply, `http.Error(w, message, code)` does all three steps: it
sets `Content-Type: text/plain; charset=utf-8`, writes the status, and writes
your message plus a newline.

```go
http.Error(w, "disk full on "+host, http.StatusInsufficientStorage)
// status 507, body "disk full on db1\n"
```

Call `return` right after it, so the handler writes nothing else.

### Looking up a map key, safely

```go
owner, ok := teams["search"]
if !ok {
	// "search" is not in the map
}
```

> **Why is it written like this?** The two-value form `v, ok := m[key]` is Go
> syntax; with one value you'd get `""` for a missing key and couldn't tell.
> The name `ok` is convention.

### Routing with a ServeMux

A *router* picks which handler answers each request. Go's router is the
*ServeMux*. You register a *pattern* (method, space, path) and a handler:

```go
mux := http.NewServeMux()
mux.HandleFunc("GET /version", Version)
mux.HandleFunc("GET /teams/{team}/owner", t.ServeOwner)
```

- `"GET /version"` matches GET (and HEAD) requests for exactly `/version`.
- `{team}` is a *wildcard* for one path segment. In the handler,
  `r.PathValue("team")` returns it, e.g. `"search"` for `/teams/search/owner`.
- Pass the handler's **name** without `()`. With `()` you would call it now
  instead of handing it to the router.
- `t.ServeOwner` is a method used as a value; it remembers which `t` to use.

The mux also answers some requests itself:

| Request | Answer |
|---|---|
| No pattern matches the path | 404 Not Found |
| Path matches, method doesn't | 405 Method Not Allowed, with an `Allow: GET, HEAD` header |

A real program then starts with `http.ListenAndServe(":8080", mux)`.

### How the tests call your handlers

The tests don't open a port. They make a fake request and a *ResponseRecorder*,
then call the handler directly:

```go
req := httptest.NewRequest("GET", "/version", nil)
rec := httptest.NewRecorder()
Version(rec, req)
// rec.Code = status, rec.Header() = headers, rec.Body.String() = body
```

For task 2 the tests call `req.SetPathValue("name", "payments")`, which fills
in the wildcard like the mux would. So task 2 works before task 3.

## Your task

Your platform team wants a tiny server that says which version of each
service is deployed. `Catalog` maps service names to versions, for example
`Catalog{"payments": "v1.4.2"}`.

### 1. Answer health checks

**Goal:** `Healthz` sets `Content-Type` to `text/plain; charset=utf-8` and writes the body `ok\n`.

The body is `ok` followed by a newline. The status stays at the default 200.
Add the `"fmt"` import.

```text
GET /healthz
=> 200, Content-Type: text/plain; charset=utf-8, body "ok\n"
```

**Think it through:**
- Headers first, then body.
- Which part goes inside the quotes: header name, header value, body text?

### 2. Look up a service

**Goal:** `ServeService` reads the name from the path and replies with its version, or a 404.

- Name comes from `r.PathValue("name")`; look it up in the catalog `c`.
- **Unknown name:** reply with `http.Error`, status 404, message
  `unknown service: ` followed by the name.
- **Known name:** set header `X-Service-Version` to the version, then write
  the name, a space, the version, and a newline. Status stays 200.

```text
GET /services/payments   (catalog has "payments": "v1.4.2")
=> 200, X-Service-Version: v1.4.2, body "payments v1.4.2\n"

GET /services/billing    (not in the catalog)
=> 404, body "unknown service: billing\n"
```

**Think it through:**
- Which parts are fixed text and which come from variables?
- Don't forget the space between name and version.
- After the 404, stop the handler.

### 3. Wire up the routes

**Goal:** `NewMux` registers `Healthz` for `GET /healthz` and `c.ServeService` for `GET /services/{name}`.

The tests send requests through the mux:

| Request | Status |
|---|---|
| `GET /healthz` | 200 |
| `GET /services/payments` | 200 |
| `POST /healthz` | 405, and the `Allow` header contains `GET` |
| `DELETE /services/payments` | 405, and the `Allow` header contains `GET` |
| `GET /nope` | 404 |

**Think it through:**
- The pattern is one string; the handler is passed by name, no `()`.
- The 404 and 405 come free from the mux.

## Learn more

- [Package net/http](https://pkg.go.dev/net/http)
- [Routing enhancements for Go 1.22 (Go blog)](https://go.dev/blog/routing-enhancements)
- [Go by Example: HTTP Server](https://gobyexample.com/http-server)
