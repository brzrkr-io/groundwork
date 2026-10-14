# Booleans

You finished *Basics* with *Practice: deploy summary*: functions that compute
and format values. This module adds decisions, loops and reusable functions,
so your code can choose what to do and repeat work.

Ops work is full of yes-or-no questions: is the port privileged, may this
deploy go out, is it work hours? Go answers them with *booleans*.

**New here:** the `bool` type, comparison operators, `&&`, `||`, `!`.

## Learn

### The bool type

A `bool` holds one of two values: `true` or `false`. Nothing else.

```go
debug := true
var maintenance bool = false
```

A function can return a `bool` like any other type:

```go
// IsDebug reports whether debug mode is on.
func IsDebug() bool {
	return true
}
```

> **Why is it written like this?** `bool` after the `)` is the *return
> type*, and Go requires it. `true` and `false` are Go keywords, so they
> never go in quotes: `"true"` is a string, not a boolean. Starting the
> comment with the function name ("IsDebug reports whether…") is a Go
> convention.

### Comparisons make booleans

A *comparison* checks two values and gives back `true` or `false`.

| Operator | Meaning | Example | Result |
|---|---|---|---|
| `==` | equal | `3 == 3` | `true` |
| `!=` | not equal | `3 != 3` | `false` |
| `<` | less than | `2 < 5` | `true` |
| `>` | greater than | `2 > 5` | `false` |
| `<=` | less than or equal | `5 <= 5` | `true` |
| `>=` | greater than or equal | `4 >= 5` | `false` |

`==` and `!=` work on strings too: `env == "staging"` is `true` only when
`env` is exactly `"staging"`.

> **Why is it written like this?** `=` gives a variable a value; `==`
> compares two values. Go requires the double `==` for comparing, and the
> compiler rejects a single `=` where it expects a condition.

### Combining booleans

*Logical operators* combine booleans into a new boolean.

| Operator | Name | `true` when |
|---|---|---|
| `a && b` | and | both `a` and `b` are true |
| `a \|\| b` | or | at least one of `a`, `b` is true |
| `!a` | not | `a` is false |

```go
healthy := true
paused := false

healthy && !paused // true: healthy, and not paused
healthy || paused  // true: at least one is true
!healthy           // false
```

You can chain more than two: `a && b && c` is true only when all three are.
In `a && b`, if `a` is false Go never looks at `b`. This is called
*short-circuit evaluation*.

Use parentheses to group parts, like in maths:

```go
(port >= 8000 && port < 9000) || port == 80
```

> **Why is it written like this?** The symbols `&&`, `||` and `!` are what Go
> requires; the words `and`, `or`, `not` don't work. The spaces around `&&`
> are convention (`gofmt` adds them for you).

### Return the expression directly

A comparison already *is* a `bool`, so you can return it straight away:

```go
// IsEmpty reports whether the name is empty.
func IsEmpty(name string) bool {
	return name == ""
}
```

You will make decisions with `if` in *if and else*. You don't need it here:
every answer is one `return` line.

## Your task

You are writing small checks for a deploy tool. Each function returns a
hard-coded `false`. Replace that `false` with the right boolean expression.
Leave the `func` line as it is.

### 1. Spot privileged ports

**Goal:** make `IsPrivilegedPort` return `true` when `port` is below 1024.

On Linux, ports below 1024 are *privileged*: only root may listen on them.
1024 itself is not privileged.

```go
IsPrivilegedPort(22)   // => true
IsPrivilegedPort(80)   // => true
IsPrivilegedPort(1023) // => true
IsPrivilegedPort(1024) // => false
IsPrivilegedPort(8080) // => false
```

**Think it through:**
- The value to check is the parameter `port`; the limit is the number 1024.
- Which operator in the table means "less than", without "or equal"?

### 2. Decide whether a deploy may go out

**Goal:** make `CanDeploy` return `true` only when the tests passed, the
deploy is approved, and no change freeze is active.

All three must hold at once. The three parameters are `testsPassed`,
`approved` and `freezeActive`, all `bool`s.

```go
CanDeploy(true, true, false)   // => true
CanDeploy(false, true, false)  // => false (tests failed)
CanDeploy(true, false, false)  // => false (not approved)
CanDeploy(true, true, true)    // => false (freeze is active)
CanDeploy(false, false, true)  // => false
```

**Think it through:**
- "All at once" means "and". Which operator is that?
- `freezeActive` is `true` when there *is* a freeze. How do you turn that into
  "no freeze"?

### 3. Check work hours

**Goal:** make `IsWorkHours` return `true` when `hour` is 9 or more, `hour`
is below 17, and `weekend` is `false`.

`hour` is the hour of the day, from 0 to 23. Work hours run from 9:00 up to,
but not including, 17:00, and only on weekdays.

```go
IsWorkHours(9, false)  // => true
IsWorkHours(12, false) // => true
IsWorkHours(16, false) // => true
IsWorkHours(17, false) // => false
IsWorkHours(8, false)  // => false
IsWorkHours(0, false)  // => false
IsWorkHours(10, true)  // => false (weekend)
```

**Think it through:**
- This is three small checks joined together, like task 2.
- 9 is included but 17 is not. Which two comparison operators match that?

## Learn more

- [A Tour of Go: Basic types](https://go.dev/tour/basics/11)
- [Go spec: Comparison operators](https://go.dev/ref/spec#Comparison_operators)
- [Go spec: Logical operators](https://go.dev/ref/spec#Logical_operators)
