## Task 1: Read the READY column

Split the value at the `"/"` into two parts. It's ready only if there was a
`"/"` and both parts are the same text.

### Closer

`strings.Cut(ready, "/")` returns three things: the part before, the part
after, and a `bool` saying whether `"/"` was found. Combine the checks with
`&&`.

### Answer

```go
running, total, found := strings.Cut(ready, "/")
return found && running == total
```

Without the `found` check, `""` would give two empty strings that are equal.

## Task 2: Find the pods that are not ready

Go through the output line by line, split each line into columns, skip lines
you don't care about, and collect the names that need attention. Sort the
names at the end.

### Closer

Loop with `for line := range strings.SplitSeq(output, "\n")` and get
`columns := strings.Fields(line)`. `continue` when `len(columns) < 3` or
`columns[0] == "NAME"`. Use `IsReady(columns[1])` and compare `columns[2]` with
`"Completed"`. Finish with `slices.Sort(names)`.

### Answer

```go
var names []string
for line := range strings.SplitSeq(output, "\n") {
	columns := strings.Fields(line)
	if len(columns) < 3 || columns[0] == "NAME" {
		continue
	}
	name, ready, status := columns[0], columns[1], columns[2]
	if !IsReady(ready) && status != "Completed" {
		names = append(names, name)
	}
}
slices.Sort(names)
return names
```

`Fields` ignores the padding between columns, and the length check also skips
the empty line after the last newline.

## Task 3: Write the summary

You already wrote the hard part: call `NotReadyPods` and decide what to say
based on how many names come back.

### Closer

If `len(names) == 0`, return `"all pods ready"`. Otherwise glue the names with
`strings.Join(names, ", ")` and put `"not ready: "` in front.

### Answer

```go
names := NotReadyPods(output)
if len(names) == 0 {
	return "all pods ready"
}
return "not ready: " + strings.Join(names, ", ")
```

`Join` puts `", "` only between names, never after the last one.
