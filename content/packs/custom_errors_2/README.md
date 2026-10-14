# Custom error types: fix the port allocator

A teammate wrote a port allocator that hands out ports to services on a shared
host. It has a custom error type, but the code has bugs. One of them is the
most famous trap in Go error handling.

**New here:** the typed-nil trap, up close.

## Learn

### Recap

From *Custom error types*:

- A custom error is a struct with an `Error() string` method on a pointer receiver.
- You return it as `&T{...}` through an `error` result.
- `v, ok := err.(*T)` gets the fields back. `ok` is `false` for `nil` and for other kinds of error.

### The typed-nil trap, up close

An `error` value has two parts: a *type* and a *value*. It is `nil` only when
**both** parts are empty.

```go
var p *DiskFullError  // type *DiskFullError, value nil
var err error = p     // err now remembers that type

fmt.Println(p == nil)   // true
fmt.Println(err == nil) // false!
```

So a function that keeps its error in a variable of a pointer error type, and
returns that variable as `error`, is broken on its success path. The caller's
`if err != nil` fires even though nothing went wrong.

The fix: don't send success through such a variable. Return the error right
where it happens, and write `return nil` for success.

```go
if diskFull {
	return &DiskFullError{Mount: "/var"}
}
return nil
```

> **Why is it written like this?** Go requires the two-part interface value;
> that's how the language works. Returning errors right where they happen is
> convention, and it also avoids this bug.

When the test reports this bug, it prints the hidden type with `%T`, so you'll
see `*customerr2.PortInUseError` even though no port was taken.

### Checking a map key

`value, ok := m[key]` gives the value and whether the key was there. You can
put it in an `if` so `value` and `ok` only exist inside that `if`:

```go
if size, ok := quotas["alice"]; ok {
	fmt.Println("alice has", size)
}
```

## Your task

The allocator keeps a `map[int]string` from each taken port to the service
that owns it. Fix it so deploys get correct answers.

### 1. Fix the error message

**Goal:** `(*PortInUseError).Error` returns the port and the owner.

The format is `port <Port> is already used by <Owner in double quotes>`:

```go
(&PortInUseError{Port: 8080, Owner: "nginx"}).Error()
// => `port 8080 is already used by "nginx"`
(&PortInUseError{Port: 5432, Owner: "postgres"}).Error()
// => `port 5432 is already used by "postgres"`
```

**Think it through:**
- Which two fields of `e` go into the message, and which verb fits each?
- Compare the old fixed text (`in use`) with the new one word by word.

### 2. Fix the typed-nil bug in Reserve

**Goal:** `Reserve(used, port, service)` returns `nil` for a free port, and a `*PortInUseError` for a taken one.

- Port is taken (it's a key in `used`): return a `*PortInUseError` with that `Port` and the `Owner` from `used`. Don't change `used`.
- Port is free: set `used[port] = service` and return `nil`.

Today, free ports are recorded correctly but still return a non-nil error.

```go
used := map[int]string{8080: "nginx"}
Reserve(used, 9090, "api") // => nil, and used[9090] is now "api"
Reserve(used, 8080, "api") // => error `port 8080 is already used by "nginx"`
```

**Think it through:**
- Look at the `var` line and the last `return`. Why is that error never truly `nil`?
- Can each branch return straight away instead of storing into a variable?

### 3. Finish Advice

**Goal:** when `err` is a `*PortInUseError`, `Advice` returns `stop "<Owner>" or choose a port other than <Port>`.

`nil` (returns `"ok"`) and other errors (returns `"unexpected error: "` plus
the message) already work. Add the missing case between them.

```go
Advice(&PortInUseError{Port: 8080, Owner: "nginx"})
// => `stop "nginx" or choose a port other than 8080`
Advice(errors.New("permission denied"))
// => "unexpected error: permission denied"
Advice(nil)
// => "ok"
```

**Think it through:**
- How do you check whether `err` holds a `*PortInUseError` *and* get its fields? See the recap.
- The owner comes first in this message and has quotes; the port comes last.

## Learn more

- [Go FAQ: Why is my nil error value not equal to nil?](https://go.dev/doc/faq#nil_error)
- [Go spec: Type assertions](https://go.dev/ref/spec#Type_assertions)
- [A Tour of Go: Interface values with nil underlying values](https://go.dev/tour/methods/12)
