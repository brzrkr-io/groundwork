## Task 1: Print replica counts

Add a method named `String` on `Replicas` with a value receiver. It returns
two numbers with a slash between them.

### Closer

The header is `func (r Replicas) String() string`. Use `fmt.Sprintf` with
`%d` for each count: `"%d/%d"`.

### Answer

```go
func (r Replicas) String() string {
	return fmt.Sprintf("%d/%d", r.Ready, r.Desired)
}
```

## Task 2: Print a deployment

`Replicas` and `Status` already know how to print themselves. Let `fmt` use
their `String` methods instead of building that text again.

### Closer

Write `func (d Deployment) String() string`. Print `d.Replicas` and
`d.Status` with `%v`; that calls the fields' `String` methods, not
`Deployment`'s, so there is no loop.

### Answer

```go
func (d Deployment) String() string {
	return fmt.Sprintf("%s/%s %v %v", d.Namespace, d.Name, d.Replicas, d.Status)
}
```

## Task 3: Build the report

You need the number of degraded deployments before you can write the summary
line. Count them first, then write the summary and one line per deployment
into a `strings.Builder`.

### Closer

Add `"strings"` to the imports. Count with a loop and `if d.Status == Degraded`.
Then use `var b strings.Builder`, `fmt.Fprintf(&b, ...)` for the summary, and
`fmt.Fprintln(&b, d)` for each deployment.

### Answer

```go
degraded := 0
for _, d := range deps {
	if d.Status == Degraded {
		degraded++
	}
}

var b strings.Builder
fmt.Fprintf(&b, "deployments: %d, degraded: %d\n", len(deps), degraded)
for _, d := range deps {
	fmt.Fprintln(&b, d)
}
return b.String()
```

`Fprintln` calls `Deployment.String` and adds the newline for you.
