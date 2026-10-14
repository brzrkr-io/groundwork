## Task 1: Resolve a surge or unavailable value

Strip a possible `%` first, then convert what's left to a number. Whether
there was a `%` decides if you return the number itself or a share of
`replicas`.

### Closer

`number, isPercent := strings.CutSuffix(value, "%")`, then
`strconv.Atoi(number)`. Reject `n < 0`. For a percentage,
`replicas * n / 100` rounds down and `(replicas*n + 99) / 100` rounds up.
Add the `fmt`, `strconv` and `strings` imports.

### Answer

```go
number, isPercent := strings.CutSuffix(value, "%")
n, err := strconv.Atoi(number)
if err != nil {
	return 0, fmt.Errorf("%q is not a number or a percentage: %w", value, err)
}
if n < 0 {
	return 0, fmt.Errorf("%q is negative", value)
}
if !isPercent {
	return n, nil
}
if roundUp {
	return (replicas*n + 99) / 100, nil
}
return replicas * n / 100, nil
```

## Task 2: Work out the rollout bounds

Two calls to `Resolve`, each with its own rounding direction, then the
"both zero" rule, then two sums.

### Closer

`surge, err := Resolve(maxSurge, replicas, true)` and
`unavailable, err := Resolve(maxUnavailable, replicas, false)`. The
built-in `max(a, b)` keeps `replicas-unavailable` from going below 0.

### Answer

```go
surge, err := Resolve(maxSurge, replicas, true)
if err != nil {
	return 0, 0, fmt.Errorf("maxSurge: %w", err)
}
unavailable, err := Resolve(maxUnavailable, replicas, false)
if err != nil {
	return 0, 0, fmt.Errorf("maxUnavailable: %w", err)
}
if surge == 0 && unavailable == 0 {
	unavailable = 1
}
return replicas + surge, max(replicas-unavailable, 0), nil
```

## Task 3: Find the revision to roll back to

Go through the lines. A line is a revision row when its first word is a
number. Keep the biggest revision whose other words include the image.

### Closer

`strings.SplitSeq(history, "\n")` for the lines, `strings.Fields(line)` for
the words, `strconv.Atoi(fields[0])` to spot a revision (an error means
"not a row, skip it"), and `slices.Contains(fields[1:], image)` for the
match. Start with `found := 0`; revisions start at 1.

### Answer

```go
found := 0
for line := range strings.SplitSeq(history, "\n") {
	fields := strings.Fields(line)
	if len(fields) < 2 {
		continue
	}
	revision, err := strconv.Atoi(fields[0])
	if err != nil {
		continue
	}
	if slices.Contains(fields[1:], image) && revision > found {
		found = revision
	}
}
if found == 0 {
	return 0, fmt.Errorf("no revision deployed %s", image)
}
return found, nil
```
