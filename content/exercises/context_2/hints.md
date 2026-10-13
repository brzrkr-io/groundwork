## Task 1: one line changes

`http.NewRequestWithContext` takes the context as its first argument, then
the same three arguments as `http.NewRequest`:

```go
req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
```

## Task 2: the usual three lines

```go
ctx, cancel := context.WithTimeout(context.Background(), timeout)
defer cancel()
return Fetch(ctx, client, url)
```

## Task 3: don't break the chain

Replace `context.Background()` in the loop with `ctx`. Once task 1 is done,
`Fetch` with a cancelled context fails before sending anything, and
`Versions` returns that error.
