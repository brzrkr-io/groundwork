# Interfaces

When something breaks, the on-call bot tells people through Slack, email and
PagerDuty. An *interface* lets one piece of code send to all of them without
caring which is which.

## Learn

### Describing what a value can do

A struct describes what data a value *holds*. An interface describes what a
value can *do*. It is a list of method signatures:

```go
type Notifier interface {
	Notify(a Alert) string
}
```

Read it as: "a `Notifier` is anything that has a method called `Notify` that
takes an `Alert` and returns a `string`."

### Satisfying an interface

Go has no `implements` keyword. A type *satisfies* an interface as soon as it
has every method the interface lists, with exactly the same name, parameters
and results. Nobody has to declare it.

```go
type Webhook struct {
	URL string
}

func (w Webhook) Notify(a Alert) string {
	return "webhook " + w.URL + ": " + a.Message
}
```

`Webhook` has a `Notify(a Alert) string` method, so `Webhook` is a `Notifier`.

### Interface values

A variable whose type is an interface can hold any value that satisfies it.
Calling a method on it runs the method of the value stored inside:

```go
var n Notifier = Webhook{URL: "https://hooks.example.com"}
fmt.Println(n.Notify(Alert{Service: "api", Message: "disk full"}))
// webhook https://hooks.example.com: disk full
```

If a type is missing the method, the compiler refuses to store it:

```go
var n Notifier = Alert{}
// error: Alert does not implement Notifier (missing method Notify)
```

### Slices of an interface type

Different types that satisfy the same interface can sit in one slice:

```go
notifiers := []Notifier{
	Slack{Channel: "ops"},
	Email{To: "oncall@example.com"},
}
for _, n := range notifiers {
	fmt.Println(n.Notify(alert))
}
```

The loop doesn't know or care which type each element is. To add Microsoft
Teams later, you write one new type with a `Notify` method. The loop never
changes.

### Small interfaces

Most Go interfaces have just one or two methods. The standard library is full
of them, and you'll meet these two again soon:

| Interface | Method | Satisfied by |
|---|---|---|
| `fmt.Stringer` | `String() string` | anything that can describe itself as text |
| `io.Writer` | `Write(p []byte) (n int, err error)` | files, network connections, buffers, HTTP responses |

### `any`

An interface with no methods is written `interface{}`, or by its short name
`any`. Every type satisfies it, because every type has at least zero methods.
So a variable of type `any` can hold any value at all. The test for task 3
uses it.

## Your task

Every line the on-call bot sends follows the pattern
`<where>: [<Service>] <Message>`. Finish the three notifiers and the function
that uses them.

### 1. Fill in Slack and Email

`Slack.Notify` returns `slack #<Channel>: [<Service>] <Message>`. `Channel`
holds just the name, so `Notify` adds the `#`:

```go
Slack{Channel: "ops"}.Notify(Alert{Service: "api", Message: "disk full"})
// => "slack #ops: [api] disk full"
```

`Email.Notify` returns `email <To>: [<Service>] <Message>`:

```go
Email{To: "oncall@example.com"}.Notify(Alert{Service: "api", Message: "disk full"})
// => "email oncall@example.com: [api] disk full"
```

There is exactly one space after the colon and one after the `]`.

### 2. Notify everyone

`NotifyAll(ns []Notifier, a Alert) []string` calls `Notify` on every notifier
in `ns`, in order, and returns the lines in a slice in the same order. With no
notifiers it returns an empty slice (`nil` is fine).

```go
NotifyAll([]Notifier{Slack{Channel: "ops"}, Email{To: "oncall@example.com"}},
	Alert{Service: "api", Message: "disk full"})
// => []string{"slack #ops: [api] disk full", "email oncall@example.com: [api] disk full"}
```

The test uses `Slack` and `Email`, so finish task 1 first.

### 3. Make PagerDuty a Notifier

`PagerDuty` has no methods yet. Write a `Notify` method for it, from scratch,
so that it satisfies `Notifier`. It returns
`pagerduty <ServiceKey>: [<Service>] <Message>`:

```go
PagerDuty{ServiceKey: "key-123"}.Notify(Alert{Service: "api", Message: "disk full"})
// => "pagerduty key-123: [api] disk full"
```

The test can't write `var n Notifier = PagerDuty{...}`: while the method is
missing, that line would stop the whole file from compiling. Instead it does
this:

```go
n, ok := any(PagerDuty{ServiceKey: "key-123"}).(Notifier)
```

It puts the value into an `any`, then asks "does this value satisfy
`Notifier`?" If it does, `ok` is `true` and `n` is the value as a `Notifier`.
This is a *type assertion*; a later exercise covers it properly.

## Learn more

- [A Tour of Go: Interfaces](https://go.dev/tour/methods/9)
- [A Tour of Go: Interfaces are implemented implicitly](https://go.dev/tour/methods/10)
- [Go by Example: Interfaces](https://gobyexample.com/interfaces)
- [Effective Go: Interfaces and other types](https://go.dev/doc/effective_go#interfaces_and_types)
