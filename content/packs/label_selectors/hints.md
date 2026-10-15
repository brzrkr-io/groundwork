## Task 1: Check whether a Service selects a Pod

You need two things: a selector built from the Service's map
(`svc.Spec.Selector`), and a `labels.Set` made from the Pod's map
(`pod.Labels`). Then ask the selector whether it matches the set.

### Closer

`labels.SelectorFromSet(...)` turns a `labels.Set` into a selector. A plain
`map[string]string` becomes a `labels.Set` with a conversion:
`labels.Set(pod.Labels)`. The selector has a `Matches` method.

### Answer

```go
sel := labels.SelectorFromSet(labels.Set(svc.Spec.Selector))
return sel.Matches(labels.Set(pod.Labels))
```

`SelectorFromSet` requires every pair in the map, and extra Pod labels are
ignored.

## Task 2: List the names of matching Pods

Loop over `pods` and keep the name of each Pod that `sel` matches. A `range`
loop visits the Pods in order, so the names stay in order.

### Closer

Start with `var names []string`. Inside a `for _, pod := range pods` loop,
check `sel.Matches(labels.Set(pod.Labels))` and `append` `pod.Name`.

### Answer

```go
var names []string
for _, pod := range pods {
	if sel.Matches(labels.Set(pod.Labels)) {
		names = append(names, pod.Name)
	}
}
return names
```

A `nil` slice is fine when nothing matches.

## Task 3: Select Pods with a selector string

`labels.Parse` returns a selector and an error, like `strconv.Atoi`. Check the
error first, then hand the selector to the function you wrote in task 2.

### Closer

Wrap the error with `fmt.Errorf("invalid selector %q: %w", ...)`, and
remember to add `"fmt"` to the imports:

```go
sel, err := labels.Parse(expr)
if err != nil {
	// return nil and the wrapped error
}
```

### Answer

```go
sel, err := labels.Parse(expr)
if err != nil {
	return nil, fmt.Errorf("invalid selector %q: %w", expr, err)
}
return MatchingPodNames(pods, sel), nil
```

`%w` keeps the original parse error inside yours, so callers can still
inspect it.
