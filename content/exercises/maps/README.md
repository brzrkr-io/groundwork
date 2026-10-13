# Maps

A *map* looks things up by a key: the port for a service, the owner of a
host, the number of errors per region. Service registries, config files and
label sets are all maps, so you'll use them constantly.

**New here:** maps, `make`, `delete`, comma ok on maps, `range` over a map.

## Learn

### Creating a map

The type `map[string]int` means "keys are strings, values are ints". You can
write a map out with a *map literal*:

```go
replicas := map[string]int{
	"web":    3,
	"worker": 5,
}
```

To start with an empty map, use `make`:

```go
replicas := make(map[string]int)
```

> **Why is it written like this?** `map[KeyType]ValueType` is required
> syntax. The comma after the last entry is required when the closing `}` is
> on its own line. Lining up the values is gofmt's convention; the names are
> your choice.

### Setting, reading, deleting

```go
replicas["cron"] = 1     // add a key, or replace its value
r := replicas["web"]     // r is 3
delete(replicas, "cron") // remove a key
n := len(replicas)       // number of keys: 2
```

Deleting a key that isn't there is fine. It does nothing, and it doesn't
tell you whether the key was there.

### Missing keys and comma ok

Reading a missing key does **not** fail. It gives you the zero value of the
value type, as in *Zero values*:

```go
r := replicas["batch"] // r is 0: there is no "batch" key
```

So is `0` a missing key, or a key whose value really is `0`? Ask for a second
value. It's `true` only when the key exists. This is the *comma ok* pattern
from *Multiple return values*:

```go
owner, found := owners["db-7"]
if !found {
	// nobody owns db-7
}
```

If you only need the `true`/`false`, throw the value away with `_`.

> **Why is it written like this?** The two-value form is built into Go. The
> name `ok` for the second value is convention; `found` works just as well.

### Counting

Because missing keys read as `0`, you can add to a key that isn't there yet:

```go
errors := make(map[string]int)
errors["timeout"] += 2 // 0 + 2: errors["timeout"] is now 2
errors["timeout"]++    // now 3
```

### Looping over a map

`range` works on maps too, giving each key and value:

```go
for name, count := range replicas {
	fmt.Println(name, count)
}
```

The order is **random**, and changes from run to run. Never rely on it.
(*Sorting* shows how to visit keys in order.)

### Maps are shared, nil maps are read-only

When you pass a map to a function, the function works on the *same* map, not
a copy. If the function deletes a key, the caller's map loses it too. (Arrays
in *Arrays* are copied instead.)

A map variable declared without a value is `nil`:

```go
var m map[string]int // m is nil
fmt.Println(m["x"])  // fine: 0
m["x"] = 1           // panic: assignment to entry in nil map
```

Reading a nil map is fine. Writing to one crashes, so create maps with `make`
or a literal before you put anything in them.

## Your task

Your team keeps a small service registry: a map from service name to port.
Write the functions that work with it.

### 1. Look up a port

**Goal:** make `PortFor` return the port for `service` and whether `service` is in `ports`.

```go
ports := map[string]int{"api": 8080, "redis": 6379, "sidecar": 0}

PortFor(ports, "api")     // => 8080, true
PortFor(ports, "sidecar") // => 0, true
PortFor(ports, "nginx")   // => 0, false
```

- Registered service: its port and `true`.
- Missing service: `0, false`.
- A service really registered on port `0` gives `0, true`.

**Think it through:**
- Why isn't the port alone enough to tell `"sidecar"` and `"nginx"` apart?
- Which form of map lookup gives you two values?

### 2. Remove a service

**Goal:** make `Deregister` delete `service` from `ports` and return whether it was registered.

```go
ports := map[string]int{"api": 8080, "redis": 6379}

Deregister(ports, "redis") // => true,  ports is now {"api": 8080}
Deregister(ports, "redis") // => false, ports is unchanged
```

- Return `true` if the service was there (even on port `0`), `false` if not.
- The caller's map must change. Change `ports` itself; don't make a new map.

**Think it through:**
- `delete` doesn't report anything. When must you check whether the key
  exists: before or after deleting?
- Do you need the port value, or only the `true`/`false`?

### 3. Count hosts per region

**Goal:** make `CountByRegion` return a map from each region to how many times it appears in `regions`.

```go
CountByRegion([]string{"eu-west-1", "us-east-1", "eu-west-1"})
// => map[string]int{"eu-west-1": 2, "us-east-1": 1}

CountByRegion(nil)
// => map[string]int{}
```

`regions` holds one entry per host. The starter already creates the empty
`counts` map and returns it; with no hosts it stays empty.

The tests compare maps with `maps.Equal`, which checks that two maps hold the
same keys and values, in any order.

**Think it through:**
- What do you loop over, and which part of each entry is the map key?
- What does a region you haven't seen yet read as?

## Learn more

- [A Tour of Go: Maps](https://go.dev/tour/moretypes/19)
- [Go by Example: Maps](https://gobyexample.com/maps)
- [Go blog: Go maps in action](https://go.dev/blog/maps)
