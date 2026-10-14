## Task 1: Fix StartLister

Asking for `Lister()` also tells the factory you want the Deployment informer.
The factory only starts informers it knows about when `StartWithContext` runs.
Which step has to come first?

### Closer

Move the `Lister()` call to the top of the function, keep it in a variable,
and return that variable at the end:

```go
lister := factory.Apps().V1().Deployments().Lister()
factory.StartWithContext(ctx)
// ... wait for sync ...
```

### Answer

```go
lister := factory.Apps().V1().Deployments().Lister()
factory.StartWithContext(ctx)
if err := factory.WaitForCacheSyncWithContext(ctx).AsError(); err != nil {
	return nil, err
}
return lister, nil
```

## Task 2: Finish TeamReplicas

Build a label selector for `team=<team>`, list the matching Deployments in
`namespace` from the lister, and add up their replicas. Remember that
`Spec.Replicas` is a pointer and may be `nil`.

### Closer

The selector is `labels.SelectorFromSet(labels.Set{"team": team})`, and the
list call is `lister.Deployments(namespace).List(selector)`. In the loop, add
`1` when `d.Spec.Replicas == nil`, otherwise add `*d.Spec.Replicas`.

### Answer

```go
selector := labels.SelectorFromSet(labels.Set{"team": team})
deployments, err := lister.Deployments(namespace).List(selector)
if err != nil {
	return 0, err
}
var total int32
for _, d := range deployments {
	if d.Spec.Replicas == nil {
		total++
	} else {
		total += *d.Spec.Replicas
	}
}
return total, nil
```

Add the import `"k8s.io/apimachinery/pkg/labels"`.

## Task 3: Fix Pause

The object from `Get` lives inside the shared cache, so changing it changes
what everyone else sees. Make a copy and change the copy instead.

### Closer

Rename the value from `Get` to `cached`, then `d := cached.DeepCopy()` before
setting `d.Spec.Paused`.

### Answer

```go
cached, err := lister.Deployments(namespace).Get(name)
if err != nil {
	return err
}
d := cached.DeepCopy()
d.Spec.Paused = true
_, err = cs.AppsV1().Deployments(namespace).Update(ctx, d, metav1.UpdateOptions{})
return err
```

`DeepCopy` copies every nested field, so the cached object stays untouched.
