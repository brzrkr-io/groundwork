## Task 1: one switch, five ranges

Use a tagless `switch` with one case per range, for example
`case code >= 200 && code <= 299:`. Put `"unknown"` in `default`.

## Task 2: list the codes in one case

```go
switch code {
case 429, 502, 503, 504:
	return true
}
return false
```

## Task 3: trim, convert, check

```go
code, err := strconv.Atoi(strings.TrimSpace(text))
if err != nil {
	return 0, err
}
if code < 100 || code > 599 {
	return 0, fmt.Errorf("status %d out of range", code)
}
return code, nil
```

## Task 4: reuse your functions

```go
code, err := ParseStatus(statusText)
if err != nil {
	return fmt.Sprintf("%s: invalid status %q", service, statusText)
}
retry := "no retry"
if ShouldRetry(code) {
	retry = "retry"
}
return fmt.Sprintf("%s: %d %s, %s", service, code, StatusClass(code), retry)
```
