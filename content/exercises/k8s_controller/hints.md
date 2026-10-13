## Work one test at a time

Do the tasks in order and watch the matching test turn green: `TestKeyAndExposed`,
`TestDesiredService`, `TestReconcileCreatesAndUpdates`, `TestReconcileRemoves`,
`TestEventHandlers`, `TestProcessNextItem`, `TestRun`. Each failure message
names the call, what you returned and what the test wanted.

## Tasks 1 and 2: plain Go

Reading a key that is missing from a map (even a nil map) gives `""`, so
`d.Labels[ExposeLabel] == "true"` is the whole of `Exposed`. For
`DesiredService`, loop over the containers, and inside that loop over each
container's `Ports`, appending one `corev1.ServicePort` per port:

```go
var ports []corev1.ServicePort
for _, container := range d.Spec.Template.Spec.Containers {
	for _, p := range container.Ports {
		ports = append(ports, corev1.ServicePort{
			Name:       p.Name,
			Port:       p.ContainerPort,
			TargetPort: intstr.FromInt32(p.ContainerPort),
			Protocol:   corev1.ProtocolTCP,
		})
	}
}
```

## Tasks 3 and 4: get first, then decide

Write each method below `Reconcile` and return its result from the matching
`TODO`:

```go
if Exposed(d) {
	return c.ensureService(ctx, d)
}
return c.removeService(ctx, d)
```

Both methods start the same way, then branch:

```go
services := c.clientset.CoreV1().Services(d.Namespace)
got, err := services.Get(ctx, d.Name, metav1.GetOptions{})
if apierrors.IsNotFound(err) {
	// task 3: Create want. task 4: return nil.
}
if err != nil {
	return err
}
if !metav1.IsControlledBy(got, d) {
	// task 3: return an error wrapping ErrNotOwned. task 4: return nil.
}
```

Task 3 then returns nil if `ServiceInSync(got, want)`, or sets
`got.Spec.Selector` and `got.Spec.Ports` from `want` and calls
`services.Update(ctx, got, metav1.UpdateOptions{})`. Task 4 calls
`services.Delete(ctx, d.Name, metav1.DeleteOptions{})` and treats a NotFound
error as success.

## Tasks 5 to 7: feeding and draining the queue

```go
func (c *Controller) onUpdate(oldD, newD *appsv1.Deployment) {
	if Exposed(oldD) || Exposed(newD) {
		c.queue.Add(Key(newD))
	}
}

func (c *Controller) processNextItem(ctx context.Context) bool {
	key, shutdown := c.queue.Get()
	if shutdown {
		return false
	}
	defer c.queue.Done(key)
	if err := c.Reconcile(ctx, key); err != nil {
		c.queue.AddRateLimited(key)
		return true
	}
	c.queue.Forget(key)
	return true
}

func (c *Controller) runWorker(ctx context.Context) {
	for c.processNextItem(ctx) {
	}
}
```
