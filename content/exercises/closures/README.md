# Functions as values

In Go a function is a value, just like a number or a string. You can store
it in a variable, pass it to another function, or build a new one and
return it. Retry helpers, middleware and health checks are all built this
way.

## Learn

### Functions are values

```go
double := func(n int) int {
	return n * 2
}
double(4) // => 8
```

`func(n int) int { ... }` is a *function literal*: a function without a
name. Here it is stored in the variable `double`, and you call it like any
other function.

### Function types

Every function value has a type that lists its parameters and results. The
type of `double` is `func(int) int`. Some more:

| Type | Meaning |
|---|---|
| `func() int` | takes nothing, returns an `int` |
| `func(int) bool` | takes an `int`, returns a `bool` |
| `func(attempt int) bool` | the same, with the parameter named for readers |

### Passing a function in

A parameter can have a function type. The caller decides what that function
does:

```go
// CountMatching reports how many of 1..n make check return true.
func CountMatching(n int, check func(int) bool) int {
	count := 0
	for i := 1; i <= n; i++ {
		if check(i) {
			count++
		}
	}
	return count
}

isEven := func(n int) bool { return n%2 == 0 }
CountMatching(10, isEven) // => 5
```

### Returning a function, and closures

A function can also build and return a function. The inner function can use
the outer function's variables, and it keeps them alive after the outer
function has returned. A function that does this is a *closure*:

```go
// NewGreeter returns a function that greets people from team.
func NewGreeter(team string) func(string) string {
	return func(name string) string {
		return "hi " + name + " from " + team
	}
}

sre := NewGreeter("sre")
sre("ana") // => "hi ana from sre"
```

A closure can also *change* the variables it captured. Each call to the
outer function creates fresh variables, so every closure gets its own:

```go
func NewTally() func() int {
	total := 0 // created once per NewTally call
	return func() int {
		total += 10
		return total
	}
}

a := NewTally()
b := NewTally()
a() // => 10
a() // => 20
b() // => 10, b has its own total
```

## Your task

You're building small helpers for a monitoring agent.

### 1. Build a threshold checker

Change `NewThresholdChecker` so it returns a function that reports `true`
when a value is at or above `limit`.

```go
tooHot := NewThresholdChecker(80)
tooHot(79) // => false
tooHot(80) // => true
tooHot(95) // => true
```

### 2. Build a request counter

Change `NewRequestCounter` so it returns a function that returns 1 the first
time it is called, then 2, then 3, and so on. Each counter counts on its
own.

```go
api := NewRequestCounter()
web := NewRequestCounter()
api() // => 1
api() // => 2
web() // => 1
```

### 3. Retry an operation

Change `Retry` so it calls `try(1)`, then `try(2)`, and so on, up to
`try(attempts)`. Stop as soon as `try` returns `true`, and return the attempt
number that succeeded. If no attempt succeeds, return `0`. Don't call `try`
again after it succeeds.

```go
succeedsOnThird := func(attempt int) bool { return attempt == 3 }
Retry(5, succeedsOnThird) // => 3
Retry(2, succeedsOnThird) // => 0
```

## Learn more

- [A Tour of Go: Function values](https://go.dev/tour/moretypes/24)
- [A Tour of Go: Function closures](https://go.dev/tour/moretypes/25)
- [Go by Example: Closures](https://gobyexample.com/closures)
