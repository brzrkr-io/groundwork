## Task 1: map to selector

You need two things: a selector built from the Service's map, and a
`labels.Set` made from the Pod's map. `labels.SelectorFromSet` gives you the
first. A conversion, `labels.Set(...)`, gives you the second.

```go
sel := labels.SelectorFromSet(labels.Set(svc.Spec.Selector))
return sel.Matches(labels.Set(pod.Labels))
```

## Task 2: loop and append

Start with `var names []string`, loop over `pods`, and append `pod.Name`
whenever `sel.Matches(labels.Set(pod.Labels))` is true. Looping over the
slice in order keeps the names in order.

## Task 3: parse, check, reuse

`labels.Parse` returns a selector and an error, like `strconv.Atoi`. Check the
error first, then hand the selector to the function you wrote in task 2:

```go
sel, err := labels.Parse(expr)
if err != nil {
	return nil, fmt.Errorf("invalid selector %q: %w", expr, err)
}
return MatchingPodNames(pods, sel), nil
```

Remember to add `"fmt"` to the imports.
