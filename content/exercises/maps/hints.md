## Task 1: ask for two values

`ports[service]` alone can't tell "port 0" from "not there". Ask for the
second value too, then return both:

```go
port, ok := ports[service]
```

## Task 2: check, then delete

`delete` doesn't tell you whether the key existed, so look first. Use comma ok
and throw the port away with `_`:

```go
_, ok := ports[service]
```

Then call `delete(ports, service)` and return `ok`.

## Task 3: count with ++

Loop over `regions` with `range`. For each region `r`, add one to its count
with `counts[r]++`. A region the map hasn't seen yet starts at `0`.
