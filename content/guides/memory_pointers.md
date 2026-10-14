# Memory, pointers and sharing

**In one sentence:** every value lives at an address in memory, and Go
copies values unless you share them on purpose with a pointer, or by
accident through a slice or map.

**Why it matters:** "I scaled the deployment to 5 but it's still 3" and "I
edited the canary's labels and prod changed too" are the same bug in two
costumes: not knowing whether you're holding a copy or a shared original.

**You'll learn:**

- What memory and addresses are, and what a variable really is.
- Why Go copies values, and how `&` and `*` let you share one instead.
- What `nil` means and the three good reasons to use a pointer.
- How slices, maps and copied structs can share data behind your back.
- What the stack, the heap and the garbage collector do, in plain words.
- How this decides between value and pointer receivers on methods.

## Memory is a row of numbered boxes

Your computer's memory (RAM) is a very long row of tiny boxes. Each box holds
one *byte*, and each has a number, its *address*. Nothing more magic than
that: a street of houses, each with a house number.

```text
  address        contents
  ...
  0xc000012000   +---------------------+
                 |          3          |   <- replicas (an int, 8 bytes)
  0xc000012008   +---------------------+
                 |    "api" (header)   |   <- name (a string header)
  0xc000012018   +---------------------+
  ...
```

Addresses are usually written in *hexadecimal* (base 16, the `0x...`
numbers). You'll never need to calculate with them.

A *variable* is a name the compiler gives to some boxes. When you write
`replicas := 3`, Go picks enough boxes for an `int` (8 bytes on a 64-bit
machine), puts 3 in them, and remembers that `replicas` means "the boxes at
that address".

Bigger values take more boxes. A struct with five fields takes roughly the
sum of its fields. A slice or string variable holds a small *header* that
points at the real data somewhere else (see the *Data structures* guide).

## Go copies values

Here is the rule that explains most surprises in this guide:

> **Assigning a value, or passing it to a function, makes a copy.**

You'll meet it in *Structs: changes that vanish*. Here it is with a deploy tool:

```go
package main

import "fmt"

type Deployment struct {
	Name     string
	Replicas int
}

// scaleCopy gets its own copy of the Deployment.
func scaleCopy(d Deployment, n int) {
	d.Replicas = n
}

// scale gets the address of the caller's Deployment.
func scale(d *Deployment, n int) {
	d.Replicas = n
}

func main() {
	api := Deployment{Name: "api", Replicas: 3}

	scaleCopy(api, 5)
	fmt.Println("after scaleCopy:", api.Replicas)

	scale(&api, 5)
	fmt.Println("after scale:", api.Replicas)
}
```

It prints:

```text
after scaleCopy: 3
after scale: 5
```

Trace it:

1. `api` lives at some address in `main`, holding `{api 3}`.
2. `scaleCopy(api, 5)` copies all of `api` into the parameter `d`: a second
   Deployment at a different address. It sets *that* copy to 5, then
   returns, and the copy is thrown away. `api` never changed.
3. `scale(&api, 5)` passes the *address* of `api`. `d` is a copy of the
   address, but a copy of an address still leads to the same house. Setting
   `d.Replicas` changes `api` itself.

It's like the difference between emailing someone a copy of the config file
and giving them SSH access to the server it lives on.

## Pointers: `&` and `*`

A *pointer* is a value that holds an address. Two operators work with them:

| You write | Meaning | Example |
|---|---|---|
| `&x` | "the address of x" | `p := &replicas` |
| `*T` (in a type) | "pointer to a T" | `var p *int` |
| `*p` (in code) | "the value at the address in p" | `*p = 10` |

```text
   p (a *int)                  replicas (an int)
  +--------------+            +-----+
  | 0xc000012000 | ---------> |  3  |   at address 0xc000012000
  +--------------+            +-----+

  *p = 10   follows the arrow and writes 10 into replicas
```

If you `fmt.Println(p)` you'll see an address like `0xc000012000`. The exact
number changes from run to run; never rely on it.

With structs, Go saves you some typing: `d.Replicas` on a `*Deployment`
means `(*d).Replicas`. The pointer is followed for you.

## nil: a pointer to nothing

The zero value of every pointer type is `nil`: an address that points
nowhere. Following it is a crash:

```text
panic: runtime error: invalid memory address or nil pointer dereference
```

If you've been on call for Go services, you've probably seen that line in a
crash-looping pod's logs. The fix is always the same: check before you
follow.

### `nil` as "not set"

