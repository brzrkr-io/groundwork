## Task 1: Print a usage line for the wrong number of arguments

Start with `len(args)`. If it isn't 3, print the usage line to `stderr` and
return `2` straight away, before you touch `args[1]` or `args[2]`.

### Closer

`fmt.Fprintf(stderr, ...)` writes to stderr, and `%s` fills in `args[0]`, the
program name. Remember the `\n` at the end.

### Answer

```go
if len(args) != 3 {
	fmt.Fprintf(stderr, "usage: %s SERVICE REPLICAS\n", args[0])
	return 2
}
```

## Task 2: Reject bad replica counts

`strconv.Atoi` gives you a number and an error. The count is bad if the error
isn't `nil` or if the number is below zero, and both cases print the same line.

### Closer

`replicas, err := strconv.Atoi(args[2])`, then check
`err != nil || replicas < 0`. `%q` prints the original text in double quotes.
Add `"strconv"` to the imports.

### Answer

```go
replicas, err := strconv.Atoi(args[2])
if err != nil || replicas < 0 {
	fmt.Fprintf(stderr, "%s: replicas must be a whole number >= 0, got %q\n", args[0], args[2])
	return 2
}
```

## Task 3: Print what you are scaling

Once the length check has passed, `args[1]` is the service and `args[2]` is
the replica count. The success line goes to `stdout`, not `stderr`.

### Closer

Use `fmt.Fprintf(stdout, "scaling %s to %d replicas\n", ...)`, then `return 0`.
Delete the `"not implemented yet"` line and the `return 1`.

### Answer

```go
fmt.Fprintf(stdout, "scaling %s to %d replicas\n", args[1], replicas)
return 0
```

`replicas` is the number you converted in task 2.
