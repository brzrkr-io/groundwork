## Task 1: Look up a port

`ports[service]` alone gives `0` both for a missing service and for one on port 0. Ask the map for a second value that says whether the key exists.

### Closer

The comma ok form reads two values at once: `port, ok := ports[service]`. Then return both.

### Answer

```go
port, ok := ports[service]
return port, ok
```

`ok` is `true` only when the key is really in the map, even if its value is `0`.

## Task 2: Remove a service

`delete` doesn't tell you whether the key was there, so check first, then delete. The map is shared with the caller, so deleting inside the function changes their map.

### Closer

Use comma ok and throw the port away with `_`: `_, ok := ports[service]`. Then call `delete(ports, service)` and return `ok`.

### Answer

```go
_, ok := ports[service]
delete(ports, service)
return ok
```

Deleting a missing key does nothing, so it's safe to call `delete` either way.

## Task 3: Count hosts per region

Loop over `regions` and add one to the count for each region. A region the map hasn't seen yet reads as `0`, so there's no special case.

### Closer

Use `for _, r := range regions` and inside it `counts[r]++`.

### Answer

```go
for _, r := range regions {
	counts[r]++
}
return counts
```

A missing key reads as `0`, so `++` turns a new region's count into `1`.
