# Kubernetes objects: building a Deployment

Kubernetes runs most modern platforms, and everything you ask it to do is
written down as an *object*. In this exercise you build a Deployment, the
object that keeps an app running, in plain Go.

## Learn

### Kubernetes in one minute

*Kubernetes* runs containers on a group of machines called a *cluster*. You
never log in to a machine to start a container. Instead you send the cluster a
description of what you want, and Kubernetes makes it happen and keeps it
that way.

Every description goes to one place: the *API server*, the cluster's front
door. The API server stores *objects*. Programs called *controllers* watch
those objects and work to make the real world match them. If a Deployment
says "3 copies" and one machine dies, a controller notices that only 2 are
running and starts another. Desired state in, actual state follows.

The objects you will meet most:

| Kind | What it is |
|---|---|
| Namespace | A folder that groups objects, such as `shop` or `monitoring`. |
| Pod | One or more containers that run together on one machine. |
| Deployment | "Keep N identical Pods running from this template." |
| Service | A stable network address in front of a set of Pods. |
| ConfigMap | Plain configuration values for Pods. |

Every object has the same outline. Here is a Deployment, written in YAML:

```yaml
apiVersion: apps/v1       # which API group and version defines this kind
kind: Deployment          # what sort of object it is
metadata:                 # who it is: name, namespace, labels
  name: checkout
  namespace: shop
  labels:
    app: checkout
spec:                     # what you want
  replicas: 3
  selector:
    matchLabels:
      app: checkout
  template:               # the Pod to make copies of
    metadata:
      labels:
        app: checkout
    spec:
      containers:
        - name: checkout
          image: registry.example.com/shop/checkout:1.4.2
```

Kubernetes adds a fourth part, `status` (what is actually happening), which
you don't write yourself.

*Labels* are short key/value tags such as `app: checkout` or
`env: production`. Kubernetes uses them to connect objects. A Deployment
finds its own Pods with its *selector*. The selector's `matchLabels` must
match the labels in the Pod `template`, otherwise the Deployment would
create Pods it can't recognise. The API server rejects a Deployment like
that.

### The same object in Go

The Go types live in `k8s.io/api` (one package per API group) and
`k8s.io/apimachinery` (the shared parts). Everyone imports them with short
names:

```go
import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)
```

Each YAML section becomes a struct field. `metadata` is `ObjectMeta`, and
`spec` is `Spec`:

```go
d := &appsv1.Deployment{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "checkout",
		Namespace: "shop",
		Labels:    map[string]string{"app": "checkout"},
	},
	Spec: appsv1.DeploymentSpec{ /* ... */ },
}
```

The struct is nested several levels deep. Take it one level at a time, and
read the field types in your editor or on pkg.go.dev.

### Pointer fields and `new(value)`

`Spec.Replicas` is a `*int32`, not an `int32`. With a pointer, "not set"
(`nil`, so Kubernetes uses the default of 1) is different from "set to 0".
Go won't let you write `&3` or `&int32(3)`, but since Go 1.26 the built-in
`new` accepts a value as well as a type. `new(v)` returns a pointer to a
fresh copy of `v`:

```go
Replicas: new(replicas),  // replicas is an int32, so this is a *int32
```

With a constant, say which type you want: `new(3)` would give an `*int`,
so write `new(int32(3))` for an `*int32`.

Older Kubernetes code uses the helper `ptr.To` from `k8s.io/utils/ptr` for
the same job. You will see it in other projects, but plain `new` needs no
import.

`Spec.Selector` is a pointer too, `*metav1.LabelSelector`, so you write
`&metav1.LabelSelector{MatchLabels: labels}`.

### Ports and resources

A container lists the ports it listens on, as a slice of
`corev1.ContainerPort`. It also asks for CPU and memory. *Requests* are what
the scheduler reserves for the container. *Limits* are the most it may use.
The amounts are `resource.Quantity` values (package
`k8s.io/apimachinery/pkg/api/resource`), parsed from the same strings you'd
write in YAML:

| String | Meaning |
|---|---|
| `"500m"` | 500 millicores, which is half a CPU core |
| `"128Mi"` | 128 mebibytes of memory |

```go
Resources: corev1.ResourceRequirements{
	Requests: corev1.ResourceList{
		corev1.ResourceCPU:    resource.MustParse("100m"),
		corev1.ResourceMemory: resource.MustParse("128Mi"),
	},
},
```

`corev1.ResourceList` is a map from a resource name to a quantity.
`resource.MustParse` panics on a bad string, so use it only with fixed
strings like these.

## Your task

The `shop` team wants Go code that builds their Deployments, so every app
gets the same labels, replicas and resource settings.

### 1. Build the labels

Make `Labels` return a map with `"app"` set to the app name and `"env"` set
to the environment:

```go
Labels("checkout", "production")
// => map[string]string{"app": "checkout", "env": "production"}
```

### 2. Fill in the Deployment

`NewDeployment` already sets the name and the Pod template. Add the missing
fields:

- `ObjectMeta.Namespace` is `ns`, and `ObjectMeta.Labels` is `labels`.
- `Spec.Replicas` points to `replicas`. Use `new(replicas)`.
- `Spec.Selector` is a `*metav1.LabelSelector` whose `MatchLabels` is `labels`,
  the same labels as the Pod template.

```go
d := NewDeployment("checkout", "shop", "production", "registry.example.com/shop/checkout:1.4.2", 3)
// d.Namespace                  => "shop"
// d.Labels                     => map[app:checkout env:production]
// *d.Spec.Replicas             => 3
// d.Spec.Selector.MatchLabels  => map[app:checkout env:production]
```

### 3. Build the container

Make `AppContainer` return a container with:

- `Name` set to `name` (already done) and `Image` set to `image`.
- `Ports`: exactly one `corev1.ContainerPort` with `Name` `"http"` and
  `ContainerPort` set to `port`.
- `Resources`: requests of `100m` CPU and `128Mi` memory, and limits of
  `500m` CPU and `256Mi` memory. Use `resource.MustParse` and add the import
  `"k8s.io/apimachinery/pkg/api/resource"`.

```go
c := AppContainer("metrics", "registry.example.com/ops/metrics:2.0.0", 9090)
// c.Image                    => "registry.example.com/ops/metrics:2.0.0"
// c.Ports[0].Name            => "http"
// c.Ports[0].ContainerPort   => 9090
// c.Resources.Limits.Memory  => 256Mi
```

## Learn more

- [Kubernetes: Objects in Kubernetes](https://kubernetes.io/docs/concepts/overview/working-with-objects/)
- [Kubernetes: Deployments](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/)
- [Kubernetes: Resource management for Pods and containers](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/)
- [pkg.go.dev: k8s.io/api/apps/v1](https://pkg.go.dev/k8s.io/api/apps/v1)
