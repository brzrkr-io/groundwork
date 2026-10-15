## Task 1: Write HTTPError's message

`TimeoutError`'s `Error` method just below shows the shape: build a string from the receiver's fields with `fmt.Sprintf`.

### Closer

Use `%s` for `e.URL` and `%d` for `e.StatusCode`, with the fixed text `GET ` and `: status ` around them.

### Answer

```go
return fmt.Sprintf("GET %s: status %d", e.URL, e.StatusCode)
```

## Task 2: Turn a status code into an error

Handle the healthy range first and return early. After that, every remaining code is a failure.

### Closer

The healthy check is `status >= 200 && status <= 299`. For failures, return a pointer: `&HTTPError{...}`.

### Answer

```go
if status >= 200 && status <= 299 {
	return nil
}
return &HTTPError{URL: url, StatusCode: status}
```

## Task 3: Decide what to retry

A type switch tells you which error type `err` holds. Timeouts are always worth retrying. HTTP errors are worth it only for some status codes. Everything else, including `nil`, is not.

### Closer

Write `switch e := err.(type) { ... }`. Inside `case *HTTPError:`, `e` is the `*HTTPError`, so you can read `e.StatusCode`. `default` also covers `nil`.

### Answer

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
