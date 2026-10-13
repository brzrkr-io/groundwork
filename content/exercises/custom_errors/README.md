# Custom error types

An error from `errors.New` or `fmt.Errorf` carries only a message. Sometimes
the caller needs details: which config field was wrong, which host timed out.
A *custom error type* keeps those details in fields, so code can react to
them instead of reading the message.

**New here:** custom error types, pointer receivers for errors, `&T{...}`, the nil-pointer trap.

## Learn

### Any type with an Error method is an error

As in *Errors are values*, `error` is an interface with one method,
`Error() string`. Any type with that method can be used as an `error`. A struct
is the usual choice, because its fields hold the details:

```go
// NotFoundError reports a service that is not in the registry.
type NotFoundError struct {
	Service string
}

func (e *NotFoundError) Error() string {
	return "service " + e.Service + " not found"
}
```

Inside the method, `e` is the error itself, so `e.Service` reads its field.

> **Why is it written like this?** Go requires the exact method name and
> signature `Error() string`. Ending the type name with `Error` is
> convention. The receiver name `e` is the author's choice (short names are
> convention).

### Pointer receiver, so create it with `&`

`(e *NotFoundError)` is a *pointer receiver*. It means the pointer type
`*NotFoundError` has the method, so only the pointer counts as an `error`.
You make a pointer to a new struct with `&`:

```go
func Lookup(name string) error {
	if name != "api" {
		return &NotFoundError{Service: name}
	}
	return nil
}
```

Without `&`, Go refuses to compile. That's useful: the error always has one
type, `*NotFoundError`, so every caller checks for that one type.

`Lookup` returns plain `error`, not `*NotFoundError`. Callers who only care
whether something failed still write `if err != nil`.

### Getting the details back out

Use a type assertion or a type switch, as in *Type assertions and switches*:

```go
nf, ok := err.(*NotFoundError)
if ok {
	fmt.Println("missing service:", nf.Service)
}
```

If `err` is `nil` or a different kind of error, `ok` is `false`. Nothing
crashes. For several kinds at once, use a type switch:

```go
switch err.(type) {
case nil:
	fmt.Println("all good")
case *NotFoundError:
	fmt.Println("register it first")
default:
	fmt.Println("failed:", err)
}
```

Cases are tried top to bottom. `case nil` matches when there is no error.

### The nil-pointer trap: always write `return nil`

Never return a nil *pointer* of your error type through an `error` result:

```go
func Lookup(name string) error {
	var nf *NotFoundError // a nil pointer
	if name != "api" {
		nf = &NotFoundError{Service: name}
	}
	return nf // ✗ even when nf is nil, the error is NOT nil
}
```

An interface value remembers the *type* of what you put in it. "A nil
`*NotFoundError`" is not the same as "no error", so `err != nil` is `true`.
On success, always write `return nil`.

## Your task

Your deploy tool reads a replica count from a config file. Give it a proper
validation error, then use that error to choose the tool's exit code.

### 1. Write the error message

**Goal:** the `Error` method of `ValidationError` returns the `Field`, then `: `, then the `Reason`.

```go
err := &ValidationError{Field: "replicas", Reason: "must be at least 1"}
err.Error() // => "replicas: must be at least 1"

err = &ValidationError{Field: "image", Reason: "must not be empty"}
err.Error() // => "image: must not be empty"
```

**Think it through:**
- Two pieces come from the fields of `e`. One piece is fixed text between them.
- The fixed text is a colon **and** a space.

### 2. Validate the replica count

**Goal:** `CheckReplicas(n)` returns `nil` for counts from 1 to 50, and a `*ValidationError` otherwise.

| `n` | `Field` | `Reason` |
|---|---|---|
| below 1 (0, -3) | `"replicas"` | `"must be at least 1"` |
| above 50 (51, 500) | `"replicas"` | `"must be at most 50"` |
| 1 to 50 | no error: `nil` | |

```go
CheckReplicas(0)  // => error "replicas: must be at least 1"
CheckReplicas(51) // => error "replicas: must be at most 50"
CheckReplicas(50) // => nil
```

The tests check that the error is a `*ValidationError` and check both fields.

**Think it through:**
- You need two separate checks, because the two `Reason`s differ.
- How do you build a `*ValidationError` with its fields filled in? See `Lookup` above.
- What should the last line return? Remember the nil-pointer trap.

### 3. Pick an exit code

**Goal:** `ExitCode(err)` returns the number a command-line tool exits with.

An *exit code* tells the shell how a tool finished. Many tools use `2` for
"you gave me bad input".

| `err` | Exit code |
|---|---|
| `nil` | `0` |
| a `*ValidationError` | `2` |
| any other error | `1` |

```go
ExitCode(nil)                                // => 0
ExitCode(CheckReplicas(0))                   // => 2
ExitCode(errors.New("registry unreachable")) // => 1
```

**Think it through:**
- Three outcomes depend on the *type* inside `err`. Which tool from Learn handles several types?
- Which case catches "any other error"?

## Learn more

- [A Tour of Go: Errors](https://go.dev/tour/methods/19)
- [Go blog: Error handling and Go](https://go.dev/blog/error-handling-and-go)
- [Go FAQ: Why is my nil error value not equal to nil?](https://go.dev/doc/faq#nil_error)
