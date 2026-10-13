## Task 1: Build the labels

`Labels` returns a `map[string]string`. Write a map literal with two entries,
one for the app name and one for the environment.

### Closer

A map literal is the type followed by `key: value` pairs in braces. The keys
are fixed strings, the values are the parameters:

```go
return map[string]string{
	"app": ...,
	// one more entry for "env"
}
```

### Answer

```go
return map[string]string{
	"app": app,
	"env": env,
}
```

The keys are quoted text, and the values come straight from the parameters.

## Task 2: Fill in the Deployment

Each YAML field has a Go field with the same name, capitalised: `namespace` →
`Namespace`, `replicas` → `Replicas`, `selector` → `Selector`. The local
variable `labels` already holds the labels you need.

### Closer

`Replicas` needs a `*int32`, so use `new(replicas)`. `Selector` needs a
`*metav1.LabelSelector`, so start it with `&`:

```go
Selector: &metav1.LabelSelector{
	MatchLabels: ...,
},
```

### Answer

```go
ObjectMeta: metav1.ObjectMeta{
	Name:      app,
	Namespace: ns,
	Labels:    labels,
},
Spec: appsv1.DeploymentSpec{
	Replicas: new(replicas),
	Selector: &metav1.LabelSelector{
		MatchLabels: labels,
	},
	Template: ... // leave as it is
```

The selector uses the same `labels` as the Pod template, so the Deployment
recognises its own Pods.

## Task 3: Build the container

`Image` is a plain string. `Ports` is a slice of structs, and `Requests` and
`Limits` are maps from a resource name to a quantity.

### Closer

```go
Ports: []corev1.ContainerPort{
	{Name: "http", ContainerPort: port},
},
Resources: corev1.ResourceRequirements{
	Requests: corev1.ResourceList{
		corev1.ResourceCPU: resource.MustParse("100m"),
		// memory next
	},
},
```

Add `"k8s.io/apimachinery/pkg/api/resource"` to the imports.

### Answer

```go
return corev1.Container{
	Name:  name,
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
			corev1.ResourceCPU:    resource.MustParse("500m"),
			corev1.ResourceMemory: resource.MustParse("256Mi"),
		},
	},
}
```

`resource.MustParse` turns the same strings you would write in YAML into
quantities.
