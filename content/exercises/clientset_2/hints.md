## Task 1: keep what Get returned

The bug is that the result of `Get` is thrown away (`_`). Store it in a
variable, change one field on it, and pass that same object to `Update`.

## Task 1: read, modify, write

```go
d, err := cs.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
if err != nil {
	return err
}
d.Spec.Replicas = new(replicas)
_, err = cs.AppsV1().Deployments(namespace).Update(ctx, d, metav1.UpdateOptions{})
return err
```

## Task 2: the AlreadyExists branch

```go
if !apierrors.IsAlreadyExists(err) {
	return nil, err
}
existing, err := cs.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
if err != nil {
	return nil, err
}
existing.Data = data
return cs.CoreV1().ConfigMaps(namespace).Update(ctx, existing, metav1.UpdateOptions{})
```

## Task 3: compare before you write

Just before `existing.Data = data`:

```go
if maps.Equal(existing.Data, data) {
	return existing, nil
}
```
