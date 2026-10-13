# Wrapping errors

When a deploy fails at 3 a.m., you want a message that says *where* it
failed, and code that can still tell *what* failed. Wrapping errors gives you
both.

## Learn

### Context with %v loses the original

In *Errors are values* you added context with `fmt.Errorf` and `%v`:

```go
return fmt.Errorf("load config: %v", err)
```

The message reads well, but the result is a brand-new error that holds only
text. Code further up can no longer ask "was this a not-found error?" without
searching the string, and string matching breaks as soon as someone rewords a
message.

### Wrapping with %w

Use `%w` instead of `%v`:

```go
err := fmt.Errorf("load config: %w", inner)
```

The message is exactly the same as with `%v`. The difference is that the new
error keeps `inner` inside it. We say the new error *wraps* `inner`. Each layer
of your program can wrap again, building an *error chain*:

```text
load config: config "web": not found     <- what the caller gets
  wraps: config "web": not found
    wraps: not found                     <- ErrNotFound
```

`errors.Unwrap(err)` returns the next error in the chain, or `nil` at the
end. You will rarely call it yourself: the two functions below walk the chain
for you. Put one `%w` in each `fmt.Errorf` call.

### Sentinel errors and errors.Is

A *sentinel error* is a package-level variable that callers can check for. By
convention its name starts with `Err`:

```go
var ErrNotFound = errors.New("not found")
```

Comparing with `==` only looks at the outermost error, so it stops working
once the error is wrapped. `errors.Is` walks the whole chain:

```go
err := fmt.Errorf("config %q: %w", "web", ErrNotFound)

err == ErrNotFound          // false: err is the wrapper
errors.Is(err, ErrNotFound) // true: ErrNotFound is in the chain
```

### errors.AsType: find an error by its type

Custom error types carry fields, like `ParseError` in this exercise, which
records the bad line number. A type assertion `err.(*ParseError)`, as in
*Custom error types*, also only looks at the outermost error. To search the
whole chain for a type, use `errors.AsType`:

```go
if perr, ok := errors.AsType[*ParseError](err); ok {
	fmt.Println("bad line:", perr.Line)
}
```

The type in square brackets is a type argument, the same syntax you used to
call generic functions. `errors.AsType` returns the first error in the chain
with that type, plus `true`. If there is none, it returns `nil, false`.

`errors.AsType` arrived in Go 1.26. Older code does the same job with
`errors.As` and a variable to fill in. You will see this form a lot:

```go
var perr *ParseError
if errors.As(err, &perr) {
	fmt.Println("bad line:", perr.Line)
}
```

| You want to know… | Use |
|---|---|
| Is this exact error value anywhere in the chain? | `errors.Is(err, ErrNotFound)` |
| Is an error of this type in the chain? I need its fields. | `errors.AsType[*ParseError](err)` |
| Is the outermost error this value? (rarely what you want) | `err == ErrNotFound` |

### Layers

Real programs pass errors up through layers. Each lower layer wraps the error
with its own context. The top layer decides what to do:

```text
store    Get         wraps ErrNotFound with the config name
service  LoadConfig  wraps that with "load config: "
handler  Respond     uses errors.Is / errors.AsType to pick a response
```

## Your task

Your team runs a small config service. Configs live in a `Store` as text in
`key=value` lines. Finish the three layers. `ErrNotFound`, `ParseError` and
`parseLines` are already written for you.

### 1. Wrap ErrNotFound in the store

`(*Store).Get` looks up a config by name. When the name is missing, return
`""` and an error that wraps `ErrNotFound` and says which config was
missing, with the name in quotes (`%q`):

```go
store.Get("cache")
// => "", error `config "cache": not found`
// errors.Is(err, ErrNotFound) == true
```

### 2. Wrap errors in LoadConfig

`LoadConfig(s, name)` should:

1. call `s.Get(name)` to get the text,
2. pass the text to `parseLines`, which returns the values as a map,
3. return the map and `nil`.

The starter already makes both calls, but it throws the errors away with `_`.
If either step fails, return `nil` and that error wrapped as
`load config: <error>`. `parseLines` fails with a `*ParseError` whose message
is `line N: expected key=value, got "TEXT"`.

```go
LoadConfig(store, "web")    // web is "# web server\nport=8080\nreplicas=3"
// => map[port:8080 replicas:3], nil

LoadConfig(store, "cache")  // no such config
// => nil, error `load config: config "cache": not found`

LoadConfig(store, "worker") // worker is "queue=jobs\nreplicas 3"
// => nil, error `load config: line 2: expected key=value, got "replicas 3"`
```

The tests also check that `errors.Is(err, ErrNotFound)` and
`errors.AsType[*ParseError](err)` still find the original errors, so use
`%w`.

### 3. Pick a response with errors.Is and errors.AsType

`Respond(err)` turns an error into a status line. Check in this order:

| Error | Return |
|---|---|
| `nil` | `"200 OK"` |
| chain contains `ErrNotFound` | `"404 not found"` |
| chain contains a `*ParseError` | `"400 bad config on line N"`, where N is its `Line` |
| anything else | `"500 internal error"` |

```go
Respond(fmt.Errorf("load config: %w", &ParseError{Line: 3, Text: "debug"}))
// => "400 bad config on line 3"
```

## Learn more

- [Go blog: Working with Errors in Go 1.13](https://go.dev/blog/go1.13-errors)
- [Package errors](https://pkg.go.dev/errors)
- [Go by Example: Errors](https://gobyexample.com/errors)
