# Interfaces: fix the health checks

A status page runs many kinds of health checks: disks, HTTP endpoints, TCP
ports. Two checks are missing from the page because their methods don't quite
match the interface. The function that lists failures is also backwards.

**New here:** exact method matching, the `var _ Interface = Type{}` check.

## Learn

### Recap: "has the method" means an exact match

As in *Interfaces*, a type satisfies an interface when it has every method
the interface lists. Nothing else is needed; Go never asks you to write
"implements".

The match must be exact: same **name**, same **parameters**, same **results**.

| Interface wants | Type has | Satisfies? |
|---|---|---|
| `Ready() bool` | `Ready() bool` | yes |
| `Ready() bool` | `IsReady() bool` | no: wrong name |
| `Ready() bool` | `Ready() string` | no: wrong result type |
| `Ready() bool` | `Ready(x int) bool` | no: extra parameter |

> **Why is it written like this?** Go requires the exact match; the compiler
> doesn't guess what you meant. Which methods an interface lists is the
> author's choice.

### Make the compiler check for you

Because satisfying is automatic, a typo in a method name gives no error where
you wrote the method. You only find out later, where the type is used as the
interface.

You can ask for that check right away with one line. Here with a made-up
`Pinger` interface and `Router` type:

```go
var _ Pinger = Router{}
```

Read it piece by piece:

- `var _` declares a variable named `_`, the *blank identifier* (as in
  `for _, v := range`). It means "throw this away".
- `Pinger` is the variable's type: the interface.
- `= Router{}` stores an empty `Router` in it.

The line does nothing at run time. It exists only so the compiler checks
"is a `Router` a `Pinger`?". If not, the build fails and says why:

```text
Router does not implement Pinger (missing method Ping)
```

```text
Router does not implement Pinger (wrong type for method Ping)
		have Ping() string
		want Ping() bool
```

> **Why is it written like this?** The `var _ I = T{}` line is a common Go
> convention, not a requirement. It goes at the top level of the file, outside
> any function.

**Try it:** add such a line for `HTTPCheck` and one for `TCPCheck` near the
top of `exercise.go`. The build breaks until both are fixed, but each error
tells you what to change. (The tests don't need these lines; they ask
`any(...).(HealthChecker)` as in *Interfaces*, so they compile either way.)

### Fixing an error: look at the declaration

When the error says `wrong type for method Check`, the line to change is the
method's own first line, `func (...) Check() ...`. That is where the result
type is declared. Then make the `return` lines inside give back that type.

## Your task

The status page treats every check as a `HealthChecker`:

```go
type HealthChecker interface {
	Name() string
	Check() bool
}
```

`DiskCheck` already satisfies it. Use it as your model.

### 1. Fix HTTPCheck

**Goal:** make `HTTPCheck` satisfy `HealthChecker` by fixing the method that
doesn't match.

- `Name()` returns `"http "` followed by the `URL` field. It already works.
- `Check()` returns `true` only when `StatusCode` is `200`.

```go
HTTPCheck{URL: "https://shop.example.com/healthz", StatusCode: 503}
// Name()  => "http https://shop.example.com/healthz"
// Check() => false
```

**Think it through:**
- Put the interface's two method lines next to `HTTPCheck`'s two methods.
  Which part differs: name, parameters or result?
- Does the logic inside the method need to change at all?

### 2. Fix TCPCheck

**Goal:** make `TCPCheck` satisfy `HealthChecker`. Its problem is different
from task 1.

- `Name()` returns `"tcp <Host>:<Port>"`. It already works.
- `Check()` returns `true` when the `Open` field is `true`, and `false`
  otherwise.

```go
TCPCheck{Host: "db-1", Port: 5432, Open: true}
// Name()  => "tcp db-1:5432"
// Check() => true
```

**Think it through:**
- The name is right this time. Compare what comes after `Check()` with the
  interface.
- After fixing that, the `return "up"` lines won't compile. What type is the
  `Open` field?

### 3. Fix FailingChecks

**Goal:** `FailingChecks(checks)` returns the `Name()` of every check whose
`Check()` is `false`. Right now it returns the healthy ones.

- Keep the order the checks appear in.
- When every check is healthy, return an empty slice (`nil` is fine).
- A `DiskCheck` is healthy when `FreePercent` is at least `10`.

```go
FailingChecks([]HealthChecker{
	DiskCheck{Path: "/", FreePercent: 42},
	DiskCheck{Path: "/var", FreePercent: 3},
	DiskCheck{Path: "/tmp", FreePercent: 9},
})
// => []string{"disk /var", "disk /tmp"}
```

**Think it through:**
- Find the `if` inside the loop. Does it pick checks that pass or that fail?
- Which operator turns `true` into `false` and back?

Once tasks 1 and 2 are done, `HTTPCheck` and `TCPCheck` values can go in the
same slice too.

## Learn more

- [A Tour of Go: Interfaces are implemented implicitly](https://go.dev/tour/methods/10)
- [Effective Go: Interface checks](https://go.dev/doc/effective_go#blank_implements)
- [Go spec: Implementing an interface](https://go.dev/ref/spec#Implementing_an_interface)
