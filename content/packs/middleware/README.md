# Middleware: logging, request IDs and panics

Every service does the same chores on every request: log it, tag it with an
ID you can search for later, and survive a bug without crashing. *Middleware*
does those chores once, around all your handlers.

**New here:** middleware, `http.HandlerFunc`, `next.ServeHTTP`. (Functions
that return functions are from *Functions as values*; `recover` from *panic
and recover*.)

## Learn

### A handler that wraps a handler

Middleware is a function that takes a handler and returns a new handler. The
new handler does something extra, then calls the original one (named `next`)
to do the real work:

```go
// Middleware wraps a handler with extra behaviour.
type Middleware func(http.Handler) http.Handler
```

A complete middleware that adds a header to every response:

```go
func AddVersion(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Version", "1.4.2") // before
		next.ServeHTTP(w, r)                 // the real handler
	})
}
```

- `http.HandlerFunc(func...)` turns an ordinary function into an
  `http.Handler`. It's a *conversion*, like `int(x)`, not a call.
- `next.ServeHTTP(w, r)` runs the wrapped handler. Code above it runs before
  the handler; code below it runs after. Skip it, and the handler never runs.
- The inner function is a *closure*, as in *Functions as values*: it can
  still use `next` after `AddVersion` has returned.

> **Why is it written like this?** Go requires a handler function to take
> exactly `(w http.ResponseWriter, r *http.Request)`. The names `w`, `r` and
> `next` are convention; Go would accept any names.

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

h := AddHeader("X-Team", "platform")(mux)
```

Read `AddHeader("X-Team", "platform")(mux)` left to right: the first call
gives a `Middleware`, and calling that with `mux` gives the wrapped handler.

All three functions in this exercise have this shape, and the starter already
has the three layers. You only fill in the innermost function.

> **Why is it written like this?** The inner function uses `key` and `value`,
> the outer function's *parameters*. Use those names, not the function's name
> `AddHeader`. The three nested `return`s are required: each layer must hand
> back the next one.

### Stacking middleware

The **last** wrapper you apply is the **outermost**, so it runs first:

```go
h := Recover(os.Stderr)(Logging(os.Stdout)(mux))
// a request goes: Recover -> Logging -> mux
```

### Recovering from a panic

One bad request shouldn't take down the whole service. As in *panic and
recover*, a deferred function can call `recover()`. It returns the panic
value, or `nil` when nothing panicked:

```go
defer func() {
	if v := recover(); v != nil {
		fmt.Println("caught:", v)
	}
}()
```

`%v` formats any panic value (a string, an error, ...) for a log line.

> **Why is it written like this?** `defer func() { ... }()` defines a function
> and calls it later. The final `()` is required: it's the call. `defer` must
> run *before* the code that might panic.

## Your task

Your team's services need three standard middlewares. Each test wraps a small
handler with yours and checks the response.

### 1. Log every request

**Goal:** `Logging(out)` writes `METHOD PATH` and a newline to `out`, then
runs the next handler.

The line is: the method, one space, the path, `\n`. The method is
`r.Method`; the path is `r.URL.Path` (without any `?query`). Add the `"fmt"`
import.

```text
GET /healthz        logs "GET /healthz\n"
POST /deploys       logs "POST /deploys\n"
GET /logs?limit=5   logs "GET /logs\n"
```

**Think it through:**
- Which `fmt` function writes formatted text to a writer like `out`?
- Don't forget the space between the two parts, and the newline.

### 2. Tag every request with an ID

**Goal:** `RequestID(newID)` makes sure the request and the response both
carry the same `X-Request-ID`.

- Request already has an `X-Request-ID` header: keep it, and **don't** call
  `newID`.
- No header: call `newID()` once, and set the result on the request with
  `r.Header.Set`, so the next handler can read it.
- Either way: set the same ID as the response's `X-Request-ID` header, before
  calling the next handler.

```text
request without X-Request-ID, newID returns "req-7f3a"
// next handler sees "req-7f3a"; response header X-Request-ID: req-7f3a

request with X-Request-ID: abc-123
// next handler sees "abc-123"; response header X-Request-ID: abc-123
```

**Think it through:**
- What does `r.Header.Get` return when the header is missing?
- `newID` is a parameter holding a function. You call it with `()`.

### 3. Recover from panics

**Goal:** `Recover(out)` turns a panic in the next handler into a 500 response
and a log line.

When the next handler panics:

- write `panic: <value>` and a newline to `out`, formatting the value with
  `%v`
- reply with `http.Error`, status 500, message `internal server error`

When nothing panics, the response passes through unchanged and nothing is
logged.

```text
handler panics with "boom"
// 500, body "internal server error\n", log "panic: boom\n"
```

**Think it through:**
- Where must the `defer` go relative to `next.ServeHTTP`?
- Which status constant means 500?

## Learn more

- [net/http: HandlerFunc](https://pkg.go.dev/net/http#HandlerFunc)
- [net/http: Handler](https://pkg.go.dev/net/http#Handler)
- [Go blog: Defer, Panic, and Recover](https://go.dev/blog/defer-panic-and-recover)
