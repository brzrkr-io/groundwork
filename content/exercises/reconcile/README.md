# Reconcile loops: desired vs actual

Every Kubernetes controller is built around one small idea: look at the
world, compare it with what you want, and fix the difference. In this
exercise you write that comparison as plain Go functions, with no cluster
involved.

## Learn

### Level-triggered, not edge-triggered

Think of a thermostat. It doesn't remember that someone opened a window ten
minutes ago and closed it again. It looks at the temperature *right now*,
compares it with the setting, and turns the heating on or off. If it misses
a change, nothing breaks: the next look sees the room as it really is.

That is *level-triggered* logic: act on the current state. The opposite is
*edge-triggered* logic: act on each change ("the window opened", "the window
closed"). Edge-triggered code goes wrong as soon as one event is lost or
arrives twice. Kubernetes controllers are level-triggered on purpose. Events
only tell a controller *when* to look. What it does depends on what it sees.

### Desired state and actual state

The *desired state* is what someone asked for, such as "3 replicas of
`api:1.5.0`". The *actual state* is what exists right now, such as a
Deployment with 2 replicas of `api:1.4.2`. A *reconcile* function brings the
two together in three steps:

1. **Observe**: read the desired state and the actual state.
2. **Diff**: work out what is different.
3. **Act**: change only what is different. If nothing is, do nothing.

Because reconcile only ever moves the actual state towards the desired state,
running it twice is safe. The second run finds nothing to do. As in
*Idempotency keys*, this property is called *idempotence*. A controller may
reconcile the same object many times, after restarts, retries or duplicate
events, so idempotence is not optional.

### ConfigMaps

A *ConfigMap* is a small Kubernetes object that holds configuration as
string key/value pairs, in its `Data` field:

```go
cm := &corev1.ConfigMap{
	ObjectMeta: metav1.ObjectMeta{Name: "rollout-config", Namespace: "shop"},
	Data: map[string]string{
		"replicas": "3",
		"image":    "registry.example.com/shop/api:1.5.0",
	},
}
```

Every value is a string, even numbers. Reading a missing key from a map
gives `""`, so `cm.Data["replicas"] == ""` covers "missing" and "empty"
(and a nil `Data` map) in one check.

To turn the string into an `int32`, use `strconv.ParseInt`. Its last
argument is the bit size, and `32` makes it return an error for numbers too
big for an `int32`. It still returns an `int64`, so convert it afterwards:

```go
n, err := strconv.ParseInt("3", 10, 32) // base 10, must fit in 32 bits
// n => 3 (an int64), err => nil
replicas := int32(n)
```

### Unset replicas and `ptr.Deref`

`Spec.Replicas` is a `*int32`. When it is `nil`, Kubernetes runs 1 replica.
`ptr.Deref` from `k8s.io/utils/ptr` reads a pointer with a default for
`nil`:

```go
ptr.Deref(d.Spec.Replicas, 1) // the value, or 1 when Replicas is nil
```

### Never change what you were given

Objects passed to a controller are often shared with other code (as with
the cache in *Informers: listers and the read-only cache*). `DeepCopy()` makes a complete
independent copy, so you can change the copy freely:

```go
updated := current.DeepCopy()
updated.Spec.Replicas = new(int32(3)) // current is untouched
```

## Your task

The `shop` team keeps the desired state of their `api` Deployment in a
ConfigMap called `rollout-config`. In this exercise the Deployment always
has exactly one container.

### 1. Read the desired state

Make `DesiredFromConfigMap` return a `Desired` built from `cm.Data`:

- `"replicas"` parsed into an `int32`. `"0"` is fine.
- `"image"` copied as it is.

Return an error wrapping one of the provided sentinel errors (with `%w`, so
`errors.Is` finds it):

| Problem | Error wraps | Example message |
|---|---|---|
| `"replicas"` or `"image"` missing or `""` | `ErrMissingKey` | `missing key: "image"` |
| `"replicas"` not a whole number, negative, or too big for `int32` | `ErrInvalidReplicas` | `invalid replicas: "three"` |

Check `"replicas"` first, then `"image"`. Add `"fmt"` and `"strconv"` to the
imports.

```go
DesiredFromConfigMap(cm) // Data: replicas "3", image "registry.example.com/shop/api:1.5.0"
// => Desired{Replicas: 3, Image: "registry.example.com/shop/api:1.5.0"}, nil

DesiredFromConfigMap(cm) // Data: replicas "-1", image "..."
// => Desired{}, error wrapping ErrInvalidReplicas
```

### 2. List the differences

Make `Diff` return one string per difference, in this order and format:

1. Replicas, if different: `"replicas: <current> -> <desired>"`. A `nil`
   `Spec.Replicas` counts as 1 (use `ptr.Deref`).
2. Image of the first container
   (`current.Spec.Template.Spec.Containers[0].Image`), if different:
   `"image: <current> -> <desired>"`.

Return an empty (or `nil`) slice when everything matches. Add
`"k8s.io/utils/ptr"` to the imports.

```go
// current: 2 replicas, image "registry.example.com/shop/api:1.4.2"
Diff(Desired{Replicas: 3, Image: "registry.example.com/shop/api:1.5.0"}, current)
// => []string{
//      "replicas: 2 -> 3",
//      "image: registry.example.com/shop/api:1.4.2 -> registry.example.com/shop/api:1.5.0",
//    }
```

### 3. Apply the desired state

Make `Apply` return an updated copy of `current` and whether anything
changed:

- Always start from `current.DeepCopy()` and return that copy, even when
  nothing changed. Never modify `current` itself.
- If the replicas differ (again, `nil` counts as 1), set `Spec.Replicas` to
  `new(desired.Replicas)`. Leave it alone otherwise.
- If the first container's image differs, set it to `desired.Image`.
- The `bool` is `true` if you changed at least one field.

The tests also check idempotence: applying the result a second time must
return `false`.

```go
// current: 2 replicas, image ".../api:1.4.2"
updated, changed := Apply(Desired{Replicas: 5, Image: ".../api:1.4.2"}, current)
// *updated.Spec.Replicas => 5, changed => true, current still has 2

_, changed = Apply(Desired{Replicas: 5, Image: ".../api:1.4.2"}, updated)
// changed => false
```

## Learn more

- [Kubernetes: Controllers](https://kubernetes.io/docs/concepts/architecture/controller/)
- [Kubernetes: ConfigMaps](https://kubernetes.io/docs/concepts/configuration/configmap/)
- [pkg.go.dev: strconv.ParseInt](https://pkg.go.dev/strconv#ParseInt)
- [pkg.go.dev: k8s.io/utils/ptr](https://pkg.go.dev/k8s.io/utils/ptr)
