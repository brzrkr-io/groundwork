# Practice: config validator

Before a deploy tool ships a service, it checks the config. A good validator
reports **every** problem in one go, so nobody has to fix one typo, re-run,
and hit the next one. In this exercise you build one with everything from the
Errors module.

## Learn

This is a practice exercise, so it only recaps what you already know.

| Tool | What it does here |
|---|---|
| Custom error type | `*FieldError` says *which* field is wrong. |
| Sentinel errors | `ErrRequired`, `ErrOutOfRange` and `ErrInvalid` say *what kind* of problem it is. |
| `%w` | Adds detail to a sentinel and keeps it findable: `fmt.Errorf("%w: 0 is not between 1 and 50", ErrOutOfRange)`. |
| `Unwrap() error` | Lets `errors.Is` look inside a `*FieldError`. |
| `errors.Join` | Combines all problems into one error, one per line. `nil` when there are none. |
| `Unwrap() []error` | Gets the list back out of a joined error. |
| `errors.AsType` | Finds a `*FieldError` in an error. |

Getting the list out of a joined error, as in the last *Reporting many
errors* exercise:

```go
joined, ok := err.(interface{ Unwrap() []error })
if ok {
	for _, e := range joined.Unwrap() {
		// e is one of the errors passed to errors.Join
	}
}
```

### Sorted map keys

Go visits map keys in a random order. To report problems in the same order
every time, sort the keys first. The `maps` and `slices` packages do this in
one line:

```go
env := map[string]string{"PORT": "80", "DEBUG": "1"}
for _, key := range slices.Sorted(maps.Keys(env)) {
	fmt.Println(key) // DEBUG, then PORT
}
```

`maps.Keys` produces the keys one by one, and `slices.Sorted` collects them
into a sorted slice. Add `"maps"` and `"slices"` to your imports when you use
them.

## Your task

Your platform team wants `deployctl validate` to print every problem in a
service config. The types and sentinel errors are already in `exercise.go`.
You write the logic.

### 1. Make FieldError a proper error

Finish the two methods of `*FieldError`:

- `Error` returns the field, a colon and a space, then `e.Err.Error()`.
- `Unwrap` returns `e.Err`.

```go
err := &FieldError{Field: "port", Err: fmt.Errorf("%w: 0 is not between 1 and 65535", ErrOutOfRange)}
err.Error()                     // => "port: out of range: 0 is not between 1 and 65535"
errors.Is(err, ErrOutOfRange)   // => true
```

### 2. Validate name, port and replicas

Finish task 1 first: the tests compare the joined message, which uses your
`Error` method. In `Validate`, append a `*FieldError` to `errs` for each
problem. Check the fields in this order:

| Field | Problem | `Field` | `Err` |
|---|---|---|---|
| `Name` | empty | `"name"` | `ErrRequired` |
| `Port` | below 1 or above 65535 | `"port"` | `fmt.Errorf("%w: %d is not between 1 and 65535", ErrOutOfRange, c.Port)` |
| `Replicas` | below 1 or above 50 | `"replicas"` | `fmt.Errorf("%w: %d is not between 1 and 50", ErrOutOfRange, c.Replicas)` |

You'll need to add `"fmt"` to the imports. The existing
`return errors.Join(errs...)` already returns `nil` when `errs` is empty.

```go
Validate(Config{Name: "api", Image: "nginx:1.27", Port: 8080, Replicas: 3})
// => nil

Validate(Config{Image: "nginx:1.27", Port: 0, Replicas: 51})
// => an error whose message is these three lines:
// name: required
// port: out of range: 0 is not between 1 and 65535
// replicas: out of range: 51 is not between 1 and 50
```

### 3. Validate the image and the environment variables

Add two more checks to `Validate`. The image check goes **between** the name
and port checks. The env check goes **last**.

**Image.** An image must have a tag after a colon, as in `nginx:1.27`.

- If `Image` is empty, add `Field: "image"` with `Err: ErrRequired`.
- Otherwise, if it has no `:` or nothing after the `:`, add `Field: "image"`
  with `fmt.Errorf("%w: %q has no tag", ErrInvalid, c.Image)`.
  `strings.Cut(c.Image, ":")` helps here.

**Env.** First finish `ValidEnvName`. It returns `true` only when the name is
not empty and every character is `A`–`Z`, `0`–`9` or `_`. Then, in
`Validate`, go through the `Env` keys **in sorted order**. For each bad name,
add `Field: "env." + key` with
`fmt.Errorf("%w: use only A-Z, 0-9 and _", ErrInvalid)`.

```go
Validate(Config{Image: "redis", Port: -1, Replicas: 0, Env: map[string]string{"api key": "secret"}})
// name: required
// image: invalid: "redis" has no tag
// port: out of range: -1 is not between 1 and 65535
// replicas: out of range: 0 is not between 1 and 50
// env.api key: invalid: use only A-Z, 0-9 and _
```

An empty key gives the field `"env."`.

### 4. List the fields with problems

`FieldNames(err)` returns the `Field` of every `*FieldError` in `err`, in
order.

- If `err` is a joined error (it has `Unwrap() []error`), look at each error
  inside it. Otherwise, look at `err` itself.
- Use `errors.AsType[*FieldError]` on each one. Skip errors that aren't
  field errors.
- For `nil`, return `nil`.

```go
FieldNames(Validate(Config{Name: "api", Image: "nginx", Port: 8080, Replicas: 0}))
// => []string{"image", "replicas"}

FieldNames(&FieldError{Field: "replicas", Err: ErrOutOfRange})
// => []string{"replicas"}
```

### 5. Filter by category

`FieldsWith(err, kind)` works like `FieldNames`, but keeps only the field
errors where `errors.Is(fieldErr, kind)` is true. The deploy tool uses it to
print "missing: name, image" separately from other problems.

```go
FieldsWith(Validate(Config{Image: "nginx:1.27", Port: 80, Replicas: 99}), ErrRequired)
// => []string{"name"}

FieldsWith(Validate(Config{Image: "nginx:1.27", Port: 80, Replicas: 99}), ErrOutOfRange)
// => []string{"replicas"}
```

Tasks 4 and 5 share most of their code. You can move the shared part into a
small helper function.

## Learn more

- [Go blog: Working with errors in Go 1.13](https://go.dev/blog/go1.13-errors)
- [errors package](https://pkg.go.dev/errors)
- [maps.Keys](https://pkg.go.dev/maps#Keys) and [slices.Sorted](https://pkg.go.dev/slices#Sorted)
- [Kubernetes: Environment variables for containers](https://kubernetes.io/docs/tasks/inject-data-application/define-environment-variable-container/)
