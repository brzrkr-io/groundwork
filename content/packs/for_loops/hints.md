## Task 1: Add up the backoff delays

You need two variables: a running `total` that starts at `0`, and the
current `delay` that starts at `baseSeconds`. Repeat `attempts` times.

### Closer

Use `for range attempts {`. In each round, add `delay` to `total` with
`total += delay`, then double it with `delay *= 2`. Return `total` after the
loop.

### Answer

```go
total, delay := 0, baseSeconds
for range attempts {
	total += delay
	delay *= 2
}
return total
```

## Task 2: Draw a progress bar

Build the bar in four pieces: `[`, the `#` characters, the `-` characters,
then `]`. Use a `strings.Builder` to collect them, and a loop for each run of
repeated characters.

### Closer

Add `import "strings"` below the `package` line. Start with
`var b strings.Builder` and `b.WriteString("[")`. One loop `for range done`
writes `"#"`; a second loop `for range total - done` writes `"-"`. Finish
with `b.WriteString("]")` and return `b.String()`.

### Answer

```go
var b strings.Builder
b.WriteString("[")
for range done {
	b.WriteString("#")
}
for range total - done {
	b.WriteString("-")
}
b.WriteString("]")
return b.String()
```

## Task 3: Count the retries that fit in a budget

Keep three variables: the time `used` so far, the next `delay`, and the
number of `retries`. Before each retry, ask: would adding the next delay
still fit in the budget?

### Closer

That question is the loop condition: `for used+delay <= budgetSeconds {`.
Inside, add `delay` to `used`, double `delay`, and add one to `retries`.

### Answer

```go
used, delay, retries := 0, baseSeconds, 0
for used+delay <= budgetSeconds {
	used += delay
	delay *= 2
	retries++
}
return retries
```
