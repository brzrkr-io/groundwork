## Task 1: Panic on an invalid built-in port

Parse the string with `strconv.Atoi`. Three things make a port bad: it isn't a
number, it is below 1, or it is above 65535. You can test all three in one
`if`.

### Closer

`port, err := strconv.Atoi(s)`, then
`if err != nil || port < 1 || port > 65535 { ... }`. Build the message with
`fmt.Sprintf("invalid port %q", s)` and pass it to `panic`.

### Answer

```go
port, err := strconv.Atoi(s)
if err != nil || port < 1 || port > 65535 {
	panic(fmt.Sprintf("invalid port %q", s))
}
return port
```

`%q` adds the double quotes around `s` for you.

## Task 2: Turn a panic into an error

`recover()` only stops a panic when it is called inside a function that you
`defer`. `err` is already a named result, so a deferred function can still
change it after the panic.

### Closer

Put this before `return job()`:

```go
defer func() {
	if r := recover(); r != nil {
		// set err here
	}
}()
```

### Answer

```go
defer func() {
	if r := recover(); r != nil {
		err = fmt.Errorf("%w: %v", ErrPanic, r)
	}
}()
return job()
```

`%w` keeps `ErrPanic` findable with `errors.Is`, and `%v` prints the panic
value whatever its type.

## Task 3: Run every job, even after a panic

Let `SafeRun` do the hard part. Because `SafeRun` never lets a panic escape,
a plain loop always reaches the next job.

### Closer

Loop with `for _, job := range jobs` and call `SafeRun(job.Run)`. Append
`job.Name + ": ok"` when the error is `nil`, otherwise
`job.Name + ": " + err.Error()`.

### Answer

```go
var report []string
for _, job := range jobs {
	if err := SafeRun(job.Run); err != nil {
		report = append(report, job.Name+": "+err.Error())
	} else {
		report = append(report, job.Name+": ok")
	}
}
return report
```
