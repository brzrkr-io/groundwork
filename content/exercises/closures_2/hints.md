## Task 1: one line moves

`count := 0` runs every time the returned function runs. Move it above the
`return func() int {` line.

## Task 2: same idea, then format

Keep an `n := 0` outside, add 1 inside, and return
`prefix + "-" + strconv.Itoa(n)`.

## Task 3: count calls, compare with limit

```go
calls := 0
return func() bool {
	calls++
	return calls <= limit
}
```
