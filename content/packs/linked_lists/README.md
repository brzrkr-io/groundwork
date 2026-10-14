# Linked lists: a deploy queue with container/list

A *linked list* keeps items in order and lets you move or remove any one of
them without shifting the rest. You'll use Go's `container/list` to run a
deploy queue where a hotfix can jump the line.

**New here:** doubly linked lists, `container/list`, `*list.Element`.

## Learn

### Why not a slice?

A slice keeps items next to each other in memory. To move the item at
index 500 to the front, every item before it shifts one place: 500 copies.
The *Data structures* guide called that O(n): the work grows with the
length.

A *doubly linked list* stores each item in its own *element*, and every
element points to the one before it and the one after it:

```text
front                                          back
[ deploy api ] ⇄ [ deploy web ] ⇄ [ deploy db ]
```

To move `deploy db` to the front, the list changes a few pointers. Nothing
else moves, however long the list is. The catch: there's no `list[2]`. To
reach an element you either walk from one end, or you kept the element
when you added it. That's why the guide says you'll rarely need one; the
LRU cache later in this module is one of the places you do.

### container/list

The standard library's `container/list` is a doubly linked list. Here it
holds recent `kubectl` contexts, newest first:

```go
import "container/list"

recent := list.New()                    // an empty *list.List
prod := recent.PushFront("prod-eu")     // returns the new *list.Element
recent.PushFront("staging")             // order: staging, prod-eu
recent.PushBack("dev")                  // order: staging, prod-eu, dev
recent.MoveToFront(prod)                // order: prod-eu, staging, dev
recent.Front()                          // the element holding "prod-eu"
oldest := recent.Back()                 // the element holding "dev"
recent.Remove(oldest)                   // order: prod-eu, staging
recent.Len()                            // => 2
```

| Method | Does |
|---|---|
| `PushFront(v)`, `PushBack(v)` | add `v` at that end, return its `*list.Element` |
| `Front()`, `Back()` | the first or last element, or `nil` if the list is empty |
| `MoveToFront(el)` | move `el` to the front |
| `Remove(el)` | unlink `el`, and return its value |
| `Len()` | how many elements there are |

`PushFront`'s result is worth keeping: it's how you find that element again
later without a walk.

### Walking a list

Each element has `Next()` (and `Prev()`), which is `nil` past the end. This
loop visits every element from front to back:

```go
for el := recent.Front(); el != nil; el = el.Next() {
	fmt.Println(el.Value) // prod-eu, then staging
}
```

> **Why is it written like this?** `container/list` is a package of the
> standard library, so Go requires the import. The three-part `for` loop
> (start; keep going while; step) is the classic Go loop, and walking a
> list with it is the convention.

### Getting your value back out

An element's `Value` field has type `any`, so the list can hold anything.
As in *Type assertions and switches*, you assert the type you put in:

```go
type alert struct{ name string }

alerts := list.New()
alerts.PushBack(&alert{name: "DiskFull"})
a := alerts.Front().Value.(*alert) // panics if Value isn't an *alert
a.name                             // => "DiskFull"
```

> **Why is it written like this?** The `x.(T)` syntax is what Go requires
> to get a concrete type out of `any`. Using the one-result form (which
> panics on the wrong type) instead of `v, ok := x.(T)` is the author's
> choice when only one type ever goes into the list. Storing pointers is a
> choice too: changing a field through the pointer changes the stored item.

## Your task

Your release train runs one deploy at a time from a queue. Usually it's
first come, first served, but a hotfix jumps to the front. Finish the
`Queue`: it keeps every waiting `*Deploy` in the list `q.deploys`, with the
next one to run at the front.

### 1. Add deploys to the queue

**Goal:** make `Push(d)` add `d` at the **back** of `q.deploys` and return
its element.

```go
q := New()
q.Push(&Deploy{Service: "api", Version: "v1.4.2"})
web := q.Push(&Deploy{Service: "web", Version: "v2.0.0"})
// q.deploys holds: api, web
web.Value // => the *Deploy for web
```

**Think it through:**
- Which end of the list does a newcomer join?
- Which method adds there, and what does it already return?

### 2. List the queue

**Goal:** make `Services()` return the `Service` of every queued deploy,
from the front of `q.deploys` to the back.

An empty queue returns an empty slice (`len` 0).

```go
q.Push(&Deploy{Service: "api", Version: "v1.4.2"})
q.Push(&Deploy{Service: "web", Version: "v2.0.0"})
q.Services() // => ["api" "web"]
```

**Think it through:**
- Which loop visits every element from front to back?
- `el.Value` is `any`. How do you get the `*Deploy` out of it?

### 3. Let a hotfix jump the line

**Goal:** make `Expedite(el)` move `el` to the front of `q.deploys`.

```go
q.Push(&Deploy{Service: "api", Version: "v1.4.2"})
q.Push(&Deploy{Service: "web", Version: "v2.0.0"})
fix := q.Push(&Deploy{Service: "db", Version: "v9.1.1-hotfix"})
q.Expedite(fix)
q.Services() // => ["db" "api" "web"]
```

**Think it through:**
- Which method moves an element you already hold?

### 4. Take the next deploy

**Goal:** make `Next()` remove the deploy at the front of `q.deploys` and
return it with `true`, or return `nil, false` when the queue is empty.

```go
// q.Services() is ["db" "api" "web"]
d, ok := q.Next() // => the *Deploy for db, true
q.Services()      // => ["api" "web"]
New().Next()      // => nil, false
```

**Think it through:**
- What does `Front()` return on an empty list?
- `Remove` returns the value it unlinked, as `any`. What do you do with it?

## Learn more

- [pkg.go.dev: container/list](https://pkg.go.dev/container/list)
- [Wikipedia: Doubly linked list](https://en.wikipedia.org/wiki/Doubly_linked_list)
- [Go by Example: For](https://gobyexample.com/for)
