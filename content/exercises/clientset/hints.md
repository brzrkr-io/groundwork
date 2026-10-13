## Pick the typed client

Deployments live in the `apps/v1` group, so start with
`cs.AppsV1().Deployments(namespace)`. ConfigMaps are in the core group:
`cs.CoreV1().ConfigMaps(namespace)`. Every call takes `ctx` first and an
empty options struct last.

## Task 1: Get, then walk the structs

```go
d, err := cs.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
if err != nil {
	return "", err
}
containers := d.Spec.Template.Spec.Containers
```

Check `len(containers) == 0` before you read `containers[0]`, and build the
error with `fmt.Errorf("deployment %s/%s has no containers", namespace, name)`.

## Task 2: set Data, then Create

Add `Data: data,` to the struct literal (or write `cm.Data = data`), then hand
the result of `Create` straight back:

```go
return cs.CoreV1().ConfigMaps(namespace).Create(ctx, cm, metav1.CreateOptions{})
```

## Task 3: check NotFound first

```go
err := cs.AppsV1().Deployments(namespace).Delete(ctx, name, metav1.DeleteOptions{})
if apierrors.IsNotFound(err) {
	return false, nil
}
if err != nil {
	return false, err
}
return true, nil
```
