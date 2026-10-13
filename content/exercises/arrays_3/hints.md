## Task 1: one loop, one subtraction

Loop over the seven indexes and set `change[i]` from `this[i]` and `last[i]`.

## Task 2: remember an index

Keep `best := 0`, the index of the busiest day so far. In the loop, move
`best` to `i` only when `pages[i]` is strictly greater than `pages[best]`.
Strictly greater keeps the earliest day on a tie. At the end, look up
`dayNames[best]`.

## Task 3: compare first, then add up

Check `this == last` first. For the totals, a small helper saves typing:

```go
func sum(pages [7]int) int {
	total := 0
	for i := range len(pages) {
		total += pages[i]
	}
	return total
}
```

## Task 3: almost the answer

```go
total := sum(this)
if this == last {
	return fmt.Sprintf("same as last week: %d pages", total)
}
change := total - sum(last)
return fmt.Sprintf("%d pages (%+d vs last week), busiest %s", total, change, BusiestDay(this))
```
