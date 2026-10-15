## Task 1: Define the remaining flags

Each flag is one line like the `-service` example, using the function for its
type. Each returns a pointer, which you read with `*` when you build `Options`.

### Closer

Use `fs.Int`, `fs.Bool` and `fs.Duration`. The `-timeout` default is a
`time.Duration`: `30*time.Second`. The leftover arguments are `fs.Args()`.

### Answer

```go
replicas := fs.Int("replicas", 1, "number of replicas")
dryRun := fs.Bool("dry-run", false, "print the plan, change nothing")
timeout := fs.Duration("timeout", 30*time.Second, "how long to wait for the rollout")

// ...after fs.Parse:
opts := Options{
	Service:  *service,
	Replicas: *replicas,
	DryRun:   *dryRun,
	Timeout:  *timeout,
	Hosts:    fs.Args(),
}
```

## Task 2: Require -service

After `fs.Parse`, check whether `*service` is empty. If it is, do three things:
print the message, show the usage text, and return an error.

### Closer

`fmt.Fprintln(stderr, ...)` prints the line, `fs.Usage()` prints the flag list,
and `errors.New(...)` makes the error. Add `"errors"` to the imports.

### Answer

```go
if *service == "" {
	fmt.Fprintln(stderr, "deploy: -service is required")
	fs.Usage()
	return Options{}, errors.New("-service is required")
}
```

## Task 3: Return the right exit code

Asking for help with `-h` makes `Parse` return an error too, but it isn't a
failure. Tell that error apart from the rest.

### Closer

`errors.Is(err, flag.ErrHelp)` is true for `-h` and `--help`. Return `0` in
that case and `2` otherwise.

### Answer

```go
if err != nil {
	if errors.Is(err, flag.ErrHelp) {
		return 0
	}
	return 2
}
```
