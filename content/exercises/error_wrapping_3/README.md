# Wrapping errors: triage a failed deploy

A deploy pipeline runs build, test and rollout stages. When one fails, the
on-call engineer needs two answers fast: which stage broke, and must the
previous version be restored? You will build an error type that answers both.

**New here:** the `Unwrap() error` method on your own error type.

## Learn

Recap from *Wrapping errors*: `fmt.Errorf("context: %w", err)` builds an error
chain. `errors.Is` searches it for a sentinel value; `errors.AsType[*T]`
searches it for a type and hands you its fields.

### How the chain is linked: Unwrap

How does `errors.Is` find the next error in a chain? It calls the error's
`Unwrap() error` method. The error from `fmt.Errorf` with `%w` has one. Your
own types can too:

```go
// QueryError reports a failed database query.
type QueryError struct {
	Query string
	Err   error // the underlying error
}

func (e *QueryError) Error() string {
	return fmt.Sprintf("query %q: %v", e.Query, e.Err)
}

func (e *QueryError) Unwrap() error {
	return e.Err
}
```

> **Why is it written like this?** The name `Unwrap` and the signature
> `() error` are required: `errors.Is` looks for exactly that method. Calling
> the field `Err` is convention. The receiver name `e` is your choice.

### What you get

A struct with an `Err` field and an `Unwrap` method gives you both:

- `errors.AsType[*QueryError](err)` finds it and gives you `Query`.
- `errors.Is(err, ErrConnRefused)` still finds the error inside `Err`.

Without `Unwrap`, the chain stops at your type. Many standard library errors
work this way: `*fs.PathError` (from `os.Open`) keeps the file path in a field
and unwraps to the real cause.

### Return a plain nil

When nothing goes wrong, return a plain `nil`, never a nil `*QueryError`
variable. That is the typed-nil gotcha from *Custom error types*.

## Your task

`exercise.go` has a `DeployError` type, a `Stage` type (a `Name` and a `Run`
function), and a sentinel `ErrUnhealthy`. Finish the error type and two
functions.

### 1. Make DeployError a link in the chain

**Goal:** fix `DeployError.Error`, and add an `Unwrap() error` method on `*DeployError` that returns `Err`.

`Error` must return `deploy SERVICE: stage STAGE: ` followed by the message of
`Err`. `SERVICE`, `STAGE` and the message come from the fields `Service`,
`Stage` and `Err`; `deploy `, `: stage ` and `: ` are fixed text.

```go
err := &DeployError{Service: "api", Stage: "build", Err: errors.New("compile main.go: exit status 1")}
err.Error()
// => "deploy api: stage build: compile main.go: exit status 1"
errors.Unwrap(err)
// => the "compile main.go: exit status 1" error
```

**Think it through:**
- `Error` already uses `fmt.Sprintf`; which pieces are missing from its format string?
- The TODO comment marks where the new method goes, outside `Error`.

### 2. Run the pipeline

**Goal:** make `Deploy(service, stages)` run each stage in order and return a `*DeployError` for the first failure.

- Call each stage's `Run` function, in order.
- At the first stage that returns an error, stop: later stages must not run.
- Return a `*DeployError` with `service`, that stage's `Name`, and the stage's
  error as `Err`. Don't wrap the stage's error in anything else first.
- If every stage succeeds, or there are no stages, return `nil`.

```go
Deploy("api", stages) // the "test" stage returns errors.New("2 tests failed")
// => error "deploy api: stage test: 2 tests failed"
//    "build" and "test" ran, "rollout" did not
```

**Think it through:**
- `Run` is a field holding a function; how do you call it?
- Where do you return so that no later stage runs?

### 3. Decide whether to roll back

**Goal:** make `ShouldRollback(err)` return `true` when the deploy left a broken version running.

Return `true` when either is true:

- the chain contains `ErrUnhealthy` (the new version is live but fails its
  health checks), or
- the chain contains a `*DeployError` whose `Stage` is `"rollout"` (some
  servers may already run the new version).

Everything else returns `false`, including `nil`. The error may be wrapped in
extra `fmt.Errorf` layers, and `ErrUnhealthy` may sit inside a `DeployError`.

```go
ShouldRollback(fmt.Errorf("release v1.4.2: %w",
	&DeployError{Service: "api", Stage: "verify", Err: fmt.Errorf("GET /healthz returned 503: %w", ErrUnhealthy)}))
// => true

ShouldRollback(&DeployError{Service: "api", Stage: "test", Err: errors.New("2 tests failed")})
// => false
```

**Think it through:**
- One rule looks for a value, the other for a type and then one of its fields.
- Finding `ErrUnhealthy` inside a `DeployError` relies on your `Unwrap` from task 1.

## Learn more

- [Package errors](https://pkg.go.dev/errors)
- [Go blog: Working with Errors in Go 1.13](https://go.dev/blog/go1.13-errors)
- [fs.PathError, a standard library error with Unwrap](https://pkg.go.dev/io/fs#PathError)
