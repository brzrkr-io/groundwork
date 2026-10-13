# Booleans

Ops work is full of yes-or-no questions: is the port privileged, may this
deploy go out, is it work hours? Go answers them with *booleans*.

## Learn

### The bool type

A `bool` holds one of two values: `true` or `false`.

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

`==` and `!=` work on strings too: `env == "production"` is `true` only when
`env` is exactly `"production"`. Note the difference: `=` assigns a value, `==`
compares two values.

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

Go stops early when it already knows the answer. This is called
*short-circuit evaluation*: in `a && b`, if `a` is false, `b` is never looked at.

Use parentheses to group parts, just like in maths:

```go
(port >= 8000 && port < 9000) || port == 80
```

### Return the expression directly

A comparison already *is* a `bool`, so you can return it straight away:

```go
// IsEmpty reports whether the name is empty.
func IsEmpty(name string) bool {
	return name == ""
}
```

You will learn to make decisions with `if` in *if and else*. For this exercise
you don't need it: every answer is one `return` line with a boolean expression.

## Your task

You are writing small checks for a deploy tool. Each function returns a
hard-coded `false`. Replace it with the right boolean expression.

### 1. Spot privileged ports

On Linux, ports below 1024 are *privileged*: only root may listen on them.
Change `IsPrivilegedPort` so it returns `true` when `port` is below 1024.

```go
IsPrivilegedPort(22)   // => true
IsPrivilegedPort(1023) // => true
IsPrivilegedPort(1024) // => false
IsPrivilegedPort(8080) // => false
```

### 2. Decide whether a deploy may go out

A deploy may go out only when the tests passed, **and** someone approved it,
**and** there is no change freeze active. Change `CanDeploy` to return that.

```go
CanDeploy(true, true, false)  // => true
CanDeploy(true, false, false) // => false (not approved)
CanDeploy(true, true, true)   // => false (freeze is active)
```

### 3. Check work hours

`hour` is the hour of the day, from 0 to 23. Work hours run from 9:00 up to,
but not including, 17:00, and only on weekdays. Change `IsWorkHours` so it
returns `true` when `hour` is 9 or more, `hour` is below 17, and `weekend` is
`false`.

```go
IsWorkHours(9, false)  // => true
IsWorkHours(16, false) // => true
IsWorkHours(17, false) // => false
IsWorkHours(8, false)  // => false
IsWorkHours(10, true)  // => false (weekend)
```

## Learn more

- [A Tour of Go: Basic types](https://go.dev/tour/basics/11)
- [Go spec: Comparison operators](https://go.dev/ref/spec#Comparison_operators)
- [Go spec: Logical operators](https://go.dev/ref/spec#Logical_operators)
