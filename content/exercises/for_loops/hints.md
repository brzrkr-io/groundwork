## Task 1: keep two variables

Keep a running `total` and the current `delay`. Start `delay` at
`baseSeconds`. Each round, add `delay` to `total`, then double `delay` with
`delay *= 2`.

## Task 2: two loops

Start with `var b strings.Builder` and `b.WriteString("[")`. One loop
`for range done` writes `"#"`, a second loop `for range total - done` writes
`"-"`. Finish with `b.WriteString("]")` and return `b.String()`. Remember
`import "strings"`.

## Task 3: loop while the next delay fits

```go
used, delay, retries := 0, baseSeconds, 0
for used+delay <= budgetSeconds {
	used += delay
	delay *= 2
	retries++
}
return retries
```

## Task 1: almost the answer

```go
total, delay := 0, baseSeconds
for range attempts {
	total += delay
	delay *= 2
}
return total
```
