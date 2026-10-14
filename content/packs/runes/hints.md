## Task 1: Count characters

`len(s)` counts bytes, but `é` takes 2 bytes and `🚀` takes 4. Look in the
`unicode/utf8` package for a function that counts runes instead.

### Closer

Add `import "unicode/utf8"` under the `package` line, then call
`utf8.RuneCountInString(...)` with the string.

### Answer

```go
import "unicode/utf8"

func CharCount(s string) int {
	return utf8.RuneCountInString(s)
}
```

`RuneCountInString` decodes the UTF-8 and counts whole characters, not bytes.

## Task 2: Truncate a chat message

Slicing a string with `msg[:n]` cuts bytes and can split an emoji in half.
Convert the string to something you can cut by character first.

### Closer

`[]rune(msg)` gives one element per character, so `len` and `[:n]` count
characters. `string(runes[:n])` turns the first `n` runes back into a string:

```go
runes := []rune(msg)
if len(runes) <= n {
	return msg
}
```

### Answer

```go
runes := []rune(msg)
if len(runes) <= n {
	return msg
}
return string(runes[:n]) + "..."
```

Working on runes means the cut always falls between whole characters.

## Task 3: Validate a label value

Start with the helper: runes are numbers, and `'a'` to `'z'` sit next to each
other, so you can compare a rune against a range. Then check the length, every
character, and the first and last character in `ValidLabel`.

### Closer

`'a' <= r && r <= 'z'` checks for a lower-case ASCII letter; join three such
checks with `||` in `isAlphaNum`. In `ValidLabel`, convert to `[]rune`, loop
with `for _, r := range runes`, and reject any rune that is not
`isAlphaNum(r)`, `'-'`, `'_'` or `'.'`. The last character is
`runes[len(runes)-1]`.

### Answer

```go
func isAlphaNum(r rune) bool {
	return ('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z') || ('0' <= r && r <= '9')
}

func ValidLabel(s string) bool {
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
}
```

The length check comes first, so `runes[0]` is safe to read at the end.
