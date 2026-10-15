# Reporting many errors: fix the firewall validator

A firewall rule validator that stops at the first problem makes you fix rules
one rerun at a time. This one has three bugs. Find and fix them.

**New here:** wrapping `nil` with `%w` does not give `nil`.

## Learn

### Recap

From *Reporting many errors*: `errors.Join(errs...)` combines errors, one
message per line. It skips `nil`s and returns `nil` when nothing is left.
`errors.Is` searches every error inside, including ones wrapped with `%w`.

### The trap: wrapping nil

`fmt.Errorf` always returns a real error, even when the value for `%w` is
`nil`. Its message shows the mistake:

```go
var problems error // nil: nothing went wrong
err := fmt.Errorf("audit users: %w", problems)
fmt.Println(err != nil) // true
fmt.Println(err)        // audit users: %!w(<nil>)
```

So before you add context to an error that might be `nil`, check it first.

> **Why is it written like this?** This is how Go defines `fmt.Errorf`: it
> always builds a new error, whatever its arguments. `%!w(<nil>)` is `fmt`'s
> way of saying "this verb got a bad value".

## Your task

Before a firewall change is applied, `ValidateAll` checks every inbound rule.
A rule can have these problems, each wrapping a sentinel error:

| Problem | Condition | Message |
|---|---|---|
| `ErrBadPort` | `Port` is below 1 or above 65535 | `rule "ssh": bad port: 70000` |
| `ErrBadProtocol` | `Protocol` is neither `"tcp"` nor `"udp"` | `rule "ssh": bad protocol: "icmp"` |
| `ErrOpenToWorld` | `Source` is `"0.0.0.0/0"` and `Port` is not 80 or 443 | `rule "ssh": open to the world: port 22` |

The message formats in `exercise.go` are already correct. Only the control
flow is wrong.

### 1. Report every problem with a rule

**Goal:** make `ValidateRule` check all three conditions and join every problem it finds.

Check in the order of the table. It still returns `nil` for a good rule.
Right now it returns as soon as it finds one problem.

```go
ValidateRule(Rule{Name: "ssh", Port: 22, Protocol: "icmp", Source: "0.0.0.0/0"})
// => error with the message:
// rule "ssh": bad protocol: "icmp"
// rule "ssh": open to the world: port 22

ValidateRule(Rule{Name: "https", Port: 443, Protocol: "tcp", Source: "0.0.0.0/0"})
// => nil
```

**Think it through:**
- The function already declares `errs` and joins it at the end. Does anything ever go into it?
- Keep each message exactly as it is; change only what happens to it.

### 2. Return nil when every rule is fine

**Goal:** make `ValidateAll` return `nil` when no rule has a problem.

When there are problems, keep returning the joined problems, in rule order,
wrapped as `check firewall: <problems>`.

```go
ValidateAll([]Rule{
	{Name: "https", Port: 443, Protocol: "tcp", Source: "0.0.0.0/0"},
	{Name: "dns", Port: 53, Protocol: "udp", Source: "10.0.0.0/8"},
})
// => nil

ValidateAll([]Rule{
	{Name: "ssh", Port: 22, Protocol: "tcp", Source: "0.0.0.0/0"},
	{Name: "db", Port: 0, Protocol: "tcp", Source: "10.0.0.0/8"},
})
// => error with the message:
// check firewall: rule "ssh": open to the world: port 22
// rule "db": bad port: 0
```

**Think it through:**
- What does the joined result look like when every rule is fine?
- Look at the trap in Learn: what must you check before wrapping?

### 3. Find open rules inside a report

**Goal:** make `NeedsReview(err)` return `true` when `ErrOpenToWorld` is anywhere in `err`.

Return `false` otherwise, including for `nil`. Right now it compares with
`==`, which only matches the bare sentinel, never a wrapped or joined one.

```go
NeedsReview(ValidateAll([]Rule{
	{Name: "ssh", Port: 22, Protocol: "tcp", Source: "0.0.0.0/0"},
}))
// => true

NeedsReview(ValidateAll([]Rule{
	{Name: "db", Port: 0, Protocol: "tcp", Source: "10.0.0.0/8"},
}))
// => false

NeedsReview(nil)
// => false
```

**Think it through:**
- Which function from the recap searches inside wrapped and joined errors?

## Learn more

- [errors.Join](https://pkg.go.dev/errors#Join)
- [errors.Is](https://pkg.go.dev/errors#Is)
- [fmt.Errorf](https://pkg.go.dev/fmt#Errorf)
