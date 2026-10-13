# Interfaces: fix the health checks

A status page runs many kinds of health checks: disks, HTTP endpoints, TCP
ports. Two of them have been left off the page because their methods don't
quite match the interface, and the function that lists failures is backwards.

## Learn

### Recap

As in *Interfaces*, an interface lists method signatures, and a type satisfies
it automatically when it has all of those methods. "Has the method" means an
exact match: same name, same parameters, same results. `Checked() bool` does
not count as `Check() bool`, and neither does `Check() string`.

### Asking the compiler to check for you

Because satisfaction is automatic, a typo in a method name doesn't produce an
error where the method is written. You only find out when you try to use the
type as the interface. You can ask for that check up front with one line:

```go
var _ HealthChecker = HTTPCheck{}
```

This declares a variable of type `HealthChecker` and stores an `HTTPCheck`
in it. The name `_` (the *blank identifier*, as in `for _, v := range`) means
the variable is thrown away; the line exists only so the compiler checks it.
If `HTTPCheck` doesn't satisfy `HealthChecker`, the build fails with a message
that says why:

```text
HTTPCheck does not implement HealthChecker (missing method Check)
```

```text
TCPCheck does not implement HealthChecker (wrong type for method Check)
		have Check() string
		want Check() bool
```

Try it: add both lines to `exercise.go` before you start. The build breaks
until both types are fixed, but the error tells you exactly what to change.
The tests use the `any(...).(HealthChecker)` question from *Interfaces*
instead, so they still compile while the types are broken.

## Your task

The status page treats every check as a `HealthChecker`:

```go
type HealthChecker interface {
	Name() string
	Check() bool
}
```

`DiskCheck` already satisfies it; use it as a model.

### 1. Fix HTTPCheck

`HTTPCheck` does not satisfy `HealthChecker`. Find the method that doesn't
match and fix it. `Check` reports `true` only when `StatusCode` is `200`, and
`Name` returns `http <URL>`:

```go
HTTPCheck{URL: "https://shop.example.com/healthz", StatusCode: 503}
// Name()  => "http https://shop.example.com/healthz"
// Check() => false
```

### 2. Fix TCPCheck

`TCPCheck` doesn't satisfy `HealthChecker` either, for a different reason.
Fix it so `Check` reports `true` when `Open` is `true`, and `false` otherwise.
`Name` returns `tcp <Host>:<Port>`:

```go
TCPCheck{Host: "db-1", Port: 5432, Open: true}
// Name()  => "tcp db-1:5432"
// Check() => true
```

### 3. Fix FailingChecks

`FailingChecks(checks []HealthChecker) []string` should return the `Name()`
of every check whose `Check()` returns `false`, in the order they appear.
Right now it returns the healthy ones instead. When every check is healthy it
returns an empty slice (`nil` is fine).

A `DiskCheck` is healthy when `FreePercent` is at least `10`:

```go
FailingChecks([]HealthChecker{
	DiskCheck{Path: "/", FreePercent: 42},
	DiskCheck{Path: "/var", FreePercent: 3},
	DiskCheck{Path: "/tmp", FreePercent: 9},
})
// => []string{"disk /var", "disk /tmp"}
```

Once tasks 1 and 2 are done, an `HTTPCheck` and a `TCPCheck` can go in the
same slice too.

## Learn more

- [A Tour of Go: Interfaces are implemented implicitly](https://go.dev/tour/methods/10)
- [Effective Go: Interface checks](https://go.dev/doc/effective_go#blank_implements)
- [Go spec: Implementing an interface](https://go.dev/ref/spec#Implementing_an_interface)
