## Copy the Upper stage

The `Upper` example in the README is the shape of all three stages: `out :=
make(...)`, `go func() { defer close(out); ... }()`, `return out`. Only the
loop inside the goroutine changes.

## Task 1: loop over the slice

```go
out := make(chan string)
go func() {
	defer close(out)
	for _, line := range lines {
		out <- line
	}
}()
return out
```

## Task 2: Cut each line

Inside `for line := range in`:

```go
level, message, _ := strings.Cut(line, " ")
out <- Entry{Level: level, Message: message}
```

## Task 3: send only matches

Inside `for entry := range in`, send the entry only `if entry.Level ==
level`. The channel type is `chan Entry`.
