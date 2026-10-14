# Generics: a Ring of recent deploys

Ops tools keep "the last few" of many things: recent deploys for a rollback
menu, recent log lines for an alert, recent health checks for a dashboard.
Here you finish one generic container that does all of them.

**New here:** generic structs, choosing the type with `NewRing[Deploy](5)`. (`slices.Clone` is from *append and copy*.)

## Learn

### Recap

As in *Generics*, a type can have a type parameter, and each method repeats
it in the receiver. As in *Pointer receivers*, a method that changes a
struct's fields needs a pointer receiver (`*`), or it changes a copy.

### Generic structs

A struct can have fields of type `T`. Here is a different container, a
stack:

```go
type Stack[T any] struct {
	items []T
}

func (s *Stack[T]) Size() int {
	return len(s.items)
}
```

> **Why is it written like this?** Go requires the `[T any]` after the type
> name, and requires `[T]` again in every receiver (`*Stack[T]`), without
> the constraint. The name `T` is convention. The receiver name `s` is the
> author's choice; Go style keeps it short and the same in every method.

### Picking the type when you make one

`NewRing` takes no argument of type `T`, so Go can't guess `T`. You write it
in square brackets:

```go
lines := NewRing[string](100)  // a ring of strings, up to 100
checks := NewRing[bool](10)    // a ring of bools, up to 10
```

### The zero value of T

Inside a generic function you don't know what `T` is, so you can't write
`""` or `0`. `var zero T` makes the zero value of whatever `T` is: `""` for
strings, `0` for numbers, an empty struct for `Deploy`.

### Ends of a slice

The first element is at index `0` and the last is at `len(s)-1`.
Re-slicing with `s[2:]` makes a slice that leaves out the first two
elements:

```go
nums := []int{10, 20, 30, 40}
nums[len(nums)-1] // => 40
nums[2:]          // => [30 40]
```

### Returning a copy

If a method returns its own slice field, the caller gets a slice that shares
the same memory. If the caller changes an element, your struct changes too.
`slices.Clone` makes a new slice with its own memory (needs `"slices"` in
the imports):

```go
func (c *Config) Hosts() []string {
	return slices.Clone(c.hosts) // the caller can't touch c.hosts
}
```

## Your task

Your deploy tool keeps the most recent deploys and log lines in a `Ring[T]`.
`NewRing`, `Len` and part of `Push` are already written. Rules for the whole
ring:

- Items are kept in `r.items`, oldest first, newest last.
- A ring never holds more than `r.limit` items. `limit` is always at least 1.

### 1. Return the newest item

**Goal:** finish `Latest` so it returns the newest item and `true`, or the zero value of `T` and `false` when the ring is empty.

```go
r := NewRing[string](3)
r.Latest() // => "", false
r.Push("deploy started")
r.Push("deploy finished")
r.Latest() // => "deploy finished", true
```

**Think it through:**
- Where in `r.items` does the newest item live: start or end?
- Handle the empty ring first. The two lines already there do exactly that case.

### 2. Forget the oldest item

**Goal:** fix `Push` so that when the ring would hold more than `limit` items, it drops the oldest one.

The `append` line stays. Add your fix after it.

```go
r := NewRing[string](2)
r.Push("v1")
r.Push("v2")
r.Push("v3")
// r now holds "v2", "v3"

r1 := NewRing[string](1)
r1.Push("v1")
r1.Push("v2")
// r1 now holds "v2"
```

**Think it through:**
- After the `append`, compare how many items there are with `r.limit`.
- The oldest item is the first one. How do you make a slice without it?
- Store the result back in `r.items`.

### 3. Hand out a copy

**Goal:** change `Items` so it returns a copy of the items, oldest first, not the ring's own slice.

Changing the returned slice must never change the ring. An empty ring returns
an empty slice (`nil` is fine). If you use `slices.Clone`, add `"slices"` to
the imports.

```go
r := NewRing[Deploy](5)
r.Push(Deploy{Service: "api", Version: "1.4.0"})
r.Push(Deploy{Service: "web", Version: "2.0.1"})
items := r.Items()
// => []Deploy{{Service: "api", Version: "1.4.0"}, {Service: "web", Version: "2.0.1"}}
items[0].Version = "9.9.9"
r.Items()[0].Version // => "1.4.0", the ring is unchanged
```

**Think it through:**
- Only the `return` line changes. What should it hand back instead of `r.items` itself?

## Learn more

- [Tutorial: Getting started with generics](https://go.dev/doc/tutorial/generics)
- [Go by Example: Generics](https://gobyexample.com/generics)
- [Package slices: Clone](https://pkg.go.dev/slices#Clone)
