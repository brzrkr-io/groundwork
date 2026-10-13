# Multiple return values

Go functions can hand back more than one value. That is how Go reports
"here is the answer, and here is whether it worked". You will use it every
time you parse a port, a config value or a command-line flag.

## Learn

### Returning two values

List the result types in parentheses, and return the values separated by
commas:

```go
// MinMax returns the smaller and the larger of a and b.
func MinMax(a, b int) (int, int) {
	if a < b {
		return a, b
	}
	return b, a
}
```

The caller receives both values at once:

```go
low, high := MinMax(9, 4) // low is 4, high is 9
```

If you don't need one of them, assign it to `_`, the *blank identifier*.
Go throws that value away:

```go
_, high := MinMax(9, 4)
```

Go insists that you use every variable you declare, so `_` is how you say
"I really don't want this one".

### Comma ok

Often the second value is a `bool` called `ok` that says whether the first
value means anything. This is the *comma ok* pattern. `strings.Cut` uses it.
It splits a string around the first separator it finds:

```go
before, after, found := strings.Cut("db:5432", ":")
// before is "db", after is "5432", found is true

before, after, found = strings.Cut("db", ":")
// before is "db", after is "", found is false
```

### Errors

Something can go wrong in many ways, and a single `bool` can't say which. So
most Go functions that can fail return an *error* as their last value. The
type is called `error`. The rule is simple:

- When everything went fine, the error is `nil` (Go's word for "nothing").
- When something went wrong, the error is not `nil` and carries a message.

`strconv.Atoi` turns text into an `int` and follows this rule:

```go
n, err := strconv.Atoi("8080") // n is 8080, err is nil
n, err = strconv.Atoi("http")  // n is 0, err is not nil
```

Check the error straight after the call. If it is not `nil`, stop and pass
it back to *your* caller:

```go
port, err := strconv.Atoi(text)
if err != nil {
	return 0, err
}
// from here on, port is a real number
```

You can make your own errors. `errors.New` takes a fixed message, and
`fmt.Errorf` works like `fmt.Sprintf` but gives you an error:

```go
import (
	"errors"
	"fmt"
)

err := errors.New("missing port")
err = fmt.Errorf("port %d is out of range", 70000)
```

When you return an error, return zero values for the other results, such as
`""` and `0`. The caller should ignore them anyway. A later module covers
errors in depth.

## Your task

Your team's status page shows how long servers have been up and which
address each one listens on. Write the helpers it needs.

### 1. Split uptime into hours and minutes

Change `SplitUptime` so it returns two values: the whole hours, and the
minutes left over. Use `/` and `%` from *Variables*.

```go
SplitUptime(135) // => 2, 15
SplitUptime(59)  // => 0, 59
SplitUptime(120) // => 2, 0
```

### 2. Split a host and a port

Change `SplitHostPort` so it uses `strings.Cut` to split the address at the
first `:`. Return the host, the port text, and `true`. If there is no `:`,
return `"", "", false`.

```go
SplitHostPort("db.internal:5432") // => "db.internal", "5432", true
SplitHostPort("localhost")        // => "", "", false
```

### 3. Parse an address

Change `ParseAddress` so it returns the host, the port as an `int`, and an
error. Return a non-nil error when:

- the address has no `:`,
- the port is not a whole number (use `strconv.Atoi`),
- the port is below 1 or above 65535.

Whenever you return an error, also return `""` and `0`. The tests only check
whether the error is `nil` or not, so the messages are up to you. You'll need
to add `errors` (or `fmt`) and `strconv` to the imports.

```go
ParseAddress("api.internal:8080") // => "api.internal", 8080, nil
ParseAddress("api.internal")      // => "", 0, (an error)
ParseAddress("api.internal:http") // => "", 0, (an error)
ParseAddress("api.internal:0")    // => "", 0, (an error)
```

## Learn more

- [A Tour of Go: Multiple results](https://go.dev/tour/basics/6)
- [Go by Example: Multiple Return Values](https://gobyexample.com/multiple-return-values)
- [Go by Example: Errors](https://gobyexample.com/errors)
- [strings.Cut](https://pkg.go.dev/strings#Cut)
