# Context: cancelling HTTP requests

A context only helps if it reaches the code that waits. For HTTP, that means
attaching it to the request. Then a cancelled deploy, or a slow endpoint,
stops the request instead of leaving it hanging.

## Learn

### Recap

`context.WithCancel` and `context.WithTimeout` create a child context and a
`cancel` function (always `defer cancel()`). `ctx.Done()` is closed when the
context ends, and `ctx.Err()` says why: `context.Canceled` or
`context.DeadlineExceeded`. Functions that might wait take `ctx` as their
first parameter and pass it on.

### Requests with a context

In *HTTP clients* you built requests with `http.NewRequest`. Its sibling
`http.NewRequestWithContext` attaches a context:

```go
req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
if err != nil {
	return "", err
}
resp, err := client.Do(req)
```

When `ctx` is cancelled or times out, `client.Do` (or reading the body) stops
and returns an error that wraps `ctx.Err()`. So `errors.Is(err,
context.DeadlineExceeded)` is `true` after a timeout. Without a context, the
request waits as long as the server does, which might be forever.

On the server side, `r.Context()` is cancelled when the client goes away. The
test servers in this exercise use that to wait for you to cancel.

### Pass the context all the way down

```text
Versions(ctx) ──▶ Fetch(ctx) ──▶ NewRequestWithContext(ctx) ──▶ client.Do
     ▲
 cancel() here stops the request at the bottom
```

If any layer in the middle swaps in `context.Background()`, the chain is
broken. Cancelling at the top then does nothing below that point.

## Your task

Your release dashboard fetches `/version` from every service.

### 1. Attach the context to the request

`Fetch(ctx, client, url)` GETs `url` and returns the response body as a
string. A status other than 200 gives the error `unexpected status <code>`.
That part works. But it ignores `ctx`, so cancelling does nothing. Build the
request with `http.NewRequestWithContext` so cancelling `ctx` stops it.

```go
Fetch(ctx, client, srv.URL+"/version") // => "v1.4.2", nil

// /slow never answers; ctx is cancelled while waiting
Fetch(ctx, client, srv.URL+"/slow") // => an error where errors.Is(err, context.Canceled)
```

### 2. Fetch with a timeout

`FetchWithTimeout(client, url, timeout)` should create a context that times
out after `timeout`, starting from `context.Background()`, and call `Fetch`
with it.

```go
FetchWithTimeout(client, srv.URL+"/version", time.Second)    // => "v1.4.2", nil
FetchWithTimeout(client, srv.URL+"/slow", 20*time.Millisecond)
// => after ~20ms, an error where errors.Is(err, context.DeadlineExceeded)
```

### 3. Pass the context through

`Versions(ctx, client, urls)` fetches each URL in order and returns the
bodies. On the first error, it stops and returns `nil` and that error. It
currently calls `Fetch` with `context.Background()`, so the caller's context
is lost. Pass `ctx` through. With an already-cancelled `ctx`, `Versions`
must return a `context.Canceled` error without sending any request.

```go
Versions(ctx, client, []string{srv.URL + "/version", srv.URL + "/version"})
// => []string{"v1.4.2", "v1.4.2"}, nil

// ctx is already cancelled
Versions(ctx, client, urls) // => nil, an error where errors.Is(err, context.Canceled)
```

## Learn more

- [http.NewRequestWithContext](https://pkg.go.dev/net/http#NewRequestWithContext)
- [Go blog: Go Concurrency Patterns: Context](https://go.dev/blog/context)
- [Go blog: Contexts and structs](https://go.dev/blog/context-and-structs)
