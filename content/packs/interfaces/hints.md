## Task 1: Fill in Slack and Email

Each line is `<where>: [<Service>] <Message>`. The receiver (`s` or `e`) holds the "where"; the parameter `a` holds `a.Service` and `a.Message`.

### Closer

`fmt.Sprintf` builds the whole line in one go, with one `%s` per piece: `"slack #%s: [%s] %s"`. Watch the single spaces after `:` and `]`.

### Answer

```go
func (s Slack) Notify(a Alert) string {
	return fmt.Sprintf("slack #%s: [%s] %s", s.Channel, a.Service, a.Message)
}

func (e Email) Notify(a Alert) string {
	return fmt.Sprintf("email %s: [%s] %s", e.To, a.Service, a.Message)
}
```

The `%s` verbs are filled in order by the arguments after the format string.

## Task 2: Notify everyone

Loop over `ns` and collect each notifier's line in a new slice. Every element is a `Notifier`, so you can call `Notify` without knowing whether it's a `Slack` or an `Email`.

### Closer

Start with `var sent []string`, use `for _, n := range ns`, and `append` the result of `n.Notify(a)`.

### Answer

```go
var sent []string
for _, n := range ns {
	sent = append(sent, n.Notify(a))
}
return sent
```

Each call runs the `Notify` method of whatever concrete type `n` holds.

## Task 3: Make PagerDuty a Notifier

A type satisfies `Notifier` just by having a method with the same name, parameters and result. Copy the shape of `Slack.Notify`, with a `PagerDuty` receiver.

### Closer

Write the method below the `PagerDuty` struct, with a doc comment:

```go
func (p PagerDuty) Notify(a Alert) string {
	// build "pagerduty <ServiceKey>: [<Service>] <Message>"
}
```

### Answer

```go
// Notify formats a for PagerDuty, like "pagerduty key-123: [api] disk full".
func (p PagerDuty) Notify(a Alert) string {
	return fmt.Sprintf("pagerduty %s: [%s] %s", p.ServiceKey, a.Service, a.Message)
}
```

No `implements` keyword is needed: having the method is enough.
