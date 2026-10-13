# Methods: host inventory

A fleet inventory is a list of hosts that every ops tool keeps asking
questions: how many are healthy, which ones run in `eu-west-1`, where is
`web-2`? Here you give an `Inventory` type methods that answer those
questions.

## Learn

### Recap

A method has a receiver: `func (inv Inventory) Count() int` belongs to
`Inventory`, and you call it as `inv.Count()`. With a value receiver the
method gets a copy of the value, which is fine for methods that only read.

### Methods on a struct that holds a slice

```go
type Inventory struct {
	Hosts []Host
}

func (inv Inventory) Count() int {
	return len(inv.Hosts)
}
```

The receiver `inv` is a copy of the `Inventory`, but copying a slice copies
only its small header, not the hosts, so this is cheap. As you saw in *Slice
gotchas: shared memory*, the copy still points at the same hosts as the
original. Your methods here only read, so they never change anything for the
caller.

### Comparing names without caring about case

Host names are case-insensitive, like DNS names: `WEB-1` and `web-1` are the
same host. `strings.EqualFold(a, b)` reports whether two strings are equal
when you ignore upper and lower case:

```go
strings.EqualFold("WEB-1", "web-1") // true
strings.EqualFold("web-1", "web-2") // false
```

## Your task

`Inventory` holds the fleet's hosts. Each `Host` has a `Name`, a `Region`, and
a `Healthy` flag from the last health check. Complete the three methods. Every
method must work on an empty inventory too.

### 1. Count healthy hosts

Complete `HealthyCount` so it returns how many hosts have `Healthy` set to
`true`.

```go
inv := Inventory{Hosts: []Host{
	{Name: "web-1", Region: "eu-west-1", Healthy: true},
	{Name: "web-2", Region: "eu-west-1", Healthy: false},
	{Name: "db-1", Region: "us-east-1", Healthy: true},
}}

inv.HealthyCount()
// => 2
```

### 2. Filter by region

Complete `InRegion` so it returns the hosts whose `Region` equals `region`
exactly, in the same order as in the inventory. If no host matches, return an
empty slice (`nil` is fine).

```go
inv.InRegion("eu-west-1")
// => []Host{
//      {Name: "web-1", Region: "eu-west-1", Healthy: true},
//      {Name: "web-2", Region: "eu-west-1", Healthy: false},
//    }

inv.InRegion("ap-south-1")
// => []Host{} (no hosts)
```

### 3. Find a host by name

Complete `Find` so it returns the first host whose `Name` matches `name`,
ignoring upper and lower case, and `true`. If no host matches, return the zero
`Host{}` and `false`. You'll need `strings.EqualFold`, so add `"strings"` to
the imports.

```go
inv.Find("WEB-2")
// => Host{Name: "web-2", Region: "eu-west-1", Healthy: false}, true

inv.Find("cache-1")
// => Host{}, false
```

## Learn more

- [A Tour of Go: Methods](https://go.dev/tour/methods/1)
- [Go by Example: Methods](https://gobyexample.com/methods)
- [strings.EqualFold](https://pkg.go.dev/strings#EqualFold)
- [AWS: Regions and Availability Zones](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-regions-availability-zones.html)
