## Task 1: fill the empty branch

The `if apierrors.IsNotFound(err) { ... }` block is empty, so the code falls
through to `if err != nil` and returns the "not found" error. Call `Create`
inside the block and `return`, so nothing below it runs.

```go
_, err = cs.AppsV1().Deployments(namespace).Create(ctx, newDeployment(namespace, name, desired), metav1.CreateOptions{})
```

## Task 2: an early return

Put one `if` before the update code. Both conditions must hold, so join them
with `&&`:

```go
if ptr.Deref(current.Spec.Replicas, 1) == desired.Replicas &&
	current.Spec.Template.Spec.Containers[0].Image == desired.Image {
	return nil
}
```

## Task 3: start from current

Replace the `newDeployment` and `ResourceVersion` lines:

```go
updated := current.DeepCopy()
updated.Spec.Replicas = new(desired.Replicas)
updated.Spec.Template.Spec.Containers[0].Image = desired.Image
```

The `Update` call below stays as it is.
