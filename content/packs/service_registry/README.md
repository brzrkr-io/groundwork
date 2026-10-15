# Practice: service registry

Before one service can call another, it has to find it. A *service registry*
answers "where is `billing-api` running, and is it healthy?" Consul, etcd and
Kubernetes Services all do this job. Here you build a small one.

**New here:** nothing; this is practice for *Types and methods* and the
interfaces from *Interfaces and generics*.

## Learn

### How the registry fits together

```
Registry ──holds──▶ map[name]Service
                         │
                         └─ Service{Name, Address, Tags, Status}
                                                          │
CheckAll(hc) ──asks──▶ HealthChecker.Check(address) ──▶ sets Status
```

- A `Registry` holds one map: service name → `Service`.
- Each method reads or changes that map.
- `CheckAll` doesn't know *how* to check health. It asks whatever
  `HealthChecker` you pass in. The tests pass a fake one, so no network is
  used.

### File map

| File | What's in it | Do you edit it? |
|---|---|---|
| `exercise.go` | `Status`, `Service`, `HealthChecker`, `Registry`, and the methods with `TODO(n)` markers | Yes: the method bodies, plus an `import` block you add under `package registry` (the starter has none yet) |
| `exercise_test.go` | The tests, including the fake health checker | No, read it if you're curious |

The types, `NewRegistry` and every method signature are already written.
You only fill in the bodies.

### Pieces you'll reuse

| Piece | From | Here |
|---|---|---|
| Struct | *Structs* | `Service` holds a name, an address, tags and a status |
| Pointer receiver | *Pointer receivers* | Every `Registry` method uses `r *Registry`, so all share one map |
| `iota` constants | *Custom types and enums* | `StatusUnknown`, `StatusHealthy`, `StatusUnhealthy` |
| `String()` method | *Stringers* | Makes `Status` print as a word |
| Interface | *Interfaces* | `HealthChecker` is anything with `Check(address string) bool` |

### Map values are copies

You can't change a field straight inside a map: `m[key].Field = x` does not
compile (Go rule). Read the value out, change it, store it back:

```go
p := people["ana"]
p.Age = 31
people["ana"] = p
```

The same is true for the `v` in `for k, v := range m`: it is a copy.

### Map order is random

Ranging over a map gives a different order each run. Sort anything you
return, as in *Sorting*. `slices.SortFunc` takes a compare function;
`cmp.Compare(a, b)` gives the right answer for it:

```go
slices.SortFunc(jobs, func(a, b Job) int {
	return cmp.Compare(a.Priority, b.Priority)
})
```

### Other helpers

- `slices.Contains(list, x)` reports whether `x` is in the slice `list`.
- `delete(m, key)` removes an entry from a map. It's built in, so it needs no import.
- `v, ok := m[key]` (*comma-ok*) tells you whether `key` was there.

> **Why is it written like this?** The `import` block with one package per
> line in `( ... )` is gofmt convention; Go rejects unused imports, so add
> each one only when you use it. You'll need `errors` and `fmt` (tasks 1–2)
> and `cmp` and `slices` (task 4). The receiver name `r` is the author's choice.

## Your task

Your platform team wants a registry so services stop hard-coding each other's
IP addresses. Fill in the method bodies in `exercise.go`. The tests for tasks
3–5 fill the registry directly, so any order works, but 1 to 5 is easiest.

### 1. Name the statuses

**Goal:** make `Status.String` return the status as a word.

| Value | String |
|---|---|
| `StatusUnknown` | `"unknown"` |
| `StatusHealthy` | `"healthy"` |
| `StatusUnhealthy` | `"unhealthy"` |
| anything else | `"Status(N)"`, where N is the number |

```go
fmt.Println(StatusHealthy) // healthy
fmt.Println(Status(7))     // Status(7)
```

**Think it through:**
- Which statement from *Switch: size, method, latency* picks one of several answers, with a fallback?
- For the fallback, use `%d` with `int(s)`. `%v` or `%s` on `s` would call `String` again, forever.

### 2. Register and look up services

**Goal:** `Register` stores a valid service; `Lookup` finds one by name.

`Register(s Service) error`, checked in this order:

1. `s.Name` is `""` → return the error `service name is empty`.
2. Name already registered → return the error
   `service "web" is already registered` (the real name, in double quotes).
   Keep the first service; don't overwrite it.
3. Otherwise set `s.Status` to `StatusUnknown` (nobody has checked it yet),
   store `s` in `r.services` under its name, and return `nil`.

`Lookup(name string) (Service, bool)` returns the stored service and `true`,
or the zero `Service` and `false` if the name isn't registered.

```go
r := NewRegistry()
r.Register(Service{Name: "web", Address: "10.0.0.9:80"})  // => nil
r.Register(Service{Name: "web", Address: "10.0.0.10:80"}) // => error: service "web" is already registered
r.Lookup("web")   // => Service{Name: "web", Address: "10.0.0.9:80", Status: StatusUnknown}, true
r.Lookup("ghost") // => Service{}, false
```

**Think it through:**
- Which error helper takes a fixed message, and which can put a name in quotes (`%q`)?
- How do you ask a map "is this key there?" without needing the value?

### 3. Deregister a service

**Goal:** `Deregister(name)` removes the service and returns `true`, or changes nothing and returns `false` if it isn't registered.

```go
r.Deregister("web") // => true
r.Deregister("web") // => false (already gone)
```

**Think it through:**
- Check whether the name is there *before* removing it, so you know what to return.

### 4. Find services by tag

**Goal:** `ByTag(tag)` returns every service whose `Tags` contain `tag`, sorted by `Name` from A to Z.

If nothing matches, return an empty slice (`nil` is fine).

```go
// web: [prod eu-west-1], billing-api: [prod us-east-1], auth: [prod]
r.ByTag("prod")      // => services auth, billing-api, web (in that order)
r.ByTag("eu-west-1") // => service web
```

**Think it through:**
- Collect the matches in a loop first, then sort once.
- Which field of `Service` do you compare when sorting?

### 5. Health-check everything

**Goal:** `CheckAll(hc)` sets every service's `Status` from `hc.Check(s.Address)` and returns how many are now healthy.

- `true` → `StatusHealthy`
- `false` → `StatusUnhealthy`

The change must be stored in the map, so `Lookup` sees it afterwards.

```go
r.CheckAll(checker) // web and cache are up, billing-api is down
// => 2
// web: healthy, cache: healthy, billing-api: unhealthy
```

**Think it through:**
- The loop gives you a copy of each service. What do you need to do after changing it?
- You need the name too, to store it back. Which two things can `range` give you?

## Learn more

- [Go by Example: Maps](https://gobyexample.com/maps)
- [Effective Go: Interfaces](https://go.dev/doc/effective_go#interfaces)
- [Package slices](https://pkg.go.dev/slices)
- [Kubernetes: Service](https://kubernetes.io/docs/concepts/services-networking/service/)
