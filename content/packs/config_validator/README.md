# Practice: config validator

Before a deploy tool ships a service, it checks the config. A good validator
reports **every** problem in one go, so nobody fixes one typo, re-runs, and
hits the next one.

**New here:** nothing; practice from the Errors module, plus sorted map keys
from *Sorting*.

## Learn

This is a practice exercise, so it only recaps what you already know.

### The tools you'll use

| Tool | What it does here |
|---|---|
| Custom error type | `*FieldError` says *which* field is wrong. |
| Sentinel errors | `ErrRequired`, `ErrOutOfRange`, `ErrInvalid` say *what kind* of problem it is. |
| `%w` in `fmt.Errorf` | Adds detail to a sentinel and keeps it findable with `errors.Is`. |
| `Unwrap() error` | Lets `errors.Is` look inside one wrapped error. |
| `errors.Join` | Combines many errors into one, one per line. Returns `nil` when there are none. |
| `Unwrap() []error` | Gets the list back out of a joined error. |
| `errors.AsType` | Finds an error of a given type, such as `*FieldError`. |

Adding detail to a sentinel, as in *Wrapping errors*:

```go
var ErrTooBig = errors.New("too big")

err := fmt.Errorf("%w: %d MB is over the limit", ErrTooBig, 900)
err.Error()               // "too big: 900 MB is over the limit"
errors.Is(err, ErrTooBig) // true
```

The sentinel's own message comes first, because `%w` sits at the start of
the format string.

> **Why is it written like this?** `%w` is required (not `%v`) if you want
> `errors.Is` to find the sentinel. Putting it first, then `: detail`, is
> convention. The wording after the colon is the author's choice.

### Opening a joined error

As in *Reporting many errors*:

```go
joined, ok := err.(interface{ Unwrap() []error })
if ok {
	for _, e := range joined.Unwrap() {
		// e is one of the errors passed to errors.Join
	}
}
```

If `ok` is `false`, `err` is a single error, not a list.

### Sorted map keys

Go visits map keys in a random order. To report problems in the same order
every time, sort the keys first, as in *Sorting*:

```go
limits := map[string]int{"memory": 512, "cpu": 2}
for _, key := range slices.Sorted(maps.Keys(limits)) {
	fmt.Println(key) // cpu, then memory
}
```

`maps.Keys` hands out the keys one by one; `slices.Sorted` collects them into
a sorted slice.

> **Why is it written like this?** The random order is Go's rule, chosen on
> purpose so nobody relies on it. You must import `"maps"` and `"slices"` to
> use them.

## Your task

Your platform team wants `deployctl validate` to print every problem in a
service config. The types and sentinel errors are already in `exercise.go`.
You write the logic.

### 1. Make FieldError a proper error

**Goal:** finish the `Error` and `Unwrap` methods of `*FieldError`.

- `Error` returns `e.Field`, then a colon and a space, then the message of
  `e.Err`.
- `Unwrap` returns the wrapped error, `e.Err`.

```go
err := &FieldError{Field: "port", Err: fmt.Errorf("%w: 0 is not between 1 and 65535", ErrOutOfRange)}
err.Error()                   // => "port: out of range: 0 is not between 1 and 65535"
errors.Is(err, ErrOutOfRange) // => true
```

**Think it through:**
- `e.Err` is an `error`, not a string. Which method turns an error into its
  message?
- Which part is fixed text, and which parts come from the struct's fields?

### 2. Validate name, port and replicas

**Goal:** in `Validate`, append a `*FieldError` to `errs` for each problem
with `Name`, `Port` and `Replicas`.

Check in this order. Each row gives the `Field` value and the error message
the joined report must show:

| Field | Problem | `Field` | Message after `<field>: ` |
|---|---|---|---|
| `Name` | empty `""` | `"name"` | `required` (just `ErrRequired`) |
| `Port` | below 1 or above 65535 | `"port"` | `out of range: <port> is not between 1 and 65535` |
| `Replicas` | below 1 or above 50 | `"replicas"` | `out of range: <replicas> is not between 1 and 50` |

`<port>` and `<replicas>` are the actual numbers from the config. The two
range errors must wrap `ErrOutOfRange`, so build them with `fmt.Errorf` and
`%w` (add `"fmt"` to the imports). 1, 65535 and 50 themselves are valid.

