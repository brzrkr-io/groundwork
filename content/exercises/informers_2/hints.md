## Task 1: ask first, start second

`Lister()` registers the Deployment informer with the factory.
`StartWithContext` only starts informers that are already registered. Move
the `Lister()` call to the top, keep it in a variable, and return that
variable at the end.

## Task 2: selector, list, sum

Build the selector with `labels.SelectorFromSet(labels.Set{"team": team})`.
Pass it to `lister.Deployments(namespace).List`. Loop over the result. When
`d.Spec.Replicas` is `nil` add 1, otherwise add `*d.Spec.Replicas`.

## Task 3: copy before you edit

```go
cached, err := lister.Deployments(namespace).Get(name)
if err != nil {
	return err
}
d := cached.DeepCopy()
d.Spec.Paused = true
```

## Task 2: almost the answer

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
