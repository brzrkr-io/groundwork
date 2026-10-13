## Task 1: a map literal

`Labels` returns a `map[string]string`. Write a map literal with two entries:
the key `"app"` with the value `app`, and the key `"env"` with the value
`env`.

## Task 2: match the YAML

Each YAML field has a Go field with the same name, capitalised:
`namespace` → `Namespace`, `replicas` → `Replicas`, `selector` → `Selector`.
`Replicas` needs a `*int32`, so use `new(replicas)`. `Selector` needs a
`*metav1.LabelSelector`, so start it with `&`.

```go
Spec: appsv1.DeploymentSpec{
	Replicas: new(replicas),
	Selector: &metav1.LabelSelector{
		MatchLabels: labels,
	},
	Template: ...
```

## Task 3: ports and resources

`Ports` is a slice of structs, and `Requests` and `Limits` are maps:

```go
Image: image,
Ports: []corev1.ContainerPort{
	{Name: "http", ContainerPort: port},
},
Resources: corev1.ResourceRequirements{
	Requests: corev1.ResourceList{
		corev1.ResourceCPU:    resource.MustParse("100m"),
		corev1.ResourceMemory: resource.MustParse("128Mi"),
	},
	Limits: corev1.ResourceList{
		// the same two keys, with "500m" and "256Mi"
	},
},
```

Remember to add the `resource` import at the top of the file.
