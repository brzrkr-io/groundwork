# Functions as values

In Go a function is a value, just like a number or a string. You can store
it in a variable, pass it to another function, or build a new one and
return it. Retry helpers, middleware and health checks are all built this
way.

**New here:** function literals, function types, closures.

## Learn

### Functions are values

```go
double := func(n int) int {
	return n * 2
}
double(4) // => 8
```

`func(n int) int { ... }` is a *function literal*: a function without a
name. Here it is stored in the variable `double`. You call it like any
other function: the variable name, then `(...)`.

> **Why is it written like this?** Go requires the same parts as a named
> function (`func`, parameters, result type, `{ body }`), just without a name
> after `func`. The variable name `double` is your choice.

### Function types

Every function value has a *type* that lists its parameters and results.
The type of `double` is `func(int) int`. Some more:

| Type | Meaning |
|---|---|
| `func() int` | takes nothing, returns an `int` |
| `func(int) bool` | takes an `int`, returns a `bool` |
| `func(attempt int) bool` | the same, with the parameter named for readers |

> **Why is it written like this?** In a type, parameter names are optional:
> `func(int) bool` and `func(attempt int) bool` are the same type. Adding the
> name is the author's choice, to tell readers what the number means.

### Passing a function in

A parameter can have a function type. The caller decides what that function
does; your code just calls it by its parameter name.

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

Notice `check(i)`: inside `CountMatching` the function is called `check`,
whatever the caller named it.

### Returning a function, and closures

A function can also build and return a function. The inner function can use
the outer function's variables and parameters, and it keeps them alive after
the outer function has returned. A function that does this is a *closure*:

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

Read the first line in two halves: `NewGreeter(team string)` is what you
pass in; `func(string) string` is what you get back, a function.

### Closures can remember changes

A closure can also *change* a variable it captured. The variable must be
declared in the outer function, outside the inner one; a variable declared
inside the inner function starts fresh on every call. Each call to the
outer function makes new variables, so every closure gets its own:

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

You're building small helpers for a monitoring agent. The starter already
has the `return func...` lines; you fill in the bodies.

### 1. Build a threshold checker

**Goal:** make the function returned by `NewThresholdChecker` report `true` when a value is at or above `limit`.

`limit` comes from the outer function's parameter; the value to check is
the inner function's parameter.

```go
tooHot := NewThresholdChecker(80)
tooHot(79) // => false
tooHot(80) // => true
tooHot(95) // => true
```

**Think it through:**
- Which comparison operator means "at or above"?
- Which two names are you comparing: one from the outer line, one from the inner?

### 2. Build a request counter

**Goal:** make the function returned by `NewRequestCounter` return 1, then 2, then 3, and so on, on each call.

Each counter counts on its own: a new counter starts at 1 again.

```go
api := NewRequestCounter()
web := NewRequestCounter()
api() // => 1
api() // => 2
web() // => 1
api() // => 3
```

**Think it through:**
- The count must survive between calls. Which function should declare it:
  the outer one or the inner one?
- Do you add 1 before or after returning, so the first call gives 1?

### 3. Retry an operation

**Goal:** make `Retry` call `try` with attempt numbers 1, 2, 3 … up to `attempts`, and return the first attempt number that succeeds.

- Stop as soon as `try` returns `true`. Don't call `try` again after that.
- If no attempt succeeds, return `0`. Every attempt has been used by then.
- If `attempts` is `0`, don't call `try` at all; return `0`.

```go
succeedsOnThird := func(attempt int) bool { return attempt == 3 }
Retry(5, succeedsOnThird) // => 3 (try called 3 times)
Retry(2, succeedsOnThird) // => 0 (try called 2 times)
Retry(0, succeedsOnThird) // => 0 (try never called)
```

**Think it through:**
- Which loop counts from 1 up to `attempts`? Look at `CountMatching` above.
- `try` is a parameter. What do you pass it, and what do you do with the `bool` it gives back?
- Where does `return 0` go so it only runs when nothing worked?

## Learn more

- [A Tour of Go: Function values](https://go.dev/tour/moretypes/24)
- [A Tour of Go: Function closures](https://go.dev/tour/moretypes/25)
- [Go by Example: Closures](https://gobyexample.com/closures)
