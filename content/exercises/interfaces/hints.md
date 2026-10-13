## Task 1: build the line with Sprintf

`fmt.Sprintf` builds the whole line in one go. The format string needs three
`%s` verbs: one for the channel (or address), one for the service and one for
the message. The receiver (`s` or `e`) holds the channel or address; the
parameter `a` holds `a.Service` and `a.Message`.

## Task 2: loop and append

`ns` is a slice, so `range` works on it. Each element is a `Notifier`, so you
can call `Notify` on it without knowing whether it is a `Slack` or an `Email`.
Start with an empty `[]string` and `append` each result.

## Task 3: copy the shape of Slack.Notify

A method that satisfies `Notifier` must look exactly like this, with your own
receiver:

```go
func (p PagerDuty) Notify(a Alert) string {
	return fmt.Sprintf("pagerduty %s: [%s] %s", p.ServiceKey, a.Service, a.Message)
}
```

Put it below the `PagerDuty` struct, and give it a doc comment.
