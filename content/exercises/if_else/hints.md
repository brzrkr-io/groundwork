## Task 1: check the biggest number first

Go runs only the first branch whose condition is true. Check `>= 90` first,
then `>= 80`, and let everything else be `"ok"`.

## Task 2: two early returns

```go
if requested < 1 {
	return 1
}
```

Add a second `if` for "above 10", then `return requested` at the end.

## Task 3: guard first, then the short statement

Return `"no replicas"` when `replicas < 1` before you divide. Then write
`if per := totalRequests / replicas; per > 100 {` and continue with
`else if per < 20` and `else`.

## Almost the answer

```go
if replicas < 1 {
	return "no replicas"
}
if per := totalRequests / replicas; per > 100 {
	return "scale up"
} else if per < 20 {
	return "scale down"
} else {
	return "steady"
}
```
