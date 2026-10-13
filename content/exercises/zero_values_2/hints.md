## Task 1: read the condition aloud

`path != ""` means "path is NOT empty". When should the default be used?

## Task 2: same pattern, float64

`if rate == 0 { return 0.1 }`, then return `rate`.

## Task 3: which variable gets assigned?

Inside the `keepFiles` check, look at the variable on the left of `=`. Then
add a check of the same shape for `maxSizeMiB`:

```go
if maxSizeMiB == 0 {
	maxSizeMiB = 100
}
if keepFiles == 0 {
	keepFiles = 5
}
```
