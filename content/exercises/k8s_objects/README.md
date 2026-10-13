# Kubernetes objects: building a Deployment

Kubernetes runs most modern platforms, and everything you ask it to do is
written down as an *object*. In this exercise you build a Deployment, the
object that keeps an app running, in plain Go.

**New here:** Kubernetes objects, the `k8s.io/api` types, pointer fields with `new(value)`, `resource.Quantity`.

## Learn

### Kubernetes in one minute

*Kubernetes* runs containers on a group of machines called a *cluster*. You
don't start containers by hand. You send the cluster a description of what
you want, and Kubernetes makes it happen and keeps it that way.

Descriptions go to the *API server*, the cluster's front door, which stores
them as *objects*. Programs called *controllers* watch the objects and fix
the real world to match. If a Deployment says "3 copies" and one dies, a
controller starts another.

| Kind | What it is |
|---|---|
| Namespace | A folder that groups objects, such as `shop` or `monitoring`. |
| Pod | One or more containers that run together on one machine. |
| Deployment | "Keep N identical Pods running from this template." |
| Service | A stable network address in front of a set of Pods. |
| ConfigMap | Plain configuration values for Pods. |

### The shape of an object

Every object has the same outline. Here is a Deployment in YAML:

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

Kubernetes adds a fourth part, `status` (what is really happening). You
never write it yourself.

*Labels* are key/value tags such as `app: checkout`. A Deployment finds its
own Pods with its *selector*. The selector's `matchLabels` must match the
labels in the Pod `template`, or the API server rejects the Deployment.

### The same object in Go

The Go types live in `k8s.io/api` (one package per API group) and
`k8s.io/apimachinery` (shared parts). Each YAML section becomes a struct
field: `metadata` is `ObjectMeta`, `spec` is `Spec`. Here is a different
object, a ConfigMap:

```go
cm := &corev1.ConfigMap{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "settings",
		Namespace: "monitoring",
	},
	Data: map[string]string{"level": "debug"},
}
```

The Go field name is the YAML name with a capital first letter:
`namespace` → `Namespace`. Take nested structs one level at a time, and
check field types in your editor or on pkg.go.dev.

> **Why is it written like this?**
> - **Go requires** capitalised field names: only capitalised names can be
>   used from another package.
> - **Convention:** the import names `appsv1`, `corev1` and `metav1` are what
>   every Kubernetes project uses. Go would accept any name.
> - **Author's choice:** one field per line, and the aligned `:` columns,
>   come from gofmt and taste, not the compiler.

### Pointer fields and `new(value)`

Some fields are pointers, like `*int32`. A pointer can be `nil`, which
means "not set, use the default". That is different from "set to 0".

Go won't let you write `&3`. Since Go 1.26 the built-in `new` accepts a
value: `new(v)` returns a pointer to a fresh copy of `v`.

```go
timeout := int64(30)
p := new(timeout)  // p is a *int64 pointing at 30
q := new(int64(5)) // with a constant, say the type you want
```

For a struct, put `&` in front of the struct literal instead:

```go
sel := &metav1.LabelSelector{MatchLabels: map[string]string{"tier": "db"}}
```

> **Why is it written like this?** **Go requires** a pointer where the field
> type starts with `*`; a plain value won't compile. Older code uses
> `ptr.To(v)` from `k8s.io/utils/ptr` for the same job (**convention** of
> its time); plain `new` needs no import.

### Ports and resources

A container lists its ports as a slice, `[]corev1.ContainerPort`. It also
asks for CPU and memory:

- *Requests*: what the scheduler reserves for the container.
- *Limits*: the most it may use.

Amounts are `resource.Quantity` values from package
`k8s.io/apimachinery/pkg/api/resource`, parsed from YAML-style strings:

| String | Meaning |
|---|---|
| `"500m"` | 500 millicores, half a CPU core |
| `"128Mi"` | 128 mebibytes of memory |

`corev1.ResourceList` is a map from a resource name (`corev1.ResourceCPU`,
`corev1.ResourceMemory`) to a quantity:

```go
list := corev1.ResourceList{
	corev1.ResourceCPU: resource.MustParse("2"),
}
```

> **Why is it written like this?** `MustParse` panics on a bad string. By
> **convention** you use it only with fixed strings you typed yourself, never
> with user input.

## Your task

The `shop` team wants Go code that builds their Deployments, so every app
gets the same labels, replicas and resource settings.

### 1. Build the labels

**Goal:** make `Labels(app, env)` return a map with two entries.

- Key `"app"` → the `app` parameter.
- Key `"env"` → the `env` parameter.

```go
Labels("checkout", "production")
// => map[string]string{"app": "checkout", "env": "production"}
```

**Think it through:**
- Which parts are fixed text (in quotes) and which are parameters (no quotes)?
- What does a map literal look like? (Seen in *Maps*.)

### 2. Fill in the Deployment

**Goal:** add the missing fields in `NewDeployment` at the two `TODO(2)` lines.

The name and Pod template are already set. Add:

- `ObjectMeta.Namespace`: the `ns` parameter.
- `ObjectMeta.Labels`: the local variable `labels`.
- `Spec.Replicas`: a pointer to the `replicas` parameter (type `*int32`).
- `Spec.Selector`: a `*metav1.LabelSelector` whose `MatchLabels` is `labels`.

```go
d := NewDeployment("checkout", "shop", "production", "registry.example.com/shop/checkout:1.4.2", 3)
// d.Namespace                  => "shop"
// d.Labels                     => map[app:checkout env:production]
// *d.Spec.Replicas             => 3
// d.Spec.Selector.MatchLabels  => map[app:checkout env:production]
```

**Think it through:**
- Which fields are pointers? Check the type: does it start with `*`?
- For a pointer to a value use `new(...)`; for a pointer to a struct use `&`.
- Why must the selector use the same `labels` as the Pod template?

### 3. Build the container

**Goal:** make `AppContainer` return a container with its image, one port and resources.

- `Name` is `name` (already done); `Image` is `image`.
- `Ports`: exactly one `corev1.ContainerPort`, `Name` `"http"`, `ContainerPort` the `port` parameter.
- `Resources.Requests`: CPU `"100m"`, memory `"128Mi"`.
- `Resources.Limits`: CPU `"500m"`, memory `"256Mi"`.

Add the import `"k8s.io/apimachinery/pkg/api/resource"` yourself.

```go
c := AppContainer("metrics", "registry.example.com/ops/metrics:2.0.0", 9090)
// c.Image                    => "registry.example.com/ops/metrics:2.0.0"
// c.Ports[0].Name            => "http"
// c.Ports[0].ContainerPort   => 9090
// c.Resources.Requests.Cpu() => 100m
// c.Resources.Limits.Memory() => 256Mi
```

**Think it through:**
- `Ports` is a slice of structs: one `{...}` inside `[]corev1.ContainerPort{...}`.
- `Requests` and `Limits` are two separate `corev1.ResourceList` maps inside `corev1.ResourceRequirements`.

## Learn more

- [Kubernetes: Objects in Kubernetes](https://kubernetes.io/docs/concepts/overview/working-with-objects/)
- [Kubernetes: Deployments](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/)
- [Kubernetes: Resource management for Pods and containers](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/)
- [pkg.go.dev: k8s.io/api/apps/v1](https://pkg.go.dev/k8s.io/api/apps/v1)
