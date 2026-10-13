# Generics: Sum, SortedKeys and GroupBy

Capacity reports add up numbers of every kind: request counts, CPU seconds,
bytes of disk. Here you meet a constraint that lets one `Sum` handle them all,
and you fix two more generic helpers.

**New here:** union constraints (`~int | ~float64`), the `~` tilde, generic functions with two type parameters, `slices.Sort`.

## Learn

### Recap

As in *Generics*, a *type parameter* goes in square brackets after the
function name, and Go usually works out (*infers*) the real type from the
arguments:

```go
func First[T any](items []T) T { ... }

First([]string{"a", "b"}) // T is string here
```

The *constraint* after the name (`any`, `comparable`, `cmp.Ordered`) says
which types are allowed.

### Your own constraint

A constraint is an interface. Besides methods, it can list types joined by
`|`. That list is called a *union*:

```go
type Whole interface {
	~int8 | ~int16 | ~int32
}

func Triple[T Whole](v T) T {
	return v * 3
}
```

`*` works on every type in the list, so Go allows it on a `T`. With plain
`any`, `v * 3` would not compile, because not every type can be multiplied.

> **Why is it written like this?** Go requires `type Name interface { ... }`
> and `|` between the types. Putting the union on its own line is gofmt's
> layout. The name `Whole` is the author's choice. A union interface can only
> be used inside `[...]`, never as the type of a normal variable (Go rule).

### The tilde

`~int64` means "`int64`, or any type built on `int64`". That includes custom
types like the ones from *Custom types and enums*:

```go
type Bytes int64
```

| Constraint | `int64` allowed? | `Bytes` allowed? |
|---|---|---|
| `int64` | yes | no |
| `~int64` | yes | yes |

The exercise file already has a `Number` constraint with tildes, so `Sum`
works for `Bytes` too. You don't need to change it.

### Two type parameters

A function over maps needs one type parameter for keys and one for values,
separated by a comma. Map keys must support `==`, so the key type is at least
`comparable`:

```go
func CountValues[K comparable, V any](m map[K]V) int {
	return len(m)
}
```

> **Why is it written like this?** The single capital letters `K`, `V`, `T`
> are a community convention, not a rule; `[Key comparable, Val any]` also
> compiles. The order inside `[...]` is the author's choice, but callers who
> write the types out by hand must follow it.

### Sorting a slice

`slices.Sort(s)` sorts `s` in place, smallest first. It returns nothing, so
you don't write `s = ...`. It needs `"slices"` in the imports and works for
any `cmp.Ordered` element type:

```go
nums := []float64{2.5, 0.1, 9}
slices.Sort(nums) // nums is now [0.1 2.5 9]
```

## Your task

You're building the helpers behind a cluster capacity report.

### 1. Fix Sum

**Goal:** make `Sum` return the total of all `values`, not just the last one.

An empty slice sums to `0`. It must work for `int`, `float64` and `Bytes`.

```go
Sum([]int{120, 80, 45})      // => 245
Sum([]float64{0.5, 1.25, 2}) // => 3.75
Sum([]Bytes{512, 1024})      // => 1536
Sum([]int{})                 // => 0
```

**Think it through:**
- Look at the one line inside the loop. What happens to `total` each time round?
- You only change that line. The function signature stays as it is.

### 2. List map keys in order

**Goal:** finish `SortedKeys` so it returns every key of `m`, sorted smallest first.

Map order is random, but a report must look the same every time. For an empty
map, return an empty slice (the `keys` slice that is already made). Sort with
`slices.Sort`, and add `"slices"` to the imports.

```go
SortedKeys(map[string]string{"replicas": "3", "image": "api:1.4", "port": "8080"})
// => []string{"image", "port", "replicas"}

SortedKeys(map[int]string{443: "https", 22: "ssh", 80: "http"})
// => []int{22, 80, 443}
```

**Think it through:**
- `keys` already exists and starts empty. How do you visit each key of a map (see *Maps*)?
- Sort once, after all keys are in, not inside the loop.

### 3. Fix GroupBy

**Goal:** make `GroupBy` keep every item in its group, not just the last one.

`GroupBy` calls the `key` function on each item and puts the item in the group
for that key. The result maps each key to its items, in their original order.
An empty input gives an empty map.

```go
pods := []Pod{
	{Name: "api-1", Node: "node-a"},
	{Name: "api-2", Node: "node-b"},
	{Name: "web-1", Node: "node-a"},
}
GroupBy(pods, func(p Pod) string { return p.Node })
// => map[string][]Pod{
//	"node-a": {{Name: "api-1", Node: "node-a"}, {Name: "web-1", Node: "node-a"}},
//	"node-b": {{Name: "api-2", Node: "node-b"}},
// }

GroupBy([]int{200, 503, 204, 500}, func(code int) int { return code / 100 })
// => map[int][]int{2: {200, 204}, 5: {503, 500}}
```

**Think it through:**
- The line under `TODO(3)` builds a brand-new one-item slice. What should it add to instead?
- What does reading a map key that isn't there yet give you for a slice value?

## Learn more

- [Go spec: Type constraints](https://go.dev/ref/spec#Type_constraints)
- [An Introduction to Generics (Go blog)](https://go.dev/blog/intro-generics)
- [Package slices: Sort](https://pkg.go.dev/slices#Sort)
