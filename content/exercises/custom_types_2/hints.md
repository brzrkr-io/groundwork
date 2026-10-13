## Task 1: Fix the constant order

With `iota`, each line's position is its value. Put the lines in the order of the table, and only the first line needs `Phase = iota`.

### Closer

Move `Unknown` to the top and give it `Phase = iota`; the other names follow on their own lines:

```go
const (
	Unknown Phase = iota
	Pending
	// ...
)
```

### Answer

```go
const (
	Unknown Phase = iota
	Pending
	Building
	Testing
	Deploying
	Done
)
```

`Unknown` is now `0`, so a `Phase` nobody set reads as `Unknown`.

## Task 2: Fix IsActive

Once the order is fixed, the active phases sit strictly between `Pending` and `Done`. The lower bound is already right; look at the upper bound.

### Closer

`p <= Done` lets `Done` in. You want "less than", not "less than or equal".

### Answer

```go
return p > Pending && p < Done
```

Only `Building`, `Testing` and `Deploying` lie strictly between `Pending` and `Done`.

## Task 3: Stop Next at the ends

Two phases must not move: `Unknown` and `Done`. Deal with them before the `p + 1` step.

### Closer

Check both with `||` and return `p` unchanged: `if p == Unknown || p == Done { ... }`.

### Answer

```go
if p == Unknown || p == Done {
	return p
}
return p + 1
```

Every other phase is followed by the next number, which is the next phase in the pipeline.
