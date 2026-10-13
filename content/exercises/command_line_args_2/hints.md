## Task 1: count from 1

For `waitport db 5432 10`, `args` is `["waitport", "db", "5432", "10"]`. HOST
is `args[1]`, PORT is `args[2]`, and TIMEOUT_SECONDS is `args[3]`. Don't
forget the `args[1]` inside the port error message.

## Task 2: two small changes

In the usage branch, change `stdout` to `stderr` in the `Fprintf` call, and
`return 1` to `return 2`.

## Task 3: default, then check

```go
timeout := 30
if len(args) == 4 {
	timeout, err = strconv.Atoi(args[3])
	if err != nil || timeout < 1 {
		fmt.Fprintf(stderr, "%s: timeout must be a whole number >= 1, got %q\n", prog, args[3])
		return 2
	}
}
```

It follows the same pattern as the port check above it.
