# Structs: changes that vanish

A health checker keeps a list of hosts and flips them between healthy and
unhealthy. Its code looks right, but none of its changes stick. The cause is
one of the most common struct bugs in Go: changing a copy.

**New here:** struct copies, changing slice elements in place with `s[i]`.

## Learn

### Recap

A struct groups named fields into one value, as in *Structs*. You create one
with a literal like `Host{Name: "web-1"}`, and read or set fields with a dot:
`h.Healthy = true`.

### Structs are copied

A struct is a value, like an `int`. Every time you put it in a new variable,
you get a separate copy. Changing the copy does not touch the original.

Three places make a copy, and they're easy to miss:

| Code | `h` is... |
|---|---|
| `h := hosts[i]` | a copy of `hosts[i]` |
| `for _, h := range hosts` | a copy of each element, one at a time |
| `func reset(h Host)` | a copy of whatever the caller passed in |

```go
hosts := []Host{{Name: "web-1", Healthy: true}}

h := hosts[0]
h.Healthy = false
fmt.Println(hosts[0].Healthy) // true: only the copy changed
```

> **Why is it written like this?** Copying is Go's rule, not a choice: `=`,
> `range` and function parameters always copy a struct. The compiler won't
> warn you, because changing a copy is legal. It just does nothing useful.

### Changing the struct in the slice

To change an element, assign *through the slice*: `hosts[i].Field = ...`.
That writes to the element itself, not a copy. A `for i := range` loop gives
you just the index:

```go
services := []Service{{Name: "api"}, {Name: "web"}} // Service from *Structs*
for i := range services {
	services[i].Port = 8080 // changes the element in the slice
}
```

> **Why is it written like this?** `for i := range s` with one variable is
> allowed by Go and gives only the index. Calling it `i` is convention.

If you did make a copy, you can write it back when you're done:
`hosts[i] = h`.

### Getting a change out of a function

A function gets a copy, so its changes stay inside it. Give the function a
*return type*, return the changed copy, and let the caller store it:

```go
func drain(s Service) Service {
	s.Port = 0
	return s
}

services[0] = drain(services[0])
```

> **Why is it written like this?** In `func drain(s Service) Service`, Go
> requires the parameter list in `()` and the return type after it. Leave
> the return type out and `return s` won't compile. The names `drain` and
> `s` are the author's choice.

## Your task

Each function in `exercise.go` changes a copy, so the caller never sees the
change. Fix them so every change lands in `hosts` itself.

### 1. Mark every host unhealthy

**Goal:** make `MarkAllUnhealthy(hosts)` set `Healthy` to `false` on every
host in the slice.

Other fields stay as they are.

```go
hosts := []Host{{Name: "web-1", Healthy: true}, {Name: "web-2", Healthy: true}}
MarkAllUnhealthy(hosts)
// hosts[0].Healthy == false, hosts[1].Healthy == false
```

**Think it through:**
- In the starter loop, which row of the table is `h`?
- What can you write instead of `h` so the change reaches the slice?

### 2. Record a failed check

**Goal:** make `RecordFailure(hosts, name)` add 1 to `Failures` on the host
whose `Name` is `name`, inside the slice.

When that host's `Failures` reaches `3` or more, also set its `Healthy` to
`false`. If no host has that name, nothing changes. Other hosts never change.

```go
hosts := []Host{{Name: "db-1", Healthy: true, Failures: 2}}
RecordFailure(hosts, "db-1")
// hosts[0] is now {Name:db-1 Healthy:false Failures:3}
```

**Think it through:**
- The loop and the `if` already find the right host. The bug is on the
  `h := hosts[i]` line.
- Either change the element directly, or write the copy back before
  `return`.

### 3. Reset every host

**Goal:** make `ResetAll(hosts)` leave every host with `Failures` `0` and
`Healthy` `true`.

The work happens in the helper `reset`. Change two places:

- `reset` (at the bottom of the file) must *return* the updated `Host`.
- `ResetAll` must store what `reset` returns back into the slice.

```go
hosts := []Host{{Name: "web-1", Failures: 4}, {Name: "web-2", Failures: 1}}
ResetAll(hosts)
// both hosts now have Healthy == true and Failures == 0
```

**Think it through:**
- Which part of the `func reset(...)` line says what it returns? Right now
  there's nothing there.
- Once `reset` returns a `Host`, where in `ResetAll` should that value go?

## Learn more

- [Go FAQ: Why are function arguments passed by value?](https://go.dev/doc/faq#pass_by_value)
- [Go spec: For statements with range clause](https://go.dev/ref/spec#For_range)
- [Go by Example: Structs](https://gobyexample.com/structs)
