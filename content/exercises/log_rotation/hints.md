## Task 1: Rename the cryptic names

Say each one out loud: `fs` is the list of files, `t` is the current time, `r` is the list of actions, `f` is one file. Use Rename Symbol so every use changes.

### Closer

`fs` → `files`, `t` → `now`, `r` → `actions`, `f` → `file`. The parameter `t` is especially confusing in Go, where `t` usually means a `*testing.T`.

### Answer

```go
func Plan(files []File, now time.Time) []string {
	var actions []string
	for _, file := range files {
		if !strings.Contains(file.Name, ".log.") {
			if file.Size >= 104857600 {
				actions = append(actions, "rotate "+file.Name)
	// ... and so on: every f becomes file, every r becomes actions
```

## Task 2: Name the numbers

`104857600` is 100 MiB, `604800*time.Second` is 7 days, and `86400*time.Second` is 1 day. Each comes from the rotation policy, so each gets a name.

### Closer

A `const ( ... )` block outside `Plan`: `rotateSize = 100 << 20`, `compressAfter = 24 * time.Hour`, `deleteAfter = 7 * 24 * time.Hour`. Then compare with the names: `now.Sub(file.Modified) > deleteAfter`.

### Answer

```go
// The rotation policy.
const (
	rotateSize    = 100 << 20          // rotate the live log at 100 MiB
	compressAfter = 24 * time.Hour     // compress rotated logs after a day
	deleteAfter   = 7 * 24 * time.Hour // delete rotated logs after a week
)
```

## Task 3: Extract the decisions

The outer `if` splits live logs from rotated ones. Everything inside its first branch becomes `liveAction`; everything inside the `else` becomes `rotatedAction`. Return the word; let `Plan` add the name.

### Closer

In the helpers, every `r = append(r, "keep "+f.Name)` becomes `return "keep"`. In `Plan`: `if strings.Contains(file.Name, ".log.") { actions = append(actions, rotatedAction(file, now)+" "+file.Name); continue }`, then the same with `liveAction` for the live log.

### Answer

```go
func Plan(files []File, now time.Time) []string {
	var actions []string
	for _, file := range files {
		if strings.Contains(file.Name, ".log.") {
			actions = append(actions, rotatedAction(file, now)+" "+file.Name)
			continue
		}
		actions = append(actions, liveAction(file)+" "+file.Name)
	}
	return actions
}
```

## Task 4: Flatten into guard clauses

Each helper can return the moment it knows the answer. Deleting wins over everything for rotated logs, so check it first.

### Closer

`liveAction`: `if file.Size >= rotateSize { return "rotate" }`, then `return "keep"`. `rotatedAction`: compute `age` once, return `"delete"` when it's over `deleteAfter`, then return `"keep"` when the file is already `.gz` **or** not older than `compressAfter`, and otherwise `"compress"`.

### Answer

```go
// liveAction decides what to do with the log the service is writing to now:
// "rotate" or "keep".
func liveAction(file File) string {
	if file.Size >= rotateSize {
		return "rotate"
	}
	return "keep"
}

// rotatedAction decides what to do with an old, rotated log: "delete",
// "compress" or "keep".
func rotatedAction(file File, now time.Time) string {
	age := now.Sub(file.Modified)
	if age > deleteAfter {
		return "delete"
	}
	if strings.HasSuffix(file.Name, ".gz") || age <= compressAfter {
		return "keep"
	}
	return "compress"
}
```
