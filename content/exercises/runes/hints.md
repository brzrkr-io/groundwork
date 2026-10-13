## Bytes or characters?

`len(s)` counts bytes. Every task here needs characters (runes) instead. For
counting, use `utf8.RuneCountInString`. For cutting, convert to `[]rune` first.

## Task 2: cut runes, not bytes

```go
runes := []rune(msg)
if len(runes) <= n {
	return msg
}
```

Then return the first `n` runes as a string, plus `"..."`. `string(runes[:n])`
turns a rune slice back into a string.

## Task 3: compare rune ranges

Runes are numbers, so `'a' <= r && r <= 'z'` checks for a lower-case ASCII
letter. Combine three such checks with `||` in your helper:

```go
func isAlphaNum(r rune) bool {
	return ('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z') || ('0' <= r && r <= '9')
}
```

In `ValidLabel`, convert to `[]rune`, check the length, loop over the runes to
reject anything that is not `isAlphaNum(r)`, `'-'`, `'_'` or `'.'`, then check
`runes[0]` and `runes[len(runes)-1]`.

## Almost the answer for task 3

```go
runes := []rune(s)
if len(runes) < 1 || len(runes) > 63 {
	return false
}
for _, r := range runes {
	if !isAlphaNum(r) && r != '-' && r != '_' && r != '.' {
		return false
	}
}
return isAlphaNum(runes[0]) && isAlphaNum(runes[len(runes)-1])
```
