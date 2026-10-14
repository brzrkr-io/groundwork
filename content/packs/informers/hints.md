## Task 1: Create the factory

The `informers` package has a function that builds a shared informer factory.
It needs the clientset and a resync period, and `0` turns resyncing off.

### Closer

Call `informers.NewSharedInformerFactory(...)` with `cs` first and the resync
period second, and return what it gives you.

### Answer

```go
return informers.NewSharedInformerFactory(cs, 0)
```

## Task 2: Record Deployment events

First give `Record` the same lock that `Events` has. Then get the typed
Deployment informer from the factory and register one handler function per
event: added, updated and deleted.

### Closer

`Record` starts with `r.mu.Lock()` and `defer r.mu.Unlock()`. The informer is
`factory.Apps().V1().Deployments().TypedInformer()`, and you register handlers
with:

```go
_, err := informer.AddTypedEventHandler(cache.TypedResourceEventHandlerFuncs[*appsv1.Deployment]{
	AddFunc:    func(d *appsv1.Deployment) { /* rec.Record(...) */ },
	UpdateFunc: func(_, d *appsv1.Deployment) { /* ... */ },
	DeleteFunc: func(deleted cache.DeletedObject[*appsv1.Deployment]) { /* deleted.GetKey() */ },
})
```

### Answer

```go
func (r *Recorder) Record(event string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, event)
}

func WatchDeployments(factory informers.SharedInformerFactory, rec *Recorder) error {
	informer := factory.Apps().V1().Deployments().TypedInformer()
	_, err := informer.AddTypedEventHandler(cache.TypedResourceEventHandlerFuncs[*appsv1.Deployment]{
		AddFunc: func(d *appsv1.Deployment) {
			rec.Record("added " + d.Namespace + "/" + d.Name)
		},
		UpdateFunc: func(_, d *appsv1.Deployment) {
			rec.Record("updated " + d.Namespace + "/" + d.Name)
		},
		DeleteFunc: func(deleted cache.DeletedObject[*appsv1.Deployment]) {
			rec.Record("deleted " + deleted.GetKey())
		},
	})
	return err
}
```

Add the imports `appsv1 "k8s.io/api/apps/v1"` and
`"k8s.io/client-go/tools/cache"`; the second parameter of `UpdateFunc` is the
new object.

## Task 3: Start and wait for sync

Two factory methods do the work: one starts the informers, the other blocks
until their caches are full. Both take `ctx`.

### Closer

Call `factory.StartWithContext(ctx)`, then return
`factory.WaitForCacheSyncWithContext(ctx)` followed by `.AsError()`.

### Answer

```go
factory.StartWithContext(ctx)
return factory.WaitForCacheSyncWithContext(ctx).AsError()
```

`AsError()` is `nil` once every informer has synced, and wraps the context's
error if `ctx` ended first.
