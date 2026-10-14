## Task 1: Put the unit in the delay's name

Rename first, with your editor's Rename Symbol. Then look at the conversion in `BackoffDelay`: the name now says milliseconds, but what is it multiplied by?

### Closer

The number of milliseconds becomes a duration with `time.Duration(n) * time.Millisecond`. Don't forget the comment: `// delayMillis returns ...`.

### Answer

```go
func BackoffDelay(attempt int) time.Duration {
	return time.Duration(delayMillis(attempt)) * time.Millisecond
}

// delayMillis returns the wait before retry number attempt, in milliseconds.
func delayMillis(attempt int) int {
```

## Task 2: Rename the helper that returns the opposite

The loop appends a node when `!node.Healthy`, so the helper returns the unhealthy names. `ServingCount` counted those as if they were the healthy ones.

### Closer

Rename to `unhealthyNodes` and write a comment like `// unhealthyNodes returns the names of the nodes that failed their last health check.` Healthy is everything else: `len(nodes)` minus the unhealthy count.

### Answer

```go
// unhealthyNodes returns the names of the nodes that failed their last
// health check.
func unhealthyNodes(nodes []Node) []string {

// ...

func ServingCount(nodes []Node) int {
	return len(nodes) - len(unhealthyNodes(nodes))
}
```

## Task 3: Make the lookup a lookup

`owners[service] = owner` writes the default into the caller's map. Delete the write, and return the default directly when the key is missing.

### Closer

Use the comma-ok form: `if owner, ok := owners[service]; ok { return owner }`, then `return "platform"` after the `if`.

### Answer

```go
func ownerOf(owners map[string]string, service string) string {
	if owner, ok := owners[service]; ok {
		return owner
	}
	return "platform"
}
```
