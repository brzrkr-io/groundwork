## Task 1: read, then decode

`os.ReadFile(path)` returns the file's bytes. Pass them to
`json.Unmarshal(data, &rs)`, where `rs` is a `[]Resource` variable. Both calls
can fail, and both errors get the same prefix:

```go
data, err := os.ReadFile(path)
if err != nil {
	return nil, fmt.Errorf("load desired state: %w", err)
}
```

## Task 2: two more loops over the maps

Creates: loop over `want` and look each ID up in `got`. Deletes: loop over
`got` and look each ID up in `want`. The update loop shows how to look an ID
up. Leave the sort at the end alone: it runs after your loops.

```go
for id := range want {
	if _, ok := got[id]; !ok {
		changes = append(changes, Change{Action: "create", ID: id})
	}
}
```

## Task 3: count while you write

In the `if err != nil` block after `Actual`, return the same thing as the
block after `LoadDesired`.
Then keep three counters and use a `switch c.Action` to pick the line format:

```go
var b strings.Builder
add, change, destroy := 0, 0, 0
for _, c := range changes {
	switch c.Action {
	case "create":
		add++
		fmt.Fprintf(&b, "+ %s\n", c.ID)
	// case "update": "~ %s (tags: %s)\n" with strings.Join(c.Keys, ", ")
	// case "delete": "- %s\n"
	}
}
fmt.Fprintf(&b, "Plan: %d to add, %d to change, %d to destroy.\n", add, change, destroy)
return b.String(), 2
```
