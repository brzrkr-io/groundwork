# HTTP servers: handlers and routes

Nearly every service you run in production speaks HTTP: the load balancer
calls its health check, Prometheus scrapes its metrics, and other services
call its API. In this exercise you build a small service catalog server with
nothing but Go's standard library.

## Learn

### Handlers

A web server receives *requests* and sends back *responses*. In Go, the code
that answers a request is a *handler*. The simplest handler is a function with
exactly this shape:

```go
func Hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "hello\n")
}
```

- `r *http.Request` is what the client sent: `r.Method` (`"GET"`, `"POST"`),
  `r.URL.Path` (`"/healthz"`), the headers and the body.
- `w http.ResponseWriter` is where you write your answer. It is an
  `io.Writer`, so `fmt.Fprint(w, ...)` writes the response body, the same way
  `fmt.Fprintf(conn, ...)` wrote to a connection in
  *TCP: a line-based control protocol*.

Officially, a handler is anything with a `ServeHTTP(w, r)` method: the
`http.Handler` interface. The type `http.HandlerFunc` turns a plain function
with the shape above into an `http.Handler`. You won't need to write it
yourself here, because `mux.HandleFunc` (below) does the conversion for you.

### Headers, then status, then body

A response has three parts, and you must set them in this order:

| Step | Code | If you skip it |
|---|---|---|
| 1. Headers | `w.Header().Set("X-Region", "eu-west-1")` | no extra headers |
| 2. Status code | `w.WriteHeader(http.StatusNotFound)` | 200 OK |
| 3. Body | `fmt.Fprint(w, "...")` | empty body |

The status line and headers travel before the body, so once you write the
status or the first byte of the body, later header changes are ignored. If you
never set `Content-Type`, Go guesses one from the first bytes of the body.
Status codes have named constants: `http.StatusOK` (200),
`http.StatusNotFound` (404), `http.StatusMethodNotAllowed` (405).

For an error reply, `http.Error` does all three steps in one call. It sets
`Content-Type: text/plain; charset=utf-8`, writes the status, and writes your
message followed by a newline:

```go
http.Error(w, "unknown service: billing", http.StatusNotFound)
// status 404, body "unknown service: billing\n"
```

Call `return` right after it, so the handler doesn't write anything else.

### Routing with a ServeMux

A *router* picks which handler answers each request. Go's router is the
*ServeMux*. You register *patterns* made of a method and a path:

```go
mux := http.NewServeMux()
mux.HandleFunc("GET /healthz", Healthz)
mux.HandleFunc("GET /services/{name}", catalog.ServeService)
```

- `"GET /healthz"` matches GET requests (and HEAD, GET's body-less twin) for
  exactly `/healthz`.
- `{name}` is a *wildcard* that matches one path segment. Inside the handler,
  `r.PathValue("name")` returns it: `"payments"` for `/services/payments`.
- `catalog.ServeService` is a method used as a function value. It remembers
  which `catalog` to use, like a closure.

The mux also answers some requests for you:

| Request | Answer |
|---|---|
| No pattern matches the path (`GET /nope`) | 404 Not Found |
| The path matches, the method doesn't (`POST /healthz`) | 405 Method Not Allowed, with an `Allow: GET, HEAD` header |

A real program would then start serving with
`http.ListenAndServe(":8080", mux)`.

### How the tests call your handlers

The tests don't open a network port. They build a fake request with
`httptest.NewRequest` and a *ResponseRecorder* with `httptest.NewRecorder`,
then call the handler directly. The recorder remembers everything the
handler wrote:

```go
req := httptest.NewRequest("GET", "/healthz", nil)
rec := httptest.NewRecorder()
Healthz(rec, req)
// rec.Code is the status, rec.Header() the headers, rec.Body.String() the body
```

The tests for task 2 call `req.SetPathValue("name", "payments")`, which fills
in the wildcard the way the mux would, so task 2 works before you write task
3. *Testing HTTP: recorders and fake servers* covers these tools properly.

## Your task

Your platform team wants a tiny server that says which version of each
service is deployed. `Catalog` maps service names to versions, for example
`Catalog{"payments": "v1.4.2"}`.

### 1. Answer health checks

Change `Healthz` so it sets the header `Content-Type` to
`text/plain; charset=utf-8` and writes the body `ok\n` (`ok` and a newline).
The status stays at the default 200. Add the `"fmt"` import.

```text
GET /healthz
=> 200, Content-Type: text/plain; charset=utf-8, body "ok\n"
```

### 2. Look up a service

Change `ServeService` so it reads the service name with `r.PathValue("name")`
and looks it up in the catalog `c`.

- Unknown name: reply with `http.Error`, status 404 and the message
  `unknown service: <name>`.
- Known name: set the header `X-Service-Version` to the version, then write
  `<name> <version>\n`. The status is the default 200.

```text
GET /services/payments   (catalog has "payments": "v1.4.2")
=> 200, X-Service-Version: v1.4.2, body "payments v1.4.2\n"

GET /services/billing    (not in the catalog)
=> 404, body "unknown service: billing\n"
```

### 3. Wire up the routes

Change `NewMux` so it registers `Healthz` for `GET /healthz` and
`c.ServeService` for `GET /services/{name}`. The tests send requests through
the mux and check the status codes:

| Request | Status |
|---|---|
| `GET /healthz` | 200 |
| `GET /services/payments` | 200 |
| `POST /healthz` | 405, and the `Allow` header contains `GET` |
| `DELETE /services/payments` | 405, and the `Allow` header contains `GET` |
| `GET /nope` | 404 |

## Learn more

- [Package net/http](https://pkg.go.dev/net/http)
- [Routing enhancements for Go 1.22 (Go blog)](https://go.dev/blog/routing-enhancements)
- [Go by Example: HTTP Server](https://gobyexample.com/http-server)
