## Task 1: Parse one line

`strings.Fields(line)` gives you a slice of words. Check how many words there
are before you index into it, then think about which index holds the level and
which words make up the message.

### Closer

Return `"", "", false` when `len(fields) < 3`. The level is `fields[1]`
(upper-cased with `strings.ToUpper`), and the message words are the sub-slice
`fields[2:]`, glued back with `strings.Join(..., " ")`.

### Answer

```go
fields := strings.Fields(line)
if len(fields) < 3 {
	return "", "", false
}
return strings.ToUpper(fields[1]), strings.Join(fields[2:], " "), true
```

## Task 2: Count lines per level

You need a map from level to count, and a loop that asks `ParseLine` about
each line. Bad lines should be skipped, not counted.

### Closer

Start with `counts := make(map[string]int)`. In the loop, call
`level, _, ok := ParseLine(line)`, `continue` when `!ok`, otherwise
`counts[level]++`.

### Answer

```go
counts := make(map[string]int)
for _, line := range lines {
	level, _, ok := ParseLine(line)
	if !ok {
		continue
	}
	counts[level]++
}
return counts
```

Reading a missing key gives `0`, so `++` works even the first time.

## Task 3: Format a summary

Build a `[]string` of `LEVEL=count` parts and join them with spaces. Use two
loops: one over `standardLevels`, then one over the other levels in the map,
in alphabetical order.

### Closer

Each part is `fmt.Sprintf("%s=%d", level, counts[level])`. For the second loop,
range over `slices.Sorted(maps.Keys(counts))` and skip levels where
`slices.Contains(standardLevels, level)` is true. Add `"fmt"`, `"maps"` and
`"slices"` to the imports.

### Answer

```go
var parts []string
for _, level := range standardLevels {
	parts = append(parts, fmt.Sprintf("%s=%d", level, counts[level]))
}
for _, level := range slices.Sorted(maps.Keys(counts)) {
	if slices.Contains(standardLevels, level) {
		continue // already written above
	}
	parts = append(parts, fmt.Sprintf("%s=%d", level, counts[level]))
}
return strings.Join(parts, " ")
```

## Task 4: List the top messages for a level

Do it in four steps: count the messages for this level in a map, sort the
messages, cut the list to `n`, then format each one. The result must be
`[]string{}`, never `nil`.

### Closer

Get sorted keys with `slices.Sorted(maps.Keys(counts))`, then re-sort with
`slices.SortFunc` and `cmp.Compare`, comparing `counts[b]` to `counts[a]` so
higher counts come first. Cut with `messages[:n]` when `len(messages) > n`.
Format with `fmt.Sprintf("%dx %s", ...)`. Add `"cmp"` to the imports.

### Answer

```go
counts := make(map[string]int)
for _, line := range lines {
	lineLevel, message, ok := ParseLine(line)
	if ok && lineLevel == level {
		counts[message]++
	}
}

messages := slices.Sorted(maps.Keys(counts))
slices.SortFunc(messages, func(a, b string) int {
	if c := cmp.Compare(counts[b], counts[a]); c != 0 {
		return c // b first: higher count wins
	}
	return cmp.Compare(a, b) // tie: alphabetical
})
if len(messages) > n {
	messages = messages[:n]
}

top := []string{}
for _, message := range messages {
	top = append(top, fmt.Sprintf("%dx %s", counts[message], message))
}
return top
```

Starting `top` as `[]string{}` means an empty result is never `nil`.
