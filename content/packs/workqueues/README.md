# Workqueues: keys, dedup and the worker loop

Every real Kubernetes controller sits on a *workqueue*. Event handlers drop
small keys into the queue, and workers take them out one at a time and do the
slow work. Here you build both ends.

**New here:** object keys (`namespace/name`), `workqueue.TypedRateLimitingInterface`, `Get`/`Done`, the worker loop.

## Learn

### Why not do the work in the event handler?

In *Informers: list, watch and cache* your handlers ran code for every event.
Real controllers only put a note in a queue and return:

| Problem | What the queue does |
|---|---|
| Handlers must be fast, or events pile up. | `Add` is instant; a worker does the slow work later. |
| One object can change ten times in a second. | Ten events for one object collapse into one queued key (*deduplication*). |
| Two workers on the same object fight each other. | A key being processed is never handed to a second worker. |
| The API server can fail partway through. | A failed key can be retried later. |

### Keys: `namespace/name`

The queue holds a *key*, not the object. `"shop/web"` is the object `web` in
namespace `shop`. Objects outside any namespace, such as a Namespace itself,
have just a name: `"monitoring"`.

The worker uses the key to read the newest copy of the object, so it never
works on stale data. Two helpers in `k8s.io/client-go/tools/cache` convert
both ways:

```go
key, err := cache.MetaNamespaceKeyFunc(pod)   // pod "db-0" in "data"
// key => "data/db-0"

ns, name, err := cache.SplitMetaNamespaceKey("data/db-0")
// ns => "data", name => "db-0"
ns, name, err = cache.SplitMetaNamespaceKey("kube-system")
// ns => "", name => "kube-system"
```

`MetaNamespaceKeyFunc` takes `any` and returns an error if the value isn't a
Kubernetes object (a plain string, say). `SplitMetaNamespaceKey` returns an
error for a key with more than one `/`, such as `"a/b/c"`.

### The workqueue

The queue lives in `k8s.io/client-go/util/workqueue`. It is generic, as in
*Generics*: `[string]` says the items are string keys.

```go
q := workqueue.NewTypedRateLimitingQueue(
	workqueue.DefaultTypedControllerRateLimiter[string](),
)
defer q.ShutDown()
```

The type of `q` is `workqueue.TypedRateLimitingInterface[string]`. The *rate
limiter* decides how long a failed key waits before a retry; you'll use it in
*Workqueues: retries with backoff*.

> **Why is it written like this?** Go requires `[string]` on the rate limiter,
> because it has no argument to infer the type from. The queue infers it from
> the rate limiter, so leaving `[string]` off there is convention (your editor
> flags it if you add it).

| Method | What it does |
|---|---|
| `q.Add(key)` | Queue `key`. If it is already waiting, nothing changes. |
| `q.Len()` | How many keys are waiting (not counting ones being processed). |
| `q.Get()` | Take the next key. **Blocks** until there is one. Returns two values: `(key, shutdown)`. |
| `q.Done(key)` | Say you have finished with `key`. **Always call it.** |
| `q.ShutDown()` | Stop the queue. Once it is empty, `Get` returns `shutdown == true`. |

### Why `Done` matters

While a key is being processed, an `Add` for it only sets a "changed again"
mark. The key goes back into the queue when you call `Done`. Forget `Done`
and that object is never processed again.

`defer` makes sure something runs however the function returns, even from an
early `return`:

```go
f, err := os.Open(path)
if err != nil {
	return err
}
defer f.Close() // runs on every return below this line
```

Put the `defer` right after you get the thing, before any early return.

### The worker loop

A worker processes one key per turn until the queue shuts down:

```go
func RunWorker(q workqueue.TypedRateLimitingInterface[string], handle func(namespace, name string)) {
	for ProcessNextItem(q, handle) {
	}
}
```

`for cond() {}` keeps calling `cond` until it returns `false`. Controllers
start a few of these in goroutines, as in *Worker pools*.

> **Why is it written like this?** `handle func(namespace, name string)` is a
> parameter whose type is a function, so callers choose what "handling" does.
> Go requires the `func(...)` type; the names `handle`, `namespace`, `name`
> are the author's choice.

## Your task

The `shop` team is writing a controller that audits Deployments. The queue,
the worker loop and `KeyFor` exist. Fill in the pieces between them.

`KeyFor` (already written) returns the key from `cache.MetaNamespaceKeyFunc(obj)`,
error included:

```go
KeyFor(web)        // the Deployment "web" in "shop"   => "shop/web", nil
KeyFor(monitoring) // the Namespace "monitoring"       => "monitoring", nil
KeyFor("shop/web") // a string is not an object        => "", error
```

### 1. Enqueue a batch of events

**Goal:** make `EnqueueAll` add the key of every object in `objs` to `q`, in order, and return `nil`.

- Get each key with `KeyFor`.
- If an object has no key, return that error straight away. Keys added before
  it stay in the queue.

```go
EnqueueAll(q, []any{web, web, web})
// => nil, and q.Len() == 1: three events, one key "shop/web"

EnqueueAll(q, []any{web, "not an object", api})
// => an error, and the queue holds only "shop/web"
```

**Think it through:**
- What does `KeyFor` return, and what do you do with each value?
- Do you need to remove duplicates yourself, or does the queue do it?

### 2. Process one key

**Goal:** make `ProcessNextItem` take one key from `q`, pass its namespace and name to `handle`, and report whether the worker should keep going.

In order:

1. Get the next key. If the queue says it is shut down, return `false` and do nothing else.
2. Make sure the key is marked done however the function returns.
3. Split the key into namespace and name. If that fails, return `true` without
   calling `handle`. A malformed key will never work, so the worker skips it.
4. Call `handle` with the namespace and name, and return `true`.

```go
q.Add("shop/web")
ProcessNextItem(q, handle) // calls handle("shop", "web"), returns true
q.Add("monitoring")
ProcessNextItem(q, handle) // calls handle("", "monitoring"), returns true
q.Add("shop/web/extra")
ProcessNextItem(q, handle) // doesn't call handle, returns true
q.ShutDown()
ProcessNextItem(q, handle) // queue is empty and shut down: returns false
```

The tests also check that each key can be added again afterwards (proof that
you called `Done`), and that `RunWorker` handles every key and returns after
`ShutDown`.

**Think it through:**
- `Get` returns two values. Which one tells you to stop?
- Which line from *Why `Done` matters* belongs right after the shutdown check?

## Learn more

- [pkg.go.dev: k8s.io/client-go/util/workqueue](https://pkg.go.dev/k8s.io/client-go/util/workqueue)
- [pkg.go.dev: cache.MetaNamespaceKeyFunc](https://pkg.go.dev/k8s.io/client-go/tools/cache#MetaNamespaceKeyFunc)
- [sample-controller: controller.go](https://github.com/kubernetes/sample-controller/blob/master/controller.go)
- [client-go under the hood](https://github.com/kubernetes/sample-controller/blob/master/docs/controller-client-go.md)
