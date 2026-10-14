## Task 1: Plan the rollout batches

Each batch starts at a host number: 1, then `1 + batchSize`, and so on. Loop
over those start numbers, and for each one work out the batch's last host.

### Closer

Loop with `for first := 1; first <= hosts; first += batchSize {`. The last
host is `first + batchSize - 1`, but never more than `hosts`, so use
`min(first+batchSize-1, hosts)`. Write `","` before every batch except the
first (`if first > 1`). The `"-"` is fixed text, so it goes in quotes;
`first` and `last` are numbers, so turn them into text with `strconv.Itoa`.
Remember to import `"strconv"` and `"strings"`.

### Answer

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

## Task 2: Plan the retry delays

Repeat `attempts` times. Keep a `delay` that starts at `baseSeconds` and
doubles each round, but what you *write* is never bigger than `maxSeconds`.

### Closer

Loop `for i := range attempts`. Write `" "` before every delay except the
first (`if i > 0`). Then write `strconv.Itoa(min(delay, maxSeconds))`,
followed by the fixed text `"s"` in quotes. Finish the round with
`delay *= 2`.

### Answer

```go
var b strings.Builder
delay := baseSeconds
for i := range attempts {
	if i > 0 {
		b.WriteString(" ")
	}
	b.WriteString(strconv.Itoa(min(delay, maxSeconds)))
	b.WriteString("s")
	delay *= 2
}
return b.String()
```