`nil` is also useful. It gives you a way to say "there is no value", which
plain numbers can't. Is a timeout of `0` "no timeout" or "the user didn't
say"? A pointer can tell them apart, as you practise in
*Pointers: optional config fields*:

```go
package main

import "fmt"

type Probe struct {
	Path    string
	Timeout *int // seconds; nil means "not set"
}

func effectiveTimeout(p Probe) int {
	if p.Timeout == nil {
		return 30 // the default
	}
	return *p.Timeout
}

func main() {
	web := Probe{Path: "/healthz"}
	batch := Probe{Path: "/ready", Timeout: new(0)}

	fmt.Println(web.Path, effectiveTimeout(web))
	fmt.Println(batch.Path, effectiveTimeout(batch))
}
```

It prints:

```text
/healthz 30
/ready 0
```

`web` never set `Timeout`, so it's `nil` and gets the default 30. `batch`
set it explicitly to 0 with `new(0)` (a pointer to a new int holding 0;
`new(value)` needs Go 1.26 or later), and that choice is respected.
Kubernetes' own Go types use this trick: `Replicas` in a Deployment spec is
an `*int32`, so "not set" and "zero replicas" mean different things.

## When to use a pointer

| Reason | Example | Without a pointer |
|---|---|---|
| Change the caller's value | `scale(&api, 5)` | the change vanishes |
| Avoid copying a big struct | a large config struct, many calls | copies it every call |
| Say "maybe absent" | `Timeout *int`; find returns `nil` | 0 looks like "set" |

And when *not* to:

- **Small values you only read.** An `int`, a `string`, a small struct:
  copying is cheap and safe, and nobody can change it under you.
- **"Pointers are faster."** That's a myth as a general rule. A pointer
  avoids a copy, but it can force the value onto the heap (below), which
  costs garbage-collector work. Use pointers for the reasons in the table,
  not for speed you haven't measured.

## Sharing surprises: aliasing

When two names reach the same memory, they are *aliases*. A change through
one shows up through the other. Pointers alias on purpose. Slices and maps
alias too, and that can surprise you, because there's no `&` in sight.

```go
package main

import "fmt"

type Deployment struct {
	Name   string
	Labels map[string]string
}

func main() {
	// 1. Two slices, one backing array.
	hosts := []string{"web-1", "web-2", "web-3", "web-4"}
	canaries := hosts[:2]
	canaries[0] = "DRAINED"
	fmt.Println(hosts)

	// 2. A copied struct still shares its map.
	prod := Deployment{
		Name:   "api",
		Labels: map[string]string{"env": "prod"},
	}
	canary := prod
	canary.Name = "api-canary"
	canary.Labels["track"] = "canary"
	fmt.Println(prod.Name, prod.Labels)
	fmt.Println(canary.Name, canary.Labels)
}
```

It prints:

```text
[DRAINED web-2 web-3 web-4]
api map[env:prod track:canary]
api-canary map[env:prod track:canary]
```

What happened:

1. `hosts[:2]` makes a new slice *header*, but it points into the same
   backing array as `hosts`. Writing `canaries[0]` writes `hosts[0]`.
2. `canary := prod` copies the struct. `Name` is copied, so renaming the
   canary leaves prod alone. But a map variable is a small reference to the
   real map; copying it copies the reference, not the entries. Both
   Deployments now share one label map, and prod just got `track: canary`.

```text
   prod.Labels   ---+
                    +--->  { env: prod, track: canary }
   canary.Labels ---+
```

**The fix is an explicit copy** when you want independence:
`slices.Clone(hosts[:2])` for a slice, `maps.Clone(prod.Labels)` for a
map. *Slice gotchas: shared memory* covers the slice cases in depth.

**A copy is "shallow".** Copying a struct copies its fields. If a field is a
slice, map or pointer, you copy the reference, and the data behind it is
still shared. This is exactly how a Kubernetes controller can corrupt its
cache: it "copies" an object, edits the labels, and has edited the cached
original. Client libraries provide `DeepCopy` methods for this reason; you'll
meet them in the *Kubernetes with client-go* module.

Sharing is also where concurrency bugs come from. When two parts of a
program run at the same time and change the same map, that's a *data race*.
That's the *Concurrency* module's story.

## Stack, heap and the garbage collector

Go puts values in one of two areas of memory. You don't choose; the compiler
does. It still helps to know what they are.

