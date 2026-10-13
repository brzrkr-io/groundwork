# panic and recover: protect the status API

A web server handles thousands of requests. If one handler hits a bug, that
request should fail with a 500 error and the bug should be logged, but the
server must keep serving everyone else. You build that safety net here.

## Learn

Recap: `recover()` stops a panic when it is called inside a deferred function,
and a deferred function can set a named result. You did both in the previous
exercises. Two new pieces come in now.

### A deferred recover inside a returned function

*Middleware* is a function that takes a handler and returns a new handler that
does something extra around it. The returned function literal can have a named
result and its own `defer`, just like a normal function:

```go
func Timed(h Handler) Handler {
	return func(req Request) (resp Response) {
		defer func() {
			// runs after h(req), and can change resp
		}()
		return h(req)
	}
}
```

The `defer` belongs to the inner function, so it runs once per request.

### What did the handler panic with?

`recover()` returns an `any`. It can be a string from `panic("...")`, an
`error`, or anything else. Use a type switch (as in *Type assertions and
switches*) to handle each kind. Go's own panics, such as an index out of
range, are `error` values whose message starts with `runtime error:`.

If the value is an `error`, wrap it with `%w`. Then `errors.Is` still finds a
sentinel such as `ErrDatabaseDown` deep inside the chain, as in *Wrapping
errors*.

For values of other types, the `%T` verb prints the value's type:

```go
fmt.Sprintf("%v (%T)", 42, 42)                   // => "42 (int)"
fmt.Sprintf("%v (%T)", []int{1, 2}, []int{1, 2}) // => "[1 2] ([]int)"
```

Real Go servers work this way: `net/http` recovers a panic in a handler so the
rest of the server keeps running.

## Your task

The on-call team runs a small status API. You add a `Protect` middleware that
turns handler panics into 500 responses and logs them, and a check that pages
on-call when the database is down.

### 1. Convert a panic value into an error

Write `PanicError(v any) error` with a type switch:

| `v` is | Return |
|---|---|
| an `error` | `fmt.Errorf("panic: %w", v)` |
| a `string` | an error with message `panic: <v>` |
| anything else | an error with message `panic: unexpected value <v> (<type>)`, using `%v` and `%T` |

Add the `fmt` import.

```go
PanicError("boom")                 // => "panic: boom"
PanicError(ErrDatabaseDown)        // => "panic: database is down", errors.Is(err, ErrDatabaseDown) is true
PanicError(42)                     // => "panic: unexpected value 42 (int)"
PanicError([]string{"web-1", "web-2"})
// => "panic: unexpected value [web-1 web-2] ([]string)"
```

### 2. Write the Protect middleware

`Protect(name, h, errs)` returns a new `Handler`. The new handler calls
`h(req)`:

- If `h` does not panic, return its response unchanged and record nothing.
- If `h` panics, recover and return
  `Response{Status: 500, Body: "internal server error"}`. Also record exactly
  one error in `errs` with `errs.Record(...)`. The recorded error is
  `fmt.Errorf("handler %q: %w", name, PanicError(r))`, so `errors.Is` still
  finds `ErrDatabaseDown`.

```go
h := Protect("users", usersHandler, errs) // usersHandler panics with ErrDatabaseDown
h(Request{Method: "GET", Path: "/users"})
// => Response{Status: 500, Body: "internal server error"}
// errs.Errors now holds one error: `handler "users": panic: database is down`

h = Protect("deploys", buggyHandler, errs) // buggyHandler reads parts[3] of a 2-part slice
// recorded: `handler "deploys": panic: runtime error: index out of range [3] with length 2`
```

### 3. Page on-call when the database is down

`ShouldPage(errs)` returns `true` if at least one error in `errs.Errors`
matches `ErrDatabaseDown` with `errors.Is`, and `false` otherwise (including
when the log is empty).

## Learn more

- [Go blog: Defer, Panic, and Recover](https://go.dev/blog/defer-panic-and-recover)
- [Go blog: Working with Errors in Go 1.13](https://go.dev/blog/go1.13-errors)
- [fmt package: printing verbs](https://pkg.go.dev/fmt#hdr-Printing)
- [net/http: Handler](https://pkg.go.dev/net/http#Handler)
