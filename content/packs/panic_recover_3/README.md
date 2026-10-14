# panic and recover: protect the status API

A web server handles thousands of requests. If one handler hits a bug, that
request should fail with a 500 error and the bug should be logged, but the
server must keep serving everyone else. You build that safety net here.

**New here:** middleware (a function that returns a function, as in *Functions as values*), type switch on a recovered value. (`%T` is from *Type assertions and switches*.)

## Learn

### Recap from the earlier *panic and recover* exercises

- `recover()` stops a panic when called inside a deferred function.
- A deferred closure can set a named result.

### Middleware: a function that returns a function

*Middleware* takes a handler and returns a **new** handler that does
something extra around it. The returned function literal can have its own
named result and its own `defer`, like any function:

```go
func Branded(h Handler) Handler {
	return func(req Request) (resp Response) {
		defer func() {
			resp.Body += " -- powered by ops" // runs after h(req), changes resp
		}()
		return h(req)
	}
}
```

The `defer` belongs to the **inner** function, so it runs once per request,
not once when `Branded` is called.

> **Why is it written like this?** `Handler` is a function type, so Go
> requires the returned literal to have exactly its shape:
> `func(Request) Response`. Naming the result (`resp`) and the parameter
> (`req`) is your choice; Go allows names in the literal as long as the types
> match.

### What did the handler panic with?

`recover()` returns an `any`: a string from `panic("...")`, an `error`, a
number, anything. Use a *type switch* (as in *Type assertions and switches*)
to handle each kind:

```go
switch x := v.(type) {
case int:
	// here x is an int
case bool:
	// here x is a bool
default:
	// here x is still an any
}
```

Go's own panics, such as an index out of range, are `error` values whose
message starts with `runtime error:`.

> **Why is it written like this?** `v.(type)` is only allowed in a `switch`;
> Go requires that. Reusing the same name (`switch v := v.(type)`) is a common
> convention; a new name like `x` works the same.

### Printing a value's type with %T

The `%T` verb prints a value's type, `%v` prints the value:

```go
fmt.Sprintf("%v is a %T", 3.5, 3.5)        // => "3.5 is a float64"
fmt.Sprintf("%v is a %T", true, true)      // => "true is a bool"
```

Real Go servers work this way: `net/http` recovers a panic in a handler so
the rest of the server keeps running.

## Your task

The on-call team runs a small status API. You add a `Protect` middleware
that turns handler panics into 500 responses and logs them, and a check that
pages on-call when the database is down.

### 1. Convert a panic value into an error

**Goal:** write `PanicError(v any) error` that turns any recovered value into an error starting with `panic: `.

| `v` is | Error message | Extra rule |
|---|---|---|
| an `error` | `panic: <its message>` | wrap it with `%w`, so `errors.Is` still finds what's inside |
| a `string` | `panic: <v>` | |
| anything else | `panic: unexpected value <v> (<type of v>)` | value with `%v`, type with `%T` |

Add the `fmt` import.

```go
PanicError("boom")          // => "panic: boom"
PanicError(ErrDatabaseDown) // => "panic: database is down", errors.Is(err, ErrDatabaseDown) is true
PanicError(42)              // => "panic: unexpected value 42 (int)"
PanicError([]string{"web-1", "web-2"})
// => "panic: unexpected value [web-1 web-2] ([]string)"
```

**Think it through:**
- Three rows in the table means three branches. Which one is `default`?
- In the last row, the same `v` is used twice: once for `%v`, once for `%T`.
- The fixed text `panic: ` has a space after the colon.

### 2. Write the Protect middleware

**Goal:** make `Protect(name, h, errs)` return a new `Handler` that calls `h(req)` and turns a panic into a 500 response plus one logged error.

- If `h` doesn't panic: return its response unchanged, record nothing.
- If `h` panics: return `Response{Status: 500, Body: "internal server error"}` and call `errs.Record` exactly once.
- The recorded error's text is `handler "<name>": ` followed by `PanicError`'s error, wrapped with `%w` so `errors.Is` still finds `ErrDatabaseDown`.

```go
h := Protect("users", usersHandler, errs) // usersHandler panics with ErrDatabaseDown
h(Request{Method: "GET", Path: "/users"})
// => Response{Status: 500, Body: "internal server error"}
// errs.Errors now holds one error: `handler "users": panic: database is down`

h = Protect("deploys", buggyHandler, errs) // buggyHandler reads parts[3] of a 2-part slice
// recorded: `handler "deploys": panic: runtime error: index out of range [3] with length 2`
```

**Think it through:**
- Pieces of the recorded error: fixed `handler `, the `name` parameter in quotes, fixed `: `, then the result of task 1 given the recovered value.
- Which named result does the deferred closure set to the 500 response?
- `name`, `h` and `errs` are parameters of `Protect`, but the inner function can still use them.

### 3. Page on-call when the database is down

**Goal:** make `ShouldPage(errs)` return `true` if any error in `errs.Errors` matches `ErrDatabaseDown` with `errors.Is`.

Return `false` otherwise, including when the log is empty.

**Think it through:**
- Can you return `true` as soon as you find one match?
- What should happen after the loop finishes without a match?

## Learn more

- [Go blog: Defer, Panic, and Recover](https://go.dev/blog/defer-panic-and-recover)
- [Go blog: Working with Errors in Go 1.13](https://go.dev/blog/go1.13-errors)
- [fmt package: printing verbs](https://pkg.go.dev/fmt#hdr-Printing)
- [net/http: Handler](https://pkg.go.dev/net/http#Handler)
