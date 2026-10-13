# Custom error types: retry failed health checks

A health checker that retries every failure wastes time on problems that will
never fix themselves, like a wrong URL. One that never retries pages someone
at 3 a.m. for a blip. Two custom error types let it tell the difference.

## Learn

Recap: a custom error type is a struct with an `Error() string` method on a
pointer receiver. You return it as `&T{...}` through an `error` result, and
always write `return nil` for success. Callers inspect it with a type
assertion or a type switch.

### Using the fields inside a type switch

In a type switch, `e := err.(type)` gives you `e` with the matching case's
type. Inside `case *HTTPError:`, `e` is a `*HTTPError`, so you can read
`e.StatusCode`:

```go
switch e := err.(type) {
case *HTTPError:
	fmt.Println("status was", e.StatusCode)
case *TimeoutError:
	fmt.Println("waited", e.Seconds, "seconds")
}
```

A case can return a whole condition, not just `true` or `false`:

```go
case *HTTPError:
	return e.StatusCode == 503
```

### Which failures are worth a retry?

HTTP status codes come in families. The checker treats them like this:

| Failure | Retry? | Why |
|---|---|---|
| status 500 or above (server errors, like 503 Service Unavailable) | yes | the service may be restarting or overloaded |
| status 429 Too Many Requests | yes | you were rate-limited; wait and try again |
| any other bad status (404, 401, 302, …) | no | the request itself is wrong; retrying gives the same answer |
| a `*TimeoutError` | yes | the network or host may just be slow |
| any other error, or `nil` | no | unknown problem, or no problem at all |

## Your task

The checker calls each service's health endpoint. `TimeoutError` is already
written. You'll write `HTTPError`'s message, produce it from a status code,
and decide what to retry.

### 1. Write HTTPError's message

Make `(*HTTPError).Error` return `GET`, a space, the URL, a colon and a space,
then `status` and the code:

```go
(&HTTPError{URL: "http://api:8080/healthz", StatusCode: 503}).Error()
// => "GET http://api:8080/healthz: status 503"
```

### 2. Turn a status code into an error

`CheckResponse(url string, status int) error` returns `nil` for codes 200 to
299 (both included). For any other code, it returns
`&HTTPError{URL: url, StatusCode: status}`.

```go
CheckResponse("http://api:8080/healthz", 204)
// => nil
CheckResponse("http://web-2/ready", 404)
// => error "GET http://web-2/ready: status 404"
CheckResponse("http://web-2/ready", 302)
// => error "GET http://web-2/ready: status 302"
```

### 3. Decide what to retry

`Retryable(err error) bool` follows the table above. Use a type switch.

```go
Retryable(&HTTPError{URL: "http://api/healthz", StatusCode: 503}) // => true
Retryable(&HTTPError{URL: "http://api/healthz", StatusCode: 429}) // => true
Retryable(&HTTPError{URL: "http://api/healthz", StatusCode: 404}) // => false
Retryable(&TimeoutError{Host: "db-1", Seconds: 5})                // => true
Retryable(errors.New("dns lookup failed"))                        // => false
Retryable(nil)                                                    // => false
```

## Learn more

- [Go spec: Type switches](https://go.dev/ref/spec#Type_switches)
- [MDN: HTTP response status codes](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status)
- [Kubernetes: Configure liveness, readiness and startup probes](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/)
