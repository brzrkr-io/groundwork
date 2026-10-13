## Task 1: Describe an event

First find out whether `ev.Object` really is a Pod. If it isn't, the event
type on its own is the answer; if it is, add the Pod's namespace and name.

### Closer

A type assertion with `ok` tells you safely: `pod, ok := ev.Object.(*corev1.Pod)`.
`ev.Type` is a string type, so `string(ev.Type)` converts it. Build the full
line with `fmt.Sprintf` and the verbs `%s %s/%s`.

### Answer

```go
pod, ok := ev.Object.(*corev1.Pod)
if !ok {
	return string(ev.Type)
}
return fmt.Sprintf("%s %s/%s", ev.Type, pod.Namespace, pod.Name)
```

The `ok` form never panics, so non-Pod events such as `ERROR` fall into the
early return.

## Task 2: Collect a few events

`range` over `w.ResultChan()` ends by itself when the channel is closed.
Count what you have collected and stop once you reach `n`, and make sure
`w.Stop()` runs however you leave.

### Closer

`defer w.Stop()` at the top covers every return. Inside
`for ev := range w.ResultChan()`, append `Describe(ev)` and `break` when
`len(lines) == n`.

### Answer

```go
defer w.Stop()
var lines []string
for ev := range w.ResultChan() {
	lines = append(lines, Describe(ev))
	if len(lines) == n {
		break
	}
}
return lines
```

Checking after appending means you never read an extra event.

## Task 3: Watch Pods until cancelled

Start the watch with `cs.CoreV1().Pods(namespace).Watch`, stop it with
`defer`, then loop forever waiting for either the context to end or the next
event.

### Closer

A `select` inside `for { ... }` waits on both channels. The two-value receive
`ev, ok := <-w.ResultChan()` gives `ok == false` once the server closes it:

```go
for {
	select {
	case <-ctx.Done():
		// ...
	case ev, ok := <-w.ResultChan():
		// ...
	}
}
```

### Answer

```go
w, err := cs.CoreV1().Pods(namespace).Watch(ctx, metav1.ListOptions{})
if err != nil {
	return err
}
defer w.Stop()

for {
	select {
	case <-ctx.Done():
		return nil
	case ev, ok := <-w.ResultChan():
		if !ok {
			return nil // the server closed the watch
		}
		handle(ev)
	}
}
```

Without the `ok` check a closed channel would deliver empty events forever.
