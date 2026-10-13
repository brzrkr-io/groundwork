# Watching resources: streaming Pod events

Dashboards, chat bots and controllers all need to know the moment something
changes in a cluster. In this exercise you ask the API server to stream Pod
changes to you, instead of asking over and over again.

## Learn

### Polling versus watching

With a `List` request you could ask "which Pods are there?" every few
seconds. That is *polling*: it notices changes late, and it is
wasteful on a big cluster. Instead, a client can open a *watch*: one
long-lived request during which the API server sends an *event* every time an
object changes. `kubectl get pods --watch` does exactly this.

Every event has a type and an object:

| Type | Meaning | Object |
|---|---|---|
| `ADDED` | An object was created. | The new object. |
| `MODIFIED` | An object changed. | The new version. |
| `DELETED` | An object was deleted. | Its last version. |
| `BOOKMARK` | A progress marker. Nothing changed. | An almost empty object. |
| `ERROR` | The watch failed. | A `*metav1.Status` that describes the error. |

A watch started without options first sends an `ADDED` event for every object
that already exists. After that come the changes, as they happen.

### Watching in Go

The typed clients from *The clientset: get, create and delete* have a `Watch`
method. It returns a `watch.Interface`, from package
`k8s.io/apimachinery/pkg/watch`:

```go
w, err := cs.CoreV1().Pods("shop").Watch(ctx, metav1.ListOptions{})
if err != nil {
	return err
}
defer w.Stop()

for ev := range w.ResultChan() {
	fmt.Println(ev.Type) // ADDED, MODIFIED, ...
}
```

`watch.Interface` has two methods:

- `ResultChan()` returns a channel of `watch.Event` values. The channel is
  closed when the watch ends.
- `Stop()` ends the watch and closes the channel. Always call it when you are
  done, usually with `defer`, so the API server can drop the connection.

`ev.Type` is a `watch.EventType`, a string type with the constants
`watch.Added`, `watch.Modified`, `watch.Deleted`, `watch.Bookmark` and
`watch.Error`. `string(ev.Type)` turns it into a plain string.

### What is in `ev.Object`?

`ev.Object` is a `runtime.Object`, an interface that every Kubernetes object
implements. To read Pod fields, use a type assertion with the comma-ok form,
as in *Type assertions and switches*:

```go
pod, ok := ev.Object.(*corev1.Pod)
if !ok {
	// not a Pod: an ERROR event carries a *metav1.Status instead
}
fmt.Println(pod.Namespace, pod.Name)
```

Without the `ok`, the first ERROR event would make your program panic.

### Stopping when you are told to

`range` waits for the next event forever, even after your context is
cancelled. To stop on time, `select` on both channels, as in *select and
timeouts* and *Context and cancellation*:

```go
for {
	select {
	case <-ctx.Done():
		return nil
	case ev, ok := <-w.ResultChan():
		if !ok {
			return nil // the server closed the watch
		}
		handle(ev)
	}
}
```

`range` notices a closed channel for you. A plain receive does not: on a
closed channel it gets the zero value straight away, again and again. `ok` is
`false` once the channel is closed.

### Watches in tests

The fake clientset supports watches: each `Create`, `Update` or `Delete` on
the fake sends an event to its watchers. Package `watch` also has fakes that a
test drives by hand, such as `watch.NewFake()`, with `fw.Add(pod)`,
`fw.Modify(pod)`, `fw.Delete(pod)` and `fw.Stop()`. The tests use both, so
you don't have to.

## Your task

The `shop` team wants a bot that posts Pod changes to their chat channel. You
write the part that watches.

### 1. Describe an event

Make `Describe` return the event type, a space, and the Pod's
`namespace/name`:

```go
Describe(watch.Event{Type: watch.Added, Object: web1}) // web1 is Pod web-1 in namespace shop
// => "ADDED shop/web-1"
```

If `ev.Object` is not a `*corev1.Pod`, return just the type:

```go
Describe(watch.Event{Type: watch.Error, Object: &metav1.Status{Message: "too old resource version"}})
// => "ERROR"
```

Add `"fmt"` and `corev1 "k8s.io/api/core/v1"` to the imports.

### 2. Collect a few events

Make `Collect` read events from `w.ResultChan()` in order and describe each
one with `Describe`. Stop as soon as you have `n` descriptions (`n` is always
at least 1), or when the channel is closed, whichever comes first. Call
`w.Stop()` before you return, then return the descriptions:

```go
// w delivers ADDED shop/web-1, ADDED shop/web-2, MODIFIED shop/web-1
Collect(w, 2)
// => []string{"ADDED shop/web-1", "ADDED shop/web-2"}, and w is stopped
```

If the channel closes after those three events, `Collect(w, 5)` returns all
three.

### 3. Watch Pods until cancelled

Make `WatchPods` watch the Pods in `namespace` and call `handle` with every
event, in order:

- Start the watch with `metav1.ListOptions{}`, and add
  `metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"` to the imports. If `Watch`
  returns an error, return it.
- Stop the watch when `WatchPods` returns.
- Return `nil` when `ctx` is cancelled, and also when the server closes the
  channel.

```go
// Pod web-1 already exists in shop. Then api-7 is created, updated and deleted.
WatchPods(ctx, cs, "shop", handle)
// handle receives: ADDED web-1, ADDED api-7, MODIFIED api-7, DELETED api-7
// after cancel(), WatchPods returns nil
```

The tests run `WatchPods` in a goroutine, change Pods through the fake
clientset, then cancel the context. Two more tests make the fake server close
the watch by itself.

## Learn more

- [Kubernetes: Efficient detection of changes](https://kubernetes.io/docs/reference/using-api/api-concepts/#efficient-detection-of-changes)
- [pkg.go.dev: k8s.io/apimachinery/pkg/watch](https://pkg.go.dev/k8s.io/apimachinery/pkg/watch)
- [pkg.go.dev: client-go fake clientset](https://pkg.go.dev/k8s.io/client-go/kubernetes/fake)
- [Go by Example: Closing channels](https://gobyexample.com/closing-channels)
