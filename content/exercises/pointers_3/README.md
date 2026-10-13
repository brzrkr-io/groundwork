# Pointers: optional config fields

How does a config tell "replicas: 0" apart from "replicas not set"? Kubernetes
answers with pointers: a `nil` pointer means "not set". In this exercise you
use the same pattern to apply defaults and layer config overrides.

## Learn

### Recap

As in *Pointers*, `*int` is a pointer to an `int`, `nil` is the zero value of
every pointer type, and `*p` reads the value `p` points to. Reading `*p` when
`p` is `nil` crashes, so check `p != nil` first.

### The problem with zero values

A plain `int` field is always set to *something*. If a user leaves `Replicas`
out, it is `0`. If they write `replicas: 0` to scale a service down, it is
also `0`. The code can't tell the two apart.

A pointer field has a third state, `nil`:

| Field value | Meaning |
|---|---|
| `nil` | not set: use the default (or keep the old value) |
| points at `0` | set, and the value is `0` |
| points at `5` | set, and the value is `5` |

Kubernetes does exactly this. A Deployment's `Replicas` field has type
`*int32`, and when it is `nil` Kubernetes uses 1 replica.

```go
type DeploymentSpec struct {
	Name     string
	Replicas *int // nil means "use the default"
}

func describe(spec DeploymentSpec) string {
	if spec.Replicas == nil {
		return "replicas not set"
	}
	return fmt.Sprintf("replicas = %d", *spec.Replicas)
}
```

### Making a pointer to a value

You can't take the address of a literal: `&5` does not compile. Since Go
1.26, the built-in `new` does it for you. `new(value)` makes a brand-new
variable holding `value` and returns a pointer to it:

```go
p := new(5)         // p is an *int, and *p == 5
s := new("web:2.0") // s is a *string

o := Overrides{Image: new("web:2.0"), Replicas: new(3)}
```

The type comes from the value: `new(5)` is an `*int`, `new(true)` a `*bool`.
Every call makes a new variable, so `new(5) == new(5)` is `false`: two
different pointers that happen to point at equal values. `new(n)` with a
variable `n` copies its current value, so changing `n` later doesn't change
`*p`.

You'll see an older style in existing code, from before Go 1.26: a tiny helper
that returns the address of its parameter, such as `func StrPtr(s string)
*string { return &s }`, or Kubernetes' `ptr.To(...)`. They do the same as
`new(...)`, and your editor offers to replace them with it.

## Your task

Your platform team deploys services from a base config plus layers of
overrides (team defaults, then the environment, then a one-off command-line
flag). Write the helpers that resolve the final values.

The exercise uses these types:

```go
const DefaultReplicas = 1

type DeploymentSpec struct {
	Name     string
	Replicas *int
}

type Config struct {
	Image    string
	Replicas int
	Debug    bool
}

type Overrides struct {
	Image    *string
	Replicas *int
	Debug    *bool
}
```

### 1. Write WithReplicas

Change `WithReplicas(spec DeploymentSpec, n int) DeploymentSpec` so it
returns a copy of `spec` whose `Replicas` points at a new `int` holding `n`.
`spec` is already a copy (Go passes structs by value), so set its field and
return it. Every call must make a new pointer, even for the same `n`.

```go
s := WithReplicas(DeploymentSpec{Name: "api"}, 3)
// s.Name == "api", *s.Replicas == 3

spec := DeploymentSpec{Name: "api"}
WithReplicas(spec, 3).Replicas == WithReplicas(spec, 3).Replicas
// => false (two different pointers)
```

### 2. Write EffectiveReplicas

Change `EffectiveReplicas` so it returns the replica count `spec` asks for.
When `spec.Replicas` is `nil`, return `DefaultReplicas` (which is `1`). A
`Replicas` pointing at `0` means "scale to zero", so it returns `0`.

```go
EffectiveReplicas(DeploymentSpec{Name: "api"})
// => 1
EffectiveReplicas(DeploymentSpec{Name: "api", Replicas: new(4)})
// => 4
EffectiveReplicas(DeploymentSpec{Name: "batch", Replicas: new(0)})
// => 0
```

### 3. Write Merge

Change `Merge(base Config, layers ...Overrides) Config` so it starts from
`base` and applies each layer in order:

- For each field, a non-`nil` pointer in the layer replaces the value in the
  result with the value it points to.
- A `nil` field leaves the value as it is.
- Later layers win over earlier ones.
- With no layers, `Merge` returns `base` unchanged.

```go
base := Config{Image: "web:1.0", Replicas: 3, Debug: true}

Merge(base, Overrides{Replicas: new(0)})
// => Config{Image: "web:1.0", Replicas: 0, Debug: true}

Merge(base,
	Overrides{Image: new("web:2.0"), Replicas: new(5)},
	Overrides{Replicas: new(2), Debug: new(false)},
)
// => Config{Image: "web:2.0", Replicas: 2, Debug: false}
```

## Learn more

- [Kubernetes: Deployment replicas](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#replicas)
- [k8s.io/api/apps/v1: DeploymentSpec](https://pkg.go.dev/k8s.io/api/apps/v1#DeploymentSpec)
- [k8s.io/utils/ptr: To](https://pkg.go.dev/k8s.io/utils/ptr#To)
- [Go 1.26 release notes: new with an expression](https://go.dev/doc/go1.26#language)
