## Task 1: Finish `Apply`

The `switch` only has a case for `watch.Added`. `MODIFIED` should do the same
thing, and `DELETED` needs its own case that removes the Pod. The switch
already ignores every other type.

### Closer

A `case` can list several values separated by commas:
`case watch.Added, watch.Modified:`. The built-in `delete(m, key)` removes a
key from a map.

### Answer

```go
switch ev.Type {
case watch.Added, watch.Modified:
	phases[pod.Name] = pod.Status.Phase
case watch.Deleted:
	delete(phases, pod.Name)
}
```

## Task 2: Stop spinning on a closed watch

A closed channel never blocks: every receive returns straight away with a
zero value, so the loop spins. You need the receive to also tell you whether
the channel is closed.

### Closer

The two-value receive `ev, ok := <-ch` sets `ok` to `false` once the channel
is closed. Change the `case` line to use it.

### Answer

```go
case ev, ok := <-w.ResultChan():
	if !ok {
		return phases, ErrWatchClosed
	}
```

Returning the sentinel error lets callers check it with `errors.Is`.

## Task 3: Report ERROR events

Check the event type before calling `Apply`. When it is `watch.Error`, turn
the event's object into an API error and return it.

### Closer

`apierrors.FromObject(ev.Object)` turns the `*metav1.Status` in an ERROR event
into an error. Wrap it with `fmt.Errorf("watch error: %w", ...)`, and add the
`"fmt"` and `apierrors "k8s.io/apimachinery/pkg/api/errors"` imports.

### Answer

```go
if ev.Type == watch.Error {
	return phases, fmt.Errorf("watch error: %w", apierrors.FromObject(ev.Object))
}
Apply(phases, ev)
```

`%w` keeps the API error inside, so `apierrors.ReasonForError` still finds
the reason.
