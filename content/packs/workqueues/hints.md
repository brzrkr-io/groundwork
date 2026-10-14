## Task 1: Enqueue a batch of events

Loop over `objs` and get each object's key with `KeyFor`. Stop at the first
error; otherwise add the key to the queue, which removes duplicates for you.

### Closer

Use `for _, obj := range objs`, then `key, err := KeyFor(obj)`. On an error,
`return err`; on success, `q.Add(key)`. Return `nil` after the loop.

### Answer

```go
for _, obj := range objs {
	key, err := KeyFor(obj)
	if err != nil {
		return err
	}
	q.Add(key)
}
return nil
```

## Task 2: Process one key

Follow the four steps in the README in order. Every key you take with `Get`
must be marked `Done`, however the function returns.

### Closer

Start with the shutdown check, then the `defer`:

```go
key, shutdown := q.Get()
if shutdown {
	return false
}
defer q.Done(key)
```

Then `namespace, name, err := cache.SplitMetaNamespaceKey(key)`.

### Answer

```go
key, shutdown := q.Get()
if shutdown {
	return false
}
defer q.Done(key)

namespace, name, err := cache.SplitMetaNamespaceKey(key)
if err != nil {
	return true
}
handle(namespace, name)
return true
```

The `defer` sits before any early return, so the key is always marked done
and can be added again.
