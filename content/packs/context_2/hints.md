## Task 1: Attach the context to the request

Only one line changes: the one that builds the request. It needs a version
that also takes `ctx`.

### Closer

`http.NewRequestWithContext` takes the context first, then the same three
arguments as `http.NewRequest`.

### Answer

```go
req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
```

Now cancelling `ctx` makes `client.Do` stop and return an error wrapping `ctx.Err()`.

## Task 2: Fetch with a timeout

Instead of passing `context.Background()` straight to `Fetch`, make a child
of it that times out after `timeout`.

### Closer

`ctx, cancel := context.WithTimeout(context.Background(), timeout)`, then
`defer cancel()`, then call `Fetch` with `ctx`.

### Answer

```go
ctx, cancel := context.WithTimeout(context.Background(), timeout)
defer cancel()
return Fetch(ctx, client, url)
```

## Task 3: Pass the context through

The loop swaps in `context.Background()`, which breaks the chain from the
caller down to the request.

### Closer

Pass the `ctx` parameter of `Versions` to `Fetch`. With task 1 done, a
cancelled `ctx` makes `Fetch` fail before sending anything.

### Answer

```go
version, err := Fetch(ctx, client, url)
```

The caller's cancellation now reaches every request.
