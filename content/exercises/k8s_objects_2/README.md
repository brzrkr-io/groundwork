# Kubernetes objects: Services and selectors

You wrote Services by hand in *Services: selectors, ports and endpoints*.
Here you build them in Go: you fix a buggy Service builder and write a
check that catches a broken Deployment before you send it to the cluster.

**New here:** `corev1.Service` in Go, `intstr.IntOrString`, sorting map keys with `slices.Sorted(maps.Keys(m))`.

## Learn

### Recap

Kubernetes objects have `metadata` (name, namespace, labels) and a `spec`
(what you want). In Go they are nested structs, such as `appsv1.Deployment`.

A Deployment has two sets of labels that matter here:

| Go path | What it is |
|---|---|
| `d.Labels` | Labels on the Deployment object itself. |
| `d.Spec.Selector.MatchLabels` | Labels the Deployment uses to find its Pods. |
| `d.Spec.Template.Labels` | Labels each new Pod gets. |

### Recap: Services

A *Service* gives a set of Pods one stable name and address; the default
type, *ClusterIP*, is reachable only from inside the cluster. It doesn't
list its Pods: every ready Pod whose labels match its *selector* gets
traffic. In Go its two port fields are:

| Field | Meaning |
|---|---|
| `Port` | The port clients connect to, such as `80`. |
| `TargetPort` | The port the container listens on, such as `8080`. |

```yaml
spec:
  type: ClusterIP
  selector:
    app: checkout
  ports:
    - name: http
      port: 80          # clients call http://checkout.shop:80
      targetPort: 8080  # traffic goes to the container's port 8080
```

Select Pods with the Deployment's **selector** labels, not all its labels.
Templates often carry labels like `version: 1.4.2` that change on every
release, and a Service selecting on them would lose its Pods mid-rollout.

### `intstr.IntOrString`

In YAML, `targetPort` may be a number (`8080`) or a port name (`http`). Go
needs one type for both, so it is `intstr.IntOrString` from
`k8s.io/apimachinery/pkg/util/intstr`:

```go
a := intstr.FromInt32(9000)   // a number
b := intstr.FromString("web") // a port name
```

> **Why is it written like this?** **Go requires** every field to have one
> type, and has no "number or string" type, so Kubernetes wraps both in a
> struct. `FromInt32` takes an `int32`, the same type as your port parameter.

### Sorted map keys

Map iteration order is random. If an error depends on which key you check
first, sort the keys so the message is the same on every run:

```go
ages := map[string]int{"zoe": 30, "al": 41}
for _, name := range slices.Sorted(maps.Keys(ages)) {
	fmt.Println(name) // al, then zoe
}
```

`maps.Keys` yields the keys; `slices.Sorted` collects them into a sorted
slice. Import `"maps"` and `"slices"`.

Reading a missing key from a map gives the *zero value*: `""` for strings.
No crash, no error.

> **Why is it written like this?** **Go requires** nothing about order;
> randomness is on purpose. Sorting for stable output is a **convention**
> that keeps error messages and tests predictable.

## Your task

The `shop` team's Service builder has bugs: Services land in the wrong
namespace, select no Pods, and send traffic to the wrong port.

### 1. Fix the namespace and selector

**Goal:** in `NewService`, fix the two `TODO(1)` lines.

- `ObjectMeta.Namespace` must be the Deployment's namespace (not `"default"`).
- `Spec.Selector` must be the Deployment's selector `MatchLabels` (not `d.Labels`).

```go
// d: shop/checkout, labels {app: checkout, team: payments},
//    selector {app: checkout}, template labels {app: checkout, version: 1.4.2}
svc := NewService(d, 8080)
// svc.Namespace      => "shop"
// svc.Spec.Selector  => map[app:checkout]
```

**Think it through:**
- Where on `d` is its namespace? Where are its selector labels? (See the Recap table.)
- Replace the wrong value on the existing line; don't add a new field.

### 2. Fix the ports

**Goal:** the single port must have `Port` 80 and `TargetPort` set to `containerPort`.

Right now the two numbers are swapped.

```go
svc := NewService(d, 3000)
// svc.Spec.Ports[0].Port        => 80
// svc.Spec.Ports[0].TargetPort  => intstr.FromInt32(3000)
```

**Think it through:**
- Which number is fixed (`80`) and which comes from the parameter?
- `TargetPort` is an `IntOrString`, `Port` is a plain `int32`.

### 3. Check a Deployment's selector

**Goal:** write `CheckSelector(d)` so it returns an error when the selector doesn't match the Pod template, else `nil`.

Check in this order:

1. If `d.Spec.Selector` is `nil` **or** its `MatchLabels` is empty, return
   `deployment <namespace>/<name>: selector is empty`.
2. Go through the selector keys in alphabetical order. For the first key
   whose value in `d.Spec.Template.Labels` differs, return
   `deployment <namespace>/<name>: template label <key> is "<got>", selector wants "<want>"`.
   A missing template label counts as `""`.
3. Otherwise return `nil`. Extra template labels are fine.

Build errors with `fmt.Errorf`: `%s` for namespace, name and key, `%q` for
the two quoted values. Add imports `"fmt"`, `"maps"` and `"slices"`.

```go
// selector {app: checkout, env: production}, template {app: checkout}
CheckSelector(d)
// => deployment shop/checkout: template label env is "", selector wants "production"

// selector {app: checkout, env: production}, template {app: cart, env: staging}
CheckSelector(d)
// => deployment shop/checkout: template label app is "cart", selector wants "checkout"
```

**Think it through:**
- Check for `nil` before you touch `MatchLabels`, or the program crashes.
- Which map do you loop over: the selector or the template?
- `"<got>"` in the format is quotes plus a value; `%q` adds the quotes for you.

## Learn more

- [Kubernetes: Service](https://kubernetes.io/docs/concepts/services-networking/service/)
- [Kubernetes: Labels and selectors](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/)
- [pkg.go.dev: intstr](https://pkg.go.dev/k8s.io/apimachinery/pkg/util/intstr)
- [pkg.go.dev: maps.Keys](https://pkg.go.dev/maps#Keys)
