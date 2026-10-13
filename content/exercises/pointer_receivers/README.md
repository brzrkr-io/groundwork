# Pointer receivers

Ops code is full of things that change: a counter of failed health checks, a
server that takes new connections, a node that's being drained. A method that
changes its value needs a *pointer receiver*.

## Learn

### The bug: a value receiver changes a copy

In *Methods* every method only read its fields. Watch what happens when a
method with a value receiver tries to change one:

```go
type Counter struct {
	Count int
}

func (c Counter) Inc() {
	c.Count++ // changes the copy, not the caller's counter
}

var hits Counter
hits.Inc()
hits.Inc()
fmt.Println(hits.Count) // 0
```

`(c Counter)` gives the method its own copy of `hits`. `c.Count++` really does
add one, but to the copy, and the copy is thrown away when `Inc` returns.
There's no error and no warning. The change just silently disappears.

### The fix: a pointer receiver

Put a `*` in front of the type in the receiver:

```go
func (c *Counter) Inc() {
	c.Count++ // changes the caller's counter
}

var hits Counter
hits.Inc()
hits.Inc()
fmt.Println(hits.Count) // 2
```

Now `c` is a pointer to `hits`, so `c.Count++` changes `hits` itself. As in
*Pointers*, you write `c.Count`, not `(*c).Count`: Go follows the pointer for
you when you use a field.

### Go adds the `&` for you

`Inc` wants a `*Counter`, but `hits` is a plain `Counter`. When you call a
pointer method on a variable, Go quietly rewrites `hits.Inc()` as
`(&hits).Inc()`. If you already hold a pointer, it works directly:

```go
p := &Counter{}     // p is a *Counter
p.Inc()             // no rewrite needed
fmt.Println(p.Count) // 1
```

So callers write `x.Method()` either way. Only the method declaration decides
whether the original value changes.

### Value or pointer receiver?

| | Value receiver `(c Counter)` | Pointer receiver `(c *Counter)` |
|---|---|---|
| The method gets | a copy of the value | a pointer to the original |
| Changes to fields | lost when the method returns | kept |
| Good for | small types whose methods only read | methods that change fields, big structs |

Three rules of thumb:

1. If the method changes a field, it **must** have a pointer receiver.
2. If the struct is big, a pointer receiver avoids copying all of it on every
   call.
3. Be consistent. If one method of a type needs a pointer receiver, give
   every method of that type a pointer receiver, even the ones that only
   read. Mixing the two is a common source of the bug above.

## Your task

Your load balancer keeps a hit counter and tracks the servers behind it. The
methods were written with value receivers, so none of their changes stick.

### 1. Make the counter count

`Inc` should add one to `Count`, and `Reset` should set `Count` back to `0`.
Both bodies are already right, but their changes are lost. Change both
receivers to pointer receivers.

```go
var c Counter
c.Inc()
c.Inc()
c.Inc()
// c.Count == 3

c.Reset()
// c.Count == 0
```

### 2. Accept a connection

Complete `Accept` and give it a pointer receiver (`*Server`).

- If the server is draining (`Draining` is `true`), it refuses: return `false`
  and leave `Connections` unchanged.
- Otherwise add one to `Connections` and return `true`.

```go
s := &Server{Name: "web-1", Connections: 4}
s.Accept()
// => true, and s.Connections == 5

d := &Server{Name: "web-2", Draining: true, Connections: 3}
d.Accept()
// => false, and d.Connections == 3
```

### 3. Drain the server

Before a deploy you *drain* a server: stop sending it new work and close what
it has. Complete `Drain` (it already has a pointer receiver) so it:

- sets `Draining` to `true`,
- sets `Connections` to `0`,
- returns how many connections it closed, which is the value `Connections`
  had **before** you set it to `0`.

```go
s := &Server{Name: "web-1", Connections: 7}
s.Drain()
// => 7, and now s.Draining == true, s.Connections == 0

s.Accept()
// => false, because the server is draining
```

## Learn more

- [A Tour of Go: Pointer receivers](https://go.dev/tour/methods/4)
- [A Tour of Go: Methods and pointer indirection](https://go.dev/tour/methods/6)
- [A Tour of Go: Choosing a value or pointer receiver](https://go.dev/tour/methods/8)
- [Go FAQ: Should I define methods on values or pointers?](https://go.dev/doc/faq#methods_on_values_or_pointers)
