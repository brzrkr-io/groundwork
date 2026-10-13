## Add the import

Below `package text`, add:

```go
import "strings"
```

## Task 1: nest the calls

Trim first, then lower-case the trimmed result:
`strings.ToLower(strings.TrimSpace(host))`.

## Task 2: return the answer

`strings.HasPrefix` already returns `true` or `false`, so return its result
directly. The prefix to check is `"registry.internal/"`.

## Task 3: almost the answer

```go
return strings.ToUpper(title) + "\n" + strings.Repeat("=", len(title))
```
