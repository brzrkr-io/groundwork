# Maps

A *map* looks things up by a key: the port for a service, the owner of a
host, the number of errors per region. Service registries, config files and
label sets are all maps, so you'll use them constantly.

## Learn

### Creating a map

The type `map[string]int` means "keys are strings, values are ints". You can
write a map out with a *map literal*:

```go
ports := map[string]int{
	"api":   8080,
	"redis": 6379,
}
```

Note the comma after the last entry. Go requires it when the closing `}` is
on its own line.

To start with an empty map, use `make`:

```go
ports := make(map[string]int)
```

### Setting, reading, deleting

```go
ports["postgres"] = 5432 // add a key, or replace its value
p := ports["api"]        // p is 8080
delete(ports, "redis")   // remove a key
n := len(ports)          // number of keys: 2
```

Deleting a key that isn't there is fine. It does nothing.

### Missing keys and comma ok

Reading a key that isn't in the map does **not** fail. It gives you the zero
value of the value type, as in *Zero values*:

```go
p := ports["nginx"] // p is 0: there is no "nginx" key
```

So is `0` a missing service, or a service on port 0? To tell the difference,
ask for a second value. It's `true` when the key exists. This is the same
*comma ok* pattern you met in *Multiple return values*:

```go
p, ok := ports["nginx"]
if !ok {
	// no such service
}
```

### Counting

Because missing keys read as `0`, counting needs no special case:

```go
counts := make(map[string]int)
counts["error"]++ // 0 + 1: counts["error"] is now 1
counts["error"]++ // now 2
```

### Looping over a map

`range` works on maps too, giving each key and value:

```go
for service, port := range ports {
	fmt.Println(service, port)
}
```

The order is **random**, and changes from run to run. Never rely on it. (A
later exercise, *Sorting*, shows how to visit keys in order.)

### Maps are shared, nil maps are read-only

When you pass a map to a function, the function works on the *same* map, not
a copy. If the function deletes a key, the caller's map loses it too. (This is
different from arrays in *Arrays*, which are copied.)

A map variable declared without a value is `nil`:

```go
var m map[string]int // m is nil
fmt.Println(m["x"])  // fine: 0
fmt.Println(len(m))  // fine: 0
m["x"] = 1           // panic: assignment to entry in nil map
```

Reading a nil map is fine. Writing to one crashes your program, so create
maps with `make` or a literal before you put anything in them.

## Your task

Your team keeps a small service registry: a map from service name to port.
Write the functions that work with it.

### 1. Look up a port

Change `PortFor` so it returns the port for `service` and `true`. If the
service is not in the map, return `0, false`. A service may really be
registered on port `0`, and then the answer is `0, true`.

```go
ports := map[string]int{"api": 8080, "redis": 6379, "sidecar": 0}

PortFor(ports, "api")     // => 8080, true
PortFor(ports, "sidecar") // => 0, true
PortFor(ports, "nginx")   // => 0, false
```

### 2. Remove a service

Change `Deregister` so it deletes `service` from `ports`. Return `true` if the
service was registered (and is now gone), or `false` if it wasn't there to
begin with. The caller's map must change.

```go
ports := map[string]int{"api": 8080, "redis": 6379}

Deregister(ports, "redis") // => true,  ports is now {"api": 8080}
Deregister(ports, "redis") // => false, ports is unchanged
```

### 3. Count hosts per region

`CountByRegion` gets the region of every host in the fleet, one entry per
host. Return a map from each region to how many hosts are in it. With no
hosts, return an empty map.

```go
CountByRegion([]string{"eu-west-1", "us-east-1", "eu-west-1"})
// => map[string]int{"eu-west-1": 2, "us-east-1": 1}

CountByRegion(nil)
// => map[string]int{}
```

The tests compare maps with `maps.Equal`, which checks that two maps hold the
same keys and values, in any order.

## Learn more

- [A Tour of Go: Maps](https://go.dev/tour/moretypes/19)
- [Go by Example: Maps](https://gobyexample.com/maps)
- [Go blog: Go maps in action](https://go.dev/blog/maps)
