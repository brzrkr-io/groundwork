## Task 1: copy the -service line

Each flag is one line like the `-service` example, using `fs.Int`, `fs.Bool`
or `fs.Duration`. The default for `-timeout` is a `time.Duration`:
`30*time.Second`. Then read each pointer with `*` when you build `Options`.

## Task 1: the struct literal

```go
replicas := fs.Int("replicas", 1, "number of replicas")
// ...after fs.Parse:
opts := Options{
	Service:  *service,
	Replicas: *replicas,
	// DryRun and Timeout the same way
	Hosts:    fs.Args(),
}
```

## Task 2: print, show usage, return

```go
if *service == "" {
	fmt.Fprintln(stderr, "deploy: -service is required")
	fs.Usage()
	return Options{}, errors.New("-service is required")
}
```

Put this right after `fs.Parse`, and add `"errors"` to the imports.

## Task 3: help is not a failure

```go
if errors.Is(err, flag.ErrHelp) {
	return 0
}
return 2
```
