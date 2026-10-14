# Struct embedding: a shared health-check base

A monitoring agent runs many kinds of health checks: HTTP, TCP, DNS. They all
share a name and a schedule. Here you put the shared part in one embedded
struct and let promotion make every check type satisfy one interface.

**New here:** promoted methods satisfying an interface. (`%`, the remainder, is from *Numbers and arithmetic*.)

## Learn

### Recap

As in *Struct embedding*: embedding `BaseCheck` in `HTTPCheck` promotes
`BaseCheck`'s fields and methods, so `h.Name` works on an `HTTPCheck`. In a
literal you write the embedded struct out: `BaseCheck: BaseCheck{...}`.

### Promoted methods satisfy interfaces

As in *Interfaces*, a type satisfies an interface when it has all of its
methods. Promoted methods **count**:

```go
type Notifier interface {
	Channel() string
	Send(msg string) string
}

type Base struct{ Name string }

func (b Base) Channel() string { return b.Name } // shared

type Slack struct{ Base }

func (s Slack) Send(msg string) string { return "slack: " + msg } // its own

var n Notifier = Slack{Base{Name: "#ops"}} // OK: Channel is promoted
```

Write the shared methods once on the base; each type only adds what makes it
different.

> **Why is it written like this?**
> - **Go requires** every interface method to exist, but doesn't care whether it was declared on the type or promoted.
> - **Author's choice:** which methods live on the base and which on each type.

### The remainder operator

`a % b` is the remainder after dividing `a` by `b`. When it is `0`, `a` is a
multiple of `b`:

```go
60 % 30 // 0: 60 is a multiple of 30
45 % 30 // 15: 45 is not
0 % 30  // 0: zero is a multiple of everything
```

## Your task

All checks implement this interface:

```go
type Check interface {
	CheckName() string
	Interval() int
	Target() string
}
```

`HTTPCheck` and `TCPCheck` both embed `BaseCheck`.

### 1. Write the shared methods

**Goal:** finish `CheckName` and `Interval` on `BaseCheck`.

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

**Think it through:**
- The receiver is `b`. Which of its fields does each method need?
- Use the constant's name `DefaultInterval`, not the number 60.

### 2. Write each check's Target

**Goal:** finish `Target` on `HTTPCheck` (return the URL) and on `TCPCheck` (return host, colon, port).

```go
HTTPCheck{URL: "https://api.internal/healthz"}.Target()
// => "https://api.internal/healthz"

TCPCheck{Host: "db.internal", Port: 5432}.Target()
// => "db.internal:5432"
```

**Think it through:**
- `Port` is an `int`, not a string, so `+` won't join it. Which `Sprintf` verb prints an `int`?
- The `:` is fixed text, with no spaces around it.
- If you use `fmt.Sprintf`, add `import "fmt"` below the `package` line.

### 3. Find the checks that are due

**Goal:** finish `DueNow(checks, elapsed)` so it returns `"<name> -> <target>"` for each check that is due.

- The agent calls it every second; `elapsed` is the seconds since it started.
- A check is due when `elapsed` is a multiple of its `Interval()`.
- Keep the order of `checks`.
- If nothing is due, return an empty slice (`nil` is fine too).

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

The tests for this task use a small fake check type from the test file, so
they work even before tasks 1 and 2 pass.

**Think it through:**
- Loop over `checks`. Each item is a `Check`, so you can only call its three methods.
- Which operator tells you "is a multiple of"?
- The text between name and target is fixed: space, `->`, space.

## Learn more

- [Effective Go: Embedding](https://go.dev/doc/effective_go#embedding)
- [Go spec: Method sets](https://go.dev/ref/spec#Method_sets)
- [Kubernetes: Configure liveness, readiness and startup probes](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/)
