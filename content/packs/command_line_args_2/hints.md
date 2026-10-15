## Task 1: Fix the off-by-one

For `waitport db 5432 10`, `args` is `["waitport", "db", "5432", "10"]`. Which
index holds HOST, and which holds PORT?

### Closer

HOST is `args[1]`, PORT is `args[2]`, and TIMEOUT_SECONDS is `args[3]`. Fix
the `args[1]` inside the port error message too.

### Answer

```go
host := args[1]
port, err := strconv.Atoi(args[2])
if err != nil || port < 1 || port > 65535 {
	fmt.Fprintf(stderr, "%s: port must be 1-65535, got %q\n", prog, args[2])
	return 2
}
```

## Task 2: Send usage errors to stderr with exit code 2

Look at the usage branch: which writer does it print to, and which number does
it return?

### Closer

Two small changes: the first argument of `Fprintf` becomes `stderr`, and
`return 1` becomes `return 2`.

### Answer

```go
if len(args) < 3 || len(args) > 4 {
	fmt.Fprintf(stderr, "usage: %s HOST PORT [TIMEOUT_SECONDS]\n", prog)
	return 2
}
```

## Task 3: Default and check the timeout

Start `timeout` at the default, and only replace it when a fourth argument is
there. The `_` throws away the conversion error, so "soon" quietly becomes `0`.

### Closer

Set `timeout := 30`. Inside `if len(args) == 4`, assign with
`timeout, err = strconv.Atoi(args[3])` and check `err != nil || timeout < 1`,
the same way the port check does.

### Answer

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
