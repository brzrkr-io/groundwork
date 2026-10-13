## Task 1: Get the host from a URL

First remove the scheme from the front, then keep only the part before the
first `"/"`. The `strings` table in the README has a function for each step.

### Closer

`strings.TrimPrefix` does nothing when the prefix isn't there, so call it once
for `"https://"` and once for `"http://"`. Then `strings.Cut(url, "/")` returns
the part before the first `"/"` as its first result.

### Answer

```go
url = strings.TrimPrefix(url, "https://")
url = strings.TrimPrefix(url, "http://")
host, _, _ := strings.Cut(url, "/")
return host
```

Only one `TrimPrefix` can match, and `Cut` returns the whole string when there
is no `"/"`.

## Task 2: Parse a label list

Loop over the comma-separated pairs, split each pair at its first `"="`, and
store the trimmed key and value in `labels`. Pairs without `"="` are skipped.

### Closer

Loop with `for pair := range strings.SplitSeq(s, ",")`. Inside,
`key, value, found := strings.Cut(pair, "=")`; if `!found`, `continue`.
Trim both sides with `strings.TrimSpace`.

### Answer

```go
for pair := range strings.SplitSeq(s, ",") {
	key, value, found := strings.Cut(pair, "=")
	if !found {
		continue
	}
	labels[strings.TrimSpace(key)] = strings.TrimSpace(value)
}
```

`Cut` splits at the first `"="` only, so `"selector=a=b"` keeps `"a=b"` as the
value.

## Task 3: Build a command line

Every argument needs a space in front of it. Arguments that contain a space
also need a double quote on each side.

### Closer

Inside the loop, write `" "` first. Then check `strings.Contains(arg, " ")`;
if true, change `arg` to `"\"" + arg + "\""` before writing it.

### Answer

```go
for _, arg := range args {
	b.WriteString(" ")
	if strings.Contains(arg, " ") {
		arg = "\"" + arg + "\""
	}
	b.WriteString(arg)
}
```

Writing the space before each argument means no trailing space, and no space at
all when there are no arguments.
