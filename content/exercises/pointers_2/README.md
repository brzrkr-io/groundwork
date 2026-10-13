# Pointers: fix the silent no-ops

The nastiest pointer bugs don't crash. The code runs, the function returns,
and nothing changed. In this exercise you hunt down three of them in a
maintenance tool.

## Learn

### Recap

As in *Pointers*, `&x` gives you the address of `x`, `*T` is the type
"pointer to `T`", and `*p` reads or writes the value `p` points to. Go passes
every argument as a copy, so a function can only change the caller's
variable if it receives a pointer to it. With a struct pointer, `p.Field`
follows the pointer for you.

### Copying a pointer vs copying the value

Assigning a pointer copies the **address**, so both pointers share one value.
Assigning `*p` copies the **value**, so you get an independent copy:

```go
a := &Host{Name: "node-1"}

b := a       // b and a point at the same Host
b.Name = "x" // a.Name is now "x" too

v := *a      // v is a separate Host with the same fields
v.Name = "y" // a.Name is still "x"
```

### Returning a pointer to a local variable

In Go it is safe for a function to return the address of one of its own
local variables. Go keeps the variable alive for as long as something
points at it:

```go
func newHost(name string) *Host {
	h := Host{Name: name}
	return &h // fine in Go
}
```

## Your task

Your team's fleet maintenance tool swaps ports during blue/green deploys,
copies host records and drains hosts before patching. All three features
quietly do the wrong thing. The exercise uses this struct:

```go
type Host struct {
	Name     string
	Drained  bool
	Sessions int
}
```

### 1. Fix SwapPorts

`SwapPorts(live, standby *int)` should swap the two **ints** they point to,
so the standby port becomes live and the old live port becomes the
standby. Right now it swaps the two local pointer variables, which changes
nothing for the caller.

```go
live, standby := 8080, 8081
SwapPorts(&live, &standby)
// live is now 8081, standby is now 8080
```

### 2. Fix CloneHost

`CloneHost(h *Host) *Host` should return a pointer to a **new** `Host` with
the same fields as `*h`. Changing the clone must not change the original,
and the returned pointer must not be equal to `h`. Right now it returns `h`
itself, so the "clone" and the original are the same host.

```go
orig := &Host{Name: "node-1", Sessions: 4}
c := CloneHost(orig)
c.Name = "node-1-canary"
// orig.Name is still "node-1"; c.Sessions is 4
```

### 3. Fix DrainHost

`DrainHost(hosts []Host, name string) bool` finds the host called `name`,
marks it drained with the helper `markDrained`, and returns `true`. If no
host has that name it returns `false` and changes nothing. A drained host has
`Drained` set to `true` and `Sessions` set to `0`.

The search and the return values already work, but the host in the slice
never changes: `markDrained` takes a `Host` value, so it only drains a copy.
Change `markDrained` to take a `*Host`, then fix the call in `DrainHost` so it
passes a pointer to the element **inside** `hosts`.

```go
hosts := []Host{{"node-1", false, 12}, {"node-2", false, 3}}
DrainHost(hosts, "node-2")
// => true
// hosts[1] is now {"node-2", true, 0}

DrainHost(hosts, "node-9")
// => false
```

## Learn more

- [A Tour of Go: Pointers](https://go.dev/tour/moretypes/1)
- [Go FAQ: When are function parameters passed by value?](https://go.dev/doc/faq#pass_by_value)
- [Go FAQ: How do I know whether a variable is allocated on the heap or the stack?](https://go.dev/doc/faq#stack_or_heap)
