# Watching resources: streaming Pod events

Dashboards, chat bots and controllers need to know the moment something
changes in a cluster. Here you ask the API server to stream Pod changes to
you, instead of asking over and over.

**New here:** watches, `watch.Interface`, `watch.Event`, reading a channel until it closes or a context ends.

## Learn

### Polling versus watching

Calling `List` every few seconds is *polling*: it notices changes late and
wastes work. A *watch* is one long-lived request during which the API server
sends an *event* every time an object changes. `kubectl get pods --watch` does
this.

| Type | Meaning | Object |
|---|---|---|
| `ADDED` | An object was created. | The new object. |
| `MODIFIED` | An object changed. | The new version. |
| `DELETED` | An object was deleted. | Its last version. |
| `BOOKMARK` | A progress marker. Nothing changed. | An almost empty object. |
| `ERROR` | The watch failed. | A `*metav1.Status`, not a Pod. |

A watch started with empty options first sends `ADDED` for every object that
already exists, then the changes as they happen.

### Watching in Go

The typed clients from *The clientset: get, create and delete* have a `Watch`
method. It takes `ctx` and a `metav1.ListOptions`, and returns a
`watch.Interface` (package `k8s.io/apimachinery/pkg/watch`) and an error.

`watch.Interface` has two methods:

- `ResultChan()` returns a channel of `watch.Event` values. It is closed when
  the watch ends.
- `Stop()` ends the watch and closes the channel. Always call it when you are
  done, usually with `defer`.

```go
w, err := cs.CoreV1().Services("shop").Watch(ctx, metav1.ListOptions{})
// ... check err, then: defer w.Stop()
```

`ev.Type` is a `watch.EventType`, a string type with constants `watch.Added`,
`watch.Modified`, `watch.Deleted`, `watch.Bookmark` and `watch.Error`.
`string(ev.Type)` turns it into a plain string; `%s` prints it directly.

### What is in `ev.Object`?

`ev.Object` is a `runtime.Object`, an interface every Kubernetes object
implements. To read Service fields, use a comma-ok type assertion, as in
*Type assertions and switches*:

```go
svc, ok := ev.Object.(*corev1.Service)
if !ok {
	// not a Service (an ERROR event, say)
}
```

> **Why is it written like this?** The `, ok` form is your choice, but without
> it a wrong type makes the program panic. The `*` is required: the watch
> sends pointers, and `corev1.Service` without `*` is a different type.

### Leaving a loop over a channel

`for x := range ch` reads until `ch` is closed. `break` leaves the loop early:

```go
for job := range jobs {
	done = append(done, job)
	if len(done) == 3 {
		break
	}
}
```

### Stopping when you are told to

`range` waits forever for the next value, even after your context is
cancelled. To wait on two things at once, put a `select` inside a `for` loop,
as in *select and timeouts*:

```go
for {
	select {
	case <-quit:
		return
	case msg, ok := <-inbox:
		if !ok {
			return // inbox was closed
		}
		fmt.Println(msg)
	}
}
```

A plain receive on a closed channel gets the zero value straight away, again
and again. The two-value form `v, ok := <-ch` sets `ok` to `false` once the
channel is closed.

> **Why is it written like this?** `for { ... }` with no condition is Go's
> infinite loop; the `return` statements are the only way out. Names like
> `msg` and `ok` are the author's choice (`ok` is the convention).

### Watches in tests

The fake clientset sends an event to watchers for each `Create`, `Update` or
`Delete`. Package `watch` also has hand-driven fakes such as `watch.NewFake()`.
The tests set these up for you.

## Your task

The `shop` team wants a bot that posts Pod changes to their chat channel. You
write the part that watches.

### 1. Describe an event

**Goal:** make `Describe` return the event type, one space, and the Pod's `namespace/name`.

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

**Think it through:**
- Three values go into the line: type, namespace, name. Which fixed text sits between them (a space, a `/`)?
- What do you return when the assertion's `ok` is `false`?

### 2. Collect a few events

**Goal:** make `Collect` describe events from `w.ResultChan()` in order and return the descriptions once it has `n` of them or the channel is closed.

- Use `Describe` for each event. `n` is always at least 1.
- Never read more events than you need.
- Call `w.Stop()` before you return, however you leave.

```go
// w delivers ADDED shop/web-1, ADDED shop/web-2, MODIFIED shop/web-1
Collect(w, 2)
// => []string{"ADDED shop/web-1", "ADDED shop/web-2"}, and w is stopped
```

If the channel closes after those three events, `Collect(w, 5)` returns all
three. If it closes with no events, return `nil`.

**Think it through:**
- Which loop ends by itself when the channel closes?
- Which keyword makes `w.Stop()` run on every return path?

### 3. Watch Pods until cancelled

**Goal:** make `WatchPods` watch the Pods in `namespace` and call `handle` with every event, in order, until `ctx` is cancelled or the channel closes.

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

The tests run `WatchPods` in a goroutine, change Pods through the fake, then
cancel the context. Two more tests make the fake server close the watch.

**Think it through:**
- Which two channels must you wait on at the same time? (`ctx.Done()` is one.)
- How do you tell a real event from the zero value of a closed channel?

## Learn more

- [Kubernetes: Efficient detection of changes](https://kubernetes.io/docs/reference/using-api/api-concepts/#efficient-detection-of-changes)
- [pkg.go.dev: k8s.io/apimachinery/pkg/watch](https://pkg.go.dev/k8s.io/apimachinery/pkg/watch)
- [Go by Example: Closing channels](https://gobyexample.com/closing-channels)
- [Go by Example: Select](https://gobyexample.com/select)
