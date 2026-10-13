## Task 1: two small methods

`Error` is one `fmt.Sprintf` call with three values. `Unwrap` is a new method
with the signature `func (e *DeployError) Unwrap() error`, and its body is one
line.

## Task 2: loop and return early

Range over `stages`. Call `stage.Run()` and check its error. The first time it
is not `nil`, return `&DeployError{...}` straight away. After the loop, return
`nil`.

```go
if err := stage.Run(); err != nil {
	return &DeployError{Service: service, Stage: stage.Name, Err: err}
}
```

## Task 3: two checks

```go
if errors.Is(err, ErrUnhealthy) {
	return true
}
derr, ok := errors.AsType[*DeployError](err)
return ok && derr.Stage == "rollout"
```
