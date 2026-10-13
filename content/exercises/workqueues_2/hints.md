## Task 1: Let keys come back

Every `Get` needs a matching call that tells the queue you have finished with
the key. Until then, adding the key again does nothing. The function has two
`return true` lines, so the call must run on both paths.

### Closer

Use `defer` with `q.Done(key)`, placed right after the shutdown check.

### Answer

```go
key, shutdown := q.Get()
if shutdown {
	return false
}
defer q.Done(key)
```

`defer` runs `Done` however the function returns.

## Task 2: Forget on success

The queue remembers how often each key failed. On success, tell it to forget
that history.

### Closer

The method is `q.Forget(key)`; call it inside `if err == nil` before
returning.

### Answer

```go
if err == nil {
	q.Forget(key)
	return false
}
```

## Task 3: Give up after MaxRetries

Before retrying, ask the queue how many times this key has already been
requeued and compare it with `MaxRetries`.

### Closer

`q.NumRequeues(key)` gives the count. Below `MaxRetries`, use
`q.AddRateLimited(key)`; otherwise `q.Forget(key)` and report the key as
dropped.

### Answer

```go
if q.NumRequeues(key) < MaxRetries {
	q.AddRateLimited(key)
	return false
}
q.Forget(key)
return true
```

`Forget` resets the count so the key starts fresh if it ever comes back.
