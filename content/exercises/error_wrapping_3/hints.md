## Task 1: Make DeployError a link in the chain

`Error` needs three values in one formatted string: the service, the stage
and the message of `Err`. `Unwrap` is a brand-new method that hands back the
error stored inside.

### Closer

Use `fmt.Sprintf("deploy %s: stage %s: %v", ...)`. The new method has the
signature `func (e *DeployError) Unwrap() error`, and its body is one line.

### Answer

```go
func (e *DeployError) Error() string {
	return fmt.Sprintf("deploy %s: stage %s: %v", e.Service, e.Stage, e.Err)
}

// Unwrap returns the stage's error, so errors.Is and errors.AsType can see it.
func (e *DeployError) Unwrap() error {
	return e.Err
}
```

`errors.Is` and `errors.AsType` call `Unwrap` to step from your error to the
one inside it.

## Task 2: Run the pipeline

Loop over `stages` and call each stage's `Run`. The first time it returns an
error, return straight away so later stages never run.

### Closer

`if err := stage.Run(); err != nil { ... }` inside a `for _, stage := range
stages` loop. Build the result with `&DeployError{Service: ..., Stage: ...,
Err: ...}`, and return `nil` after the loop.

### Answer

```go
for _, stage := range stages {
	if err := stage.Run(); err != nil {
		return &DeployError{Service: service, Stage: stage.Name, Err: err}
	}
}
return nil
```

Returning a plain `nil` (not a nil `*DeployError`) avoids the typed-nil gotcha.

## Task 3: Decide whether to roll back

There are two separate rules. One looks for a sentinel value in the chain,
the other looks for a type and then checks one of its fields.

### Closer

Use `errors.Is(err, ErrUnhealthy)` for the first rule. For the second,
`derr, ok := errors.AsType[*DeployError](err)` and then check `derr.Stage`.
Both return `false` for `nil`.

### Answer

```go
if errors.Is(err, ErrUnhealthy) {
	return true
}
derr, ok := errors.AsType[*DeployError](err)
return ok && derr.Stage == "rollout"
```

Your `Unwrap` from task 1 lets `errors.Is` find `ErrUnhealthy` inside a
`DeployError`.
