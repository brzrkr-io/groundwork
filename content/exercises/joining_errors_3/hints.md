## Task 1: collect, don't return

Keep a `var errs []error`. Inside the loop, when `s.Run()` returns an error,
append a `&StepError{...}` and carry on with the next step. Join after the
loop.

```go
for _, s := range steps {
	if err := s.Run(); err != nil {
		errs = append(errs, &StepError{Step: s.Name, Err: err})
	}
}
return errors.Join(errs...)
```

## Task 2: the type assertion

Handle `nil` first. Then ask whether `err` has the list method:

```go
if joined, ok := err.(interface{ Unwrap() []error }); ok {
	return joined.Unwrap()
}
return []error{err}
```

## Task 3: Split, then AsType

```go
var names []string
for _, e := range Split(err) {
	if stepErr, ok := errors.AsType[*StepError](e); ok {
		names = append(names, stepErr.Step)
	}
}
return names
```
