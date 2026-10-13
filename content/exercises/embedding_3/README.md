# Struct embedding: a shared health-check base

A monitoring agent runs many kinds of health checks: HTTP, TCP, DNS. They all
share a name and a schedule. Here you put the shared part in one embedded
struct and let promotion make every check type satisfy one interface.

## Learn

### Recap

As in *Struct embedding*, embedding `BaseCheck` in
`HTTPCheck` promotes `BaseCheck`'s fields and methods, so `h.Name` and
`h.CheckName()` work on an
`HTTPCheck`. In a literal you write the embedded struct out:
`HTTPCheck{BaseCheck: BaseCheck{Name: "api-health"}, URL: "..."}`.

### Promoted methods satisfy interfaces

As in *Interfaces*, a type satisfies an interface when it has all of the
interface's methods. Promoted methods count:

```go
type Check interface {
	CheckName() string
	Target() string
}

func (b BaseCheck) CheckName() string { ... } // shared by every check
func (h HTTPCheck) Target() string { ... }    // HTTPCheck's own method

var c Check = HTTPCheck{...} // OK: CheckName is promoted from BaseCheck
```

`HTTPCheck` gets `CheckName` from `BaseCheck` and adds its own `Target`, so it
has everything `Check` asks for. Write the shared methods once on `BaseCheck`;
each check type only writes what makes it different.

### The remainder operator

`a % b` is the remainder after dividing `a` by `b`. When it is `0`, `a` is a
multiple of `b`:

```go
60 % 30 // 0: 60 is a multiple of 30
45 % 30 // 15: 45 is not
0 % 30  // 0: zero is a multiple of everything
```

## Your task

The agent's checks all implement this interface:

```go
type Check interface {
	CheckName() string
	Interval() int
	Target() string
}
```

`HTTPCheck` and `TCPCheck` embed `BaseCheck`.

### 1. Write the shared methods

Finish the two methods on `BaseCheck`:

- `CheckName` returns `Name`.
- `Interval` returns `IntervalSeconds`, or `DefaultInterval` (60) when
  `IntervalSeconds` is 0.

The tests call them on `HTTPCheck` and `TCPCheck` values, through promotion:

```go
var c Check = HTTPCheck{BaseCheck: BaseCheck{Name: "api-health", IntervalSeconds: 30}}
c.CheckName() // => "api-health"
c.Interval()  // => 30

c = TCPCheck{BaseCheck: BaseCheck{Name: "postgres"}}
c.Interval()  // => 60
```

### 2. Write each check's Target

Finish `Target` on both check types:

- `HTTPCheck.Target` returns the `URL`.
- `TCPCheck.Target` returns the host, a colon and the port.

```go
HTTPCheck{URL: "https://api.internal/healthz"}.Target()
// => "https://api.internal/healthz"

TCPCheck{Host: "db.internal", Port: 5432}.Target()
// => "db.internal:5432"
```

If you use `fmt.Sprintf` with `%d` for the port, add `import "fmt"` below the
`package` line.

### 3. Find the checks that are due

Finish `DueNow(checks []Check, elapsed int) []string`. The agent calls it
every second with the number of seconds since it started. A check is due when
`elapsed` is a multiple of its `Interval()`. For each due check, in the order
they appear in `checks`, add the line `<name> -> <target>`. If no check is
due, return an empty slice.

The tests for this task use a small fake check type from the test file, so
they work even before tasks 1 and 2 pass. Any type with the three methods is a
`Check`.

```go
checks := []Check{
	HTTPCheck{BaseCheck: BaseCheck{Name: "api-health", IntervalSeconds: 30}, URL: "https://api.internal/healthz"},
	TCPCheck{BaseCheck: BaseCheck{Name: "postgres", IntervalSeconds: 10}, Host: "db.internal", Port: 5432},
}
DueNow(checks, 10)
// => []string{"postgres -> db.internal:5432"}
DueNow(checks, 30)
// => []string{"api-health -> https://api.internal/healthz", "postgres -> db.internal:5432"}
DueNow(checks, 45)
// => []string{}
```

## Learn more

- [Effective Go: Embedding](https://go.dev/doc/effective_go#embedding)
- [Go spec: Method sets](https://go.dev/ref/spec#Method_sets)
- [Kubernetes: Configure liveness, readiness and startup probes](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/)
