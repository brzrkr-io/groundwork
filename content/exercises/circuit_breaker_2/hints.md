## Task 1: what does a success forget?

Read the `err == nil` branch of `record`. It closes the breaker, but the
failure count from earlier calls is still there. One line fixes it.

## Task 2: remember the old state

Save `from := b.state` before you overwrite it. Then call the hook only when
`from != to && b.OnStateChange != nil`.

## Task 3: capture the response in a closure

Add `var errServerError = errors.New("server error")` at package level. Then
declare `var resp *http.Response` before calling `t.Breaker.Call`. Inside the
function you pass, assign it with `resp, err = t.Base.RoundTrip(req)` (declare
`var err error` first so you don't shadow `resp` with `:=`). After `Call`
returns, turn `errServerError` back into a `nil` error.

## Near-solution

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
