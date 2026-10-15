## Task 1: Day-by-day change

Each day's change only needs that same day from both weeks. Loop over the seven indexes and fill `change` one element at a time.

### Closer

Use `for i := range len(change)` and set `change[i]` from `this[i]` and `last[i]`.

### Answer

```go
for i := range len(change) {
	change[i] = this[i] - last[i]
}
```

Index `i` means the same day in all three arrays, so subtracting element by element gives the change per day.

## Task 2: Busiest day

`best` is the index of the busiest day found so far. Walk through all the days and move `best` when you find a day with more pages.

### Closer

Compare the counts, not the indexes: `pages[i] > pages[best]`. Use strictly greater, so the earliest day keeps the win on a tie.

### Answer

```go
for i := range len(pages) {
	if pages[i] > pages[best] {
		best = i
	}
}
```

With `>` a later day only wins if it has more pages, and an all-zero week leaves `best` at `0` (`"Mon"`).

## Task 3: The weekly summary

First check if the two weeks are exactly equal; arrays can be compared with `==`. You also need the total of each week, and the busiest day from task 2.

### Closer

A small helper that adds up a week saves typing. Put it at the top level of `exercise.go` (for example at the end of the file), not inside `WeekSummary`:

```go
func sum(pages [7]int) int {
	total := 0
	for i := range len(pages) {
		total += pages[i]
	}
	return total
}
```

Then use `fmt.Sprintf` with `%d` for the total and `%+d` for the change. Remember `import "fmt"`.

### Answer

```go
total := sum(this)
if this == last {
	return fmt.Sprintf("same as last week: %d pages", total)
}
change := total - sum(last)
return fmt.Sprintf("%d pages (%+d vs last week), busiest %s", total, change, BusiestDay(this))
```

`this == last` is only true when every day matches, so equal totals on different days still get the long line.
