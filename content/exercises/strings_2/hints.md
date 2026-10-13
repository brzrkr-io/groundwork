## Task 1: start or end?

`HasPrefix` checks the start of a string. The production domain is at the
**end** of a host name. Which function checks the end?

## Task 2: replace, then add

`strings.ReplaceAll(service, "-", "_")` swaps every dash. Add
`"_requests_total"` to its result with `+`.

## Task 3: lower-case the line

The search word is already lowercase. Lower-case the line before searching it:

```go
return strings.Contains(strings.ToLower(line), "timeout")
```
