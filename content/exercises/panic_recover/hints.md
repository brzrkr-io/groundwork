## Task 1: one check, one panic

Call `strconv.Atoi(s)`. You can test all three problems in one `if`: the error
is not `nil`, the port is below 1, or the port is above 65535. Build the
message with `fmt.Sprintf("invalid port %q", s)` and pass it to `panic`.

## Task 2: recover lives in a deferred function

`recover()` only stops a panic when it is called inside a function that you
`defer`. Put the `defer` before `return job()`. Because `err` is a named
result, the deferred function can still change it after the panic.

```go
defer func() {
	if r := recover(); r != nil {
		// set err here
	}
}()
```

## Task 3: let SafeRun do the work

Loop over the jobs and call `SafeRun(job.Run)` for each one. Because
`SafeRun` never panics, the loop always reaches the next job. Use
`err.Error()` to get the message text.

## Almost the answer

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

For task 2, the line inside the `if` is
`err = fmt.Errorf("%w: %v", ErrPanic, r)`.
