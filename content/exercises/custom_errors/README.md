# Custom error types

An error made with `errors.New` or `fmt.Errorf` carries only a message. Some
callers need more: which config field was wrong, which port was taken, which
host timed out. A custom error type carries that data in fields, so code can
react to it instead of parsing a message.

## Learn

### Any type with an Error method is an error

As in *Errors are values*, `error` is an interface with one method:

```go
type error interface {
	Error() string
}
```

So any type with an `Error() string` method can be used as an `error`. A
struct is the usual choice, because its fields hold the details:

```go
// NotFoundError reports a service that is not in the registry.
type NotFoundError struct {
	Service string
}

func (e *NotFoundError) Error() string {
	return "service " + e.Service + " not found"
}
```

Name custom error types with an `Error` suffix: `NotFoundError`,
`ValidationError`, `TimeoutError`.

### Pointer receiver, so return `&T{...}`

The method above has a *pointer receiver*, `(e *NotFoundError)`. That means
the pointer type `*NotFoundError` has the method and satisfies `error`. The
plain struct `NotFoundError` does not. So you create the error with `&`:

```go
func Lookup(name string) error {
	if name != "api" {
		return &NotFoundError{Service: name} // ✓ a *NotFoundError is an error
	}
	return nil
}
```

Writing `return NotFoundError{Service: name}` (without `&`) does not compile.
That is a good thing: the error always has the same type, `*NotFoundError`, so
everyone who checks for it checks for that one type. Most custom errors in Go
code use a pointer receiver for this reason.

Notice that `Lookup` returns `error`, not `*NotFoundError`. Callers that only
care whether something went wrong use the usual `if err != nil`.

### Getting the details back out

Callers that want the fields use a type assertion or a type switch, as you did
with interfaces in Module 4:

```go
err := Lookup("web")

nf, ok := err.(*NotFoundError)
if ok {
	fmt.Println("missing service:", nf.Service) // missing service: web
}
```

If `err` is `nil` or a different kind of error, `ok` is `false`. Nothing
crashes. To handle several kinds at once, use a type switch:

```go
switch e := err.(type) {
case nil:
	fmt.Println("all good")
case *NotFoundError:
	fmt.Println("register", e.Service, "first")
default:
	fmt.Println("failed:", err)
}
```

Cases are checked from top to bottom, and `case nil` matches when there is no
error at all.

### Always write `return nil`

One trap catches almost everyone once. Never return a nil *pointer* of your
error type through an `error` result:

```go
func Lookup(name string) error {
	var nf *NotFoundError // a nil pointer
	if name != "api" {
		nf = &NotFoundError{Service: name}
	}
	return nf // ✗ even when nf is nil, the error is NOT nil
}
```

An interface value remembers the *type* of what you put in it. Here it holds
"a `*NotFoundError` that happens to be nil", and that is not the same as no
error, so `err != nil` is `true`. The fix is simple: on success, always write
`return nil`.

## Your task

Your deploy tool reads a replica count from a config file. You'll give it a
proper validation error and use that error to choose the tool's exit code.

### 1. Write the error message

`ValidationError` has two fields, `Field` and `Reason`. Make its `Error`
method return the field, a colon and a space, then the reason:

```go
err := &ValidationError{Field: "replicas", Reason: "must be at least 1"}
err.Error()
// => "replicas: must be at least 1"
```

### 2. Validate the replica count

`CheckReplicas(n int) error` accepts counts from 1 to 50.

- Below 1, return `&ValidationError{Field: "replicas", Reason: "must be at least 1"}`.
- Above 50, return `&ValidationError{Field: "replicas", Reason: "must be at most 50"}`.
- Otherwise, return `nil`.

```go
CheckReplicas(0)
// => error "replicas: must be at least 1"
CheckReplicas(51)
// => error "replicas: must be at most 50"
CheckReplicas(3)
// => nil
```

The tests use a type assertion to check that the error is a
`*ValidationError` and then check both fields.

### 3. Pick an exit code

Command-line tools report success or failure through an *exit code*. Many
tools use 2 for "you gave me bad input". Make `ExitCode(err error) int` return:

| `err` | Exit code |
|---|---|
| `nil` | `0` |
| a `*ValidationError` | `2` |
| any other error | `1` |

```go
ExitCode(nil)                                  // => 0
ExitCode(CheckReplicas(0))                     // => 2
ExitCode(errors.New("registry unreachable"))   // => 1
```

## Learn more

- [A Tour of Go: Errors](https://go.dev/tour/methods/19)
- [Go blog: Error handling and Go](https://go.dev/blog/error-handling-and-go)
- [Go FAQ: Why is my nil error value not equal to nil?](https://go.dev/doc/faq#nil_error)
