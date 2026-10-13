## Task 1: two small changes

Both bugs are on or just above the `selector :=` line. Swap `svc.Labels` for
the Service's selector map, and add an early return before you build the
selector.

## Task 1: the guard

`len` works on a `nil` map too, and returns 0:

```go
if len(svc.Spec.Selector) == 0 {
	return nil, nil
}
selector := labels.SelectorFromSet(svc.Spec.Selector).String()
```

## Task 2: build the selector, then List

Build the selector text from a one-entry map, list in `metav1.NamespaceAll`,
and loop over `.Items`. You no longer need `var pods []corev1.Pod` or the
`corev1` import:

```go
selector := labels.SelectorFromSet(labels.Set{"team": team}).String()
pods, err := cs.CoreV1().Pods(metav1.NamespaceAll).List(ctx, metav1.ListOptions{LabelSelector: selector})
if err != nil {
	return nil, err
}

counts := map[string]int{}
for _, pod := range pods.Items {
	// ... the counting code that was already there
}
```
