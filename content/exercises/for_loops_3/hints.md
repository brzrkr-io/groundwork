## Task 1: step through the batch starts

Loop `for first := 1; first <= hosts; first += batchSize`. The last host of
the batch is `first + batchSize - 1`, but never more than `hosts`.

## Task 1: almost the answer

```go
var b strings.Builder
for first := 1; first <= hosts; first += batchSize {
	last := min(first+batchSize-1, hosts)
	if first > 1 {
		b.WriteString(",")
	}
	b.WriteString(strconv.Itoa(first))
	b.WriteString("-")
	b.WriteString(strconv.Itoa(last))
}
return b.String()
```

Import both `"strconv"` and `"strings"` at the top of the file.

## Task 2: cap with min

Loop `for i := range attempts`. Keep a `delay` that starts at `baseSeconds`.
Each round, write `min(delay, maxSeconds)`, then `delay *= 2`. Add `" "`
before every delay except the first (`if i > 0`).
