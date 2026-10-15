# Pointers: optional config fields

How does a config tell "replicas: 0" apart from "replicas not set"? Kubernetes
answers with pointers: a `nil` pointer means "not set". Here you use the same
pattern to apply defaults and layer config overrides.

**New here:** pointer fields as "optional" values, `new(value)`, checking for `nil` before reading.

## Learn

### Recap

As in *Pointers*: `*int` is a pointer to an `int`, `nil` is the zero value of
every pointer type, and `*p` reads the value `p` points to. Reading `*p` when
`p` is `nil` **crashes** the program, so check `p != nil` first.

### The problem with zero values

A plain `int` field is always *something*. Leave `Replicas` out and it is `0`.
Write `replicas: 0` to scale down and it is also `0`. The code can't tell
them apart.

A pointer field has a third state, `nil`:

| Field value | Meaning |
|---|---|
| `nil` | not set: use the default (or keep the old value) |
| points at `0` | set, and the value is `0` |
| points at `5` | set, and the value is `5` |

Kubernetes does exactly this: a Deployment's `Replicas` has type `*int32`,
and `nil` means 1 replica.

```go
type Probe struct {
	Path           string
	TimeoutSeconds *int // nil means "not set"
}

func describe(p Probe) string {
	if p.TimeoutSeconds == nil {
		return p.Path + ": no timeout"
	}
	return fmt.Sprintf("%s: %ds timeout", p.Path, *p.TimeoutSeconds)
}
```

> **Why is it written like this?** Go requires the `*` to read the `int`
> behind the pointer. Checking `nil` first and returning early is convention
> (it keeps the "normal" path unindented). The names are the author's choice.

### Making a pointer to a value

You can't take the address of a literal: `&5` does not compile. Since Go
1.26, the built-in `new(value)` makes a brand-new variable holding `value` and
returns a pointer to it:

```go
port := new(8080)      // port is an *int, and *port == 8080
tag := new("v1.4")     // tag is a *string
p := Probe{Path: "/healthz", TimeoutSeconds: new(5)}
```

- The type comes from the value: `new(5)` is an `*int`, `new(true)` a `*bool`.
- Every call makes a **new** variable, so `new(5) == new(5)` is `false`: two different pointers to equal values.
- `new(x)` with a variable `x` copies its current value; changing `x` later doesn't change what the pointer sees.

Older code (before Go 1.26) uses a tiny helper instead, like
`func StrPtr(s string) *string { return &s }` or Kubernetes' `ptr.To(...)`.
They do the same thing as `new(...)`.

### Changing a struct parameter

Go passes a struct by value, so a struct parameter is already the
function's own copy. You can set its fields and return it; the caller's
original is untouched:

```go
func WithPath(p Probe, path string) Probe {
	p.Path = path
	return p
}
```

## Your task

Your platform team deploys services from a base config plus layers of
overrides (team defaults, then the environment, then a one-off flag). Write
the helpers that resolve the final values. The exercise uses these types:

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

**Goal:** make `WithReplicas(spec DeploymentSpec, n int)` return `spec` with its `Replicas` field pointing at a new `int` holding `n`.

- `Name` stays as it was.
- Every call makes a **new** pointer, even for the same `n`.
- Any `n` works, including `0` and negatives.

```go
s := WithReplicas(DeploymentSpec{Name: "api"}, 3)
// s.Name == "api", *s.Replicas == 3

spec := DeploymentSpec{Name: "api"}
WithReplicas(spec, 3).Replicas == WithReplicas(spec, 3).Replicas
// => false (two different pointers)
```

**Think it through:**
- The field's type is `*int` but `n` is an `int`. Which built-in turns a value into a pointer to a fresh copy of it?
- The `return spec` line is already there. What do you set just before it?

### 2. Write EffectiveReplicas

**Goal:** make `EffectiveReplicas(spec DeploymentSpec) int` return the replica count `spec` asks for, or `DefaultReplicas` when it isn't set.

- `spec.Replicas` is `nil` → return the constant `DefaultReplicas` (which is `1`).
- Otherwise → return the `int` it points to. Pointing at `0` means "scale to zero", so return `0`.

```go
EffectiveReplicas(DeploymentSpec{Name: "api"})
// => 1
EffectiveReplicas(DeploymentSpec{Name: "api", Replicas: new(4)})
// => 4
EffectiveReplicas(DeploymentSpec{Name: "batch", Replicas: new(0)})
// => 0
```

**Think it through:**
- Which case must you handle first so you never read through a `nil` pointer?
- In the other case, `spec.Replicas` is an address. What symbol gets you the number?

### 3. Write Merge

**Goal:** make `Merge(base Config, layers ...Overrides) Config` start from `base` and apply each layer in order.

- For each field, a non-`nil` pointer in a layer replaces that field in the result with the value it points to.
- A `nil` field leaves the value as it is.
- Later layers win over earlier ones.
- With no layers (or only empty layers), `Merge` returns `base` unchanged.

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

**Think it through:**
- `layers` is a slice (as in *Variadic functions*). Which loop visits each layer in order?
- You need a result variable that starts as a copy of `base`. Each of the three fields then needs its own `nil` check.
- `Config` fields are plain values, `Overrides` fields are pointers. Which side needs the `*`?

## Learn more

- [Kubernetes: Deployment replicas](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#replicas)
- [k8s.io/api/apps/v1: DeploymentSpec](https://pkg.go.dev/k8s.io/api/apps/v1#DeploymentSpec)
- [k8s.io/utils/ptr: To](https://pkg.go.dev/k8s.io/utils/ptr#To)
- [Go 1.26 release notes: new with an expression](https://go.dev/doc/go1.26#language)
