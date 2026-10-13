## Tasks 1 and 2: use what the Deployment says

The Deployment already knows its namespace (`d.Namespace`) and the labels its
Pods carry (`d.Spec.Selector.MatchLabels`). For the ports, clients call port
`80`, and the container listens on `containerPort`.

```go
Port:       80,
TargetPort: intstr.FromInt32(containerPort),
```

## Task 3: two checks

First handle the empty selector: `d.Spec.Selector == nil` or
`len(d.Spec.Selector.MatchLabels) == 0`. Then loop over the selector's keys in
sorted order and compare each value with the template's value for that key.
Reading a missing key from a map gives `""`.

## Task 3: almost the answer

```go
want := d.Spec.Selector.MatchLabels
for _, key := range slices.Sorted(maps.Keys(want)) {
	if got := d.Spec.Template.Labels[key]; got != want[key] {
		return fmt.Errorf("deployment %s/%s: template label %s is %q, selector wants %q",
			d.Namespace, d.Name, key, got, want[key])
	}
}
return nil
```
