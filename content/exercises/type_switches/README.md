# Type assertions and switches

An event bus carries many kinds of messages: deploys, scaling changes, alerts.
When they all arrive through one channel as plain values of type `any`, your
code has to ask "what is this, really?" before it can use it. Type assertions
and type switches are how Go asks that question.

## Learn

### `any` holds a value of any type

`any` is the empty interface. A variable of type `any` can hold an `int`, a
`string`, a struct, anything:

```go
var e any
e = 42
e = "hello"
e = DeployEvent{Service: "api", Version: "v1.4.2"}
```

The catch: while the value sits in an `any`, Go only knows it as "something".
You can't read `e.Service`, because `any` has no fields. You first have to get
the concrete value back out.

### Type assertions: `e.(T)`

A *type assertion* says "I expect `e` to hold a `T`; give me that `T`":

```go
d := e.(DeployEvent) // d has type DeployEvent
fmt.Println(d.Service)
```

If `e` holds something else, this single-result form **panics** and crashes
the program. Only use it when you are certain of the type.

### The comma-ok form

The safe form returns a second value, a `bool`, just like a map lookup:

```go
d, ok := e.(DeployEvent)
if !ok {
	// e was not a DeployEvent. d is the zero DeployEvent{}.
}
```

It never panics. When `ok` is `false`, the first value is the zero value of
the type. If `e` is `nil` (holds nothing at all), `ok` is `false` too.

### Type switches

When there are several possible types, a *type switch* checks them all in
one go:

```go
switch v := e.(type) {
case DeployEvent:
	// here v is a DeployEvent
	return "deploying " + v.Version
case int:
	// here v is an int
	return fmt.Sprintf("number %d", v)
default:
	// none of the cases matched; v has the same type as e (any)
	return "no idea"
}
```

- `e.(type)` only works inside a `switch`. The word `type` is literal.
- In each `case`, `v` has that case's type, so you can use its fields.
- `default` catches every type you did not list, including `nil`.

If you only need to know the type and not the value, leave out `v :=`:

```go
switch e.(type) {
case DeployEvent:
	deploys++
}
```

### Printing a type with `%T`

The `%T` verb prints the type of a value, which is handy in messages about
values you don't recognise:

| Value | `fmt.Sprintf("%T", value)` |
|---|---|
| `42` | `int` |
| `"api"` | `string` |
| `3.5` | `float64` |
| `DeployEvent{}` (in package `typeswitches`) | `typeswitches.DeployEvent` |
| `&DeployEvent{}` | `*typeswitches.DeployEvent` |

### Gotcha: pointers are a different type

`*DeployEvent` (a pointer, as in *Pointers*) and `DeployEvent` are two
different types. A pointer to a deploy event does **not** match
`case DeployEvent:`. It falls through to `default` unless you add a
`case *DeployEvent:` as well.

## Your task

Your deploy bot receives events as `any` values. Three event types are
defined in `exercise.go`: `DeployEvent`, `ScaleEvent` and `AlertEvent`.

### 1. Pick out alerts

Change `AsAlert` so it uses a comma-ok type assertion. If `e` holds an
`AlertEvent`, return it and `true`. For anything else (including `nil`),
return the zero `AlertEvent{}` and `false`.

```go
AsAlert(AlertEvent{Service: "db", Severity: "warning", Message: "slow queries"})
// => AlertEvent{Service: "db", Severity: "warning", Message: "slow queries"}, true

AsAlert(DeployEvent{Service: "api", Version: "v1.4.2"})
// => AlertEvent{}, false
```

### 2. Describe any event

Change `Describe` so it uses a type switch and returns one line of text:

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

### 3. Tally a batch of events

Change `Tally` so it counts how many events of each kind are in the slice.
Use the keys `"deploy"`, `"scale"`, `"alert"` and `"unknown"` (for every other
type, including `nil`). Only kinds that appear get a key, so an empty slice
gives an empty map.

```go
Tally([]any{
	DeployEvent{Service: "api", Version: "v1.4.2"},
	AlertEvent{Service: "db", Severity: "warning", Message: "slow queries"},
	DeployEvent{Service: "web", Version: "v2.0.0"},
	42,
})
// => map[string]int{"deploy": 2, "alert": 1, "unknown": 1}
```

## Learn more

- [A Tour of Go: Type assertions](https://go.dev/tour/methods/15)
- [A Tour of Go: Type switches](https://go.dev/tour/methods/16)
- [Go spec: Type switches](https://go.dev/ref/spec#Type_switches)
- [Go by Example: Interfaces](https://gobyexample.com/interfaces)
