# Custom error types: retry failed health checks

A health checker that retries every failure wastes time on problems that
never fix themselves, like a wrong URL. One that never retries pages someone
at 3 a.m. for a blip. Two custom error types let it tell the difference.

**New here:** using fields inside a type switch, returning a condition.

## Learn

### Recap

From *Custom error types*: a custom error is a struct with an `Error() string`
method on a pointer receiver. Return it as `&T{...}` through `error`, and
always `return nil` for success. Callers inspect it with a type assertion or a
type switch.

### Using the fields inside a type switch

Write `e := err.(type)` and each case gets its own `e`, already of that case's
type. Inside `case *DiskFullError:`, `e` is a `*DiskFullError`, so you can read
its fields:

```go
switch e := err.(type) {
case *DiskFullError:
	fmt.Println("disk", e.Mount, "is full")
case *QuotaError:
	fmt.Println("user", e.User, "is over quota")
default:
	fmt.Println("something else")
}
```

> **Why is it written like this?** Go requires the `.(type)` form; it only
> works inside a `switch`. The name `e` is the author's choice. `default` is
> optional, and it also catches `nil` when there is no `case nil`.

### Returning a condition

A `return` in a `bool` function can return a whole comparison, not only
`true` or `false`. Join comparisons with `||` (or) and `&&` (and).

```go
case *DiskFullError:
	return e.UsedPercent >= 95 || e.Mount == "/"
```

### Which failures are worth a retry?

HTTP status codes come in families. The checker treats them like this:

| Failure | Retry? | Why |
|---|---|---|
| `*HTTPError`, status 500 or above (like 503) | yes | the service may be restarting or overloaded |
| `*HTTPError`, status 429 Too Many Requests | yes | you were rate-limited; wait and try again |
| `*HTTPError`, any other status (404, 401, 302, …) | no | the request itself is wrong |
| `*TimeoutError` | yes | the host may just be slow |
| any other error, or `nil` | no | unknown problem, or no problem at all |

## Your task

The checker calls each service's health endpoint. `TimeoutError` is already
written; use its `Error` method as a model. You'll write `HTTPError`'s message,
create it from a status code, and decide what to retry.

### 1. Write HTTPError's message

**Goal:** `(*HTTPError).Error` returns `GET <URL>: status <StatusCode>`.

The fixed text is `GET ` (with a space) and `: status ` (colon, space, `status`, space).

```go
(&HTTPError{URL: "http://api:8080/healthz", StatusCode: 503}).Error()
// => "GET http://api:8080/healthz: status 503"
(&HTTPError{URL: "http://web-2/ready", StatusCode: 404}).Error()
// => "GET http://web-2/ready: status 404"
```

**Think it through:**
- Which two fields of `e` fill the gaps? Which verb fits text, which fits a number?
- The URL has no quotes around it.

### 2. Turn a status code into an error

**Goal:** `CheckResponse(url, status)` returns `nil` for codes 200 to 299 (both included), and a `*HTTPError` for anything else.

The `*HTTPError` holds the `url` and `status` you were given.

```go
CheckResponse("http://api:8080/healthz", 204) // => nil
CheckResponse("http://web-2/ready", 404)      // => error "GET http://web-2/ready: status 404"
CheckResponse("http://cache:6379/ping", 199)  // => error "GET http://cache:6379/ping: status 199"
CheckResponse("http://cache:6379/ping", 300)  // => error "GET http://cache:6379/ping: status 300"
```

**Think it through:**
- "From 200 to 299" needs two comparisons. Which operator joins them?
- Which parameters go into which fields of the error?

### 3. Decide what to retry

**Goal:** `Retryable(err)` returns `true` only for the failures marked "yes" in the retry table.

Use a type switch.

```go
Retryable(&HTTPError{URL: "http://api/healthz", StatusCode: 503}) // => true
Retryable(&HTTPError{URL: "http://api/healthz", StatusCode: 429}) // => true
Retryable(&HTTPError{URL: "http://api/healthz", StatusCode: 404}) // => false
Retryable(&TimeoutError{Host: "db-1", Seconds: 5})                // => true
Retryable(errors.New("dns lookup failed"))                        // => false
Retryable(nil)                                                    // => false
```

**Think it through:**
- How many cases do you need? One per row with a different *type*, plus one for the rest.
- In the `*HTTPError` case, two status rules mean "yes". Can one returned condition cover both?

## Learn more

- [Go spec: Type switches](https://go.dev/ref/spec#Type_switches)
- [MDN: HTTP response status codes](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status)
- [Kubernetes: Configure liveness, readiness and startup probes](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/)
