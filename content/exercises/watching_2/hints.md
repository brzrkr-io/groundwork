## Task 1: one more case, and a delete

A `case` can list several values, separated by commas. Add `watch.Modified`
next to `watch.Added`, and add a second `case` for `watch.Deleted` that calls
`delete(phases, pod.Name)`. The switch already ignores every other type.

## Task 2: the two-value receive

Change the receive in the `select` so it also gives you `ok`:

```go
case ev, ok := <-w.ResultChan():
	if !ok {
		return phases, ErrWatchClosed
	}
```

## Task 3: check the type before Apply

```go
if ev.Type == watch.Error {
	err := apierrors.FromObject(ev.Object)
	return phases, fmt.Errorf("watch error: %w", err)
}
Apply(phases, ev)
```
