# Custom error types: fix the port allocator

A teammate wrote a small port allocator that hands out ports to services on a
shared host. It has a custom error type, but the code has bugs. One of them is
the most famous trap in Go error handling.

## Learn

Recap: a custom error type is a struct with an `Error() string` method,
usually on a pointer receiver, so you return it as `&T{...}` through an
`error` result. Callers get the fields back with a type assertion,
`perr, ok := err.(*T)`, which gives `ok == false` for `nil` and for other
kinds of error.

### The typed-nil trap, up close

An `error` value has two parts: a *type* and a *value*. It is `nil` only when
**both** parts are empty.

```go
var p *PortInUseError // type *PortInUseError, value nil
var err error = p     // err now holds that type

fmt.Println(p == nil)   // true
fmt.Println(err == nil) // false!
```

Putting a nil pointer into an `error` fills in the type part, so the error is
no longer `nil`. Any function that returns a pointer variable of an error type
through `error` has this bug on its success path. The caller's
`if err != nil` fires even though nothing went wrong.

The fix is to never route success through such a variable. Return the error
directly where it happens, and write `return nil` for success:

```go
if somethingWrong {
	return &PortInUseError{...}
}
return nil
```

When a test reports this bug, `%T` prints the hidden type, so the message
mentions `*customerr2.PortInUseError` even though no port was taken.

## Your task

The allocator keeps a `map[int]string` from each taken port to the service
that owns it. Fix it so deploys get correct answers.

### 1. Fix the error message

`(*PortInUseError).Error` leaves out the owner. Make it return the port, then
`is already used by`, then the owner in double quotes:

```go
(&PortInUseError{Port: 8080, Owner: "nginx"}).Error()
// => `port 8080 is already used by "nginx"`
```

`%q` in `fmt.Sprintf` adds the double quotes for you.

### 2. Fix the typed-nil bug in Reserve

`Reserve(used, port, service)` should:

- If `port` is in `used`, return `&PortInUseError{Port: port, Owner: <owner from used>}`
  and leave `used` unchanged.
- Otherwise, set `used[port] = service` and return `nil`.

Right now a free port is recorded correctly but still returns a non-nil error.
Fix it.

```go
used := map[int]string{8080: "nginx"}
Reserve(used, 9090, "api")
// => nil, and used[9090] is now "api"
Reserve(used, 8080, "api")
// => error `port 8080 is already used by "nginx"`
```

### 3. Finish Advice

`Advice(err)` already handles `nil` (returns `"ok"`) and unknown errors
(returns `"unexpected error: "` plus the message). Add the missing case: when
`err` is a `*PortInUseError`, return

```text
stop "<owner>" or choose a port other than <port>
```

```go
Advice(&PortInUseError{Port: 8080, Owner: "nginx"})
// => `stop "nginx" or choose a port other than 8080`
Advice(errors.New("permission denied"))
// => "unexpected error: permission denied"
Advice(nil)
// => "ok"
```

## Learn more

- [Go FAQ: Why is my nil error value not equal to nil?](https://go.dev/doc/faq#nil_error)
- [Go spec: Type assertions](https://go.dev/ref/spec#Type_assertions)
- [A Tour of Go: Interface values with nil underlying values](https://go.dev/tour/methods/12)
