## Task 1: trim, lower, join

`strings.ToLower(strings.TrimSpace(service))` cleans the name. Then add
`" "` and `version` with `+`.

## Task 2: the ceiling trick

As in *Numbers and arithmetic*: `(a + b - 1) / b` rounds `a / b` up.

## Task 3: one Sprintf

Use `%s` for the release and environment and `%d` for the three numbers:

```go
fmt.Sprintf("%s -> %s: %d replicas x %d MiB = %d MiB",
	Release(service, version), env, replicas, memoryMiB, replicas*memoryMiB)
```

## Task 4: convert before dividing

```go
gib := float64(replicas*memoryMiB) / 1024
return fmt.Sprintf("$%.2f/month", gib*dollarsPerGiB)
```

`$` has no special meaning in a format string, so it is printed as is.
