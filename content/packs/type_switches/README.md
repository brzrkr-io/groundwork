# Type assertions and switches

An event bus carries many kinds of messages: deploys, scaling changes, alerts.
When they all arrive as plain values of type `any`, your code has to ask
"what is this, really?" before it can use them.

**New here:** type assertions `x.(T)`, the comma-ok form, type switches `switch v := x.(type)`, the `%T` verb. (`any` is from *Interfaces*.)

## Learn

### `any` holds a value of any type

`any` is the *empty interface*. A variable of type `any` can hold an `int`, a
`string`, a struct, anything:

```go
var box any
box = 42
box = "hello"
box = Pod{Name: "web-1", Ready: true}
```

The catch: while the value sits in an `any`, Go only knows it as
"something". You can't write `box.Name`, because `any` has no fields. You
first have to get the real value back out.

> **Why is it written like this?** `any` is a built-in name Go provides
> (Go requires it to mean "any type"). Older code writes `interface{}`; it
> means exactly the same. Using `any` is the modern convention.

### Type assertions: `x.(T)`

A *type assertion* says "I expect `x` to hold a `T`; give me that `T`":

```go
p := box.(Pod) // p has type Pod
fmt.Println(p.Name)
```

If `box` holds something else, this one-result form **panics** (crashes the
program). Only use it when you are certain of the type.

### The comma-ok form

The safe form gives back a second value, a `bool`, just like a map lookup:

```go
p, ok := box.(Pod)
if !ok {
	// box was not a Pod. p is the zero Pod{}.
}
```

It never panics. When `ok` is `false`, the first value is the *zero value* of
the type. If `box` is `nil` (holds nothing at all), `ok` is `false` too.

> **Why is it written like this?** The dot, the brackets and the type
> inside them are required syntax. The names `p` and `ok` are your choice;
> `ok` is the community convention for the `bool`.

### Type switches

When there are several possible types, a *type switch* checks them all at
once:

```go
switch v := box.(type) {
case Pod:
	// here v is a Pod, so v.Name works
	return "pod " + v.Name
case int:
	// here v is an int
	return fmt.Sprintf("number %d", v)
default:
	// no case matched; v is still an any
	return "no idea"
}
```

- `x.(type)` only works inside a `switch`. The word `type` is typed literally.
- In each `case`, `v` has that case's type, so you can use its fields.
- `default` catches every type you did not list, including `nil`.

If you only need to know the type, not the value, leave out `v :=`:

```go
switch box.(type) {
case Pod:
	pods++
}
```

> **Why is it written like this?** `switch`, `.(type)`, `case` and the colons
> are required by Go. Calling the variable `v` is just a common choice. gofmt
> lines `case` up with `switch`, not indented further.

### Printing a type with `%T`

The `%T` verb prints the type of a value:

| Value | `fmt.Sprintf("%T", value)` |
|---|---|
| `42` | `int` |
| `"api"` | `string` |
| `3.5` | `float64` |
| `DeployEvent{}` (in package `typeswitches`) | `typeswitches.DeployEvent` |
| `&DeployEvent{}` | `*typeswitches.DeployEvent` |

### Gotcha: pointers are a different type

`*DeployEvent` (a pointer, as in *Pointers*) and `DeployEvent` are two
different types. A pointer does **not** match `case DeployEvent:`; it falls
to `default`.

## Your task

Your deploy bot receives events as `any` values. Three event types are
defined in `exercise.go`: `DeployEvent`, `ScaleEvent` and `AlertEvent`. Look
at their fields before you start.

### 1. Pick out alerts

**Goal:** make `AsAlert(e)` return the `AlertEvent` inside `e` and `true`, or `AlertEvent{}` and `false`.

Use a comma-ok type assertion. Anything that is not an `AlertEvent`
(including `nil`) gives the zero `AlertEvent{}` and `false`.

```go
AsAlert(AlertEvent{Service: "db", Severity: "warning", Message: "slow queries"})
// => AlertEvent{Service: "db", Severity: "warning", Message: "slow queries"}, true

AsAlert(DeployEvent{Service: "api", Version: "v1.4.2"})
// => AlertEvent{}, false
```

**Think it through:**
- Which variable holds the `any` value? It's the parameter, not the function name.
- The comma-ok form already gives you both values the function must return.

### 2. Describe any event

**Goal:** make `Describe(e)` return one line of text, chosen by the type of `e`.

Use a type switch. The text comes from the event's fields; the words and
symbols around them are fixed text:

| Type of `e` | Format | Example |
|---|---|---|
| `DeployEvent` | `deploy <Service> <Version>` | `deploy api v1.4.2` |
| `ScaleEvent` | `scale <Service> <From> -> <To>` | `scale api 2 -> 5` |
| `AlertEvent` | `alert <Service> [<Severity>]: <Message>` | `alert api [critical]: disk full` |
| anything else | `unknown event <type>`, using `%T` | `unknown event int` |

The starter already returns the "anything else" line. Keep it as your
`default` case.

```go
Describe(ScaleEvent{Service: "api", From: 2, To: 5})
// => "scale api 2 -> 5"

Describe("deploy api")
// => "unknown event string"

Describe(&DeployEvent{Service: "api", Version: "v1.4.2"})
// => "unknown event *typeswitches.DeployEvent"  (a pointer is a different type)
```

**Think it through:**
- `From` and `To` are `int`s. Which `fmt` verb prints a number?
- Mind the single spaces, the `->`, the square brackets and the colon.

### 3. Tally a batch of events

**Goal:** make `Tally(events)` return a map counting how many events of each kind are in the slice.

Keys are `"deploy"`, `"scale"`, `"alert"` and `"unknown"` (every other type,
including `nil`). Only kinds that appear get a key, so an empty slice gives an
empty map. The starter already counts `"unknown"` in its `default` case.

```go
Tally([]any{
	DeployEvent{Service: "api", Version: "v1.4.2"},
	AlertEvent{Service: "db", Severity: "warning", Message: "slow queries"},
	DeployEvent{Service: "web", Version: "v2.0.0"},
	42,
})
// => map[string]int{"deploy": 2, "alert": 1, "unknown": 1}
```

**Think it through:**
- Look at how the `default` case adds one to `"unknown"`. The other cases work the same way.
- Do you need the value `v` here, or only the type?

## Learn more

- [A Tour of Go: Type assertions](https://go.dev/tour/methods/15)
- [A Tour of Go: Type switches](https://go.dev/tour/methods/16)
- [Go spec: Type switches](https://go.dev/ref/spec#Type_switches)
- [Go by Example: Interfaces](https://gobyexample.com/interfaces)
