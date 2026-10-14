## Task 1: Read the desired state

Read `cm.Data["replicas"]` first: is it empty, and does it parse as a
non-negative whole number? Then check `cm.Data["image"]` the same way for
emptiness. Each problem returns a `fmt.Errorf` that wraps a sentinel with `%w`.

### Closer

`strconv.ParseInt(raw, 10, 32)` parses base 10 and rejects anything too big
for `int32`. A parse error *or* a negative result means `ErrInvalidReplicas`:

```go
if raw == "" {
	return Desired{}, fmt.Errorf("%w: %q", ErrMissingKey, "replicas")
}
n, err := strconv.ParseInt(raw, 10, 32)
```

### Answer

```go
raw := cm.Data["replicas"]
if raw == "" {
	return Desired{}, fmt.Errorf("%w: %q", ErrMissingKey, "replicas")
}
n, err := strconv.ParseInt(raw, 10, 32)
if err != nil || n < 0 {
	return Desired{}, fmt.Errorf("%w: %q", ErrInvalidReplicas, raw)
}
image := cm.Data["image"]
if image == "" {
	return Desired{}, fmt.Errorf("%w: %q", ErrMissingKey, "image")
}
return Desired{Replicas: int32(n), Image: image}, nil
```

`ParseInt` returns an `int64`, so convert it with `int32(n)` once you know it fits.

## Task 2: List the differences

Build the slice step by step: compare replicas, then the first container's
image, and `append` a message for each one that differs. A `nil` replicas
pointer counts as 1.

### Closer

Start with `var changes []string`. `ptr.Deref(current.Spec.Replicas, 1)` gives
the replica count with the default filled in. Format with
`fmt.Sprintf("replicas: %d -> %d", have, desired.Replicas)`, and `%s` for images.

### Answer

```go
var changes []string
if have := ptr.Deref(current.Spec.Replicas, 1); have != desired.Replicas {
	changes = append(changes, fmt.Sprintf("replicas: %d -> %d", have, desired.Replicas))
}
if have := current.Spec.Template.Spec.Containers[0].Image; have != desired.Image {
	changes = append(changes, fmt.Sprintf("image: %s -> %s", have, desired.Image))
}
return changes
```

When nothing differs, `changes` stays `nil`, which counts as empty.

## Task 3: Apply the desired state

Make a deep copy first and change only the copy. Track a `changed` flag that
becomes `true` whenever you set a field.

### Closer

```go
updated := current.DeepCopy()
changed := false
if ptr.Deref(updated.Spec.Replicas, 1) != desired.Replicas {
	updated.Spec.Replicas = new(desired.Replicas)
	changed = true
}
```

Do the same for the image, through `updated`, never `current`.

### Answer

```go
updated := current.DeepCopy()
changed := false
if ptr.Deref(updated.Spec.Replicas, 1) != desired.Replicas {
	updated.Spec.Replicas = new(desired.Replicas)
	changed = true
}
c := &updated.Spec.Template.Spec.Containers[0]
if c.Image != desired.Image {
	c.Image = desired.Image
	changed = true
}
return updated, changed
```

Taking `&...Containers[0]` gives a pointer into the copy, so the assignment changes the copy itself.
