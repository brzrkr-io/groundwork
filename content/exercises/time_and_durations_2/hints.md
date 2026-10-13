## Task 1: one piece per field

The layout needs the reference time's year, month, day, hour and minute:
`2006`, `01`, `02`, `15`, `04`. Keep the `-` between the date and the time.

## Task 2: convert, then multiply by the unit

`time.Duration(attempt)` is a count of nanoseconds. Multiply it by
`30 * time.Second` to get the right unit. The cap check can stay as it is.

## Task 3: almost the answer

```go
func NextRun(now time.Time, every time.Duration) time.Time {
	return now.Truncate(every).Add(every)
}
```
