## Task 1: move lines, keep iota on the first

Only the first constant needs `Phase = iota`. Move `Unknown` to the top, give
it `Phase = iota`, and leave the other names on their own lines in pipeline
order:

```go
const (
	Unknown Phase = iota
	Pending
	// ...
)
```

## Task 2: which end is wrong?

Once the order is fixed, the active phases sit strictly between `Pending`
and `Done`. The lower bound already uses `>`. Look closely at the upper
bound: `<=` lets `Done` in.

## Task 3: handle the ends first

Check for the two special phases before doing the arithmetic, and return
`p` unchanged for them:

```go
if p == Unknown || p == Done {
	return p
}
return p + 1
```
