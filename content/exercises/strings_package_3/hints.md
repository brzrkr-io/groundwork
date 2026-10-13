## Task 1: Cut gives you three answers

`running, total, found := strings.Cut(ready, "/")`. The pod is ready only if
`found` is true **and** `running == total`.

## Task 2: the loop

```go
var names []string
for line := range strings.SplitSeq(output, "\n") {
	columns := strings.Fields(line)
	// skip short lines and the header with continue
	// check columns[1] with IsReady and columns[2] against "Completed"
}
slices.Sort(names)
return names
```

## Task 3: reuse task 2

Call `NotReadyPods(output)`. If the result has length 0, return
`"all pods ready"`. Otherwise return
`"not ready: " + strings.Join(names, ", ")`.
