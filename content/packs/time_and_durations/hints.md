## Task 1: Days until a certificate expires

Turn the string into a `time.Time` first, then work out how long it is from
`now` until then. Dividing a duration by a duration gives a count.

### Closer

`expiry, err := time.Parse(time.RFC3339, notAfter)` parses the string, and
`expiry.Sub(now)` is the `time.Duration` left. Divide that by
`24 * time.Hour` and wrap it in `int(...)`.

### Answer

```go
const day = 24 * time.Hour

expiry, err := time.Parse(time.RFC3339, notAfter)
if err != nil {
	return 0, err
}
return int(expiry.Sub(now) / day), nil
```

Integer division truncates toward zero, which is exactly the rounding the
README asks for.

## Task 2: Format an uptime line

Get the uptime as a duration, then split it into whole days and the hours left
over after those days.

### Closer

`up := now.Sub(booted)`. Whole days are `up / day`; the leftover is
`up % day`, and dividing that by `time.Hour` gives whole hours. Format the
boot time with `booted.UTC().Format("2006-01-02 15:04 MST")`, and add
`"fmt"` to the imports.

### Answer

```go
const day = 24 * time.Hour

up := now.Sub(booted)
days := int(up / day)
hours := int(up % day / time.Hour)
return fmt.Sprintf("up %dd%dh since %s", days, hours,
	booted.UTC().Format("2006-01-02 15:04 MST"))
```

`%` keeps only what is left after the whole days are taken out.

## Task 3: Check a maintenance window

Parse both strings first and return early on any error. Then check the two
edges of the window separately.

### Closer

`time.Parse(time.RFC3339, start)` and `time.ParseDuration(length)` give you
the start and the length. The end is `begin.Add(d)`. "At or after start" is
`!now.Before(begin)`; "before end" is `now.Before(end)`.

### Answer

```go
begin, err := time.Parse(time.RFC3339, start)
if err != nil {
	return false, err
}
d, err := time.ParseDuration(length)
if err != nil {
	return false, err
}
return !now.Before(begin) && now.Before(begin.Add(d)), nil
```

`!Before` includes the start; `Before` on the end excludes it.
