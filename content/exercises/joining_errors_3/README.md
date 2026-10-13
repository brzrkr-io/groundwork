# Reporting many errors: run every shutdown step

When a service shuts down it should drain traffic, flush logs, close database
connections and leave the load balancer, even if one of those steps fails.
Skipping the rest leaves a mess for the next deploy. Here you run every
cleanup step, report all failures, and then pick the joined error apart again.

## Learn

Recap: collect failures in a `[]error` and return `errors.Join(errs...)`. It
is `nil` when the slice holds no errors. `errors.Is` and `errors.AsType` search
every error inside a joined error.

### Getting the list back out

Sometimes you need the separate errors again, for example to count them or to
print each one on its own. A joined error has a method that returns them:

```go
Unwrap() []error
```

Compare it with the `Unwrap() error` method from *Wrapping errors*: one
wrapped error versus a list of them.

The type `errors.Join` returns is not exported, so you can't name it. Instead
you ask "does this error have an `Unwrap() []error` method?" with a type
assertion to an interface, as in *Type assertions and switches*. The
interface can be written right inside the assertion, without giving it a
name:

```go
joined, ok := err.(interface{ Unwrap() []error })
if ok {
	for _, e := range joined.Unwrap() {
		fmt.Println("problem:", e)
	}
}
```

- `ok` is `true` when `err` has the method. `joined.Unwrap()` then gives you
  the errors in the order they were joined, without the skipped `nil`s.
- `ok` is `false` for a plain error, such as one from `errors.New`, and for
  `nil`.
- `errors.Join` with a single non-nil error still returns a joined error: a
  list of one.
- Wrapping a joined error with `fmt.Errorf("...: %w", joined)` hides the list:
  the outer error only has `Unwrap() error`. Look for the list on the error
  that `errors.Join` returned.

## Your task

The shutdown hook of a service runs a list of cleanup `Step`s. Each step has a
`Name` and a `Run` function. The `StepError` type is ready for you: its
message is `step "<name>": <reason>`, and it has an `Unwrap() error` method.

### 1. Run every step

Change `Shutdown` so it calls `Run` on every step, in order, even after a step
fails. Wrap each failure as `&StepError{Step: s.Name, Err: err}` and join all
of them. Return `nil` when every step succeeds. You need to add the `errors`
import.

```go
Shutdown([]Step{
	{Name: "drain", Run: drain},           // fails: 3 requests still in flight
	{Name: "flush-logs", Run: flushLogs},  // succeeds
	{Name: "deregister", Run: deregister}, // fails: load balancer timed out
})
// => error with the message:
// step "drain": 3 requests still in flight
// step "deregister": load balancer timed out
```

### 2. Split a joined error

Change `Split` so it returns the errors inside `err`:

| `err` | `Split` returns |
|---|---|
| `nil` | `nil` |
| has an `Unwrap() []error` method | the result of that method |
| any other error | a list of just `err` |

```go
Split(errors.Join(errDrain, errFlush)) // => [errDrain errFlush]
Split(errDrain)                        // => [errDrain]
Split(nil)                             // => nil
```

### 3. List the failed steps

Change `FailedSteps` so it returns the `Step` name of every `*StepError` in
`err`, in order. Use `Split`, then `errors.AsType[*StepError]` on each error,
so a `*StepError` that was wrapped with `%w` is still found. Skip errors that
contain no `*StepError`. Return `nil` (an empty list) for a `nil` error.

```go
FailedSteps(Shutdown(steps)) // from the example in task 1
// => ["drain" "deregister"]

FailedSteps(errors.Join(errors.New("signal lost"), flushErr))
// flushErr is a *StepError for "flush-logs"
// => ["flush-logs"]
```

## Learn more

- [errors package: wrapping multiple errors](https://pkg.go.dev/errors#pkg-overview)
- [errors.Join](https://pkg.go.dev/errors#Join)
- [errors.AsType](https://pkg.go.dev/errors#AsType)
- [A Tour of Go: Type assertions](https://go.dev/tour/methods/15)
