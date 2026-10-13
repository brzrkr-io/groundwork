# Methods: host inventory

A fleet inventory is a list of hosts that every ops tool keeps asking
questions: how many are healthy, which ones run in `eu-west-1`, where is
`web-2`? Here you give an `Inventory` type methods that answer those
questions.

**New here:** methods on a struct that holds a slice, `strings.EqualFold`.

## Learn

### Recap: the parts of a method line

```go
func (inv Inventory) Count() int {
```

| Part | What it is | Can you change it? |
|---|---|---|
| `func` | keyword | No, Go requires it |
| `(inv Inventory)` | the *receiver*: a name and the type the method belongs to | The name `inv` is your choice; the type is fixed |
| `Count` | the method name | Already chosen by the starter; tests call it |
| `()` | parameters (none here) | Fixed by the starter |
| `int` | the result type | Fixed by the starter |

You call it as `inv.Count()`. Inside the method, `inv` is the value it was
called on.

> **Why is it written like this?** Go requires the receiver in brackets
> before the name. Calling it a short name like `inv` (not `this` or `self`)
> is Go convention.

### Methods on a struct that holds a slice

```go
type Shelf struct {
	Books []string
}

func (s Shelf) Size() int {
	return len(s.Books)
}
```

`s.Books` is the slice inside the struct. You can `range` over it, index it
and take its `len`, just like any other slice.

The receiver `s` is a copy of the `Shelf`. Copying a slice copies only its
small header, not the items, so this is cheap. Your methods here only read,
so the copy never matters.

### Looping and counting inside a method

A loop over a field looks the same as a loop over any slice. Here each item
is a struct, so you reach its fields with a dot:

```go
type Book struct {
	Title string
	Read  bool
}

for _, b := range shelf.Books {
	if b.Read {
		fmt.Println(b.Title)
	}
}
```

> **Why is it written like this?** `_` throws away the index, because Go
> rejects unused variables. The name `b` is your choice.

### Comparing names without caring about case

Host names are case-insensitive, like DNS names: `WEB-1` and `web-1` are the
same host. `strings.EqualFold(a, b)` reports whether two strings are equal
when you ignore upper and lower case:

```go
strings.EqualFold("Go", "GO") // true
strings.EqualFold("go", "gopher") // false
```

> **Why is it written like this?** `strings.` is the package name; Go
> requires it in front of everything from another package, and requires
> `"strings"` in the file's `import` list. Forgetting the import is a compile
> error.

## Your task

`Inventory` holds the fleet's hosts. Each `Host` has a `Name`, a `Region`, and
a `Healthy` flag from the last health check. Complete the three methods. Every
method must also work on an empty inventory (no hosts).

The examples below use this inventory:

```go
inv := Inventory{Hosts: []Host{
	{Name: "web-1", Region: "eu-west-1", Healthy: true},
	{Name: "web-2", Region: "eu-west-1", Healthy: false},
	{Name: "db-1", Region: "us-east-1", Healthy: true},
}}
```

### 1. Count healthy hosts

**Goal:** `HealthyCount` returns how many hosts have `Healthy` set to `true`.

```go
inv.HealthyCount()
// => 2
```

An empty inventory returns `0`.

**Think it through:**
- Where do the hosts live? (Which field of the receiver?)
- You need a number that starts at `0` and grows. When should it grow?

### 2. Filter by region

**Goal:** `InRegion(region)` returns the hosts whose `Region` equals the
`region` parameter.

- Compare exactly (case matters here).
- Keep the same order as in the inventory.
- If no host matches, return an empty slice (`nil` is fine).

```go
inv.InRegion("eu-west-1")
// => []Host{
//      {Name: "web-1", Region: "eu-west-1", Healthy: true},
//      {Name: "web-2", Region: "eu-west-1", Healthy: false},
//    }

inv.InRegion("ap-south-1")
// => []Host{} (no hosts)
```

**Think it through:**
- Two different things are called "region": the parameter `region` and each
  host's `Region` field. Which do you compare with which?
- How did you add items to a slice in *Slices*?

### 3. Find a host by name

**Goal:** `Find(name)` returns the first host whose `Name` matches `name`,
ignoring case, plus `true`.

- If no host matches, return the zero `Host{}` and `false`.
- Add `"strings"` to the imports so you can use `strings.EqualFold`.

```go
inv.Find("WEB-2")
// => Host{Name: "web-2", Region: "eu-west-1", Healthy: false}, true

inv.Find("cache-1")
// => Host{}, false
```

**Think it through:**
- "First" match: once you find one, can you return straight away?
- The "not found" answer belongs *after* the loop. Why?

## Learn more

- [A Tour of Go: Methods](https://go.dev/tour/methods/1)
- [Go by Example: Methods](https://gobyexample.com/methods)
- [strings.EqualFold](https://pkg.go.dev/strings#EqualFold)
- [AWS: Regions and Availability Zones](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-regions-availability-zones.html)
