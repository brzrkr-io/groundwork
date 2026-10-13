## Task 1: Read a Deployment's image

Deployments live in the `apps/v1` group. `Get` the Deployment, then walk down its structs to the containers, and make sure there is at least one before you read it.

### Closer

Start with `cs.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})`. The containers are at `d.Spec.Template.Spec.Containers`; check `len(containers) == 0` before reading `containers[0]`, and add `"fmt"` to the imports.

### Answer

```go
d, err := cs.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
if err != nil {
	return "", err
}
containers := d.Spec.Template.Spec.Containers
if len(containers) == 0 {
	return "", fmt.Errorf("deployment %s/%s has no containers", namespace, name)
}
return containers[0].Image, nil
```

Returning `Get`'s error unchanged keeps `apierrors.IsNotFound` working for callers.

## Task 2: Create a ConfigMap

Two steps: put `data` into the ConfigMap, then send it to the API server. ConfigMaps are in the core group, so the typed client comes from `cs.CoreV1()`.

### Closer

Add `Data: data,` to the struct literal (or write `cm.Data = data`). Then call `Create` with `ctx` first and empty options last:

```go
cs.CoreV1().ConfigMaps(namespace).Create(ctx, cm, ...)
```

### Answer

```go
cm := &corev1.ConfigMap{
	ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
	Data:       data,
}
return cs.CoreV1().ConfigMaps(namespace).Create(ctx, cm, metav1.CreateOptions{})
```

`Create` already returns `(*corev1.ConfigMap, error)`, so you can return its results directly.

## Task 3: Delete a Deployment, if it is there

Call `Delete` and then sort the result into the three cases in the README table. Check for NotFound before the general `err != nil` check.

### Closer

`Delete` returns only an error. `apierrors.IsNotFound(err)` is `false` for `nil`, so it is safe to call first. Add `apierrors "k8s.io/apimachinery/pkg/api/errors"` to the imports.

```go
err := cs.AppsV1().Deployments(namespace).Delete(ctx, name, metav1.DeleteOptions{})
if apierrors.IsNotFound(err) {
	// ...
}
```

### Answer

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

A missing Deployment means the cleanup is already done, so it is not an error.
