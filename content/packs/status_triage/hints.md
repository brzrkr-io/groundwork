## Task 1: Classify a status code

Each class is a range of codes, like 200 to 299. A tagless `switch` lets you
check one range per `case`, top to bottom, with a fallback for everything
else.

### Closer

Write `switch {` with cases like `case code >= 200 && code <= 299:` that
return the class name in quotes. Use `default:` to return `"unknown"`.

### Answer

```go
switch {
case code >= 100 && code <= 199:
	return "informational"
case code >= 200 && code <= 299:
	return "success"
case code >= 300 && code <= 399:
	return "redirect"
case code >= 400 && code <= 499:
	return "client error"
case code >= 500 && code <= 599:
	return "server error"
default:
	return "unknown"
}
```

## Task 2: Decide whether to retry

Only four exact codes mean "retry". A `switch` on `code` can list all four in
one `case`.

### Closer

Write `switch code {` with `case 429, 502, 503, 504:` returning `true`, and a
`default:` returning `false`.

### Answer

```go
switch code {
case 429, 502, 503, 504:
	return true
default:
	return false
}
```

## Task 3: Parse a status code from text

Three steps: trim the spaces, turn the text into a number, check the number is
in range. Stop and return an error as soon as a step fails.

### Closer

`code, err := strconv.Atoi(strings.TrimSpace(text))`, then
`if err != nil { return 0, err }`. Next check `code < 100 || code > 599` and
return `0` with a `fmt.Errorf(...)`. Add `strconv` and `strings` to the
imports.

### Answer

```go
code, err := strconv.Atoi(strings.TrimSpace(text))
if err != nil {
	return 0, err
}
if code < 100 || code > 599 {
	return 0, fmt.Errorf("status code %d is outside 100-599", code)
}
return code, nil
```

## Task 4: Write the triage line

Use your three functions: `ParseStatus` first, and stop with the "invalid"
message if it fails. Otherwise pick `"retry"` or `"no retry"` with
`ShouldRetry`, and get the class from `StatusClass`.

### Closer

In `fmt.Sprintf`, fixed text goes inside the quoted format string; the values
go after it. `%q` puts the original text in double quotes:
`fmt.Sprintf("%s: invalid status %q", service, statusText)`. For the normal
line, set `retry := "no retry"` and change it to `"retry"` when
`ShouldRetry(code)` is true.

### Answer

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
