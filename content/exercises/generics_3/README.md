# Generics: a Ring of recent deploys

Ops tools keep "the last few" of many things: recent deploys for a rollback
menu, recent log lines for an alert, recent health checks for a dashboard.
Here you build one generic container that does all of them.

## Learn

### Recap

As in *Generics*, a type can have a type parameter, and each method repeats it
in the receiver: `func (s Set[T]) Add(v T)`. That `Set` could use a value
receiver because a map shares its data. A struct is different: as in
*Pointer receivers*, a method that changes a struct's fields needs a pointer
receiver, or it changes a copy.

### Generic structs

A struct can hold fields of type `T`:

```go
// Ring keeps the most recent items, up to limit.
type Ring[T any] struct {
	items []T
	limit int
}

func (r *Ring[T]) Len() int {
	return len(r.items)
}
```

To make one, you name the type explicitly, because `NewRing` takes no argument
of type `T` for Go to infer from:

```go
deploys := NewRing[Deploy](5)
lines := NewRing[string](100)
```

### Returning a copy

When a method returns `r.items`, the caller gets a slice that shares the
ring's backing array. If the caller changes an element, the ring changes too.
Return a copy instead. `slices.Clone` makes one:

```go
func (c *Config) Hosts() []string {
	return slices.Clone(c.hosts) // the caller can't touch c.hosts
}
```

## Your task

Your deploy tool keeps the most recent deploys and log lines in a `Ring[T]`.
`NewRing`, `Len` and part of `Push` are written. A ring never holds more than
its `limit` items, and `limit` is always at least 1. Items are kept oldest
first.

### 1. Return the newest item

Finish `Latest`. It returns the most recently pushed item and `true`. For an
empty ring it returns the zero value of `T` and `false`.

```go
r := NewRing[string](3)
r.Latest() // => "", false
r.Push("deploy started")
r.Push("deploy finished")
r.Latest() // => "deploy finished", true
```

### 2. Forget the oldest item

`Push` adds an item at the end, but the ring grows past its limit. Fix `Push`
so that when the ring would hold more than `limit` items, it drops the oldest
one.

```go
r := NewRing[string](2)
r.Push("v1")
r.Push("v2")
r.Push("v3")
// r now holds "v2", "v3"
```

### 3. Hand out a copy

`Items` returns the items, oldest first, but it returns the ring's own slice.
Change it to return a copy, so changing the result never changes the ring. An
empty ring returns an empty slice (`nil` is fine). If you use `slices.Clone`,
add `"slices"` to the imports.

```go
r := NewRing[Deploy](5)
r.Push(Deploy{Service: "api", Version: "1.4.0"})
r.Push(Deploy{Service: "web", Version: "2.0.1"})
items := r.Items()
// => []Deploy{{Service: "api", Version: "1.4.0"}, {Service: "web", Version: "2.0.1"}}
items[0].Version = "9.9.9"
r.Items()[0].Version // => "1.4.0", the ring is unchanged
```

## Learn more

- [Tutorial: Getting started with generics](https://go.dev/doc/tutorial/generics)
- [Go by Example: Generics](https://gobyexample.com/generics)
- [Package slices: Clone](https://pkg.go.dev/slices#Clone)
