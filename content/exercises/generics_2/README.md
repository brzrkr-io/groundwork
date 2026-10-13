# Generics: Sum, SortedKeys and GroupBy

Capacity reports add up numbers of every kind: request counts, CPU seconds,
bytes of disk. Here you write your own constraint so one `Sum` handles them
all, and you fix two more generic helpers.

## Learn

### Recap

As in *Generics*, a type parameter goes in square brackets, like
`func Max[T cmp.Ordered](items []T) (T, bool)`, and Go usually infers `T` from
the arguments. The constraint says which types are allowed: `any`,
`comparable` or `cmp.Ordered`. A function can take several type parameters,
separated by commas.

### Your own constraint

A constraint is an interface. Besides methods, a constraint interface can list
types, separated by `|`. This list is called a *union*:

```go
type Number interface {
	~int | ~int64 | ~float64
}

func Double[T Number](v T) T {
	return v + v
}
```

`+` works on every type in the list, so it is allowed on a `T`. An interface
with a union can only be used as a constraint, not as the type of a normal
variable.

### The tilde

`~int64` means "`int64`, or any type whose underlying type is `int64`". That
includes custom types like the ones from *Custom types and enums*:

```go
type Bytes int64
```

| Constraint | `int64` allowed? | `Bytes` allowed? |
|---|---|---|
| `int64` | yes | no |
| `~int64` | yes | yes |

Without the `~`, `Sum([]Bytes{512, 1024})` would not compile.

### Maps with two type parameters

A generic function over maps needs one type parameter for the keys and one
for the values. Map keys must support `==`, so `K` is at least `comparable`:

```go
func Values[K comparable, V any](m map[K]V) []V {
	values := make([]V, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}
	return values
}
```

## Your task

You're building the helpers behind a cluster capacity report.

### 1. Fix Sum

`Sum` should return the total of all `values`, but it returns only the last
one. Fix the bug. An empty slice sums to `0`.

```go
Sum([]int{120, 80, 45})      // => 245
Sum([]float64{0.5, 1.25, 2}) // => 3.75
Sum([]Bytes{512, 1024})      // => 1536
Sum([]int{})                 // => 0
```

### 2. List map keys in order

Map order is random, but a report must be the same every time. Finish
`SortedKeys` so it returns every key of `m`, sorted smallest first. For an
empty map it returns an empty slice. Sort with `slices.Sort`, which means
adding `"slices"` to the imports.

```go
SortedKeys(map[string]string{"replicas": "3", "image": "api:1.4", "port": "8080"})
// => []string{"image", "port", "replicas"}

SortedKeys(map[int]string{443: "https", 22: "ssh", 80: "http"})
// => []int{22, 80, 443}
```

### 3. Fix GroupBy

`GroupBy` calls `key` on each item and puts the item in the group for that
key. The result maps each key to its items, in their original order. The
current code keeps only the last item of each group. Fix it. An empty input
gives an empty map.

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

## Learn more

- [Go spec: Type constraints](https://go.dev/ref/spec#Type_constraints)
- [An Introduction to Generics (Go blog)](https://go.dev/blog/intro-generics)
- [Package slices: Sort](https://pkg.go.dev/slices#Sort)
