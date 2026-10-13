## Task 1: convert both, then divide

Convert `healthy` and `total` to `float64` separately, divide, then multiply
by 100:

```go
return float64(healthy) / float64(total) * 100
```

## Task 2: Itoa

After adding `import "strconv"`, join `host`, `":"` and
`strconv.Itoa(port)` with `+`.

## Task 3: reuse task 1

Call `HealthyPercent(healthy, total)` and format its result with
`strconv.FormatFloat(pct, 'f', 1, 64)`. Use `strconv.Itoa` for the two
counts.

## Almost the answer

```go
pct := HealthyPercent(healthy, total)
return strconv.Itoa(healthy) + "/" + strconv.Itoa(total) +
	" pods healthy (" + strconv.FormatFloat(pct, 'f', 1, 64) + "%)"
```
