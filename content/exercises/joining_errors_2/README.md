# Reporting many errors: fix the firewall validator

A firewall rule validator that stops at the first problem makes you fix rules
one rerun at a time. This one has three bugs that make it less useful than it
should be. Find and fix them.

## Learn

Recap: `errors.Join(errs...)` combines errors into one, one message per line.
It skips `nil` errors and returns `nil` when there is nothing left to join.
`errors.Is` searches every error inside a joined error, including ones wrapped
with `%w`.

One new trap: wrapping `nil` with `%w` does **not** give you `nil`.
`fmt.Errorf` always returns a real error, and its message shows the mistake:

```go
err := fmt.Errorf("check firewall: %w", nil)
fmt.Println(err != nil) // true
fmt.Println(err)        // check firewall: %!w(<nil>)
```

So when you want to add context to a joined error, check that it is not `nil`
before you wrap it.

## Your task

Before a firewall change is applied, `ValidateAll` checks every inbound rule.
A rule can have these problems, each wrapping a sentinel error:

| Problem | Condition | Message |
|---|---|---|
| `ErrBadPort` | `Port` is below 1 or above 65535 | `rule "ssh": bad port: 70000` |
| `ErrBadProtocol` | `Protocol` is not `"tcp"` or `"udp"` | `rule "ssh": bad protocol: "icmp"` |
| `ErrOpenToWorld` | `Source` is `"0.0.0.0/0"` and `Port` is not 80 or 443 | `rule "ssh": open to the world: port 22` |

The message formats in `exercise.go` are already correct. Only the control
flow is wrong.

### 1. Report every problem with a rule

`ValidateRule` returns as soon as it finds one problem. Change it so it checks
all three conditions, in the order of the table, and joins every problem it
finds. It still returns `nil` for a good rule.

```go
ValidateRule(Rule{Name: "ssh", Port: 22, Protocol: "icmp", Source: "0.0.0.0/0"})
// => error with the message:
// rule "ssh": bad protocol: "icmp"
// rule "ssh": open to the world: port 22

ValidateRule(Rule{Name: "https", Port: 443, Protocol: "tcp", Source: "0.0.0.0/0"})
// => nil
```

### 2. Return nil when every rule is fine

`ValidateAll` returns an error even when every rule is valid. Fix it so it
returns `nil` when no rule has a problem. When there are problems, it keeps
returning the joined problems wrapped as `check firewall: <problems>`, in rule
order:

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

### 3. Find open rules inside a report

`NeedsReview` is meant to report whether a validation error contains
`ErrOpenToWorld` anywhere, so a security engineer can look at the change. It
compares with `==`, which only matches the bare sentinel, never one that is
wrapped or joined. Fix it.

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

## Learn more

- [errors.Join](https://pkg.go.dev/errors#Join)
- [errors.Is](https://pkg.go.dev/errors#Is)
- [fmt.Errorf](https://pkg.go.dev/fmt#Errorf)
