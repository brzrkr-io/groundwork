## Task 1: loop, key, add

Range over `objs`. For each one, call `KeyFor`. On an error, return it. On
success, `q.Add(key)`. The queue does the deduplication for you.

## Task 2: the shape of ProcessNextItem

```go
key, shutdown := q.Get()
if shutdown {
	return false
}
defer q.Done(key)
```

Put the `defer` before anything that can return early, so the key is always
marked done.

## Task 2: almost the answer

```go
namespace, name, err := cache.SplitMetaNamespaceKey(key)
if err != nil {
	return true
}
handle(namespace, name)
return true
```