```text
  STACK                                  HEAP (shared)
  +----------------------------+        +---------------------------+
  | main:  api, hosts          |        |  values that outlive the  |
  +----------------------------+        |  function that made them  |
  | scale: d, n                |        |                           |
  +----------------------------+        |  freed later by the       |
    ^ a frame is pushed per call,       |  garbage collector        |
      popped when the call returns      +---------------------------+
```

**The stack** holds a function's local variables in a *frame*. When the
function returns, the whole frame is discarded at once. That's very cheap,
like the undo stack in the *Data structures* guide.

**The heap** holds values that must outlive the function that created them,
for example a value whose address is returned, or stored somewhere
long-lived. Deciding which values go where is called *escape analysis*, and
the compiler does it for you.

```go
func newDeployment(name string) *Deployment {
	d := Deployment{Name: name, Replicas: 1}
	return &d // safe in Go: d moves to the heap
}
```

In some older languages, returning the address of a local variable is a
serious bug. In Go it's safe and normal: the compiler sees `&d` escape and
puts `d` on the heap.

**The garbage collector (GC)** cleans the heap. Every so often it starts
from what the program can still reach (global variables, the stack) and
follows every pointer. Anything it can't reach is garbage, and
its memory is reused. It runs mostly alongside your program, but it does use
CPU time.

Two ops consequences:

- **Reachable means kept.** A small slice of a huge array keeps the whole
  array alive. A map you keep adding to and never clear (a cache with no
  expiry, say) grows forever. In a container with a memory limit, that ends
  in an `OOMKilled` pod.
- **Fewer heap values, less GC work.** That's why the "pointers are always
  faster" myth is a myth. Measure before tuning.

## Methods: value or pointer receiver?

This section looks ahead to *Methods* and *Pointer receivers*; come back
to it after those if it moves too fast. A *method* is a function attached to
a type: `func (d Deployment) Summary() string` is called as `d.Summary()`.
The `(d Deployment)` before the name is the *receiver*. It's just a
parameter, so the same copy rule applies:

- A **value receiver** `(d Deployment)` gets a copy. Fine for methods that
  only read.
- A **pointer receiver** `(d *Deployment)` gets the address. Required for
  methods that change the value.

Here's a rolling deploy helper that scales every deployment to zero:

```go
package main

import "fmt"

type Deployment struct {
	Name     string
	Replicas int
}

// Summary only reads, so a value receiver (a copy) is fine.
func (d Deployment) Summary() string {
	return fmt.Sprintf("%s=%d", d.Name, d.Replicas)
}

// ScaleTo changes the Deployment, so it needs a pointer receiver.
func (d *Deployment) ScaleTo(n int) {
	d.Replicas = n
}

func main() {
	deps := []Deployment{
		{Name: "api", Replicas: 3},
		{Name: "web", Replicas: 2},
	}

	for _, d := range deps {
		d.ScaleTo(0) // d is a copy of the element
	}
	fmt.Println(deps[0].Summary(), deps[1].Summary())

	for i := range deps {
		deps[i].ScaleTo(0) // deps[i] is the element itself
	}
	fmt.Println(deps[0].Summary(), deps[1].Summary())
}
```

It prints:

```text
api=3 web=2
api=0 web=0
```

Trace it:

1. In the first loop, `d` is a fresh copy of each element. `d.ScaleTo(0)`
   is shorthand for `(&d).ScaleTo(0)`: Go takes the address of the *copy*,
   so the copy is scaled and then discarded. Nothing in `deps` changes, and
   the compiler doesn't warn you.
2. In the second loop, `deps[i]` names the element inside the slice's
   backing array, so `ScaleTo` changes the real thing.

**Rule of thumb:** if any method on a type needs a pointer receiver, give
them all pointer receivers, so the type behaves one way. The Go FAQ linked
below says the same.

## Common mistakes

- **Changing a copy.** A function or method that takes a struct by value
  and edits it does nothing visible. Take a `*T`, or return the new value.
- **Editing the `range` variable.** `for _, d := range deps { d.X = ... }`
  edits a copy. Use `for i := range deps { deps[i].X = ... }`.
- **Following a nil pointer.** Check `if p == nil` before `*p` or `p.Field`
  whenever a pointer can be missing, such as the result of a "find".
- **Assuming a struct copy is independent.** Copied structs still share
  their slices and maps. Clone them when you need a separate copy.
- **Sub-slicing and then appending or writing.** The new slice shares the
  backing array, so writes land in the original.
