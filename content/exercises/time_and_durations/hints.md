## Task 1: parse, subtract, divide

`time.Parse(time.RFC3339, notAfter)` gives you a `time.Time`. `expiry.Sub(now)`
gives the `time.Duration` left. Dividing that by `24 * time.Hour` counts whole
days.

## Task 2: days, then leftover hours

Get the uptime with `now.Sub(booted)`. Whole days are `up / day`. The leftover
is `up % day`, and dividing the leftover by `time.Hour` gives whole hours.
Wrap each count in `int(...)` before printing it with `%d`.

## Task 3: two comparisons

Parse both strings first and return early on any error. The end of the window
is `begin.Add(length)`. "At or after start" is `!now.Before(begin)`, and
"before end" is `now.Before(end)`.

## Almost the answer

```go
const day = 24 * time.Hour

up := now.Sub(booted)
days := int(up / day)
hours := int(up % day / time.Hour)
return fmt.Sprintf("up %dd%dh since %s", days, hours,
	booted.UTC().Format("2006-01-02 15:04 MST"))
```

For task 3: `return !now.Before(begin) && now.Before(begin.Add(length)), nil`.
