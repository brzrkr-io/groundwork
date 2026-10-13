# Structs: changes that vanish

A health checker keeps a list of hosts and flips them between healthy and
unhealthy. Its code looks right, but none of its changes stick. The cause is
one of the most common struct bugs in Go: changing a copy.

## Learn

### Recap

A struct groups named fields into one value, as in *Structs*. You create one
with a literal like `Host{Name: "web-1"}`, and read or set fields with a dot:
`h.Healthy = true`.

### Structs are copied

A struct is a value, like an `int`. Every time you put it in a new variable,
you get a separate copy. Changing the copy does not touch the original. That
happens in three places that are easy to miss:

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

### Changing the struct in the slice

To change an element of a slice, assign through the slice itself with
`hosts[i]`. A `for i := range hosts` loop gives you just the index:

```go
for i := range hosts {
	hosts[i].Healthy = false // changes the element in the slice
}
```

If you did make a copy, write it back when you're done: `hosts[i] = h`.

For a function, return the changed struct and let the caller store it:

```go
func drain(h Host) Host {
	h.Healthy = false
	return h
}

hosts[0] = drain(hosts[0])
```

## Your task

Each function in `exercise.go` changes a copy instead of the hosts in the
slice, so the caller never sees the change. Fix them so every change lands in
`hosts` itself.

### 1. Mark every host unhealthy

`MarkAllUnhealthy(hosts)` sets `Healthy` to `false` on every host in `hosts`.
Other fields stay as they are.

```go
hosts := []Host{{Name: "web-1", Healthy: true}, {Name: "web-2", Healthy: true}}
MarkAllUnhealthy(hosts)
// hosts[0].Healthy == false, hosts[1].Healthy == false
```

### 2. Record a failed check

`RecordFailure(hosts, name)` finds the host whose `Name` is `name` and adds 1
to its `Failures`. When `Failures` reaches `3` or more, it also sets `Healthy`
to `false`. If no host has that name, nothing changes.

```go
hosts := []Host{{Name: "db-1", Healthy: true, Failures: 2}}
RecordFailure(hosts, "db-1")
// hosts[0] is now {Name:db-1 Healthy:false Failures:3}
```

### 3. Reset every host

`ResetAll(hosts)` calls the helper `reset` on every host. `reset` should set
`Failures` to `0` and `Healthy` to `true`. Change `reset` so it returns the
updated `Host`, and have `ResetAll` store that result back in the slice.

```go
hosts := []Host{{Name: "web-1", Failures: 4}, {Name: "web-2", Failures: 1}}
ResetAll(hosts)
// both hosts now have Healthy == true and Failures == 0
```

## Learn more

- [Go FAQ: Why are function arguments passed by value?](https://go.dev/doc/faq#pass_by_value)
- [Go spec: For statements with range clause](https://go.dev/ref/spec#For_range)
- [Go by Example: Structs](https://gobyexample.com/structs)
