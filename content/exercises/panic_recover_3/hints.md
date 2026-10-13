## Task 1: the type switch

`switch v := v.(type)` gives you `v` with the right type in each `case`. You
need three branches: `case error:`, `case string:` and `default:`. Only the
`error` branch uses `%w`.

## Task 2: a function that returns a function

`Protect` doesn't call `h` itself. It returns
`func(req Request) (resp Response) { ... }`. Inside that function, defer a
recover, then `return h(req)`. Because `resp` is a named result, the deferred
function can replace it with the 500 response.

## Task 3: loop and ask errors.Is

Range over `errs.Errors` and return `true` as soon as
`errors.Is(err, ErrDatabaseDown)` is true. After the loop, return `false`.

## Almost the answer

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
