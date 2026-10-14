## Task 1: Slugify a display name

Loop over the name rune by rune and decide for each one: keep it (ASCII letter
or digit) or treat it as a separator. Don't write `-` the moment you see a
separator; remember that you saw one instead.

### Closer

Turn `A`–`Z` into lower case with `r += 'a' - 'A'`. Keep a `gap` flag: set it
to `true` on a separator. When you next keep a character, write `-` first only
if `gap` is set **and** `b.Len() > 0`, then clear the flag. Add `"strings"` to
the imports.

```go
var b strings.Builder
gap := false
for _, r := range name {
	// lower-case, then keep or set gap
}
return b.String()
```

### Answer

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

Writing the `-` only before a kept character handles runs of separators and
both ends at once.

## Task 2: Make a valid DNS label

Start from `Slug(name)` and apply the three rules in order: cut, drop a
trailing `-`, fall back to `"unnamed"`. A slug is pure ASCII, so cutting bytes
is safe here.

### Closer

`label[:63]` keeps the first 63 characters. `strings.TrimSuffix(label, "-")`
removes a trailing dash if there is one and does nothing otherwise.

### Answer

```go
label := Slug(name)
if len(label) > 63 {
	label = label[:63]
}
label = strings.TrimSuffix(label, "-")
if label == "" {
	return "unnamed"
}
return label
```

Trimming after the cut catches the case where the 63rd character is a `-`.
