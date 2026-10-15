## Task 1: Name the release

First clean up the service name: remove the spaces at both ends, then make
it lowercase. Then add a space and the version. Remember to add
`import "strings"`.

### Closer

`strings.TrimSpace(service)` removes the spaces, and `strings.ToLower(...)`
lowercases, so put one call inside the other. Join with `+`: the space is
fixed text in quotes (`" "`); `version` is a parameter, so it has none.

### Answer

```go
return strings.ToLower(strings.TrimSpace(service)) + " " + version
```

## Task 2: Size the deploy

This is plain division, except the answer must round **up**: 1000 requests
at 300 each needs 4 replicas, not 3. Remember the ceiling trick from
*Numbers and arithmetic*.

### Closer

`(a + b - 1) / b` rounds `a / b` up for whole numbers. Here `a` is `peakRPS`
and `b` is `rpsPerReplica`.

### Answer

```go
return (peakRPS + rpsPerReplica - 1) / rpsPerReplica
```

## Task 3: Write the summary line

Use one `fmt.Sprintf` call with the README's line as the format. Each `<...>`
slot becomes a verb, and the values follow in the same order. Add
`import "fmt"`.

### Closer

Fixed text stays inside the quoted format; the values go after it. Use `%s`
for text (the release and `env`) and `%d` for whole numbers:

```go
fmt.Sprintf("%s -> %s: %d replicas x %d MiB = %d MiB", ...)
```

The release is `Release(service, version)` and the total is
`replicas*memoryMiB`.

### Answer

```go
return fmt.Sprintf("%s -> %s: %d replicas x %d MiB = %d MiB",
	Release(service, version), env, replicas, memoryMiB, replicas*memoryMiB)
```

## Task 4: Estimate the monthly cost

Follow the README's four steps one at a time. The tricky step is the
division: convert to `float64` before dividing by 1024, or you lose the
fraction.

### Closer

`float64(replicas*memoryMiB) / 1024` gives the GiB. Multiply by
`dollarsPerGiB`, then format with `%.2f` for two decimals. `$` and `/month`
are fixed text, so they go inside the quoted format: `"$%.2f/month"`.

### Answer

```go
gib := float64(replicas*memoryMiB) / 1024
return fmt.Sprintf("$%.2f/month", gib*dollarsPerGiB)
```

`$` has no special meaning in a format string, so it is printed as is.
