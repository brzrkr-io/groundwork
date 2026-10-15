# Watching resources: closed watches and ERROR events

A watch that works in a test can still fall over in production, because real
API servers end watches and report errors through them. Here you fix a Pod
monitor so it copes with both.

**New here:** closed-channel spinning, `ERROR` events, `apierrors.FromObject`. (Multi-value `case` and `delete` on a map are practice.)

## Learn

### Recap

As in *Watching resources: streaming Pod events*: `w.ResultChan()` delivers
`watch.Event` values with a `Type` and an `Object`, and `w.Stop()` ends the
watch. `pod, ok := ev.Object.(*corev1.Pod)` reaches the Pod. A watch loop
`select`s on `ctx.Done()` and the result channel.

### A closed channel never blocks

A receive from a closed channel returns the zero value straight away, every
time. For a watch that is `watch.Event{}`: `Type` is `""`, `Object` is `nil`.

A loop that ignores this wakes millions of times a second, does nothing and
never returns. One CPU core sits at 100% and the program looks frozen. The fix
is the two-value receive, which works in a `select` `case` too:

```go
case line, ok := <-lines:
	if !ok {
		// lines is closed
	}
```

### ERROR events

When a watch fails, the API server sends one `ERROR` event and closes the
channel. Its object is a `*metav1.Status` with an HTTP code, a *reason* and a
message. `apierrors.FromObject` turns it into an ordinary error that works
with the `apierrors` helpers:

```go
err := apierrors.FromObject(status)
fmt.Println(err)                              // too old resource version: 5 (12)
fmt.Println(apierrors.IsResourceExpired(err)) // true
fmt.Println(apierrors.ReasonForError(err))    // Expired
```

These helpers still work after you wrap the error with `%w`, as in *Wrapping
errors*. With `%v` they would not.

### Keeping state from events

Events only describe changes, so a watcher keeps its own picture of the
cluster. A map from Pod name to phase answers "how many Pods are Running?".
The phase is `pod.Status.Phase`, a `corev1.PodPhase` such as
`corev1.PodRunning`.

The built-in `delete(m, key)` removes a key from a map (no error if it is
missing). A `switch` `case` can list several values, separated by commas:

```go
switch day {
case "Sat", "Sun":
	fmt.Println("weekend")
case "Mon":
	fmt.Println("ugh")
}
```

> **Why is it written like this?** Go needs the comma list for "any of
> these" because, unlike C, a `case` never falls through into the next one.
> Values that match no `case` simply do nothing.

### Why raw watches are fragile

- Connections drop, and the API server ends every watch after a few minutes
  on purpose. You must start a new one.
- Every object has a *resourceVersion*. To restart without missing changes,
  watch from the last one you saw: `metav1.ListOptions{ResourceVersion: rv}`.
- Ask for one that is too old and you get an `ERROR` with reason `Expired`
  (HTTP 410 Gone). Then you must `List` again and watch from there.

client-go's *informers* handle all of this for you. They are next.

## Your task

The `payments` team's rollout monitor follows the Pods in their namespace and
counts how many are Running. A teammate wrote it, and it misbehaves in
production.

### 1. Finish `Apply`

**Goal:** make `Apply` update `phases` for `MODIFIED` and `DELETED` events, not just `ADDED`.

`phases` maps Pod name to phase:

| Event | What to do |
|---|---|
| `ADDED`, `MODIFIED` | set `phases[pod.Name]` to the Pod's phase |
| `DELETED` | remove the Pod from the map |
| anything else (`BOOKMARK`, `ERROR`, not a Pod) | leave the map alone |

```go
phases := map[string]corev1.PodPhase{"checkout-1": corev1.PodPending}
Apply(phases, watch.Event{Type: watch.Modified, Object: checkout1}) // checkout1 is now Running
// phases => map[checkout-1:Running]
Apply(phases, watch.Event{Type: watch.Deleted, Object: checkout1})
// phases => map[]
```

**Think it through:**
- The existing `switch` has one `case`. Does `MODIFIED` need new code, or just a place in an existing `case`?

### 2. Stop spinning on a closed watch

**Goal:** when the watch channel is closed, make `Follow` return the phases so far and `ErrWatchClosed`.

After an API server upgrade the monitor pinned a CPU core at 100% and never
reported again: `Follow` never checks whether the channel is closed.

```go
// w delivers ADDED checkout-1 (Pending), ADDED checkout-2 (Pending) and
// MODIFIED checkout-1 (Running), then the server closes it.
Follow(ctx, w)
// => map[checkout-1:Running checkout-2:Pending], ErrWatchClosed
```

The tests check with `errors.Is(err, ErrWatchClosed)`. If `Follow` is still
running 2s after the watch closed, the test fails.

**Think it through:**
- Which line receives from the channel, and which form of receive also tells you it is closed?

### 3. Report ERROR events

**Goal:** when an `ERROR` event arrives, make `Follow` return the phases so far and the API error, prefixed with `watch error: `.

- Turn the event's object into an error with `apierrors.FromObject(ev.Object)`.
- Wrap it with `%w` after the fixed text `watch error: ` (note the space).
- Do this before the event reaches `Apply`.
- Add `"fmt"` and `apierrors "k8s.io/apimachinery/pkg/api/errors"` to the imports.

```text
watch error: too old resource version: 5 (12)
watch error: etcd leader changed
```

The tests check that text, and that `apierrors.ReasonForError(err)` still
finds the reason (`Expired`, or `InternalError` for a server error).

**Think it through:**
- Which field of `ev` tells you it is an ERROR event, and which constant do you compare it with?
- Why must it be `%w` and not `%v`?

## Learn more

- [Kubernetes: Resource versions](https://kubernetes.io/docs/reference/using-api/api-concepts/#resource-versions)
- [Kubernetes: Efficient detection of changes](https://kubernetes.io/docs/reference/using-api/api-concepts/#efficient-detection-of-changes)
- [pkg.go.dev: k8s.io/apimachinery/pkg/api/errors](https://pkg.go.dev/k8s.io/apimachinery/pkg/api/errors)
- [Go spec: Receive operator](https://go.dev/ref/spec#Receive_operator)