```go
Validate(Config{Name: "api", Image: "nginx:1.27", Port: 8080, Replicas: 3})
// => nil

Validate(Config{Image: "nginx:1.27", Port: 0, Replicas: 51})
// => an error whose message is these three lines:
// name: required
// port: out of range: 0 is not between 1 and 65535
// replicas: out of range: 51 is not between 1 and 50
```

The existing `return errors.Join(errs...)` already returns `nil` when `errs`
is empty. Finish task 1 first: the joined message uses your `Error` method.

**Think it through:**
- Each rule is one `if` and one `append`. What does `append` need back?
- "Below 1 **or** above 65535": which operator says "or"?
- Which value goes in the `%d` spot: the limit, or the number from `c`?

### 3. Validate the image and the environment variables

**Goal:** add an image check **between** name and port, and an env check
**last**; finish `ValidEnvName` too.

**Image.** An image needs a tag after a colon, as in `nginx:1.27`.

- Empty `Image`: `Field: "image"`, error `ErrRequired` → `image: required`.
- Otherwise, no `:` or nothing after the `:`: `Field: "image"`, wrapping
  `ErrInvalid` → `image: invalid: "nginx" has no tag`. The image is shown in
  double quotes (`%q` does that). Only one image problem is reported, never
  both.

**Env names.** `ValidEnvName(name)` returns `true` only when `name` is not
empty and every character is `A`–`Z`, `0`–`9` or `_`.

**Env check.** In `Validate`, go through the `Env` keys **in sorted order**.
For each bad name, add `Field: "env." + key`, wrapping `ErrInvalid` →
`env.db_host: invalid: use only A-Z, 0-9 and _`. An empty key gives the field
`"env."`.

```go
Validate(Config{Image: "redis", Port: -1, Replicas: 0, Env: map[string]string{"api key": "secret"}})
// name: required
// image: invalid: "redis" has no tag
// port: out of range: -1 is not between 1 and 65535
// replicas: out of range: 0 is not between 1 and 50
// env.api key: invalid: use only A-Z, 0-9 and _
```

`"nginx:"` (colon, empty tag) also has no tag.

**Think it through:**
- `strings.Cut` (from *The strings package*) splits at the first `:` and
  tells you whether it found one. What do you check in its results?
- In `ValidEnvName`, loop over the characters. When can you return `false`
  early?
- Why must the keys be sorted before the loop, not after?

### 4. List the fields with problems

**Goal:** `FieldNames(err)` returns the `Field` of every `*FieldError` in
`err`, in order.

- If `err` is a joined error (it has `Unwrap() []error`), look at each error
  inside it. Otherwise, look at `err` itself.
- Skip errors that aren't field errors.
- For `nil`, return `nil`.

```go
FieldNames(Validate(Config{Name: "api", Image: "nginx", Port: 8080, Replicas: 0}))
// => []string{"image", "replicas"}

FieldNames(&FieldError{Field: "replicas", Err: ErrOutOfRange})
// => []string{"replicas"}

FieldNames(errors.Join(errors.New("disk full"), &FieldError{Field: "image", Err: ErrRequired}))
// => []string{"image"}
```

**Think it through:**
- Start with a list holding just `err`. When should you replace it with the
  joined list?
- `errors.AsType[*FieldError](e)` gives you a value and an `ok`. What do you
  do when `ok` is `false`?

### 5. Filter by category

**Goal:** `FieldsWith(err, kind)` works like `FieldNames`, but keeps only the
fields whose error matches `kind` according to `errors.Is`.

The deploy tool uses it to print "missing: name, image" apart from other
problems. A `nil` error gives `nil`.

```go
cfg := Config{Image: "nginx:1.27", Port: 80, Replicas: 99}
FieldsWith(Validate(cfg), ErrRequired)   // => []string{"name"}
FieldsWith(Validate(cfg), ErrOutOfRange) // => []string{"replicas"}
```

**Think it through:**
- Tasks 4 and 5 share most of their steps. Could a small helper return the
  `*FieldError` values, so both functions use it?
- Pass the `*FieldError` itself to `errors.Is`. Why does that work? (Look at
  task 1.)

## Learn more

- [Go blog: Working with errors in Go 1.13](https://go.dev/blog/go1.13-errors)
- [errors package](https://pkg.go.dev/errors)
- [maps.Keys](https://pkg.go.dev/maps#Keys) and [slices.Sorted](https://pkg.go.dev/slices#Sorted)
- [Kubernetes: Environment variables for containers](https://kubernetes.io/docs/tasks/inject-data-application/define-environment-variable-container/)
