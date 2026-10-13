## Task 1: fields and slicing

`strings.Fields(line)` gives you a slice of words. Check `len(fields) < 3`
first. The level is `fields[1]`, and the message words are the sub-slice
`fields[2:]`, glued back together with `strings.Join(..., " ")`.

## Task 2: counting with a map

Start with `counts := make(map[string]int)`. In a `range` loop, call
`level, _, ok := ParseLine(line)`. Use `continue` when `!ok`, otherwise
`counts[level]++`. Return the map even if it stays empty.

## Task 3: two loops into one slice of parts

Collect strings like `fmt.Sprintf("%s=%d", level, counts[level])` into a
`[]string`, then `strings.Join(parts, " ")`.

```go
for _, level := range standardLevels { /* add LEVEL=count */ }
for _, level := range slices.Sorted(maps.Keys(counts)) {
	if slices.Contains(standardLevels, level) {
		continue
	}
	/* add LEVEL=count */
}
```

## Task 4: count, sort, cut, format

Count messages for the level in a `map[string]int`. Get the keys with
`slices.Sorted(maps.Keys(counts))`, then sort them by count:

```go
slices.SortFunc(messages, func(a, b string) int {
	if c := cmp.Compare(counts[b], counts[a]); c != 0 {
		return c // b first: higher count wins
	}
	return cmp.Compare(a, b) // tie: alphabetical
})
```

Cut to `messages[:n]` when there are more than `n`. Start the result as
`top := []string{}` so it is never `nil`, and append
`fmt.Sprintf("%dx %s", counts[m], m)` for each message.
