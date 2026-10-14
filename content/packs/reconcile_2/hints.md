## Task 1: Create the missing Deployment

The `if apierrors.IsNotFound(err) { ... }` block is empty, so the code falls
through to `if err != nil` and returns the "not found" error. Create the
Deployment inside that block and `return`, so nothing below it runs.

### Closer

Build the object with `newDeployment(namespace, name, desired)` and pass it to
`cs.AppsV1().Deployments(namespace).Create(ctx, ..., metav1.CreateOptions{})`.
Wrap a failure like the other errors: `fmt.Errorf("create deployment %s/%s: %w", ...)`.

### Answer

```go
if apierrors.IsNotFound(err) {
	_, err = cs.AppsV1().Deployments(namespace).Create(ctx, newDeployment(namespace, name, desired), metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("create deployment %s/%s: %w", namespace, name, err)
	}
	return nil
}
```

The early `return nil` stops the function before the update code runs.

## Task 2: Do nothing when in sync

Before the update code, check whether the live Deployment already matches
`desired`. If it does, return early without writing anything.

### Closer

Both the replicas and the image must match, so join two comparisons with `&&`.
Use `ptr.Deref(current.Spec.Replicas, 1)` so a `nil` pointer counts as 1, and
add `"k8s.io/utils/ptr"` to the imports.

### Answer

```go
if ptr.Deref(current.Spec.Replicas, 1) == desired.Replicas &&
	current.Spec.Template.Spec.Containers[0].Image == desired.Image {
	return nil
}
```

Returning before `Update` means a reconcile of an in-sync Deployment makes no writes.

## Task 3: Update a copy of the live Deployment

Instead of building a fresh object with `newDeployment`, start from the
Deployment you just read and change only the two fields you own.

### Closer

`current.DeepCopy()` gives you a copy with all labels, annotations, env vars
and the `ResourceVersion` already set. Delete the `newDeployment` and
`ResourceVersion` lines and set fields on the copy instead.

### Answer

```go
updated := current.DeepCopy()
updated.Spec.Replicas = new(desired.Replicas)
updated.Spec.Template.Spec.Containers[0].Image = desired.Image
```

The `Update` call below stays as it is; everything else on the live object survives because you started from it.
