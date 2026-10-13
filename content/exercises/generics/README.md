# Generics

Ops code is full of "the same job on different data": keep the healthy hosts
from a list of names, keep the open ports from a list of numbers. *Generics*
let you write that function once and use it with any type.

**New here:** type parameters `[T any]`, constraints (`any`, `comparable`, `cmp.Ordered`), `var zero T`, generic types.

## Learn

### The problem

Without generics you write one function per type:

```go
func CountStrings(items []string) int { return len(items) }
func CountInts(items []int) int       { return len(items) }
```

The bodies are identical. Only the type changes.

### Type parameters

A generic function has a *type parameter* in square brackets, between its
name and its normal parameters:

```go
func Last[T any](items []T) T {
	return items[len(items)-1]
}
```

Read `[T any]` as "`T` stands for some type, and any type is allowed". Inside
the function you use `T` like any other type name: `[]T`, `func(T) bool`,
`var x T`.

> **Why is it written like this?** The square brackets and their position are
> required by Go. The name `T` is convention (short, capital); `T, U` for a
> second one. You could write `[Item any]` and it would still compile.

### Calling a generic function

Go usually works out `T` from the arguments. This is *type inference*:

```go
Last([]string{"web-1", "web-2"}) // T is string, returns "web-2"
Last([]int{80, 443})             // T is int, returns 443
```

You can also spell it out, as in `Last[int](ports)`. You only need to when
there is no argument to infer it from.

### Constraints

The word after `T` is its *constraint*: which types are allowed. It also
decides what you may do with a `T` inside the function.

| Constraint | Allowed types | What you can do with a `T` |
|---|---|---|
| `any` | every type | store it, pass it, return it |
| `comparable` | types that work with `==`: numbers, strings, bools, structs of those | also `==`, `!=`, use as a map key |
| `cmp.Ordered` | numbers and strings | also `<`, `>`, `<=`, `>=` |

`cmp.Ordered` lives in the `cmp` package, so the file needs `import "cmp"`
(the starter has it). With `any` you can't write `a < b`; the compiler stops
you, because not every type can be ordered.

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

Sometimes you must return "nothing" of type `T`. You don't know whether `T`
is `int` (zero is `0`) or `string` (zero is `""`), so declare a variable and
let Go fill in the zero value: `var zero T`. The starter for `Max` shows it.

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
type Stack[T any] []T

func (s Stack[T]) Size() int {
	return len(s)
}
```

`Stack[string]` holds strings and `Stack[int]` holds ints.

A map value refers to shared data, so a method with a value receiver on a map
type can still add keys, just like a function you pass a map to.

> **Why is it written like this?** Go requires `[T]` in the receiver
> (`Stack[T]`), but not the constraint; that's already on the type.

### You've used generics already

`slices.Contains`, `slices.Index` and `slices.Sort` are generic functions from
the standard library. That's why one `slices.Sort` works on `[]int` and
`[]string` alike.

## Your task

You're writing helpers for a deploy dashboard. They must work for host names,
port numbers and latencies alike.

### 1. Keep the items that match

**Goal:** make `Filter(items, keep)` return a new slice of the items for which `keep` returns `true`.

Keep the original order. If nothing matches, return an empty slice (`nil` is
fine).

```go
Filter([]string{"web-1", "db-1", "web-2"}, func(h string) bool {
	return strings.HasPrefix(h, "web-")
})
// => []string{"web-1", "web-2"}

Filter([]int{22, 80, 443, 8080}, func(p int) bool { return p >= 443 })
// => []int{443, 8080}
```

**Think it through:**
- `keep` is a parameter holding a function. How do you call it with one item?
- Look at the `Map` example: it loops and appends. What changes when you only want some items?

### 2. Find the largest item

**Goal:** make `Max(items)` return the largest item and `true`, or the zero value of `T` and `false` for an empty slice.

Strings compare alphabetically. Watch out for negative numbers: the largest
of `-5`, `-2` and `-9` is `-2`.

```go
Max([]float64{120.5, 340.2, 98.7})          // => 340.2, true
Max([]string{"web-01", "web-03", "web-02"}) // => "web-03", true
Max([]int{-5, -2, -9})                      // => -2, true
Max([]int{})                                // => 0, false
```

**Think it through:**
- Handle the empty slice first; the starter's last two lines already return that answer.
- What should you start "largest so far" at? Starting at `0` breaks the negative example.

### 3. Remember what you've seen

**Goal:** finish the `Add` and `Has` methods of `Set[T]`.

`Set[T]` is a `map[T]bool`. `NewSet` is already written; read it to see how
it puts items in.

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

**Think it through:**
- Inside a method, the receiver `s` *is* the map. Use it like any map.
- What does looking up a missing key in a `map[T]bool` give you?

## Learn more

- [Tutorial: Getting started with generics](https://go.dev/doc/tutorial/generics)
- [An Introduction to Generics (Go blog)](https://go.dev/blog/intro-generics)
- [Package cmp: Ordered](https://pkg.go.dev/cmp#Ordered)
- [Go by Example: Generics](https://gobyexample.com/generics)
