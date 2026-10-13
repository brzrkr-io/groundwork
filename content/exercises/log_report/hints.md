## Task 1: one helper for every group

The starter's `group` closure saves a lot of typing: `group("status")` is
`"200"`, `group("time")` is `"11/Oct/2026:13:55:36 +0000"`, and so on. Copy the
shape of the time step (call, check `err`, keep the value) for
`strconv.Atoi(...)`, `time.ParseDuration(group("duration") + "s")` and
`strings.Cut(group("path"), "?")`.

## Task 2: the same loop as in Reading input line by line

Count every line with `n++` at the top of the loop, `continue` on blank lines,
and on a parse error append `fmt.Errorf("line %d: %w", n, err)` to
`log.Problems` instead of returning. After the loop, check `sc.Err()`.

## Tasks 3 and 4: count into a map, then sort a slice

`fmt.Sprintf("%dxx", e.Status/100)` turns 404 into `"4xx"`. For `TopPaths`,
count into a `map[string]int`, copy it into a `[]PathCount`, and sort:

```go
slices.SortFunc(top, func(a, b PathCount) int {
	if a.Count != b.Count {
		return cmp.Compare(b.Count, a.Count) // b first: biggest first
	}
	return cmp.Compare(a.Path, b.Path)
})
```

Finish with `if len(top) > n { top = top[:n] }`.

## Tasks 5 and 6: near-solutions

```go
sorted := slices.Clone(durations)
slices.Sort(sorted)
rank := (p*len(sorted) + 99) / 100
return sorted[rank-1]
```

For the text report, write into a `strings.Builder` with `fmt.Fprintf(&b, ...)`
and finish with `_, err := io.WriteString(w, b.String())`. For JSON, keep the encoder in a variable so you can call `SetIndent` before
encoding:

```go
enc := json.NewEncoder(w)
enc.SetIndent("", "  ")
return enc.Encode(jsonReport{Requests: r.Requests, P50Ms: r.P50.Milliseconds() /* ... */})
```
