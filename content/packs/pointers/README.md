# Pointers

A pointer lets a function change a variable that belongs to someone else.
Ops code uses them everywhere: to update a deployment in place, to find one
host in a long list and edit it, or to mark a config field as "not set".

**New here:** `&` (address of), `*T` (pointer type), `*p` (follow a
pointer), `nil` pointers.

## Learn

### Variables live at addresses

Think of a variable as a labelled box in memory. Every box has an *address*:
where it lives. The `&` operator gives you a variable's address. A value that
holds an address is a *pointer*.

```go
count := 3
p := &count        // p points at count
fmt.Println(*p)    // 3: *p reads the value p points to
*p = 10            // write through the pointer...
fmt.Println(count) // 10: ...and count changed
```

The symbols mean different things depending on where they appear:

| You write | Where | Meaning |
|---|---|---|
| `*int` | in a type | "pointer to an int" |
| `&x` | in an expression | the address of `x` |
| `*p` | in an expression | the value `p` points to (read it, or assign to it) |

So `p := &count` gives `p` the type `*int`.

> **Why is it written like this?** `&` and `*` are Go's required operators;
> there is no other spelling. Putting `*p` on the *left* of `=` writes to
> the pointed-at variable; plain `p = ...` would change which box `p`
> points at instead. Naming a pointer `p` is just the author's choice.

### nil: a pointer to nothing

The zero value of every pointer type is `nil`: "points nowhere". Reading or
writing `*p` when `p` is `nil` crashes the program with a *panic*. When a
pointer might be missing, compare it with `nil` before you follow it:

```go
func describe(port *int) string {
	if port == nil {
		return "no port set"
	}
	return fmt.Sprint("port ", *port)
}
```

A function with no return values can stop early with a plain `return`, as
in *Structs: changes that vanish*: `if p == nil { return }`.

### Why functions need pointers

Go passes every argument as a **copy**. A function that changes its
parameter only changes its own copy:

```go
func double(n int)     { n = n * 2 }   // bug: changes the copy only
func doublePtr(n *int) { *n = *n * 2 } // changes the caller's variable

x := 5
double(x)     // x is still 5
doublePtr(&x) // x is now 10
```

`doublePtr` gets a copy of the *address*, but that copy still points at the
caller's `x`, so writing to `*n` writes into `x`.

> **Why is it written like this?** In `func doublePtr(n *int)`, the `*int`
> is the parameter's type and is required: it tells callers to pass an
> address (`&x`), not a number.

### Pointers to structs

A pointer can point at a struct. Go lets you write `p.Field` instead of the
clumsy `(*p).Field`; it follows the pointer for you.

```go
s := Service{Name: "api", Port: 80}
p := &s
p.Port = 8080      // same as (*p).Port = 8080
fmt.Println(s.Port) // 8080
```

### Pointing into a slice

`&list[i]` is the address of an element *inside* the slice. Changing
something through that pointer changes the slice.

The variable in `for _, v := range list` is a **copy** of each element, as in
*Structs: changes that vanish*. A pointer to it, `&v`, points at the copy, so
changes through it never reach the slice.

```go
ports := []int{80, 443}
p := &ports[1]
*p = 8443
fmt.Println(ports) // [80 8443]
```

## Your task

Your deploy tool tracks a restart counter and a list of deployments. Write the
helpers that change them in place. `exercise.go` declares this struct:

```go
type Deployment struct {
	Name     string
	Replicas int
}
```

### 1. Reset a counter

**Goal:** make `ResetCounter(count)` set the `int` that `count` points to
back to `0`.

You can assume `count` is never `nil`.

```go
restarts := 7
ResetCounter(&restarts)
// restarts is now 0
```

**Think it through:**
- `count` is an address, not a number. How do you reach the number it
  points to?

### 2. Scale a deployment

**Goal:** make `Scale(d, replicas)` set the `Replicas` field of the
deployment `d` points to.

- If `d` is `nil`, do nothing. It must not crash.
- If `replicas` is negative, set `Replicas` to `0` instead.

```go
d := Deployment{Name: "api", Replicas: 2}
Scale(&d, 5)
// d.Replicas is now 5
Scale(&d, -1)
// d.Replicas is now 0
Scale(nil, 3)
// nothing happens
```

**Think it through:**
- What must you check *first*, before touching any field?
- The new count is `replicas`. What should it become when it's below
  zero?

### 3. Find a deployment

**Goal:** make `FindDeployment(ds, name)` return a pointer to the element of
`ds` whose `Name` equals `name`, or `nil` if there is none.

The pointer must point **into** `ds`, so a change made through it shows up
in the slice. The tests check this. An empty or `nil` `ds` gives `nil`.

```go
ds := []Deployment{{Name: "api", Replicas: 2}, {Name: "web", Replicas: 3}}
p := FindDeployment(ds, "web")
p.Replicas = 10
// ds[1].Replicas is now 10

FindDeployment(ds, "db")
// => nil
```

**Think it through:**
- Which kind of `range` loop gives you a way to name the real element, not
  a copy?
- The function's return type is `*Deployment`. Which operator turns an
  element into its address?

## Learn more

- [A Tour of Go: Pointers](https://go.dev/tour/moretypes/1)
- [A Tour of Go: Pointers to structs](https://go.dev/tour/moretypes/4)
- [Go by Example: Pointers](https://gobyexample.com/pointers)
- [Go spec: Address operators](https://go.dev/ref/spec#Address_operators)
