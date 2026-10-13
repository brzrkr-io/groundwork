## Task 1: copy the pattern

`TimeoutError`'s `Error` method right below shows the shape. Use
`fmt.Sprintf` with `%s` for the URL and `%d` for the status code:
`"GET %s: status %d"`.

## Task 2: healthy first

Check the healthy range with `status >= 200 && status <= 299` and
`return nil`. After that `if`, every other code is a failure, so return
`&HTTPError{URL: url, StatusCode: status}`.

## Task 3: the switch

```go
switch e := err.(type) {
case *TimeoutError:
	return true
case *HTTPError:
	return e.StatusCode >= 500 || e.StatusCode == 429
default:
	return false
}
```

`default` also covers `nil`.
