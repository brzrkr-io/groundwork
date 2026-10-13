## Task 1: wrap it in defer

Put the existing `if r := recover(); r != nil { ... }` inside
`defer func() { ... }()`, and place that before `err = handler()`.

## Task 2: name the result

Change the signature to `func Guard(command string, handler func() error) (err error)`.
Then delete `var err error`, because `err` already exists as the result. You
can also replace the last two lines with `return handler()`.

## Task 3: a nil map can't be written to

`var counts map[string]int` declares a map but doesn't create one, so
`counts[a.Severity]++` panics. Use `counts := make(map[string]int)` instead,
and remove the whole `defer func() { recover() }()` block.

## Almost the answer

```go
func Guard(command string, handler func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("command %q panicked: %v", command, r)
		}
	}()
	return handler()
}
```
