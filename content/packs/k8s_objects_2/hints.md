## Task 1: Fix the namespace and selector

The Deployment already knows its namespace and which labels its Pods carry.
Look at where `d` keeps each of those, instead of the hard-coded values.

### Closer

The namespace is `d.Namespace`. The Pod labels to select on are in the
selector: `d.Spec.Selector.MatchLabels`, not `d.Labels`.

### Answer

```go
ObjectMeta: metav1.ObjectMeta{
	Name:      d.Name,
	Namespace: d.Namespace,
	Labels:    d.Labels,
},
Spec: corev1.ServiceSpec{
	Type:     corev1.ServiceTypeClusterIP,
	Selector: d.Spec.Selector.MatchLabels,
	// Ports unchanged here
```

The selector labels are the ones every Pod of the Deployment is guaranteed to
carry.

## Task 2: Fix the ports

The two numbers are swapped. Clients connect to `Port`, and traffic goes to
the container's `TargetPort`.

### Closer

`Port` is a plain number, `80`. `TargetPort` is an `intstr.IntOrString`, so
wrap the container port with `intstr.FromInt32(...)`.

### Answer

```go
Port:       80,
TargetPort: intstr.FromInt32(containerPort),
```

Clients call port 80, and the Service forwards to the port the container
listens on.

## Task 3: Check a Deployment's selector

Do two checks: first, is there a selector at all? Then go through the
selector's keys in alphabetical order and compare each value with the
template's value for that key.

### Closer

The empty check is `d.Spec.Selector == nil || len(d.Spec.Selector.MatchLabels) == 0`.
Keep the selector labels in a short variable, `want := d.Spec.Selector.MatchLabels`,
and loop over its sorted keys with `slices.Sorted(maps.Keys(want))` (import
`"maps"`, `"slices"` and `"fmt"`). Reading a missing key from a map gives `""`.

```go
want := d.Spec.Selector.MatchLabels
for _, key := range slices.Sorted(maps.Keys(want)) {
	if got := d.Spec.Template.Labels[key]; got != want[key] {
		// return the error
	}
}
```

### Answer

```go
if d.Spec.Selector == nil || len(d.Spec.Selector.MatchLabels) == 0 {
	return fmt.Errorf("deployment %s/%s: selector is empty", d.Namespace, d.Name)
}
want := d.Spec.Selector.MatchLabels
for _, key := range slices.Sorted(maps.Keys(want)) {
	if got := d.Spec.Template.Labels[key]; got != want[key] {
		return fmt.Errorf("deployment %s/%s: template label %s is %q, selector wants %q",
			d.Namespace, d.Name, key, got, want[key])
	}
}
return nil
```

Sorting the keys makes the error the same on every run.
