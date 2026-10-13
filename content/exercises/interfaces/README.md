# Interfaces

When something breaks, the on-call bot tells people through Slack, email and
PagerDuty. An *interface* lets one piece of code send to all of them without
caring which is which.

**New here:** `interface` types, satisfying an interface, slices of an interface type, `any`.

## Learn

### Describing what a value can do

A struct describes what data a value *holds*. An interface describes what a
value can *do*. It is a list of method signatures, with no bodies:

```go
type Pinger interface {
	Ping(host string) bool
}
```

Read it as: "a `Pinger` is anything with a method called `Ping` that takes a
`string` and returns a `bool`."

> **Why is it written like this?** `type Name interface { ... }` is required
> syntax. Each line inside is a method name, its parameters and its result,
> exactly as in a `func` line but without `func`, receiver or body. The name
> `Pinger` is the author's choice; ending one-method interfaces in `-er` is
> convention.

### Satisfying an interface

Go has no `implements` keyword. A type *satisfies* an interface as soon as it
has every method the interface lists, with exactly the same name, parameters
and results. Nobody has to declare it.

```go
type ICMP struct {
	Timeout int
}

func (c ICMP) Ping(host string) bool {
	return host != ""
}
```

`ICMP` has a `Ping(host string) bool` method, so `ICMP` is a `Pinger`.

### The parts of a method line

| Part | In the example | Required? |
|---|---|---|
| `func` | `func` | Go requires it |
| receiver in `( )` | `(c ICMP)` | required for a method; the name `c` is your choice (one short letter is convention) |
| method name | `Ping` | must match the interface exactly |
| parameters in `( )` | `(host string)` | types must match; the name `host` is your choice |
| result type | `bool` | must match the interface |
| body in `{ }` | `{ return ... }` | required; a `return` gives back a value, without types |

> **Why is it written like this?** The result type goes on the `func` line
> only. Inside the body, `return` is followed by just the value.

### Interface values

A variable whose type is an interface can hold any value that satisfies it.
Calling a method on it runs the method of the value stored inside:

```go
var p Pinger = ICMP{Timeout: 2}
fmt.Println(p.Ping("db-1")) // true
```

If a type is missing the method, the compiler refuses to store it:

```go
var p Pinger = Alert{}
// error: Alert does not implement Pinger (missing method Ping)
```

### Slices of an interface type

Different types that satisfy the same interface can sit in one slice, and one
loop can call the method on each:

```go
pingers := []Pinger{ICMP{Timeout: 2}, TCP{Port: 22}}
for _, p := range pingers {
	fmt.Println(p.Ping("db-1"))
}
```

The loop doesn't know or care which type each element is. To add a new kind
later, you write one new type with the method. The loop never changes.

### Small interfaces

Most Go interfaces have just one or two methods. You'll meet these two again
soon:

| Interface | Method | Satisfied by |
|---|---|---|
| `fmt.Stringer` | `String() string` | anything that can describe itself as text |
| `io.Writer` | `Write(p []byte) (n int, err error)` | files, network connections, buffers, HTTP responses |

### `any`

An interface with no methods is written `interface{}`, or by its short name
`any`. Every type satisfies it, so an `any` variable can hold any value. The
test for task 3 uses it.

## Your task

Every line the on-call bot sends follows the same pattern:

```
<where>: [<Service>] <Message>
```

`<where>` comes from the notifier (the receiver). `Service` and `Message` come
from the `Alert` parameter `a`. The colon, the square brackets and the two
single spaces (after `:` and after `]`) are fixed text.

### 1. Fill in Slack and Email

**Goal:** make `Slack.Notify` and `Email.Notify` return the full line for the alert `a`.

| Notifier | `<where>` part | Notes |
|---|---|---|
| `Slack` | `slack #<Channel>` | `Channel` is just the name, so `Notify` adds the `#` |
| `Email` | `email <To>` | |

```go
Slack{Channel: "ops"}.Notify(Alert{Service: "api", Message: "disk full"})
// => "slack #ops: [api] disk full"

Email{To: "oncall@example.com"}.Notify(Alert{Service: "api", Message: "disk full"})
// => "email oncall@example.com: [api] disk full"
```

`Slack.Notify` already returns the start of its line; finish it.

**Think it through:**
- Which value comes from the receiver, and which two come from `a`?
- Write the expected line out, then mark which bits are fixed text and which are values.

### 2. Notify everyone

**Goal:** make `NotifyAll` call `Notify` on every notifier in `ns`, in order, and return the lines in a slice.

The slice keeps the same order as `ns`. With no notifiers, return an empty
slice (`nil` is fine).

```go
NotifyAll([]Notifier{Slack{Channel: "ops"}, Email{To: "oncall@example.com"}},
	Alert{Service: "api", Message: "disk full"})
// => []string{"slack #ops: [api] disk full", "email oncall@example.com: [api] disk full"}
```

The test uses `Slack` and `Email`, so finish task 1 first.

**Think it through:**
- You need a slice to collect lines in. Where does it start, and how does it grow (as in *Slices*)?
- Each element of `ns` is a `Notifier`. What method can you call on it?

### 3. Make PagerDuty a Notifier

**Goal:** write a brand-new `Notify` method for `PagerDuty` so it satisfies `Notifier`.

It returns `pagerduty <ServiceKey>: [<Service>] <Message>`. Put it where the
`TODO(3)` comment is, below the `PagerDuty` struct.

```go
PagerDuty{ServiceKey: "key-123"}.Notify(Alert{Service: "api", Message: "disk full"})
// => "pagerduty key-123: [api] disk full"
```

The test can't write `var n Notifier = PagerDuty{...}`: while the method is
missing, that line would stop the whole file from compiling. Instead it asks
"does this value satisfy `Notifier`?" with a *type assertion*:

```go
n, ok := any(PagerDuty{ServiceKey: "key-123"}).(Notifier)
```

If it does, `ok` is `true`. A later exercise covers type assertions properly.

**Think it through:**
- Look at the `Notifier` interface. Which parts of your method line must match it exactly?
- Which parts of the line can you choose yourself? (See *The parts of a method line*.)
- The other two `Notify` methods are a model for the shape; only the receiver type and the `<where>` part differ.

## Learn more

- [A Tour of Go: Interfaces](https://go.dev/tour/methods/9)
- [A Tour of Go: Interfaces are implemented implicitly](https://go.dev/tour/methods/10)
- [Go by Example: Interfaces](https://gobyexample.com/interfaces)
- [Effective Go: Interfaces and other types](https://go.dev/doc/effective_go#interfaces_and_types)
