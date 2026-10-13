## Pointer receivers everywhere

Change every `(r Rollout)` to `(r *Rollout)`, like `Pause`. `Advance` and
`Run` change `r.Updated`, so with a value receiver their work is lost. `Status`
only reads, but it gets one too, to keep the type consistent.

## Task 1: order your ifs

Check `r.Paused` first and return early, then check
`r.Updated == r.Replicas`. Build the other two strings with
`fmt.Sprintf("paused at %d/%d", r.Updated, r.Replicas)` and the same idea
for `"rolling out %d/%d"`.

## Task 2: how many are left?

`r.Replicas - r.Updated` replicas still need updating. The batch you can
actually do is `min(batch, r.Replicas-r.Updated)`. Return early with `0` when
`r.Paused || batch <= 0`.

## Task 3: a loop that stops on zero

```go
var steps []string
for r.Advance(batch) > 0 {
	steps = append(steps, r.Status())
}
return steps
```

Each pass around the loop calls `Advance` once. When it returns `0`, the
condition is false and the loop ends.
