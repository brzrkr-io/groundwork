# Workqueues: keys, dedup and the worker loop

Every real Kubernetes controller sits on a *workqueue*. Event handlers drop
small keys into the queue, and workers take them out one at a time and do the
slow work. In this exercise you build both ends.

## Learn

### Why not do the work in the event handler?

In *Informers: list, watch and cache* your event handlers ran code for every Add,
Update and Delete. Real controllers do almost nothing there. They put a
note into a queue and return, for four reasons:

| Problem | What the queue does |
|---|---|
| Handlers must be fast, or events pile up behind them. | `Add` is instant. The slow work happens later in a worker. |
| One Deployment can change ten times in a second. | Ten events for the same object collapse into one queued key (*deduplication*). |
| Two workers working on the same object at once fight each other. | A key that is being processed is never handed to a second worker. |
| The API server can fail half way through your work. | A failed key can go back into the queue and be retried later. |

### Keys: `namespace/name`

The queue doesn't hold whole objects. It holds a *key*, a string that names
one object: `"shop/web"` is the Deployment `web` in the namespace `shop`.
Objects that don't live in a namespace, such as a Namespace itself, have just
a name: `"monitoring"`. The worker uses the key to read the newest copy of the
object, so it never works on stale data.

client-go has helpers for both directions, in `k8s.io/client-go/tools/cache`:

```go
key, err := cache.MetaNamespaceKeyFunc(d) // d is *appsv1.Deployment in shop
// key => "shop/web"

ns, name, err := cache.SplitMetaNamespaceKey("shop/web")
// ns => "shop", name => "web"
ns, name, err = cache.SplitMetaNamespaceKey("monitoring")
// ns => "", name => "monitoring"
```

`MetaNamespaceKeyFunc` takes `any` (handlers get objects as `any`) and returns
an error if the value isn't a Kubernetes object, like a plain string.
`SplitMetaNamespaceKey` returns an error for a key with more than one `/`,
such as `"shop/web/extra"`.

### The workqueue

The queue lives in `k8s.io/client-go/util/workqueue`. It is generic, as in
*Generics*: `[string]` says the items are string keys.

```go
q := workqueue.NewTypedRateLimitingQueue(
	workqueue.DefaultTypedControllerRateLimiter[string](),
)
defer q.ShutDown()
```

The rate limiter has no argument to learn the type from, so it needs
`[string]`. `NewTypedRateLimitingQueue` gets its type from the rate limiter,
so Go infers it and you leave `[string]` off. Your editor flags type
arguments that Go could infer on its own.

The *rate limiter* decides how long a failed key waits before a retry. You
will use it in *Workqueues: retries with backoff*. The type of `q` is
`workqueue.TypedRateLimitingInterface[string]`.

| Method | What it does |
|---|---|
| `q.Add(key)` | Queue `key`. If it is already waiting, nothing changes. |
| `q.Len()` | How many keys are waiting (not counting ones being processed). |
| `q.Get()` | Take the next key. **Blocks** until there is one. Returns `(key, shutdown)`. |
| `q.Done(key)` | Say you have finished with `key`. **Always call it.** |
| `q.ShutDown()` | Stop the queue. Waiting `Get` calls return `shutdown == true` once the queue is empty. |
| `q.ShutDownWithDrain()` | Like `ShutDown`, but also waits until every key handed out is `Done`. |

`Done` matters more than it looks. While a key is being processed, an `Add`
for it only sets a "changed again" mark. The key goes back into the queue when
you call `Done`. Forget `Done` and that object is never processed again. Use
`defer q.Done(key)` straight after `Get`, so every return path calls it.

### The worker loop

A worker is a loop that processes one key per turn until the queue shuts
down:

```go
func RunWorker(q workqueue.TypedRateLimitingInterface[string], handle func(namespace, name string)) {
	for ProcessNextItem(q, handle) {
	}
}
```

`for cond() {}` keeps calling `cond` until it returns `false`. Controllers
start a few of these in goroutines, as in *Worker pools*.

## Your task

The `shop` team is writing a controller that audits Deployments. The queue,
the worker loop and `KeyFor` exist. Fill in the pieces between them.

`KeyFor` returns the key from `cache.MetaNamespaceKeyFunc(obj)`, error
included:

```go
KeyFor(web)        // web is the Deployment "web" in "shop"
// => "shop/web", nil
KeyFor(monitoring) // the Namespace "monitoring"
// => "monitoring", nil
KeyFor("shop/web") // a string is not a Kubernetes object
// => "", error
```

### 1. Enqueue a batch of events

Make `EnqueueAll` add the key of every object in `objs` to `q`, in order.
Use `KeyFor`. If an object has no key, return that error straight away. Keys
added before it stay in the queue.

```go
EnqueueAll(q, []any{web, web, web})
// => nil, and q.Len() == 1: three events, one key "shop/web"

EnqueueAll(q, []any{web, "not an object", api})
// => error, and the queue holds only "shop/web"
```

### 2. Process one key

Make `ProcessNextItem`:

1. Call `q.Get()`. If `shutdown` is `true`, return `false` and do nothing else.
2. `defer q.Done(key)`.
3. Split the key with `cache.SplitMetaNamespaceKey`. If that fails, return
   `true` without calling `handle`. A malformed key will never work, so the
   worker skips it and moves on.
4. Call `handle(namespace, name)` and return `true`.

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

The tests also check that each key can be added again after
`ProcessNextItem` (so you called `Done`), and that `RunWorker` handles every
key and returns after `ShutDown`.

## Learn more

- [pkg.go.dev: k8s.io/client-go/util/workqueue](https://pkg.go.dev/k8s.io/client-go/util/workqueue)
- [pkg.go.dev: cache.MetaNamespaceKeyFunc](https://pkg.go.dev/k8s.io/client-go/tools/cache#MetaNamespaceKeyFunc)
- [sample-controller: controller.go](https://github.com/kubernetes/sample-controller/blob/master/controller.go)
- [client-go under the hood](https://github.com/kubernetes/sample-controller/blob/master/docs/controller-client-go.md)
