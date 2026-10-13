## Task 1: Fix the failure count

Read the `err == nil` branch of `record`. It closes the breaker, but what
happens to the failures counted before this success?

### Closer

The count lives in `b.failures`. A success should set it back to zero before
returning.

### Answer

```go
if err == nil {
	b.setState(Closed)
	b.failures = 0
	return
}
```

## Task 2: Call the hook on every state change

You need the old state to pass as `from`, so grab it before you overwrite
`b.state`. Then only call the hook when it's set and something changed.

### Closer

Save `from := b.state` first. The condition is
`from != to && b.OnStateChange != nil`.

### Answer

```go
from := b.state
b.state = to
if from != to && b.OnStateChange != nil {
	b.OnStateChange(from, to)
}
```

## Task 3: Send requests through the breaker

`Call` only gives back an error, so the function you pass it must store the
response somewhere outside itself. A 5xx response should look like a failure
to the breaker but not to the caller.

### Closer

Add `var errServerError = errors.New("server error")` at package level.
Declare `var resp *http.Response` before `t.Breaker.Call`, and inside the
closure assign with `resp, err = t.Base.RoundTrip(req)` (declare `var err error`
first so `:=` doesn't shadow `resp`). After `Call`, turn `errServerError`
back into a `nil` error.

### Answer

```go
var resp *http.Response
err := t.Breaker.Call(func() error {
	var err error
	resp, err = t.Base.RoundTrip(req)
	if err == nil && resp.StatusCode >= http.StatusInternalServerError {
		return errServerError
	}
	return err
})
if errors.Is(err, errServerError) {
	return resp, nil
}
return resp, err
```

When the breaker is open the closure never runs, so `resp` stays `nil` and
the caller gets `ErrOpen`.
