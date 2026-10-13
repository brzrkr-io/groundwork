## Task 1: Load the desired state

Two steps: read the file's bytes, then decode the JSON into a `[]Resource`.
Both can fail, and both errors get the same `load desired state: ` prefix.

### Closer

`os.ReadFile(path)` returns the bytes. `json.Unmarshal(data, &rs)` fills a
`var rs []Resource`. Wrap each error with
`fmt.Errorf("load desired state: %w", err)`, and add `"encoding/json"`,
`"fmt"` and `"os"` to the imports.

### Answer

```go
data, err := os.ReadFile(path)
if err != nil {
	return nil, fmt.Errorf("load desired state: %w", err)
}
var rs []Resource
if err := json.Unmarshal(data, &rs); err != nil {
	return nil, fmt.Errorf("load desired state: %w", err)
}
return rs, nil
```

## Task 2: Find the differences

You need two more loops over the maps. A "create" is an ID in `want` that is
missing from `got`; a "delete" is the other way round. The update loop shows
how to look an ID up.

### Closer

Use the two-value lookup `if _, ok := got[id]; !ok { ... }` inside
`for id := range want`. Leave the sort at the end alone: it runs after your
loops.

### Answer

```go
for id := range want {
	if _, ok := got[id]; !ok {
		changes = append(changes, Change{Action: "create", ID: id})
	}
}

// ... the update loop stays here ...

for id := range got {
	if _, ok := want[id]; !ok {
		changes = append(changes, Change{Action: "delete", ID: id})
	}
}
```

## Task 3: Render the plan

The error case after `Actual` returns the same thing as the one after
`LoadDesired`. For the report, write one line per change and count each kind
as you go, so you can print the summary line at the end.

### Closer

Build the text in a `strings.Builder` with `fmt.Fprintf(&b, ...)`. Keep three
counters and use `switch c.Action` to pick the line format: `"+ %s\n"`,
`"~ %s (tags: %s)\n"` with `strings.Join(c.Keys, ", ")`, and `"- %s\n"`.

### Answer

```go
if err != nil {
	return "Error: " + err.Error() + "\n", 1
}

changes := Diff(desired, actual)
if len(changes) == 0 {
	return "No changes. AWS matches the desired state.\n", 0
}

var b strings.Builder
add, change, destroy := 0, 0, 0
for _, c := range changes {
	switch c.Action {
	case "create":
		add++
		fmt.Fprintf(&b, "+ %s\n", c.ID)
	case "update":
		change++
		fmt.Fprintf(&b, "~ %s (tags: %s)\n", c.ID, strings.Join(c.Keys, ", "))
	case "delete":
		destroy++
		fmt.Fprintf(&b, "- %s\n", c.ID)
	}
}
fmt.Fprintf(&b, "Plan: %d to add, %d to change, %d to destroy.\n", add, change, destroy)
return b.String(), 2
```

Exit code 2 tells a CI job "there is drift", just like `terraform plan -detailed-exitcode`.
