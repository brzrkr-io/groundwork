# Reconcile loops: desired vs actual

Every Kubernetes controller is built around one small idea: look at the
world, compare it with what you want, and fix the difference. Here you write
that comparison as plain Go functions, with no cluster involved.

**New here:** desired vs actual state, level-triggered logic, ConfigMap `Data`, `strconv.ParseInt` with a bit size, `ptr.Deref`.

## Learn

### Level-triggered, not edge-triggered

A thermostat doesn't remember that a window opened ten minutes ago. It looks
at the temperature *right now*, compares it with the setting, and acts. If it
misses a change, nothing breaks.

That is *level-triggered* logic: act on the current state. *Edge-triggered*
logic acts on each change, and goes wrong as soon as one event is lost or
duplicated. Kubernetes controllers are level-triggered: events only say
*when* to look.

### Desired, actual, reconcile

The *desired state* is what someone asked for ("3 replicas of `api:1.5.0`").
The *actual state* is what exists ("2 replicas of `api:1.4.2`"). A
*reconcile* function brings them together:

1. **Observe**: read the desired and actual state.
2. **Diff**: work out what is different.
3. **Act**: change only what is different. If nothing is, do nothing.

Running it twice is safe: the second run finds nothing to do. As in
*Idempotency keys*, that is *idempotence*, and controllers need it because
they reconcile the same object many times.

### ConfigMaps

A *ConfigMap* holds configuration as string key/value pairs in its `Data`
field:

```go
cm := &corev1.ConfigMap{
	Data: map[string]string{"timeout": "30", "mode": "fast"},
}
mode := cm.Data["mode"] // "fast"
```

Every value is a string, even numbers. A missing key reads as `""`, so one
`== ""` check covers "missing", "empty" and a nil `Data` map.

### Parsing a number that must fit in `int32`

`strconv.ParseInt(s, base, bitSize)` turns text into a number. With bit size
`32`, it returns an error for numbers too big for an `int32`. It still
returns an `int64`, so convert afterwards:

```go
n, err := strconv.ParseInt("30", 10, 32) // base 10, must fit in 32 bits
// n => 30 (an int64), err => nil
timeout := int32(n)
```

`ParseInt` happily accepts `"-1"`, so check for negatives yourself.

> **Why is it written like this?** Go requires the explicit `int32(n)`
> conversion; it never converts number types silently. Base `10` is what
> humans write; the bit size is your choice of how strict to be.

### Wrapping a sentinel error

As in earlier error exercises, `%w` in `fmt.Errorf` wraps an error so
`errors.Is` can find it, and `%q` puts a value in quotes:

```go
fmt.Errorf("%w: %q", ErrNotFound, "db-0")
// message: not found: "db-0"; errors.Is(err, ErrNotFound) => true
```

### Unset replicas and `ptr.Deref`

`Spec.Replicas` is a `*int32`. When it is `nil`, Kubernetes runs 1 replica.
`ptr.Deref` from `k8s.io/utils/ptr` reads a pointer, with a default for `nil`:

```go
ptr.Deref(pod.Spec.Priority, 0) // the value, or 0 when Priority is nil
```

To set a pointer field, `new(value)` makes a pointer to a copy of the value:
`p.Spec.Priority = new(int32(10))`.

### Never change what you were given

Objects passed to a controller are often shared (as with the cache in
*Informers: listers and the read-only cache*). `DeepCopy()` makes a complete
independent copy that you can change freely.

One trap: `c := d.Spec.Containers[0]` copies the container *struct*, so
changing `c.Image` changes nothing in `d`. Assign through the full path,
or take a pointer with `&`:

```go
c := &pod.Spec.Containers[0] // c points into pod
c.Image = "busybox:1.36"     // pod is changed
```

## Your task

The `shop` team keeps the desired state of their `api` Deployment in a
ConfigMap called `rollout-config`. Here the Deployment always has exactly one
container.

### 1. Read the desired state

**Goal:** make `DesiredFromConfigMap` return a `Desired` built from `cm.Data`, or an error.

- `Replicas`: the `"replicas"` value parsed into an `int32`. `"0"` is fine.
- `Image`: the `"image"` value, copied as it is.
- Check `"replicas"` completely (present, then a valid number) before you
  look at `"image"`, so a bad replicas value is reported even when the image
  is missing too.

On a problem, return `Desired{}` and an error wrapping one of the provided
sentinel errors with `%w`:

| Problem | Error wraps | Example message |
|---|---|---|
| `"replicas"` or `"image"` missing or `""` (also when `Data` is nil) | `ErrMissingKey` | `missing key: "image"` |
| `"replicas"` not a whole number, negative, or too big for `int32` (e.g. `"5000000000"`) | `ErrInvalidReplicas` | `invalid replicas: "three"` |

Add `"fmt"` and `"strconv"` to the imports.

```go
DesiredFromConfigMap(cm) // Data: replicas "3", image "registry.example.com/shop/api:1.5.0"
// => Desired{Replicas: 3, Image: "registry.example.com/shop/api:1.5.0"}, nil

DesiredFromConfigMap(cm) // Data: replicas "-1", image "..."
// => Desired{}, error wrapping ErrInvalidReplicas
```

**Think it through:**
- The keys `"replicas"` and `"image"` are fixed text. Which value goes in the `%q` of each error: the key name or the bad value?
- Which two checks does a parsed number need before it counts as valid?

### 2. List the differences

**Goal:** make `Diff` return one string per difference between `desired` and `current`, replicas first.

1. Replicas, if different: `"replicas: <current> -> <desired>"`. A `nil`
   `Spec.Replicas` counts as 1.
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

**Think it through:**
- In each string, which parts are fixed text (`"replicas: "`, `" -> "`) and which are values?
- Where does the "nil counts as 1" rule come in?

### 3. Apply the desired state

**Goal:** make `Apply` return an updated copy of `current` plus `true` if it changed at least one field.

- Always work on `current.DeepCopy()` and return that copy, even when nothing
  changed. Never modify `current` itself.
- If the replicas differ (`nil` counts as 1), set `Spec.Replicas` to a pointer
  to `desired.Replicas`. Leave it alone otherwise.
- If the first container's image differs, set it to `desired.Image`.

The tests also check idempotence: applying the result again returns `false`.

```go
// current: 2 replicas, image ".../api:1.4.2"
updated, changed := Apply(Desired{Replicas: 5, Image: ".../api:1.4.2"}, current)
// *updated.Spec.Replicas => 5, changed => true, current still has 2

_, changed = Apply(Desired{Replicas: 5, Image: ".../api:1.4.2"}, updated)
// changed => false
```

**Think it through:**
- The comparisons are the same ones you wrote in `Diff`.
- Does your image change land in the copy, or in a copy of the container? (See *Never change what you were given*.)

## Learn more

- [Kubernetes: Controllers](https://kubernetes.io/docs/concepts/architecture/controller/)
- [Kubernetes: ConfigMaps](https://kubernetes.io/docs/concepts/configuration/configmap/)
- [pkg.go.dev: strconv.ParseInt](https://pkg.go.dev/strconv#ParseInt)
- [pkg.go.dev: k8s.io/utils/ptr](https://pkg.go.dev/k8s.io/utils/ptr)
