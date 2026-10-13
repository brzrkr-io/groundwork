## Task 1: the missing call

Every `Get` needs a matching `Done`. The function has two `return true`
lines, so use `defer` right after the shutdown check:

```go
defer q.Done(key)
```

## Task 2: one line

When `err == nil`, call `q.Forget(key)` before `return false`.

## Task 3: check the count first

Ask the queue how many times this key has already been requeued, and compare
with `MaxRetries` before deciding.

## Task 3: almost the answer

```go
if q.NumRequeues(key) < MaxRetries {
	q.AddRateLimited(key)
	return false
}
q.Forget(key)
return true
```
