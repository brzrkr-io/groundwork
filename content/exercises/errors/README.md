# Errors are values

Things go wrong all the time in ops work: a config has a typo, a port number is
not a number, a hostname is missing. In Go, a function tells its caller that
something went wrong by returning an *error* value. In this exercise you write
your first functions that return errors.

## Learn

### The error type

`error` is a built-in interface type. It has one method:

```go
type error interface {
	Error() string
}
```

Any value with an `Error() string` method is an error. Calling `err.Error()`
gives you the message as a string. `fmt.Println(err)` prints the same message.

### nil means success

A function that can fail returns an `error` as its **last** result. If
everything worked, it returns `nil` (no error). If something failed, it
returns a non-nil error that describes the problem.

You met this already with `strconv.Atoi`:

```go
n, err := strconv.Atoi("8080")
if err != nil {
	// the text was not a number
}
```

`if err != nil` is the most common line in Go. It reads "if something went
wrong".

### Making errors

There are two ways to make your own error value:

```go
emptyErr := errors.New("service name is empty")                 // a fixed message
rangeErr := fmt.Errorf("port %d is out of range 1-65535", 70000) // a formatted message
```

`errors.New` lives in the `errors` package. `fmt.Errorf` works like
`fmt.Sprintf`, but gives you an error instead of a string. Use it when the
message should include the bad value:

| Verb | Prints | Example output |
|---|---|---|
| `%q` | a string in double quotes | `port "http" is not a number` |
| `%d` | an integer | `port 70000 is out of range 1-65535` |
| `%v` | any value, including an error's message | `endpoint "db:http": port "http" is not a number` |

The last row is how you add *context*: you take an error from a function you
called and put a short note in front of it, so the reader knows where it came
from.

### Return early, with zero values

When a function fails, it stops right away and returns the error. Its other
results get their *zero value*: `0` for numbers, `""` for strings. The caller
must not use them anyway.

```go
func ParseReplicas(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("replicas %q is not a number", s)
	}
	return n, nil
}
```

The happy path ends with `return n, nil`. Each problem gets its own `if` that
returns early. Your code stays flat instead of nesting deeper and deeper.

### Message style

Error messages often end up glued together ("deploy: load config: port ...")
and printed in logs. Go code follows a few rules so that works well:

| Good | Bad | Why |
|---|---|---|
| `service name is empty` | `Service name is empty` | start lowercase |
| `port "x" is not a number` | `port "x" is not a number.` | no trailing punctuation |
| `port "x" is not a number` | `invalid input` | say what was wrong, show the value |
| `endpoint "db:x": port "x" ...` | `port "x" ...: endpoint "db:x"` | context goes first |

## Your task

A deploy tool reads service names and network endpoints like `db:5432` from
its settings. Make it report clear errors when the input is bad.

Before you start, add these imports at the top of `exercise.go` (Go refuses to
compile a file that imports a package it doesn't use, so add each one when you
first need it):

```go
import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)
```

### 1. Reject an empty service name

`RequireName` returns an error made with `errors.New` when `name` is `""`, and
`nil` otherwise. The message is exactly:

```text
service name is empty
```

```go
RequireName("api") // => nil
RequireName("")    // => error "service name is empty"
```

### 2. Parse a port number

`ParsePort` turns text like `"8080"` into the number `8080`. A valid port is
between 1 and 65535, inclusive. On any error it returns `0` and the error:

| Input | Problem | Error message |
|---|---|---|
| `"http"` | `strconv.Atoi` fails | `port "http" is not a number` |
| `"70000"` | above 65535 | `port 70000 is out of range 1-65535` |
| `"0"` | below 1 | `port 0 is out of range 1-65535` |
| `"-1"` | below 1 | `port -1 is out of range 1-65535` |

```go
ParsePort("8080") // => 8080, nil
ParsePort("http") // => 0, error "port \"http\" is not a number"
```

### 3. Parse an endpoint

`ParseEndpoint` splits `"db:5432"` into the host `"db"` and the port `5432`.
Use `strings.Cut(s, ":")`, which returns the text before the first `:`, the
text after it, and `found` (a bool that is `false` when there is no `:`).

Check these in order. On any error, return `""`, `0` and the error:

| Input | Problem | Error message |
|---|---|---|
| `"db"` | no `:` | `endpoint "db" has no port` |
| `":5432"` | empty host | `endpoint ":5432" has no host` |
| `"db:http"` | `ParsePort` fails | `endpoint "db:http": port "http" is not a number` |
| `"db:0"` | `ParsePort` fails | `endpoint "db:0": port 0 is out of range 1-65535` |

For the last two rows, call your `ParsePort` and add context in front of its
error with `fmt.Errorf("endpoint %q: %v", s, err)`.

```go
ParseEndpoint("db:5432") // => "db", 5432, nil
ParseEndpoint("db")      // => "", 0, error "endpoint \"db\" has no port"
```

## Learn more

- [A Tour of Go: Errors](https://go.dev/tour/methods/19)
- [Go by Example: Errors](https://gobyexample.com/errors)
- [Go blog: Error handling and Go](https://go.dev/blog/error-handling-and-go)
- [Go Code Review Comments: Error strings](https://go.dev/wiki/CodeReviewComments#error-strings)
