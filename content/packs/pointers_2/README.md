# Pointers: fix the silent no-ops

The nastiest pointer bugs don't crash. The code runs, the function returns,
and nothing changed. Here you hunt down three of them in a maintenance tool.

**New here:** copying a pointer vs copying the value it points to, returning the address of a local variable, the address of a slice element.

## Learn

### Recap

As in *Pointers*:

| Syntax | Means |
| --- | --- |
| `*T` (in a type) | "pointer to a `T`" |
| `&x` | the address of `x` |
| `*p` (in code) | the value `p` points to (read **or** write) |
| `p.Field` | for a struct pointer, Go follows the pointer for you |

Go passes every argument as a **copy**. A function can only change the
caller's variable if it gets a pointer to it.

### Changing the pointer vs changing the value

A pointer parameter is itself a local copy of an address. Assigning to the
*name* only changes that local copy. Assigning through `*` changes the value
the caller sees:

```go
func resetWrong(count *int) {
	count = new(int) // only this function's copy of the address changes
}

func resetRight(count *int) {
	*count = 0 // the caller's int changes
}
```

Go's multiple assignment (`a, b = x, y`) works with `*` too. Each side may be
any assignable thing: a variable, `*p`, `s[i]`, `h.Name`.

> **Why is it written like this?** Go requires the `*` to reach the pointed-to
> value; without it you're working with the address. Which variable goes
> where in a swap is your choice; the compiler just evaluates the whole right
> side first, then assigns.

### Copying a pointer vs copying the value

Assigning a pointer copies the **address**, so both names share one value.
Assigning `*p` copies the **value**, so you get an independent copy:

```go
cfg := &Config{Region: "eu-west-1"}

same := cfg           // same address: one Config, two names
same.Region = "x"     // cfg.Region is now "x" too

snapshot := *cfg      // a separate Config with the same fields
snapshot.Region = "y" // cfg.Region is still "x"
```

### Returning a pointer to a local variable

In Go it is safe for a function to return `&` of its own local variable. Go
keeps the variable alive as long as something points at it:

```go
func newConfig(region string) *Config {
	c := Config{Region: region}
	return &c // fine in Go (not in C!)
}
```

> **Why is it written like this?** Go requires the return type `*Config` to
> match what you return, so you return an address (`&c`), not `c`. Naming the
> local `c` is the author's choice.

### The address of a slice element

`s[i]` is a real variable stored inside the slice, so `&s[i]` points at it.
`for _, v := range s` gives you `v`, a **copy**, so `&v` would not point
into the slice:

```go
ports := []int{80, 443}
p := &ports[1]
*p = 8443 // ports is now [80 8443]
```

## Your task

Your fleet maintenance tool swaps ports during blue/green deploys, copies host
records and drains hosts before patching. All three quietly do nothing. The
exercise uses this struct:

```go
type Host struct {
	Name     string
	Drained  bool
	Sessions int
}
```

### 1. Fix SwapPorts

**Goal:** make `SwapPorts(live, standby *int)` swap the two **ints** the pointers point to.

Right now it swaps the two local pointer variables, which the caller never
sees. After the call, the caller's `live` holds the old standby port and
`standby` holds the old live port. Equal ports stay equal.

```go
live, standby := 8080, 8081
SwapPorts(&live, &standby)
// live is now 8081, standby is now 8080
```

**Think it through:**
- In the existing line, is it changing the addresses or the ints?
- Which symbol turns "a pointer" into "the value it points at"? Where must it appear, left side, right side, or both?

### 2. Fix CloneHost

**Goal:** make `CloneHost(h *Host) *Host` return a pointer to a **new** `Host` with the same fields as `*h`.

Right now it returns `h` itself, so the "clone" is the original. Requirements:

- the returned pointer is not `nil` and is not equal to `h`;
- every field (`Name`, `Drained`, `Sessions`) matches the original;
- changing the clone does not change the original.

```go
orig := &Host{Name: "node-1", Sessions: 4}
c := CloneHost(orig)
c.Name = "node-1-canary"
// orig.Name is still "node-1"; c.Sessions is 4
```

**Think it through:**
- Which gives you a separate copy: assigning the pointer, or assigning the value it points to?
- The return type is `*Host`. Once you have your copy in a local variable, what do you put in front of it to return its address?

### 3. Fix DrainHost

**Goal:** make `DrainHost(hosts []Host, name string) bool` actually drain the matching host **inside the slice**.

Behaviour (the search and the return values already work):

- finds the host called `name`, drains it, returns `true`;
- if no host has that name, returns `false` and changes nothing;
- a drained host has `Drained` set to `true` and `Sessions` set to `0`.

The bug: the helper `markDrained` takes a `Host` value, so it only drains a
copy. This task has two edits, both marked `TODO(3)`:

1. change `markDrained`'s parameter type so it receives a pointer;
2. change the call in `DrainHost` so it passes the address of the element inside `hosts`.

```go
hosts := []Host{{"node-1", false, 12}, {"node-2", false, 3}}
DrainHost(hosts, "node-2")
// => true
// hosts[1] is now {"node-2", true, 0}

DrainHost(hosts, "node-9")
// => false
```

**Think it through:**
- Does the body of `markDrained` need to change? Remember what `p.Field` does with a struct pointer.
- After you change the parameter type, the compiler will flag the call. What must you add in front of the slice element so its type matches?

## Learn more

- [A Tour of Go: Pointers](https://go.dev/tour/moretypes/1)
- [Go FAQ: When are function parameters passed by value?](https://go.dev/doc/faq#pass_by_value)
- [Go FAQ: How do I know whether a variable is allocated on the heap or the stack?](https://go.dev/doc/faq#stack_or_heap)
