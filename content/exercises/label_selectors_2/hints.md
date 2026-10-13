## Task 1: Fix `PodsForService`

Both bugs are on or just above the `selector :=` line. Swap `svc.Labels` for
the Service's selector map, and add an early return before you build the
selector.

### Closer

The selector lives in `svc.Spec.Selector`. `len` works on a `nil` map too and
returns 0, so `len(svc.Spec.Selector) == 0` spots "no selector".

### Answer

```go
if len(svc.Spec.Selector) == 0 {
	return nil, nil
}
selector := labels.SelectorFromSet(svc.Spec.Selector).String()
```

An empty selector would match every Pod, so you must return before listing.

## Task 2: Finish `CountByEnv`

Build the selector text `team=<team>`, list Pods in every namespace with it,
and loop over the result's `.Items` instead of the empty `pods` slice.

### Closer

A one-entry map makes the selector:
`labels.SelectorFromSet(labels.Set{"team": team}).String()`. Pass
`metav1.NamespaceAll` to `cs.CoreV1().Pods(...)` and put the text in
`metav1.ListOptions{LabelSelector: ...}`. Delete `var pods []corev1.Pod` and
the now-unused `corev1` import.

### Answer

```go
selector := labels.SelectorFromSet(labels.Set{"team": team}).String()
pods, err := cs.CoreV1().Pods(metav1.NamespaceAll).List(ctx, metav1.ListOptions{LabelSelector: selector})
if err != nil {
	return nil, fmt.Errorf("list pods for team %s: %w", team, err)
}

counts := map[string]int{}
for _, pod := range pods.Items {
	// ... the counting code that was already there
}
```

The API server does the filtering, so you only receive the team's Pods.
