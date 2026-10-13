# Middleware: logging, request IDs and panics

Every service needs the same chores on every request: log it, tag it with an
ID you can search for later, and survive a bug without crashing. *Middleware*
does those chores once, around all your handlers.

## Learn

### A handler that wraps a handler

Middleware is a function that takes a handler and returns a new handler. The
new handler does something extra, and calls the original one (called `next`)
to do the real work:

```go
// Middleware wraps a handler with extra behaviour.
type Middleware func(http.Handler) http.Handler
```

Here is a complete middleware that adds a header to every response:

```go
func AddVersion(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Version", "1.4.2") // before
		next.ServeHTTP(w, r)                 // the real handler
	})
}
```

- `http.HandlerFunc(func(w, r) {...})` turns an ordinary function into an
  `http.Handler`. It's a conversion, like `int(x)`, not a call.
- `next.ServeHTTP(w, r)` runs the wrapped handler. Code before it runs before
  the handler, code after it runs after. If you never call it, the handler
  never runs (that's how an auth check blocks a request).
- The inner function is a *closure*, as in the functions-as-values
  exercises: it can use `next` long after `AddVersion` has returned.

### Middleware with settings

Middleware often needs settings, such as where to write logs. Add one more
function on the outside. It takes the settings and returns a `Middleware`:

```go
func AddHeader(key, value string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set(key, value)
			next.ServeHTTP(w, r)
		})
	}
}
```

Read it from the outside in: `AddHeader("X-Version", "1.4.2")` gives you a
`Middleware`. Calling that with a handler gives you the wrapped handler:

```go
h := AddHeader("X-Version", "1.4.2")(mux)
```

All three functions in this exercise have exactly this shape, and the starter
already has the three layers. You only fill in the middle.

### Stacking middleware

You wrap a handler several times. The **last** wrapper you apply is the
**outermost** one, so it runs first:

```go
h := Recover(os.Stderr)(Logging(os.Stdout)(mux))
// a request goes: Recover -> Logging -> mux
```

Put `Recover` outermost, so it catches panics from everything inside it.

### Recovering from a panic

A panic in a handler is a bug, but one bad request shouldn't take down a whole
service. As in the panic and recover exercises, a deferred function can call
`recover()`. It returns the panic value, or `nil` when there was no panic:

```go
defer func() {
	if v := recover(); v != nil {
		// v is whatever was passed to panic: a string, an error, ...
	}
}()
```

`%v` formats any panic value for a log line.

## Your task

Your team's services need three standard middlewares. Each test wraps a small
handler with your middleware and checks the response.

### 1. Log every request

`Logging(out)` writes one line to `out` for every request: the method, a
space, the path, and a newline. Use `r.Method` and `r.URL.Path` (the path
without any `?query`). Then call the next handler as usual.

```text
GET /healthz        logs "GET /healthz\n"
POST /deploys       logs "POST /deploys\n"
GET /logs?limit=5   logs "GET /logs\n"
```

### 2. Tag every request with an ID

`RequestID(newID)` makes sure each request has an `X-Request-ID`:

- If the request already has an `X-Request-ID` header, keep it. Don't call
  `newID`.
- Otherwise call `newID()` once and set the result as the request's
  `X-Request-ID` header (`r.Header.Set`), so the next handler can read it.
- Either way, set the same ID as the response's `X-Request-ID` header, before
  calling the next handler.

```text
request without X-Request-ID, newID returns "req-7f3a"
// next handler sees "req-7f3a"; response header X-Request-ID: req-7f3a

request with X-Request-ID: abc-123
// next handler sees "abc-123"; response header X-Request-ID: abc-123
```

### 3. Recover from panics

`Recover(out)` catches a panic from the next handler. When one happens:

- write `panic: <value>` and a newline to `out` (format the value with `%v`)
- reply with `http.Error`, status 500 and the message `internal server error`

When nothing panics, the response passes through unchanged and nothing is
logged.

```text
handler panics with "boom"
// 500, body "internal server error\n", log "panic: boom\n"
```

## Learn more

- [net/http: HandlerFunc](https://pkg.go.dev/net/http#HandlerFunc)
- [net/http: Handler](https://pkg.go.dev/net/http#Handler)
- [Go blog: Defer, Panic, and Recover](https://go.dev/blog/defer-panic-and-recover)
