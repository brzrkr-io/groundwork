## Task 1: count, then divide

Return 0 early when `len(checks) == 0`, so you never divide by zero. Range
over `checks`, count the `false` ones, then return `failed * 100 / len(checks)`.

## Task 2: three counters and a switch

Keep three `int` counters. Range over `results` and use a `switch` on each
result to bump the right counter. Finish with `fmt.Sprintf` and three `%d`
verbs (add `import "fmt"` at the top).

## Task 3: dots plus one

A name with no dots has 1 label; every dot adds one more. Handle `""` first,
then range over the string and compare each rune with `'.'`.

## Almost the answer

```go
labels := 1
for _, r := range name {
	if r == '.' {
		labels++
	}
}
return labels
```
