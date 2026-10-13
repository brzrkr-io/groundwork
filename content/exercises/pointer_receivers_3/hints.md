## Task 1: Report the status

Give `Status` a pointer receiver like `Pause`, to keep the type consistent.
Then check the situations in the README's order: paused first, then done,
then everything else.

### Closer

Use early returns: `if r.Paused { ... }`, then `if r.Updated == r.Replicas {
return "done" }`. Build the numbers into the text with
`fmt.Sprintf("paused at %d/%d", r.Updated, r.Replicas)` (add `import "fmt"`).

### Answer

```go
func (r *Rollout) Status() string {
	if r.Paused {
		return fmt.Sprintf("paused at %d/%d", r.Updated, r.Replicas)
	}
	if r.Updated == r.Replicas {
		return "done"
	}
	return fmt.Sprintf("rolling out %d/%d", r.Updated, r.Replicas)
}
```

## Task 2: Advance by one batch

`Advance` changes `r.Updated`, so it needs a pointer receiver. Handle the
"do nothing" cases first, then work out how many replicas are left.

### Closer

Return `0` early when `r.Paused || batch <= 0`. `r.Replicas - r.Updated`
replicas are left, so the batch you can really do is
`min(batch, r.Replicas-r.Updated)`.

### Answer

```go
func (r *Rollout) Advance(batch int) int {
	if r.Paused || batch <= 0 {
		return 0
	}
	n := min(batch, r.Replicas-r.Updated)
	r.Updated += n
	return n
}
```

## Task 3: Run the whole rollout

Give `Run` a pointer receiver too. You want a loop that keeps calling
`r.Advance(batch)` and stops the first time it returns `0`.

### Closer

The call itself can be the loop condition: `for r.Advance(batch) > 0 { ... }`.
Inside, `append` `r.Status()` to a `var steps []string`.

### Answer

```go
func (r *Rollout) Run(batch int) []string {
	var steps []string
	for r.Advance(batch) > 0 {
		steps = append(steps, r.Status())
	}
	return steps
}
```

Each pass calls `Advance` once, so when it returns `0` the loop ends.
