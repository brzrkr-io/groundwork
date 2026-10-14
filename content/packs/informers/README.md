# Informers: list, watch and cache

Every real Kubernetes controller learns about the cluster through *informers*.
Here you set one up, react to Deployments coming and going, and wait until its
cache is ready.

**New here:** shared informer factories, typed event handlers, `cache.DeletedObject`, starting and syncing informers.

## Learn

### Why not just watch?

In *Watching resources: closed watches and ERROR events* you saw that a raw
watch is fragile. The API server closes it whenever it likes, and you must
list again and re-watch.

Also, every "which Deployments exist right now?" question costs a call to the
API server.

### What an informer does

An *informer* does that bookkeeping for one kind of object:

1. It **lists** every object once, at startup.
2. It **watches** for changes, and re-lists and re-watches when the watch drops.
3. It keeps every object in a local in-memory *cache*, so reads are instant.
4. It calls your *event handlers*: "added", "updated" or "deleted".

### The shared informer factory

One program often runs several controllers that all want Deployments. A
*shared informer factory* hands every caller the same informer per kind, so
there is one watch per kind.

This is the function that builds one:

```go
func NewSharedInformerFactory(client kubernetes.Interface, defaultResync time.Duration) SharedInformerFactory
```

Read it left to right: the name, then two parameters (each `name Type`), then
the return type. When you *call* it, you pass only values, in that order. You
never write the types or the word `func`.

The *resync period* replays every cached object to your "updated" handler on a
timer, as a safety net. `5*time.Minute` means every five minutes. `0` turns it
off.

You get an informer for one kind by walking the API group and version:

```go
pods := factory.Core().V1().Pods()        // group "core", version v1, kind Pod
pods.Lister()                              // reads from the cache
pods.TypedInformer()                       // where you add handlers
```

Inside this exercise your package is also called `informers`. That's fine:
`informers.` always means client-go's package.

> **Why is it written like this?** The chain `factory.Core().V1().Pods()` is
> client-go's design: each method returns a smaller object. Go requires the
> `()` on every call. Passing `0` instead of a named constant is the author's
> choice for this exercise.

### Event handlers

You register handlers on a typed informer. You fill in a struct with up to
three fields, each holding a function:

```go
informer := factory.Core().V1().Pods().TypedInformer()
_, err := informer.AddTypedEventHandler(cache.TypedResourceEventHandlerFuncs[*corev1.Pod]{
	AddFunc: func(p *corev1.Pod) {
		log.Println("pod appeared:", p.Name)
	},
})
```

The other two fields have these shapes:

| Field | Function shape | Notes |
|---|---|---|
| `AddFunc` | `func(obj *T)` | also fires for objects found by the first list |
| `UpdateFunc` | `func(oldObj, newObj *T)` | the **second** parameter is the new version |
| `DeleteFunc` | `func(deleted cache.DeletedObject[*T])` | call `deleted.GetKey()` |

A delete handler gets a `cache.DeletedObject`, because the informer sometimes
misses the final state of a deleted object. Its `GetKey()` always works and
returns `"<namespace>/<name>"`.

Every object has `.Namespace` and `.Name` fields. They hold the values, so you
join them with fixed text such as `"/"`.

`AddTypedEventHandler` returns two values. Write `_` for one you don't need.

Handlers run on the informer's own goroutines. If they touch shared data,
protect it with a mutex, as in *Mutexes*.

> **Why is it written like this?** `[*corev1.Pod]` is a *type argument*: Go
> requires it so the handlers get real `*corev1.Pod` values. The `cache` and
> `corev1` import names are convention. Many tutorials use the older untyped
> `AddEventHandler`, whose handlers receive `any`; both work.

### Start, sync, stop

Request every informer and register every handler **before** starting. The
factory only starts informers it already knows about.

| Method | What it does |
|---|---|
| `StartWithContext(ctx)` | starts each informer in a goroutine and returns at once |
| `WaitForCacheSyncWithContext(ctx)` | blocks until every started informer has *synced* (loaded its first list); returns a `cache.SyncResult` |
| `SyncResult.AsError()` | `nil` on success, or an error wrapping the context's error if `ctx` ended first |
| `Shutdown()` | after you cancel `ctx`, waits for the informer goroutines to finish |

You can call a method directly on the value another call returns, like
`time.Now().Year()`. No temporary variable needed.

## Your task

The `shop` team wants a feed of Deployment changes for their chat channel.

### 1. Create the factory

**Goal:** make `NewFactory` return a shared informer factory for `cs` that never resyncs.

```go
factory := NewFactory(fake.NewClientset(web, api))
// once started, factory.Apps().V1().Deployments().Lister() sees shop/web and shop/api
```

**Think it through:**
- Which parameter of `NewFactory` is the clientset?
- Which resync period means "off"?

### 2. Record Deployment events

**Goal:** fix `Recorder.Record`, then make `WatchDeployments` record one string per Deployment event.

Step 1: in `Record`, lock `r.mu` while appending, the same way `Events` does.
Without it the race detector fails the test.

Step 2: in `WatchDeployments`, register handlers on the factory's
**Deployment** informer that call `rec.Record` with:

| Event | Recorded string |
|---|---|
| added | `"added <namespace>/<name>"` |
| updated | `"updated <namespace>/<name>"` (from the new object) |
| deleted | `"deleted <namespace>/<name>"` (from `GetKey()`) |

Note the single space after the word. Return the error from
`AddTypedEventHandler`.

Add the imports `appsv1 "k8s.io/api/apps/v1"` and
`"k8s.io/client-go/tools/cache"`. The Deployment type is `appsv1.Deployment`.

```go
// cluster starts with shop/web; then staging/worker is created,
// shop/web is updated, and staging/worker is deleted
rec.Events()
// => ["added shop/web", "added staging/worker", "updated shop/web", "deleted staging/worker"]
```

Objects already in the cluster arrive as "added" events when the first list
fills the cache.

**Think it through:**
- Deployments live in the `apps` group, version `v1`.
- Which parts of each string are fixed text, and which come from the object?
- In `UpdateFunc`, which of the two parameters is the new object?

### 3. Start and wait for sync

**Goal:** make `StartAndSync` start the factory with `ctx`, wait for the caches to sync, and return the result of `AsError()`.

```go
err := StartAndSync(ctx, factory)
// => nil, and the Deployment informer's HasSynced() is true

// the API server never answers, and the caller has cancelled ctx:
err := StartAndSync(ctx, factory)
// => an error wrapping context.Canceled (errors.Is(err, context.Canceled) is true)
```

Don't cancel or shut down anything yourself; the caller does that.

**Think it through:**
- Look at the *Start, sync, stop* table: which two methods, in which order?
- What does the wait method return, and what do you call on it?

## Learn more

- [pkg.go.dev: k8s.io/client-go/informers](https://pkg.go.dev/k8s.io/client-go/informers)
- [pkg.go.dev: k8s.io/client-go/tools/cache](https://pkg.go.dev/k8s.io/client-go/tools/cache)
- [Kubernetes: Efficient detection of changes](https://kubernetes.io/docs/reference/using-api/api-concepts/#efficient-detection-of-changes)
- [GitHub: kubernetes/sample-controller](https://github.com/kubernetes/sample-controller)
