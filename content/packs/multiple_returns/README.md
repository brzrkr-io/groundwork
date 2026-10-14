# Multiple return values

Go functions can hand back more than one value. That is how Go reports
"here is the answer, and here is whether it worked". You will use it every
time you parse a port, a config value or a command-line flag.

**New here:** multiple results, `_`, comma ok, `strings.Cut`, `error`, `nil`, `strconv.Atoi`, `errors.New`, `fmt.Errorf`.

## Learn

### Returning two values

List the result types in brackets, and return the values separated by commas:

```go
// MinMax returns the smaller and the larger of a and b.
func MinMax(a, b int) (int, int) {
	if a < b {
		return a, b
	}
	return b, a
}
```

> **Why is it written like this?** Go requires the brackets around two or
> more result types; one result needs none. Each `return` must give exactly
> as many values as the function line lists, in the same order.

The caller receives both values at once:

```go
low, high := MinMax(9, 4) // low is 4, high is 9
```

### Throwing a value away: `_`

Go insists that you use every variable you declare. If you don't need one
result, assign it to `_`, the *blank identifier*:

```go
_, high := MinMax(9, 4)
```

### Comma ok

Often the last value is a `bool`, usually named `ok` or `found`, that says
whether the other values mean anything. This is the *comma ok* pattern.
`strings.Cut` uses it to split a string around the first separator:

```go
key, value, found := strings.Cut("env=prod", "=")
// key is "env", value is "prod", found is true

key, value, found = strings.Cut("env", "=")
// key is "env", value is "", found is false
```

Notice: when nothing is found, `Cut` still gives back the whole input as the
first value. `!found` means "not found".

### Errors

A single `bool` can't say *what* went wrong. So most Go functions that can
fail return an `error` as their last value:

- Everything went fine: the error is `nil` (Go's word for "nothing").
- Something went wrong: the error is not `nil` and carries a message.

`strconv.Atoi` turns text into an `int` and follows this rule:

```go
n, err := strconv.Atoi("3")    // n is 3, err is nil
n, err = strconv.Atoi("three") // n is 0, err is not nil
```

Check the error straight after the call. If it isn't `nil`, stop and pass it
back to *your* caller, with zero values for the other results:

```go
replicas, err := strconv.Atoi(text)
if err != nil {
	return "", err
}
// from here on, replicas is a real number
```

> **Why is it written like this?** `if err != nil` right after the call is
> a convention every Go programmer follows. The error going *last* in the
> result list is also convention. The name `err` is the usual choice.

### Making your own errors

`errors.New` takes a fixed message. `fmt.Errorf` works like `fmt.Sprintf`
but gives you an error:

```go
import (
	"errors"
	"fmt"
)

err := errors.New("missing name")
err = fmt.Errorf("size %d is too big", 900)
```

Each package you use must be in the `import` list, or the program won't
compile. A later module covers errors in depth.

## Your task

Your team's status page shows how long servers have been up and which
address each one listens on. Write the helpers it needs.

### 1. Split uptime into hours and minutes

**Goal:** make `SplitUptime(totalMinutes int) (int, int)` return the whole hours, then the minutes left over.

```go
SplitUptime(135)  // => 2, 15
SplitUptime(59)   // => 0, 59
SplitUptime(120)  // => 2, 0
SplitUptime(1441) // => 24, 1
```

**Think it through:**
- Which operator from *Variables* gives whole hours, and which gives the remainder?
- One `return` line, two values, separated by a comma.

### 2. Split a host and a port

**Goal:** make `SplitHostPort(addr string) (string, string, bool)` split `addr` at the first `:` using `strings.Cut`.

- Colon found: return the host, the port text, and `true`.
- No colon: return `"", "", false` (not the whole input as host).

Keep the host exactly as given; don't change its letter case.

```go
SplitHostPort("db.internal:5432") // => "db.internal", "5432", true
SplitHostPort("Web-1:80")         // => "Web-1", "80", true
SplitHostPort("localhost")        // => "", "", false
SplitHostPort("")                 // => "", "", false
```

**Think it through:**
- What does `strings.Cut` give back as its first value when there's no colon? Why can't you return that directly?
- The starter's `strings.ToLower` doesn't belong here.

### 3. Parse an address

**Goal:** make `ParseAddress(addr string) (string, int, error)` return the host, the port as an `int`, and `nil` — or an error when the address is bad.

Return `""`, `0` and a non-nil error when:

- the address has no `:`,
- the port is not a whole number (use `strconv.Atoi`),
- the port is below 1 or above 65535.

The tests only check whether the error is `nil` or not, so the messages are
up to you. Add `strconv` and `errors` (or `fmt`) to the imports.

```go
ParseAddress("api.internal:8080")  // => "api.internal", 8080, nil
ParseAddress("bastion:22")         // => "bastion", 22, nil
ParseAddress("cache:65535")        // => "cache", 65535, nil
ParseAddress("api.internal")       // => "", 0, (an error)
ParseAddress("api.internal:http")  // => "", 0, (an error)
ParseAddress("api.internal:0")     // => "", 0, (an error)
ParseAddress("api.internal:65536") // => "", 0, (an error)
```

**Think it through:**
- You already wrote a function that splits and reports a missing colon. Reuse it.
- Three things can go wrong, so expect three checks, each with its own early `return`.
- Is 65535 allowed? Is 0?

## Learn more

- [A Tour of Go: Multiple results](https://go.dev/tour/basics/6)
- [Go by Example: Multiple Return Values](https://gobyexample.com/multiple-return-values)
- [Go by Example: Errors](https://gobyexample.com/errors)
- [strings.Cut](https://pkg.go.dev/strings#Cut)
