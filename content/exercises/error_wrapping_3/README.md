# Wrapping errors: triage a failed deploy

A deploy pipeline runs build, test and rollout stages. When one fails, the
on-call engineer needs two answers fast: which stage broke, and does the
previous version have to be restored? You will build an error type that
answers both.

## Learn

Recap from the previous exercises: `fmt.Errorf("context: %w", err)` wraps an
error into a chain. `errors.Is` searches the chain for a sentinel value, and
`errors.AsType[*T]` searches it for a type and hands you its fields.

### Your own type in the chain: Unwrap

How does `errors.Is` find the next error in a chain? It calls the error's
`Unwrap() error` method. The error that `fmt.Errorf` returns for `%w` has that
method. Your own error types can have it too:

```go
// QueryError reports a failed database query.
type QueryError struct {
	Query string
	Err   error // the underlying error
}

func (e *QueryError) Error() string {
	return fmt.Sprintf("query %q: %v", e.Query, e.Err)
}

// Unwrap returns the underlying error.
func (e *QueryError) Unwrap() error {
	return e.Err
}
```

A struct with an `Err` field and an `Unwrap` method gives you the best of
both:

- `errors.AsType[*QueryError](err)` finds it and gives you `Query`, like any
  custom error type from *Custom error types*.
- `errors.Is(err, ErrConnRefused)` still finds the error stored in `Err`,
  because `Unwrap` links it into the chain.

Without `Unwrap`, the chain stops at your type: everything inside `Err` is
invisible to `errors.Is` and `errors.AsType`. Many standard library errors
work this way. For example, `*fs.PathError` (from `os.Open`) keeps the file
path in a field and unwraps to the real cause.

When nothing goes wrong, return a plain `nil`, never a nil `*QueryError`
variable. That is the typed-nil gotcha from *Custom error types*.

## Your task

`exercise.go` has a `DeployError` type, a `Stage` type with a `Name` and a
`Run` function, and a sentinel `ErrUnhealthy`. Finish the error type and the
two functions.

### 1. Make DeployError a link in the chain

Change `Error` so it returns `deploy SERVICE: stage STAGE: ` followed by the
message of `Err`. Then add an `Unwrap() error` method on `*DeployError` that
returns `Err`.

```go
err := &DeployError{Service: "api", Stage: "build", Err: errors.New("compile main.go: exit status 1")}
err.Error()
// => "deploy api: stage build: compile main.go: exit status 1"
errors.Unwrap(err)
// => the "compile main.go: exit status 1" error
```

### 2. Run the pipeline

`Deploy(service, stages)` runs each stage's `Run` function, in order. At the
first stage that returns an error, stop (later stages must not run) and return
a `*DeployError` with the service name, the stage's `Name`, and the stage's
error as `Err`. Don't wrap the stage's error in anything else first. If every
stage succeeds, or there are no stages, return `nil`.

```go
Deploy("api", stages) // "test" stage returns errors.New("2 tests failed")
// => error "deploy api: stage test: 2 tests failed"
//    "build" and "test" ran, "rollout" did not
```

### 3. Decide whether to roll back

`ShouldRollback(err)` returns `true` when the deploy left a broken version
running:

- the chain contains `ErrUnhealthy` (the new version is live but failing its
  health checks), or
- the chain contains a `*DeployError` whose `Stage` is `"rollout"` (some
  servers may already run the new version).

Everything else returns `false`, including `nil`: a failed build or test never
reached production. The error may be wrapped in extra `fmt.Errorf` layers, and
`ErrUnhealthy` may sit inside a `DeployError`, so this relies on your `Unwrap`
from task 1.

```go
ShouldRollback(fmt.Errorf("release v1.4.2: %w",
	&DeployError{Service: "api", Stage: "verify", Err: fmt.Errorf("GET /healthz returned 503: %w", ErrUnhealthy)}))
// => true

ShouldRollback(&DeployError{Service: "api", Stage: "test", Err: errors.New("2 tests failed")})
// => false
```

## Learn more

- [Package errors](https://pkg.go.dev/errors)
- [Go blog: Working with Errors in Go 1.13](https://go.dev/blog/go1.13-errors)
- [fs.PathError, a standard library error with Unwrap](https://pkg.go.dev/io/fs#PathError)
