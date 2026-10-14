# Reporting many errors: run every shutdown step

When a service shuts down it should drain traffic, flush logs, close database
connections and leave the load balancer, even if one step fails. Skipping the
rest leaves a mess for the next deploy. Here you run every cleanup step,
report all failures, then pick the joined error apart again.

**New here:** the `Unwrap() []error` method, type assertion to an unnamed interface.

## Learn

Recap from *Reporting many errors*: collect failures in a `[]error` and return
`errors.Join(errs...)`, which is `nil` when nothing failed. `errors.Is` and
`errors.AsType` search every error inside a joined error.

### Getting the list back out

Sometimes you need the separate errors again, to count them or print each one.
A joined error has a method that returns them:

```go
Unwrap() []error
```

Compare it with `Unwrap() error` from *Wrapping errors: triage a failed
deploy*: one wrapped error versus a list.

### Asking "does it have this method?"

The type `errors.Join` returns is not exported, so you can't name it. Instead,
ask whether the error has the method, with a type assertion to an interface
(as in *Type assertions and switches*). The interface can be written inline:

```go
func countProblems(err error) int {
	if multi, ok := err.(interface{ Unwrap() []error }); ok {
		return len(multi.Unwrap())
	}
	return 0
}
```

> **Why is it written like this?** `interface{ Unwrap() []error }` is a type
> written in place, with no name (Go allows any type there). The method name
> and signature must match exactly, or `ok` is `false`. The variable name
> `multi` is your choice.

What you get back:

- `ok` is `true` when `err` has the method. `Unwrap()` then gives the errors
  in the order they were joined, without the skipped `nil`s.
- `ok` is `false` for a plain error (say, from `errors.New`) and for `nil`.
- `errors.Join` with a single non-nil error still returns a joined error: a
  list of one.
- Wrapping a joined error with `fmt.Errorf("...: %w", joined)` hides the list:
  the outer error only has `Unwrap() error`.

## Your task

The shutdown hook runs a list of cleanup `Step`s, each with a `Name` and a
`Run` function. `StepError` is ready: its message is `step "<name>": <reason>`,
and it has an `Unwrap() error` method.

### 1. Run every step

**Goal:** make `Shutdown(steps)` run every step and return all failures joined, as `*StepError`s.

- Call `Run` on every step, in order, even after a step fails.
- Wrap each failure as `&StepError{Step: s.Name, Err: err}`.
- Join them all. Return `nil` when every step succeeds.
- You need to add the `errors` import.

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

**Think it through:**
- Where do you keep failures until the loop is done?
- `Run` is a field holding a function; how do you call it?

### 2. Split a joined error

**Goal:** make `Split(err)` return the errors inside `err` as a slice.

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

**Think it through:**
- Handle the rows top to bottom.
- How do you write a slice holding one error?

### 3. List the failed steps

**Goal:** make `FailedSteps(err)` return the `Step` name of every `*StepError` in `err`, in order.

- Use your `Split`, then `errors.AsType[*StepError]` on each error, so a
  `*StepError` wrapped with `%w` is still found.
- Skip errors that contain no `*StepError`.
- Return `nil` (an empty list) for a `nil` error.

```go
FailedSteps(Shutdown(steps)) // from the example in task 1
// => ["drain" "deregister"]

FailedSteps(errors.Join(errors.New("signal lost"), flushErr))
// flushErr is a *StepError for "flush-logs"
// => ["flush-logs"]
```

**Think it through:**
- Which field of the found `*StepError` holds the name?
- How do you declare the names slice so it stays `nil` when nothing is added?

## Learn more

- [errors package: wrapping multiple errors](https://pkg.go.dev/errors#pkg-overview)
- [errors.Join](https://pkg.go.dev/errors#Join)
- [errors.AsType](https://pkg.go.dev/errors#AsType)
- [A Tour of Go: Type assertions](https://go.dev/tour/methods/15)
