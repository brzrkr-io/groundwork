# Functions as values: state that sticks

A closure is only useful if its state survives between calls. Here you fix a
counter that forgets everything, then build an ID generator and a rate
limiter for an API gateway.

**New here:** nothing; practice. (`strconv.Itoa` is recapped below.)

## Learn

### Recap: where the variable lives

A *closure* is a function literal that uses variables from the function
around it, as in *Functions as values*. Where you declare a variable decides
how long it lasts:

| Declared… | Created… | So it… |
|---|---|---|
| in the outer function, before `return func…` | once per outer call | keeps its value between calls |
| inside the returned function | every time the closure runs | starts over on every call |

```go
func NewLog() func(string) string {
	lines := "" // outer: created once, remembered
	return func(msg string) string {
		lines += msg + ";" // changes the outer lines
		return lines
	}
}

add := NewLog()
add("boot") // => "boot;"
add("ready") // => "boot;ready;"
```

> **Why is it written like this?** Go requires a variable to exist before you
> use it, and `:=` creates a *new* variable each time that line runs. So the
> line's position is what matters: outside the closure it runs once, inside
> it runs on every call. Names like `lines` are your choice.

### Recap: numbers into text

`+` only joins strings with strings. To put a number into a string, turn it
into text first:

```go
import "strconv"

port := 8080
"listening on " + strconv.Itoa(port) // => "listening on 8080"
```

`fmt.Sprintf("listening on %d", port)` does the same. Whichever you use, add
its package to the `import` line at the top of the file.

## Your task

You're writing helpers for an API gateway.

### 1. Fix the error counter

**Goal:** make the function from `NewErrorCounter` return 1, 2, 3, … on successive calls.

Right now it returns 1 every time. Each counter counts on its own.

```go
count := NewErrorCounter()
count() // => 1
count() // => 2
count() // => 3

db := NewErrorCounter()
cache := NewErrorCounter()
db(); db()
cache() // => 1
```

**Think it through:**
- Look at the line that sets `count` to 0. Is it inside or outside the returned function?
- Using the table above, where should it go instead? This task only moves a line.

### 2. Build an ID generator

**Goal:** make the function from `NewIDGenerator` return `prefix`, a `-`, and a number that starts at 1 and grows by one per call.

Three pieces: `prefix` comes from the parameter, `-` is fixed text, and the
number is state the closure keeps. There are no spaces.

```go
next := NewIDGenerator("req")
next() // => "req-1"
next() // => "req-2"
next() // => "req-3"
NewIDGenerator("job")() // => "job-1"
```

**Think it through:**
- Which piece needs to survive between calls, and where do you declare it?
- Which piece goes in quotes, and which are names?
- The number is an `int`. What turns it into text, and which import does that need?

### 3. Build a rate limiter

**Goal:** make the function from `NewRateLimiter` return `true` for the first `limit` calls and `false` for every call after that.

With a `limit` of 0, every call returns `false`.

```go
allow := NewRateLimiter(2)
allow() // => true
allow() // => true
allow() // => false
allow() // => false

never := NewRateLimiter(0)
never() // => false
```

**Think it through:**
- What does the closure need to remember between calls?
- Call number 2 with `limit` 2 is allowed; call 3 isn't. Which comparison operator gives that?

## Learn more

- [A Tour of Go: Function closures](https://go.dev/tour/moretypes/25)
- [Go by Example: Closures](https://gobyexample.com/closures)
- [strconv.Itoa](https://pkg.go.dev/strconv#Itoa)
- [Go spec: Function literals](https://go.dev/ref/spec#Function_literals)
