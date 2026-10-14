## Task 1: Convert a panic value into an error

A type switch lets you handle each kind of value differently. You need one
branch for errors, one for strings, and one for everything else.

### Closer

`switch v := v.(type)` gives `v` the right type in each branch. Use
`case error:`, `case string:` and `default:`. Only the `error` branch uses
`%w`; the `default` branch uses `%v` and `%T`.

### Answer

```go
switch v := v.(type) {
case error:
	return fmt.Errorf("panic: %w", v)
case string:
	return errors.New("panic: " + v)
default:
	return fmt.Errorf("panic: unexpected value %v (%T)", v, v)
}
```

## Task 2: Write the Protect middleware

`Protect` doesn't call `h` itself. It returns a new function, and that inner
function calls `h`, recovers a panic, and can change its own named result.

### Closer

Return `func(req Request) (resp Response) { ... }`. Inside it, defer a
function that calls `recover()`, then `return h(req)`. In the deferred
function, call `errs.Record(...)` and set `resp` to the 500 response.

### Answer

```go
return func(req Request) (resp Response) {
	defer func() {
		if r := recover(); r != nil {
			errs.Record(fmt.Errorf("handler %q: %w", name, PanicError(r)))
			resp = Response{Status: 500, Body: "internal server error"}
		}
	}()
	return h(req)
}
```

The `defer` belongs to the inner function, so it runs once per request.

## Task 3: Page on-call when the database is down

Look at every recorded error and ask `errors.Is` whether `ErrDatabaseDown` is
somewhere inside it. One match is enough.

### Closer

Range over `errs.Errors`, return `true` as soon as
`errors.Is(err, ErrDatabaseDown)` is true, and return `false` after the loop.

### Answer

```go
for _, err := range errs.Errors {
	if errors.Is(err, ErrDatabaseDown) {
		return true
	}
}
return false
```
