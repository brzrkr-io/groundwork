## Task 1: Generate

Copy the shape of `Upper` from the README: make a channel, fill it from a
goroutine, and return it straight away. Here the goroutine loops over the
`lines` slice instead of an input channel.

### Closer

`out := make(chan string)`, then `go func() { defer close(out); ... }()`.
Inside, `for _, line := range lines` and send each `line` on `out`. Finish
with `return out`.

### Answer

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

`defer close(out)` closes the channel once every line is sent, so the receiver's `range` ends.

## Task 2: Parse

Same shape as `Generate`, but the channel carries `Entry` values and the loop
ranges over `in`. Split each line at its first space with `strings.Cut`.

### Closer

Make `out := make(chan Entry)` and add `"strings"` to the imports. Inside
`for line := range in`, call `level, message, _ := strings.Cut(line, " ")`
and send an `Entry` built from them.

### Answer

```go
out := make(chan Entry)
go func() {
	defer close(out)
	for line := range in {
		level, message, _ := strings.Cut(line, " ")
		out <- Entry{Level: level, Message: message}
	}
}()
return out
```

With no space, `strings.Cut` returns the whole line as `level` and `""` as `message`.

## Task 3: Filter by level

Same shape once more. This time the stage forwards an entry only when its
`Level` matches.

### Closer

Inside `for entry := range in`, wrap the send in
`if entry.Level == level { ... }`. The output channel is `chan Entry`.

### Answer

```go
out := make(chan Entry)
go func() {
	defer close(out)
	for entry := range in {
		if entry.Level == level {
			out <- entry
		}
	}
}()
return out
```

Entries that don't match are simply not sent, and order is kept because one goroutine handles them in turn.
