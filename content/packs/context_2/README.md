# Context: cancelling HTTP requests

A context only helps if it reaches the code that waits. For HTTP, that means
attaching it to the request. Then a cancelled deploy, or a slow endpoint,
stops the request instead of leaving it hanging.

**New here:** `http.NewRequestWithContext`; passing `ctx` down.

## Learn

### Recap

`context.WithCancel` and `context.WithTimeout` create a child context and a
`cancel` function (always `defer cancel()`). `ctx.Err()` says why a context
ended: `context.Canceled` or `context.DeadlineExceeded`. Functions that might
wait take `ctx` first and pass it on.

### Requests with a context

In *HTTP clients* you built requests with `http.NewRequest(method, url,
body)`. Its sibling `http.NewRequestWithContext` takes a context **first**,
then the same three arguments:

```go
req, err := http.NewRequestWithContext(ctx, http.MethodDelete, apiURL+"/pods/web-1", nil)
```

When `ctx` is cancelled or times out, `client.Do` stops and returns an error
that wraps `ctx.Err()`. So `errors.Is(err, context.DeadlineExceeded)` is
`true` after a timeout. Without a context, the request waits as long as the
server does, maybe forever.

On the server side, `r.Context()` is cancelled when the client goes away. The
test servers use that to wait for you to cancel.

> **Why is it written like this?** Putting `ctx` first is a Go-wide
> convention; the other three arguments match `http.NewRequest`. `nil` means
> "no body", normal for GET.

### Pass the context all the way down

```text
Versions(ctx) ──▶ Fetch(ctx) ──▶ NewRequestWithContext(ctx) ──▶ client.Do
     ▲
 cancel() here stops the request at the bottom
```

If any layer swaps in `context.Background()`, the chain breaks: cancelling at
the top does nothing below that point.

## Your task

Your release dashboard fetches `/version` from every service.

### 1. Attach the context to the request

**Goal:** make `Fetch(ctx, client, url)` stop its request when `ctx` is
cancelled.

`Fetch` GETs `url` and returns the body as a string. A status other than 200
gives the error `unexpected status <code>`. That part works already; only
the line that builds the request ignores `ctx`.

```go
Fetch(ctx, client, srv.URL+"/version") // => "v1.4.2", nil

// /slow never answers; ctx is cancelled while waiting
Fetch(ctx, client, srv.URL+"/slow") // => an error where errors.Is(err, context.Canceled)
```

**Think it through:**
- Which line creates the request? Which function also accepts a context?

### 2. Fetch with a timeout

**Goal:** make `FetchWithTimeout(client, url, timeout)` call `Fetch` with a
context that times out after `timeout`.

Start from `context.Background()`, and cancel the context when you're done.

```go
FetchWithTimeout(client, srv.URL+"/version", time.Second)    // => "v1.4.2", nil
FetchWithTimeout(client, srv.URL+"/slow", 20*time.Millisecond)
// => after ~20ms, an error where errors.Is(err, context.DeadlineExceeded)
```

**Think it through:**
- Which `context` function makes a child that ends on its own?
- Which of its two results goes to `Fetch`, and which gets deferred?

### 3. Pass the context through

**Goal:** make `Versions(ctx, client, urls)` pass its own `ctx` to `Fetch`.

`Versions` fetches each URL in order and returns the bodies. On the first
error it stops and returns `nil` and that error. With an already-cancelled
`ctx`, it must return a `context.Canceled` error without sending any request.

```go
Versions(ctx, client, []string{srv.URL + "/version", srv.URL + "/version"})
// => []string{"v1.4.2", "v1.4.2"}, nil

// ctx is already cancelled
Versions(ctx, client, urls) // => nil, an error where errors.Is(err, context.Canceled)
```

**Think it through:**
- Which argument in the loop breaks the chain? What's the name of the
  context `Versions` was given?

## Learn more

- [http.NewRequestWithContext](https://pkg.go.dev/net/http#NewRequestWithContext)
- [Go blog: Go Concurrency Patterns: Context](https://go.dev/blog/context)
- [Go blog: Contexts and structs](https://go.dev/blog/context-and-structs)
