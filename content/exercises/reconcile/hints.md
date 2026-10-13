## Task 1: check, parse, check

Read `cm.Data["replicas"]` and return the `ErrMissingKey` error if it is
`""`. Parse it with `strconv.ParseInt(raw, 10, 32)`. If that returns an error
*or* a negative number, return the `ErrInvalidReplicas` error. Then do the
same `""` check for `"image"`.

```go
if raw == "" {
	return Desired{}, fmt.Errorf("%w: %q", ErrMissingKey, "replicas")
}
```

## Task 2: build the slice step by step

Start with `var changes []string`. Compare
`ptr.Deref(current.Spec.Replicas, 1)` with `desired.Replicas` and `append`
a `fmt.Sprintf("replicas: %d -> %d", ...)` when they differ. Then do the
same for the image with `%s`. Return `changes`. When nothing differs it is
still `nil`, which counts as empty.

## Task 3: change the copy

```go
updated := current.DeepCopy()
changed := false
if ptr.Deref(updated.Spec.Replicas, 1) != desired.Replicas {
	updated.Spec.Replicas = new(desired.Replicas)
	changed = true
}
// Same idea for updated.Spec.Template.Spec.Containers[0].Image
return updated, changed
```

Always assign through `updated`, never through `current`.
