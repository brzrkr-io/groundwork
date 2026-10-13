## Check the length first

Everything starts with `len(args)`. If it isn't 3, print the usage line to
`stderr` and `return 2` straight away, before you touch `args[1]` or
`args[2]`.

## Print to the right writer

`fmt.Fprintf(stderr, "usage: %s SERVICE REPLICAS\n", args[0])` writes to
stderr. Use `stdout` as the first argument for the success line. Don't forget
the `\n` with `Fprintf`.

## Convert and validate

```go
replicas, err := strconv.Atoi(args[2])
if err != nil || replicas < 0 {
	fmt.Fprintf(stderr, "%s: replicas must be a whole number >= 0, got %q\n", args[0], args[2])
	return 2
}
```

Remember to add `"strconv"` to the imports.

## The success line

```go
fmt.Fprintf(stdout, "scaling %s to %d replicas\n", args[1], replicas)
return 0
```

Delete the `"not implemented yet"` line and the `return 1`.
