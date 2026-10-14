## Task 1: Rename the helper functions

Each helper appears in three places: its declaration, its call inside `Pages`, and the first word of its comment. Your editor's Rename Symbol handles the code; check the comment yourself.

### Closer

Put the cursor on `prs`, press `F2` (or your editor's rename), type `criticalService`. Do the same for `fmtP` → `pageLine`. Then check the comments: `// prs returns ...` must become `// criticalService returns ...`. Fix any your editor left alone.

### Answer

```go
		service, ok := criticalService(line)
		...
			pages = append(pages, pageLine(service, counts[service]))

// criticalService returns the service named in a critical alert line, and
// false for any other line.
func criticalService(l string) (string, bool) {

// pageLine formats one page line, such as "api: 3 critical".
func pageLine(s string, n int) string {
```

## Task 2: Rename the variables in Pages

Read `Pages` top to bottom and say out loud what each one-letter name holds: "m maps a service to its count", "o is the services in order"... The sentence gives you the name.

### Closer

`m` holds counts per service, `o` the services in first-seen order, `l` one log line, `s` and `x` a service name, `r` the page lines. Rename each with Rename Symbol so every use changes at once.

### Answer

```go
counts := map[string]int{}
var services []string // in the order they first appear
for _, line := range lines {
	service, ok := criticalService(line)
	if !ok {
		continue
	}
	if counts[service] == 0 {
		services = append(services, service)
	}
	counts[service]++
}
var pages []string
for _, service := range services {
	if counts[service] >= limit {
		pages = append(pages, pageLine(service, counts[service]))
	}
}
return pages
```

## Task 3: Name the parts of an alert line

Two jobs here. Give `l`, `f`, `s` and `n` names that say what they hold, and pull `fields[1]` and `fields[2]` out into their own variables. Mind the order: the line `"garbage"` has only one field, so reading `fields[1]` too early crashes.

### Closer

Rename `l` to `line`, `f` to `fields`, and `pageLine`'s `s` and `n` to `service` and `count`. Then split the long `if` in two: after `if len(fields) < 3 { return "", false }`, it's safe to write `level, service := fields[1], fields[2]`. Then compare `level` with `"critical"` and return `service`.

### Answer

```go
func criticalService(line string) (string, bool) {
	fields := strings.Fields(line)
	if len(fields) < 3 {
		return "", false
	}
	level, service := fields[1], fields[2]
	if level != "critical" {
		return "", false
	}
	return service, true
}

func pageLine(service string, count int) string {
	return fmt.Sprintf("%s: %d critical", service, count)
}
```
