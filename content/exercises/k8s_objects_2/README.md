# Kubernetes objects: Services and selectors

A Deployment keeps Pods running, but nothing can reach them yet. In this
exercise you fix a buggy Service builder and write a check that catches a
broken Deployment before you send it to the cluster.

## Learn

### Recap

Kubernetes objects have `metadata` (name, namespace, labels) and a `spec`
(what you want). In Go they are nested structs from `k8s.io/api`, such as
`appsv1.Deployment`. A Deployment's `Spec.Selector.MatchLabels` says which
labels its Pods carry, and its Pod template must carry those labels.

### Services

Pods come and go, and each new Pod gets a new IP address. A *Service* gives
a set of Pods one stable name and address. The default type, *ClusterIP*,
is reachable only from inside the cluster, which suits internal APIs.

A Service doesn't list its Pods. It has a *selector*: every Pod whose labels
match receives traffic. A Service has two port numbers:

| Field | Meaning |
|---|---|
| `Port` | The port clients connect to, such as `80`. |
| `TargetPort` | The port the container listens on, such as `8080`. |

```yaml
apiVersion: v1
kind: Service
metadata:
  name: checkout
  namespace: shop
spec:
  type: ClusterIP
  selector:
    app: checkout
  ports:
    - name: http
      protocol: TCP
      port: 80          # clients call http://checkout.shop:80
      targetPort: 8080  # traffic goes to the container's port 8080
```

`targetPort` may be a number or a port name, so in Go it has the type
`intstr.IntOrString` (package `k8s.io/apimachinery/pkg/util/intstr`). Build
one from a number with `intstr.FromInt32(8080)`.

Select Pods with the Deployment's selector labels, not with all of the
template labels. Templates often carry extra labels such as `version: 1.4.2`
that change on every release. A Service that selects on them would lose its
Pods during a rollout.

### Sorted map keys

Map iteration order is random. When an error message depends on which key
you look at first, sort the keys so the message is the same on every run:

```go
for _, key := range slices.Sorted(maps.Keys(m)) {
	// keys in alphabetical order
}
```

`maps.Keys` yields the keys, and `slices.Sorted` collects them into a sorted
slice. Import `"maps"` and `"slices"`.

## Your task

The `shop` team's Service builder has bugs: the Services land in the wrong
namespace, select no Pods, and send traffic to the wrong port.

### 1. Fix the namespace and selector

In `NewService`, the Service must use the Deployment's namespace, and its
`Spec.Selector` must be the Deployment's `Spec.Selector.MatchLabels`.

```go
// d: shop/checkout, labels {app: checkout, team: payments},
//    selector {app: checkout}, template labels {app: checkout, version: 1.4.2}
svc := NewService(d, 8080)
// svc.Namespace      => "shop"
// svc.Spec.Selector  => map[app:checkout]
```

### 2. Fix the ports

The single port must have `Port` 80 and `TargetPort` set to `containerPort`:

```go
svc := NewService(d, 3000)
// svc.Spec.Ports[0].Port        => 80
// svc.Spec.Ports[0].TargetPort  => intstr.FromInt32(3000)
```

### 3. Check a Deployment's selector

Write `CheckSelector(d)`. It returns:

- `deployment <namespace>/<name>: selector is empty` if `d.Spec.Selector` is
  `nil` or its `MatchLabels` map is empty.
- `deployment <namespace>/<name>: template label <key> is "<got>", selector wants "<want>"`
  for the first selector key, in alphabetical order, whose value in
  `d.Spec.Template.Labels` differs. A missing label counts as `""`.
- `nil` when every selector label is in the template with the same value.
  Extra template labels are fine.

```go
// selector {app: checkout, env: production}, template {app: checkout}
CheckSelector(d)
// => deployment shop/checkout: template label env is "", selector wants "production"

// selector {app: checkout, env: production}, template {app: cart, env: staging}
CheckSelector(d)
// => deployment shop/checkout: template label app is "cart", selector wants "checkout"
```

Build the errors with `fmt.Errorf`, using `%q` for the two quoted values.

## Learn more

- [Kubernetes: Service](https://kubernetes.io/docs/concepts/services-networking/service/)
- [Kubernetes: Labels and selectors](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/)
- [pkg.go.dev: intstr](https://pkg.go.dev/k8s.io/apimachinery/pkg/util/intstr)
