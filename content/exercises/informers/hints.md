## Task 1: one call

`informers.NewSharedInformerFactory` takes the clientset and the resync
period. Return what it gives you:

```go
return informers.NewSharedInformerFactory(cs, 0)
```

## Task 2: lock, then register

`Record` needs the same two lines that start `Events`:

```go
r.mu.Lock()
defer r.mu.Unlock()
```

In `WatchDeployments`, get the typed informer with
`factory.Apps().V1().Deployments().TypedInformer()`, then call
`AddTypedEventHandler` with a `cache.TypedResourceEventHandlerFuncs[*appsv1.Deployment]`
that sets `AddFunc`, `UpdateFunc` and `DeleteFunc`. Build each string with `+`.

## Task 2: almost the answer

```go
_, err := informer.AddTypedEventHandler(cache.TypedResourceEventHandlerFuncs[*appsv1.Deployment]{
	AddFunc: func(d *appsv1.Deployment) {
		rec.Record("added " + d.Namespace + "/" + d.Name)
	},
	UpdateFunc: func(oldD, newD *appsv1.Deployment) {
		rec.Record("updated " + newD.Namespace + "/" + newD.Name)
	},
	DeleteFunc: func(deleted cache.DeletedObject[*appsv1.Deployment]) {
		rec.Record("deleted " + deleted.GetKey())
	},
})
return err
```

## Task 3: start, then wait

```go
factory.StartWithContext(ctx)
return factory.WaitForCacheSyncWithContext(ctx).AsError()
```
