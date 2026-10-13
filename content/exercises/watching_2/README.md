# Watching resources: closed watches and ERROR events

A watch that works in a test can still fall over in production, because real
API servers end watches and report errors through them. Here you fix a Pod
monitor so it copes with both.

## Learn

### Recap

A watch streams events from the API server. `w.ResultChan()` delivers
`watch.Event` values with a `Type` (`ADDED`, `MODIFIED`, `DELETED`,
`BOOKMARK` or `ERROR`) and an `Object`, and `w.Stop()` ends the watch. You
reach the Pod with `pod, ok := ev.Object.(*corev1.Pod)`. A watch loop selects
on `ctx.Done()` and the result channel.

### A closed channel never blocks

A receive from a closed channel returns the zero value straight away, every
time:

```go
case ev := <-w.ResultChan(): // closed: ev is watch.Event{}, with Type "" and Object nil
```

A loop like that wakes up millions of times a second, does nothing useful and
never returns. One CPU core sits at 100% and the program looks frozen. It is a
classic watch bug, and it only shows up when the server ends the watch. Real
servers do that all the time. The fix is the two-value receive,
`ev, ok := <-w.ResultChan()`: `ok` is `false` once the channel is closed.

### ERROR events

When a watch fails, the API server sends one `ERROR` event and closes the
channel. The event's object is a `*metav1.Status` with an HTTP code, a
*reason* and a message. `apierrors.FromObject` turns it into an ordinary
error that works with the `apierrors` helpers from *The clientset: get,
create and delete*:

```go
err := apierrors.FromObject(ev.Object)
fmt.Println(err)                              // too old resource version: 5 (12)
fmt.Println(apierrors.IsResourceExpired(err)) // true
fmt.Println(apierrors.ReasonForError(err))    // Expired
```

These helpers still work after you wrap the error with `%w`, as in *Wrapping
errors*.

### Keeping state from events

Events only describe changes, so a program that watches keeps its own picture
of the cluster and updates it with every event. A map from Pod name to phase
is enough to answer "how many Pods are Running?". The phase is
`pod.Status.Phase`, a `corev1.PodPhase` such as `corev1.PodPending`,
`corev1.PodRunning` or `corev1.PodFailed`. `delete(m, key)` removes a key from
a map.

### Why raw watches are fragile

- Connections drop, and the API server ends every watch after a few minutes on
  purpose. You must start a new one.
- Every object and every list has a *resourceVersion*, a marker for "the
  cluster as of this moment". To restart without missing changes, you watch
  from the last one you saw: `metav1.ListOptions{ResourceVersion: rv}`.
- The server only remembers recent versions. Ask for an old one and you get an
  `ERROR` event with reason `Expired` (HTTP 410 Gone). Then you must `List`
  everything again and watch from the list's resourceVersion, or you miss
  whatever changed in between.

Getting all of that right by hand is hard. client-go's *informers* do it for
you: they list, watch, reconnect, re-list and keep a local cache. They are
next.

## Your task

The `payments` team's rollout monitor follows the Pods in their namespace and
counts how many are Running. A teammate wrote it, and it misbehaves in
production.

### 1. Finish `Apply`

`Apply` updates `phases`, a map from Pod name to phase, with one event. So far
it only handles `ADDED`. Make it handle:

- `ADDED` and `MODIFIED`: set `phases[pod.Name]` to the Pod's phase.
- `DELETED`: remove the Pod from the map.
- Anything else (`BOOKMARK`, `ERROR`, or an object that is not a Pod): leave
  the map alone.

```go
phases := map[string]corev1.PodPhase{"checkout-1": corev1.PodPending}
Apply(phases, watch.Event{Type: watch.Modified, Object: checkout1}) // checkout1 is now Running
// phases => map[checkout-1:Running]
Apply(phases, watch.Event{Type: watch.Deleted, Object: checkout1})
// phases => map[]
```

### 2. Stop spinning on a closed watch

After an API server upgrade, the monitor pinned a CPU core at 100% and never
reported again. `Follow` reads from the channel without checking whether it
is closed. Fix it: when the channel is closed, return the phases so far and
`ErrWatchClosed`.

```go
// w delivers ADDED checkout-1 (Pending), ADDED checkout-2 (Pending) and
// MODIFIED checkout-1 (Running), then the server closes it.
Follow(ctx, w)
// => map[checkout-1:Running checkout-2:Pending], ErrWatchClosed
```

The tests check the error with `errors.Is(err, ErrWatchClosed)`. If `Follow`
is still running 2s after the watch closed, the test fails.

### 3. Report ERROR events

`Follow` hands ERROR events to `Apply`, which ignores them, so the monitor
never says why its watch ended. When an `ERROR` event arrives, return the
phases so far and the error from `apierrors.FromObject(ev.Object)`, wrapped
with `%w` after the prefix `watch error: `:

```text
watch error: too old resource version: 5 (12)
```

The tests check that text, and that `apierrors.ReasonForError(err)` still
finds the reason (`Expired`, or `InternalError` for a server error). Add
`"fmt"` and `apierrors "k8s.io/apimachinery/pkg/api/errors"` to the imports.

## Learn more

- [Kubernetes: Resource versions](https://kubernetes.io/docs/reference/using-api/api-concepts/#resource-versions)
- [Kubernetes: Efficient detection of changes](https://kubernetes.io/docs/reference/using-api/api-concepts/#efficient-detection-of-changes)
- [pkg.go.dev: k8s.io/apimachinery/pkg/api/errors](https://pkg.go.dev/k8s.io/apimachinery/pkg/api/errors)
- [Go spec: Receive operator](https://go.dev/ref/spec#Receive_operator)
