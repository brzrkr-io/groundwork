# Practice: service registry

Before one service can call another, it has to find it. A *service registry*
answers that question: "where is `billing-api` running, and is it healthy?"
Consul, etcd and Kubernetes Services all do this job. In this exercise you build
a small one, using everything from *Types and methods*.

## Learn

This is a practice exercise, so there is nothing new. A quick recap of the
pieces you'll use:

| Piece | Where it's from | Here |
|---|---|---|
| Struct | *Structs* | `Service` holds a name, an address, tags and a status. |
| Pointer receiver | *Pointer receivers* | Every `Registry` method uses `r *Registry`, so all of them share one map. |
| Custom type with `iota` | *Custom types and enums* | `Status` is `StatusUnknown`, `StatusHealthy` or `StatusUnhealthy`. |
| `String()` method | *Stringers* | Makes `Status` print as a word. |
| Interface | *Interfaces* | `HealthChecker` is anything with `Check(address string) bool`. |

Three reminders that will save you time:

- **Map values are copies.** You can't write `r.services[name].Status = x`;
  Go refuses to compile it. Read the value out, change it, then store it
  back:

  ```go
  s := r.services[name]
  s.Status = StatusHealthy
  r.services[name] = s
  ```

- **Map order is random.** When you range over a map, the order changes from
  run to run. Sort anything you return, as in *Sorting*:

  ```go
  slices.SortFunc(found, func(a, b Service) int {
  	return cmp.Compare(a.Name, b.Name)
  })
  ```

- **`slices.Contains(list, x)`** reports whether `x` is in the slice `list`.
  It's handy for checking tags.

Some tasks need imports that the starter file doesn't have yet: `errors` and
`fmt` for task 2, and `cmp` and `slices` for task 4. Add them to an `import`
block at the top of `exercise.go`.

## Your task

Your platform team wants a registry so services stop hard-coding each other's
IP addresses. The types and method signatures are ready in `exercise.go`. Fill
in the method bodies. The tests for tasks 3 to 5 fill the registry directly,
so you can do the tasks in any order. Going from 1 to 5 is easiest, though.

### 1. Name the statuses

Make `Status.String` return the status as a word:

| Value | String |
|---|---|
| `StatusUnknown` | `"unknown"` |
| `StatusHealthy` | `"healthy"` |
| `StatusUnhealthy` | `"unhealthy"` |
| anything else | `"Status(N)"`, where N is the number |

```go
fmt.Println(StatusHealthy)
// healthy
fmt.Println(Status(7))
// Status(7)
```

Build `"Status(7)"` with `%d` and `int(s)`. Don't use `%v` or `%s` on `s`
inside `String`, because that calls `String` again forever.

### 2. Register and look up services

`Register(s Service) error`:

- If `s.Name` is `""`, return the error `service name is empty`.
- If a service with that name is already registered, return the error
  `service "web" is already registered` (using the real name, in double
  quotes). Keep the first service; don't overwrite it.
- Otherwise, set `s.Status` to `StatusUnknown`, because nobody has checked the
  new service yet. Store it in `r.services` under its name and return `nil`.

`Lookup(name string) (Service, bool)` returns the stored service and `true`,
or the zero `Service` and `false` when the name isn't registered.

```go
r := NewRegistry()
r.Register(Service{Name: "web", Address: "10.0.0.9:80"})
// => nil
r.Register(Service{Name: "web", Address: "10.0.0.10:80"})
// => error: service "web" is already registered
r.Lookup("web")
// => Service{Name: "web", Address: "10.0.0.9:80", Status: StatusUnknown}, true
r.Lookup("ghost")
// => Service{}, false
```

### 3. Deregister a service

`Deregister(name string) bool` removes the service and returns `true`. If the
service isn't registered, it changes nothing and returns `false`. Use the
built-in `delete(map, key)` to remove an entry.

```go
r.Deregister("web")   // => true
r.Deregister("web")   // => false (already gone)
```

### 4. Find services by tag

`ByTag(tag string) []Service` returns every service whose `Tags` contain
`tag`, sorted by `Name` from A to Z. If nothing matches, return an empty
slice (`nil` is fine).

```go
// web: [prod eu-west-1], billing-api: [prod us-east-1], auth: [prod]
r.ByTag("prod")
// => services auth, billing-api, web (in that order)
r.ByTag("eu-west-1")
// => service web
```

### 5. Health-check everything

`CheckAll(hc HealthChecker) int` calls `hc.Check(s.Address)` for every
registered service:

- `true` sets that service's `Status` to `StatusHealthy`.
- `false` sets it to `StatusUnhealthy`.

Store the updated service back in the map, and return how many services are
now healthy.

The test passes in a fake checker that answers from a map of address to
`true`/`false`, so no network is involved. That's a big reason to depend on
an interface: in production you pass a checker that makes real HTTP calls, and
in tests you pass a fake.

```go
r.CheckAll(checker) // web and cache are up, billing-api is down
// => 2
// web: healthy, cache: healthy, billing-api: unhealthy
```

## Learn more

- [Go by Example: Maps](https://gobyexample.com/maps)
- [Effective Go: Interfaces](https://go.dev/doc/effective_go#interfaces)
- [Package slices](https://pkg.go.dev/slices)
- [Kubernetes: Service](https://kubernetes.io/docs/concepts/services-networking/service/)
