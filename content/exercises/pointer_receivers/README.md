# Pointer receivers

Ops code is full of things that change: a counter of failed health checks, a
server that takes new connections, a node that's being drained. A method that
changes its value needs a *pointer receiver*.

**New here:** pointer receivers `(x *T)`, choosing value or pointer receiver.

## Learn

### The bug: a value receiver changes a copy

So far every method only read its fields. Watch what happens when a method
with a value receiver tries to change one:

```go
type Disk struct {
	UsedGB int
}

func (d Disk) Write(gb int) {
	d.UsedGB += gb // changes the copy, not the caller's disk
}

var sda Disk
sda.Write(5)
fmt.Println(sda.UsedGB) // 0
```

`(d Disk)` gives the method its own copy of `sda`. The `+=` really happens,
but to the copy, which is thrown away when `Write` returns. No error, no
warning: the change just disappears.

### The fix: a pointer receiver

Put a `*` in front of the type in the receiver:

```go
func (d *Disk) Write(gb int) {
	d.UsedGB += gb // changes the caller's disk
}

var sda Disk
sda.Write(5)
fmt.Println(sda.UsedGB) // 5
```

Now `d` points at `sda`, so the change lands on `sda` itself. The body did
not change at all: only the receiver line did.

> **Why is it written like this?** Go requires the `*` to go on the type
> (`*Disk`), not the name. You still write `d.UsedGB`, not `(*d).UsedGB`: as
> in *Pointers*, Go follows the pointer for you when you use a field.

### Go adds the `&` for you

`Write` wants a `*Disk`, but `sda` is a plain `Disk`. When you call a pointer
method on a variable, Go quietly rewrites `sda.Write(5)` as
`(&sda).Write(5)`. If you already hold a pointer, it works directly:

```go
p := &Disk{}  // p is a *Disk
p.Write(2)    // no rewrite needed
```

So callers write `x.Method()` either way. Only the method's receiver line
decides whether the original changes.

### Value or pointer receiver?

| | Value receiver `(d Disk)` | Pointer receiver `(d *Disk)` |
|---|---|---|
| The method gets | a copy of the value | a pointer to the original |
| Changes to fields | lost when the method returns | kept |
| Good for | small types whose methods only read | methods that change fields, big structs |

Rules of thumb:

1. If the method changes a field, it **must** have a pointer receiver (Go
   accepts a value receiver, but your change is lost).
2. Be consistent: if one method of a type needs a pointer receiver, give
   every method of that type one. This is Go convention.

## Your task

Your load balancer keeps a hit counter and tracks the servers behind it. The
methods were written with value receivers, so none of their changes stick.

### 1. Make the counter count

**Goal:** make `Inc` and `Reset` change the caller's `Counter` by changing
only their receiver lines.

`Inc` adds one to `Count`; `Reset` sets `Count` back to `0`. Both bodies are
already right. Don't touch them.

```go
var c Counter
c.Inc()
c.Inc()
c.Inc()
// c.Count == 3

c.Reset()
// c.Count == 0
```

**Think it through:**
- Compare the `Disk` receiver before and after the fix. Which one character
  changed?
- Where in the receiver does that character go: next to the name or next to
  the type?

### 2. Accept a connection

**Goal:** `Accept` takes a new connection and returns `true`, unless the
server is draining. Give it a pointer receiver (`*Server`).

- If `Draining` is `true`: return `false` and leave `Connections` unchanged.
- Otherwise: add one to `Connections` and return `true`.

```go
s := &Server{Name: "web-1", Connections: 4}
s.Accept()
// => true, and s.Connections == 5

d := &Server{Name: "web-2", Draining: true, Connections: 3}
d.Accept()
// => false, and d.Connections == 3
```

**Think it through:**
- Which case do you check first, so you can return early?
- Which field changes, and only in which case?

### 3. Drain the server

**Goal:** `Drain` marks the server as draining, closes all its connections
and returns how many it closed. It already has a pointer receiver.

Before a deploy you *drain* a server: stop sending it new work and close what
it has. `Drain` must:

- set `Draining` to `true`,
- set `Connections` to `0`,
- return the value `Connections` had **before** you set it to `0`.

```go
s := &Server{Name: "web-1", Connections: 7}
s.Drain()
// => 7, and now s.Draining == true, s.Connections == 0

s.Accept()
// => false, because the server is draining
```

**Think it through:**
- Once `Connections` is `0`, the old number is gone. Where can you keep it
  first?

## Learn more

- [A Tour of Go: Pointer receivers](https://go.dev/tour/methods/4)
- [A Tour of Go: Methods and pointer indirection](https://go.dev/tour/methods/6)
- [A Tour of Go: Choosing a value or pointer receiver](https://go.dev/tour/methods/8)
- [Go FAQ: Should I define methods on values or pointers?](https://go.dev/doc/faq#methods_on_values_or_pointers)