- **Using pointers "for speed".** Use them to share, to change, or to mean
  "absent". Measure before using them to make code faster.

## Check yourself

1. `func bump(n int) { n++ }` is called as `bump(count)`. Why doesn't
   `count` change, and what two edits make it work?
2. What's the difference between `*int` in `var p *int` and `*p` in
   `*p = 5`?
3. A config struct has `MaxRetries int`. Why can't you tell "the user set 0"
   from "the user said nothing", and how does a pointer fix it?
4. You copy a Deployment struct, change the copy's `Labels["env"]`, and the
   original's labels change too. Why, given that Go copies values?
5. Is returning `&d` for a local variable `d` a bug in Go? What happens to
   `d`?
6. A method `Reset()` sets a counter field to zero, but the counter never
   resets. What is most likely wrong with its receiver?

## Answers

1. `n` is a copy of `count`; `n++` changes only the copy. Change the
   parameter to `n *int` with body `*n++`, and call it as `bump(&count)`.
2. In a type, `*int` means "pointer to an int". In code, `*p` means "follow
   the pointer": here, write 5 into the int that `p` points to.
3. An `int` always holds a number, and its zero value is 0, so "unset" and
   "set to 0" look identical. With `MaxRetries *int`, `nil` means "not set"
   and a pointer to 0 means "explicitly zero".
4. Go did copy the struct, but the `Labels` field holds a reference to a
   map, and copying the field copies only the reference. Both structs point
   at one map. Use `maps.Clone` to give the copy its own.
5. No. The compiler sees the address escape the function and puts `d` on
   the heap. It stays alive as long as something can reach it; then the
   garbage collector frees it.
6. It has a value receiver, `(c Counter)`, so it resets a copy. It needs a
   pointer receiver, `(c *Counter)`.

## Key terms

| Term | Meaning |
|---|---|
| memory (RAM) | A long row of byte-sized boxes the program uses. |
| address | The number of a box in memory. |
| variable | A name for a value stored at some address. |
| pointer | A value holding the address of another value. |
| `&x` | The address of `x`. |
| `*p` | The value at the address held in `p`. |
| nil | The zero value of a pointer: points nowhere. |
| copy semantics | Assignment and calls copy the value. |
| alias | Two names reaching the same memory. |
| shallow copy | Copying fields only; shared data behind them stays shared. |
| stack | Per-call memory, freed instantly when a function returns. |
| heap | Memory for values that outlive their function. |
| escape analysis | The compiler deciding which values go on the heap. |
| garbage collector | Frees heap memory the program can no longer reach. |
| receiver | The value a method is called on: `d` in `func (d T)`. |

## Watch

- [Go Syntax - Pointers Part 1](https://www.youtube.com/watch?v=i5nyPaAwM3s) · Ardan Labs · 5 min: Bill Kennedy on how Go passes values by copying them, the rule behind *Go copies values*.
- [Go Class: 14 Reference & Value Semantics](https://www.youtube.com/watch?v=904pyovPvXM) · Matt KØDVB · 24 min: when Go copies and when it shares, with the slice and map aliasing surprises from *Sharing surprises: aliasing*.
- [Understanding Allocations: the Stack and the Heap - GopherCon SG 2019](https://www.youtube.com/watch?v=ZMZpH4yT7M0) · Singapore Gophers · 22 min: the deeper talk, with clear diagrams of what the compiler puts on the stack and the heap, and why.
- [CS50x 2025 - Lecture 4 - Memory](https://www.youtube.com/watch?v=kcRdFGbzR1I) · CS50 · 147 min: Harvard's classic lecture on addresses and pointers in C, including the swap that fails on copies; watch the first hour if that's all you have.

## Go deeper

- [A Tour of Go: Pointers](https://go.dev/tour/moretypes/1): short
  interactive pages on `&`, `*` and pointers to structs.
- [Go FAQ: when are function parameters passed by value?](https://go.dev/doc/faq#pass_by_value):
  the official word on copies, slices and maps.
- [Go FAQ: should I define methods on values or pointers?](https://go.dev/doc/faq#methods_on_values_or_pointers):
  the reasoning behind the receiver rule of thumb.
- [A Guide to the Go Garbage Collector](https://go.dev/doc/gc-guide):
  stack, heap, escape analysis and GC costs, with interactive graphs.
- [OSTEP: The Abstraction: Address Spaces](https://pages.cs.wisc.edu/~remzi/OSTEP/vm-intro.pdf):
  a free OS textbook chapter on how a program's memory is laid out.
