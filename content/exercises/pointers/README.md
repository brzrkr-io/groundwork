# Pointers

A pointer lets a function change a variable that belongs to someone else.
Ops code uses them everywhere: to update a deployment in place, to find one
host in a long list and edit it, or to mark a config field as "not set".

## Learn

### Variables live at addresses

Think of a variable as a labelled box in memory. Every box has an *address*:
where it lives. The `&` operator gives you a variable's address, and a value
that holds an address is called a *pointer*.

```go
count := 3
p := &count        // p points at count
fmt.Println(*p)    // 3: *p reads the value p points to
*p = 10            // write through the pointer...
fmt.Println(count) // 10: ...and count changed
```

The `*` means two things, depending on where it appears:

| You write | Where | Meaning |
|---|---|---|
| `*int` | in a type | "pointer to an int" |
| `&x` | in an expression | the address of `x` |
| `*p` | in an expression | the value `p` points to (read it or assign to it) |

So `p := &count` gives `p` the type `*int`.

### nil: a pointer to nothing

The zero value of every pointer type is `nil`, which means "points nowhere".
Reading or writing `*p` when `p` is `nil` crashes the program with a *panic*,
so check first when a pointer might be missing:

```go
if p == nil {
	return
}
```

### Why functions need pointers

Go passes every argument as a **copy**. A function that changes its
parameter only changes its own copy:

```go
func reset(n int)     { n = 0 }  // bug: changes the copy only
func resetPtr(n *int) { *n = 0 } // changes the caller's variable

x := 5
reset(x)     // x is still 5
resetPtr(&x) // x is now 0
```

`resetPtr` gets a copy of the *address*, but that copy still points at the
caller's `x`, so `*n = 0` writes into `x`.

### Pointers to structs

A pointer can point at a struct. Go lets you write `p.Field` instead of the
clumsy `(*p).Field`; it follows the pointer for you.

```go
d := Deployment{Name: "api", Replicas: 2}
p := &d
p.Replicas = 5          // same as (*p).Replicas = 5
fmt.Println(d.Replicas) // 5
```

### Pointing into a slice

`&ds[i]` is the address of an element inside the slice. Changing something
through that pointer changes the slice. The variable in `for _, d := range ds`
is a **copy** of each element, so a pointer to it points at the copy:

```go
for _, d := range ds {
	if d.Name == "api" {
		return &d // wrong: points at a copy, not into ds
	}
}

for i := range ds {
	if ds[i].Name == "api" {
		return &ds[i] // right: points at the element in ds
	}
}
```

## Your task

Your deploy tool tracks a restart counter and a list of deployments. Write the
helpers that change them in place.

The exercise uses this struct:

```go
type Deployment struct {
	Name     string
	Replicas int
}
```

### 1. Reset a counter

Change `ResetCounter` so it sets the `int` that `count` points to back to `0`.
You can assume `count` is never `nil`.

```go
restarts := 7
ResetCounter(&restarts)
// restarts is now 0
```

### 2. Scale a deployment

Change `Scale` so it sets the deployment's `Replicas` field to `replicas`.

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

### 3. Find a deployment

Change `FindDeployment` so it returns a pointer to the element of `ds` whose
`Name` equals `name`, or `nil` if there is no such element. The pointer must
point **into** `ds`, so that a change made through it shows up in the slice.
The tests check this.

```go
ds := []Deployment{{"api", 2}, {"web", 3}}
p := FindDeployment(ds, "web")
p.Replicas = 10
// ds[1].Replicas is now 10

FindDeployment(ds, "db")
// => nil
```

## Learn more

- [A Tour of Go: Pointers](https://go.dev/tour/moretypes/1)
- [A Tour of Go: Pointers to structs](https://go.dev/tour/moretypes/4)
- [Go by Example: Pointers](https://gobyexample.com/pointers)
- [Go spec: Address operators](https://go.dev/ref/spec#Address_operators)
