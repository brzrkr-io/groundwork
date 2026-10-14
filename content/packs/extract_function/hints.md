## Task 1: Extract validate

The five `if` blocks under `// Check the request.` need only `r`, and each one ends by returning an error. That's the new function's parameter and result.

### Closer

In `validate`, each `return nil, errors.New(...)` becomes `return errors.New(...)`, and the function ends with `return nil`. In `Plan`, replace the block with `if err := validate(r); err != nil { return nil, err }`.

### Answer

```go
func Plan(r Request) ([]string, error) {
	if err := validate(r); err != nil {
		return nil, err
	}
	// ... the rest of Plan, unchanged
}

// validate returns the first problem with r, or nil if it can be deployed.
func validate(r Request) error {
	if r.Service == "" {
		return errors.New("service is required")
	}
	// ... the other four checks, each returning just the error
	return nil
}
```

## Task 2: Extract imageRef

The block reads `r.Registry`, `r.Service` and `r.Tag`, and the rest of `Plan` only needs `image`. So `imageRef` takes `r` and returns that string.

### Closer

Move `registry := ...`, the `if` and the line that builds the reference into `imageRef`, ending with `return registry + "/" + ...`. In `Plan`, call `imageRef(r)` where `image` was used.

### Answer

```go
// imageRef returns the full image reference, such as
// "docker.io/api:v1.4.2".
func imageRef(r Request) string {
	registry := r.Registry
	if registry == "" {
		registry = "docker.io"
	}
	return registry + "/" + r.Service + ":" + r.Tag
}
```

And in `Plan`:

```go
steps := []string{fmt.Sprintf("deploy %s to %d replicas", imageRef(r), r.Replicas)}
```

## Task 3: Extract batchSteps

The loop needs `r` and appends lines. In the helper, start with an empty slice of its own (`var steps []string`), append to it, and return it.

### Closer

In `Plan`, `steps = append(steps, batchSteps(r)...)` adds every batch line after the first step. The `...` spreads the returned slice.

### Answer

```go
func Plan(r Request) ([]string, error) {
	if err := validate(r); err != nil {
		return nil, err
	}
	steps := []string{fmt.Sprintf("deploy %s to %d replicas", imageRef(r), r.Replicas)}
	steps = append(steps, batchSteps(r)...)
	steps = append(steps, "verify "+r.Service)
	return steps, nil
}

// batchSteps returns one step per batch, replacing at most MaxSurge pods at a
// time.
func batchSteps(r Request) []string {
	var steps []string
	done := 0
	for batch := 1; done < r.Replicas; batch++ {
		// ... the loop body, unchanged
	}
	return steps
}
```
