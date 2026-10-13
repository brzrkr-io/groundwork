## Task 1: one rune at a time

Loop with `for _, r := range name`. For each rune, decide: keep it (ASCII
letter or digit) or treat it as a separator. Turn `A`–`Z` into lower case
first: `r += 'a' - 'A'` moves an upper-case letter to its lower-case partner.

## Task 1: remember the gap

Don't write `-` the moment you see a separator. Instead, set a `gap := true`
flag. When you next keep a character, write `-` first, but only if the flag is
set **and** the builder already has something in it (`b.Len() > 0`). Then
clear the flag. This handles runs of separators and both ends at once.

## Almost the answer for task 1

```go
var b strings.Builder
gap := false
for _, r := range name {
	if 'A' <= r && r <= 'Z' {
		r += 'a' - 'A'
	}
	if ('a' <= r && r <= 'z') || ('0' <= r && r <= '9') {
		if gap && b.Len() > 0 {
			b.WriteRune('-')
		}
		b.WriteRune(r)
		gap = false
	} else {
		gap = true
	}
}
return b.String()
```

Add `"strings"` to the imports.

## Task 2: cut, trim, fall back

A slug is pure ASCII, so `label[:63]` cuts exactly 63 characters. Then
`strings.TrimSuffix(label, "-")` removes a trailing dash if there is one.
Finally, return `"unnamed"` if `label == ""`.
