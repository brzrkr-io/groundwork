# Wrapping errors

When a deploy fails at 3 a.m., you want a message that says *where* it failed,
and code that can still tell *what* failed. Wrapping errors gives you both.

**New here:** the `%w` verb, error chains, `errors.Is`, `errors.AsType`.

## Learn

### %v copies only the text

In *Errors are values* you added context with `fmt.Errorf` and `%v`:

```go
return fmt.Errorf("read cert: %v", err)
```

The message reads well, but the new error holds only text. Code higher up can
no longer ask "was this a not-found error?" except by searching the string,
and that breaks as soon as someone rewords a message.

### %w wraps the error

Use `%w` instead of `%v`:

```go
err := fmt.Errorf("read cert: %w", inner)
```

The printed message is the same as with `%v`. The difference: the new error
keeps `inner` inside it. We say it *wraps* `inner`.

Each layer of your program can wrap again. That builds an *error chain*:

```text
renew cert: read cert "edge": expired    <- what the caller gets
  wraps: read cert "edge": expired
    wraps: expired                       <- ErrExpired
```

> **Why is it written like this?** `%w` is a verb that only `fmt.Errorf`
> understands (Go requires it there; `Sprintf` just prints `%!w`). Use one `%w`
> per call (convention). The words before `%w` are your choice of context.

### Sentinel errors and errors.Is

A *sentinel error* is a package-level variable that callers can check for:

```go
var ErrExpired = errors.New("expired")
```

`==` only looks at the outermost error, so it stops working once the error is
wrapped. `errors.Is` walks the whole chain:

```go
err := fmt.Errorf("read cert %q: %w", "edge", ErrExpired)

err == ErrExpired          // false: err is the wrapper
errors.Is(err, ErrExpired) // true: ErrExpired is in the chain
```

> **Why is it written like this?** The `Err` prefix on sentinel names is
> convention. `var` (not `const`) is required: `errors.New` runs at start-up,
> and constants can't hold an error.

### errors.AsType: find an error by its type

Custom error types carry fields. This exercise's `ParseError`, for example,
records the bad line number. A type assertion such as `err.(*ParseError)`, from
*Custom error types*, also only looks at the outermost error. To search the
whole chain, use `errors.AsType`:

```go
if terr, ok := errors.AsType[*TimeoutError](err); ok {
	fmt.Println("waited:", terr.Seconds)
}
```

It returns the first error in the chain of that type, plus `true`. If there is
none, it returns `nil, false`.

> **Why is it written like this?** The type in square brackets is a *type
> argument*, the same syntax as calling a generic function (Go requires it).
> The `if x, ok := ...; ok` shape is convention. The name `terr` is your
> choice.

Older code (before Go 1.26) uses `errors.As` with a variable to fill in. You
will see it a lot:

```go
var terr *TimeoutError
if errors.As(err, &terr) { ... }
```

| You want to know… | Use |
|---|---|
| Is this exact error value anywhere in the chain? | `errors.Is(err, ErrExpired)` |
| Is an error of this type in the chain? I need its fields. | `errors.AsType[*TimeoutError](err)` |
| Is the outermost error this value? (rarely what you want) | `err == ErrExpired` |

`errors.Unwrap(err)` returns the next error in a chain, or `nil` at the end.
You rarely call it yourself: `errors.Is` and `errors.AsType` do it for you.

## Your task

Your team runs a small config service. Configs live in a `Store` as text in
`key=value` lines. Three layers pass errors up:

```text
store    Get         wraps ErrNotFound with the config name
service  LoadConfig  wraps that with "load config: "
handler  Respond     uses errors.Is / errors.AsType to pick a response
```

`ErrNotFound`, `ParseError` and `parseLines` are already written for you.

### 1. Wrap ErrNotFound in the store

**Goal:** make `(*Store).Get` return an error that names the missing config and wraps `ErrNotFound`.

When the name is missing, return `""` and an error with the message
`config "NAME": not found`. The name comes from the `name` parameter, in
quotes (`%q`). The `not found` part is the text of `ErrNotFound`.

```go
store.Get("cache")
// => "", error `config "cache": not found`
// errors.Is(err, ErrNotFound) == true
```

**Think it through:**
- Which fixed text is around the name, and which part comes from `ErrNotFound`?
- Which verb keeps `ErrNotFound` findable, and which verb adds the quotes?

### 2. Wrap errors in LoadConfig

**Goal:** make `LoadConfig` return `nil` and a `load config: ` error when either step fails.

`LoadConfig(s, name)` does two steps: `s.Get(name)` gets the text, then
`parseLines(text)` turns it into a map. The starter makes both calls but
throws the errors away with `_`.

- If either step fails, return `nil` and that step's error wrapped with
  `load config: ` in front.
- If both succeed, return the map and `nil`.
- `parseLines` fails with a `*ParseError`, message
  `line N: expected key=value, got "TEXT"`.

```go
LoadConfig(store, "web")    // web is "# web server\nport=8080\nreplicas=3"
// => map[port:8080 replicas:3], nil

LoadConfig(store, "cache")  // no such config
// => nil, error `load config: config "cache": not found`

LoadConfig(store, "worker") // worker is "queue=jobs\nreplicas 3"
// => nil, error `load config: line 2: expected key=value, got "replicas 3"`
```

The tests also check that `errors.Is(err, ErrNotFound)` and
`errors.AsType[*ParseError](err)` still find the original errors.

**Think it through:**
- Replace each `_` with a name, then check it right after its call.
- Should the step after a failure still run?

### 3. Pick a response with errors.Is and errors.AsType

**Goal:** make `Respond(err)` return the status line that matches what is in the error chain.

Check in this order:

| Error | Return |
|---|---|
| `nil` | `"200 OK"` |
| chain contains `ErrNotFound` | `"404 not found"` |
| chain contains a `*ParseError` | `"400 bad config on line N"`, where N is its `Line` |
| anything else | `"500 internal error"` |

```go
Respond(fmt.Errorf("reload: %w", &ParseError{Line: 3, Text: "debug"}))
// => "400 bad config on line 3"
```

**Think it through:**
- Which row needs a value check, and which needs a type check?
- For the 400 row, the number comes from a field of the error you found.

## Learn more

- [Go blog: Working with Errors in Go 1.13](https://go.dev/blog/go1.13-errors)
- [Package errors](https://pkg.go.dev/errors)
- [Go by Example: Errors](https://gobyexample.com/errors)
