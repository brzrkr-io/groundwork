# Informers: list, watch and cache

Every real Kubernetes controller learns about the cluster through *informers*.
In this exercise you set one up, react to Deployments coming and going, and
wait until its cache is ready.

## Learn

### Why not just watch?

In *Watching resources: closed watches and ERROR events* you saw that a raw watch is fragile. The API server
closes watches whenever it likes, so your code has to notice, list everything
again, and start a new watch. And if you want to know "which Deployments exist
right now?", you have to ask the API server every single time.

### The informer pattern

An *informer* does that bookkeeping for you:

1. It **lists** every object of one kind once, at startup.
2. It **watches** for changes from then on, and quietly re-lists and
   re-watches when the watch drops.
3. It keeps every object in a local in-memory *cache* (client-go calls it the
   *indexer*), so reading is instant and costs the API server nothing.
4. It calls your *event handlers*: "added", "updated" or "deleted".

One program often runs several controllers, and many of them want
Deployments. A *shared informer factory* hands every caller the same informer
per kind, so there is one watch per kind, not one per controller.

```go
import "k8s.io/client-go/informers"

factory := informers.NewSharedInformerFactory(cs, 0) // cs is a kubernetes.Interface
deployments := factory.Apps().V1().Deployments()    // the Deployment informer
```

The `0` is the *resync period*. With a value like `10*time.Minute`, the
informer would replay every cached object to your "updated" handler every ten
minutes, as a safety net. Plenty of controllers don't need that, and `0` turns
it off.

Inside this exercise the package is also called `informers`. That's fine: the
import `informers.` always means client-go's package.

### Event handlers

You register handlers on the typed informer. The handler functions get real
`*appsv1.Deployment` values, so you don't need type assertions:

```go
informer := factory.Apps().V1().Deployments().TypedInformer()
_, err := informer.AddTypedEventHandler(cache.TypedResourceEventHandlerFuncs[*appsv1.Deployment]{
	AddFunc: func(d *appsv1.Deployment) {
		fmt.Println("new deployment", d.Namespace+"/"+d.Name)
	},
	UpdateFunc: func(oldD, newD *appsv1.Deployment) { /* ... */ },
	DeleteFunc: func(deleted cache.DeletedObject[*appsv1.Deployment]) {
		fmt.Println("gone:", deleted.GetKey()) // "shop/web"
	},
})
```

`cache` is `k8s.io/client-go/tools/cache`. A delete handler gets a
`cache.DeletedObject`, because the informer sometimes misses the final state
of a deleted object. Its `GetKey()` always works and returns
`"<namespace>/<name>"`, the *key* the cache stores the object under.

Many tutorials use the older untyped form,
`Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{...})`. Its
handlers receive `any`, so they start with a type assertion. It does the same
job.

Handlers run on the informer's own goroutines. If they touch data that other
code reads, protect it with a mutex, as in *Mutexes*.

### Start, sync, stop

Ask the factory for every informer and register every handler **before**
starting it. The factory only starts informers it already knows about.

```go
ctx, cancel := context.WithCancel(context.Background())
factory.StartWithContext(ctx)  // runs each informer in a goroutine and returns
err := factory.WaitForCacheSyncWithContext(ctx).AsError()
// ... use the cache ...
cancel()           // tells the informers to stop
factory.Shutdown() // waits until their goroutines have finished
```

`StartWithContext` returns straight away. Until the first list has finished,
the cache is empty or half full. `WaitForCacheSyncWithContext` blocks until
every started informer has *synced* (loaded its first list). It returns a
`cache.SyncResult`. Its `AsError()` method gives `nil` on success, or an error
wrapping the context's error if `ctx` ended first.

## Your task

The `shop` team wants a feed of Deployment changes for their chat channel.

### 1. Create the factory

Make `NewFactory` return a shared informer factory for `cs` with a resync
period of `0`.

```go
factory := NewFactory(fake.NewClientset(web, api))
// factory.Apps().V1().Deployments().Lister() sees shop/web and shop/api once started
```

### 2. Record Deployment events

First fix `Recorder.Record`: lock `r.mu` while appending, the same way
`Events` does. Handlers call `Record` from informer goroutines while the
tests call `Events`, so without the lock the race detector fails the test.

Then make `WatchDeployments` register handlers on the factory's Deployment
informer that call `rec.Record` with:

| Event | Recorded string |
|---|---|
| added | `"added <namespace>/<name>"` |
| updated | `"updated <namespace>/<name>"` (use the new object) |
| deleted | `"deleted <namespace>/<name>"` (use `GetKey()`) |

Return the error from `AddTypedEventHandler`. Add the imports
`appsv1 "k8s.io/api/apps/v1"` and `"k8s.io/client-go/tools/cache"`.

```go
// cluster starts with shop/web; later staging/worker is created,
// shop/web is updated and staging/worker is deleted
rec.Events()
// => ["added shop/web", "added staging/worker", "updated shop/web", "deleted staging/worker"]
```

Objects already in the cluster arrive as "added" events too, when the first
list fills the cache.

### 3. Start and wait for sync

Make `StartAndSync` start the factory with `ctx`, wait for the caches to
sync, and return the result of `AsError()`.

```go
err := StartAndSync(ctx, factory)
// => nil, and the Deployment informer's HasSynced() is true

// The API server never answers, and the caller cancels ctx:
err := StartAndSync(ctx, factory)
// => an error wrapping context.Canceled (check it with errors.Is)
```

## Learn more

- [pkg.go.dev: k8s.io/client-go/informers](https://pkg.go.dev/k8s.io/client-go/informers)
- [pkg.go.dev: k8s.io/client-go/tools/cache](https://pkg.go.dev/k8s.io/client-go/tools/cache)
- [Kubernetes: Efficient detection of changes](https://kubernetes.io/docs/reference/using-api/api-concepts/#efficient-detection-of-changes)
- [GitHub: kubernetes/sample-controller](https://github.com/kubernetes/sample-controller)
