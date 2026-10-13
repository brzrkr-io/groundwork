# Generics

Ops code is full of "the same job on different data": keep the healthy hosts
from a list of names, keep the open ports from a list of numbers. *Generics*
let you write that function once and use it with any type.

## Learn

### The problem

Without generics you write one function per type:

```go
func FilterStrings(items []string, keep func(string) bool) []string { ... }
func FilterInts(items []int, keep func(int) bool) []int { ... }
```

The bodies are identical. Only the type changes.

### Type parameters

A generic function has a *type parameter* in square brackets, before its
normal parameters:

```go
func Last[T any](items []T) T {
	return items[len(items)-1]
}
```

Read `[T any]` as "`T` stands for some type, and any type is allowed". Inside
the function you use `T` like any other type name: `[]T`, `func(T) bool`,
`var x T`.

### Calling a generic function

Go usually works out `T` from the arguments. This is called *type inference*:

```go
Last([]string{"web-1", "web-2"}) // T is string, returns "web-2"
Last([]int{80, 443})             // T is int, returns 443
```

You can also spell the type out, as in `Last[int](ports)`. You only need to
when there is no argument to infer it from.

### Constraints

The word after `T` is its *constraint*: which types are allowed. The
constraint also decides what you can do with a `T` inside the function.

| Constraint | Allowed types | What you can do with a `T` |
|---|---|---|
| `any` | every type | store it, pass it, return it |
| `comparable` | types that work with `==`: numbers, strings, bools, structs of those | also `==`, `!=`, use it as a map key |
| `cmp.Ordered` | numbers and strings | also `<`, `>`, `<=`, `>=` |

`cmp.Ordered` lives in the `cmp` package, so the file needs `import "cmp"`.
With `any` you can't write `a < b`, because not every type can be ordered. The
compiler stops you.

```go
func Contains[T comparable](items []T, want T) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}
```

### The zero value of T

Sometimes you have to return "nothing" of type `T`. You don't know whether `T`
is `int` (zero is `0`) or `string` (zero is `""`), so declare a variable and
let Go fill in the zero value:

```go
var zero T
return zero, false
```

### More than one type parameter

Separate type parameters with commas. `Map` turns each item into something
else, possibly of a different type:

```go
func Map[T, U any](items []T, f func(T) U) []U {
	out := make([]U, 0, len(items))
	for _, item := range items {
		out = append(out, f(item))
	}
	return out
}

Map([]int{80, 443}, strconv.Itoa) // []string{"80", "443"}
```

### Generic types

Types can have type parameters too. Each method repeats `[T]` in its receiver:

```go
type Set[T comparable] map[T]bool

func (s Set[T]) Size() int {
	return len(s)
}
```

`Set[string]` is a set of strings and `Set[int]` a set of ints. A map value
refers to shared data, so a method with a value receiver can still add keys,
just like a function you pass a map to.

### You've used generics already

`slices.Contains`, `slices.Index` and `slices.Sort` are generic functions from
the standard library. That's why one `slices.Sort` works on `[]int` and
`[]string` alike.

## Your task

You're writing helpers for a deploy dashboard. They must work for host names,
port numbers and latencies alike.

### 1. Keep the items that match

Finish `Filter`. It returns a new slice holding the items for which `keep`
returns `true`, in their original order. If nothing matches, return an empty
slice (`nil` is fine).

```go
Filter([]string{"web-1", "db-1", "web-2"}, func(h string) bool {
	return strings.HasPrefix(h, "web-")
})
// => []string{"web-1", "web-2"}

Filter([]int{22, 80, 443, 8080}, func(p int) bool { return p >= 443 })
// => []int{443, 8080}
```

### 2. Find the largest item

Finish `Max`. It returns the largest item and `true`. For an empty slice it
returns the zero value of `T` and `false`. Strings compare alphabetically.
Watch out for negative numbers: the largest of `-5`, `-2` and `-9` is `-2`.

```go
Max([]float64{120.5, 340.2, 98.7})          // => 340.2, true
Max([]string{"web-01", "web-03", "web-02"}) // => "web-03", true
Max([]int{-5, -2, -9})                      // => -2, true
Max([]int{})                                // => 0, false
```

### 3. Remember what you've seen

`Set[T]` is a set built on `map[T]bool`. `NewSet` is already written. Finish
its two methods:

- `Add(v T)` puts `v` in the set. Adding the same value twice is fine.
- `Has(v T) bool` reports whether `v` is in the set.

```go
seen := NewSet("us-east-1")
seen.Add("eu-west-1")
seen.Has("eu-west-1")  // => true
seen.Has("ap-south-1") // => false

ports := NewSet(80, 443)
ports.Has(443) // => true
```

## Learn more

- [Tutorial: Getting started with generics](https://go.dev/doc/tutorial/generics)
- [An Introduction to Generics (Go blog)](https://go.dev/blog/intro-generics)
- [Package cmp: Ordered](https://pkg.go.dev/cmp#Ordered)
- [Go by Example: Generics](https://gobyexample.com/generics)
