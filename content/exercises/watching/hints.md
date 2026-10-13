## Task 1: is it a Pod?

Try the type assertion first. If it fails, the event has no Pod, so the type
on its own is the answer:

```go
pod, ok := ev.Object.(*corev1.Pod)
if !ok {
	return string(ev.Type)
}
```

Then build the rest with `fmt.Sprintf` and the verbs `%s %s/%s`.

## Task 2: range, then break

`range` over the channel ends by itself when the channel is closed. Count
what you have collected and `break` once you reach `n`:

```go
defer w.Stop()
var lines []string
for ev := range w.ResultChan() {
	lines = append(lines, Describe(ev))
	// stop once len(lines) == n
}
return lines
```

## Task 3: start, defer, loop

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
		// return nil when !ok, otherwise call handle(ev)
	}
}
```
