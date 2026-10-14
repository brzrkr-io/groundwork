## Task 1: Recognise production, however it is typed

`==` only matches the exact text. You need to drop the whitespace around `env`
and then compare without caring about upper or lower case.

### Closer

`strings.TrimSpace(env)` removes the spaces, tabs and newlines; compare the
result with `strings.EqualFold(..., "production")` instead of `==`.

### Answer

```go
return strings.EqualFold(strings.TrimSpace(env), "production")
```

`EqualFold` ignores case, and `TrimSpace` removes the noise first.

## Task 2: Keep the whole value of an env line

`strings.Split` cuts at every `"="`, so `parts[1]` only holds the text up to
the second `"="`. You want to cut at the first `"="` only.

### Closer

`key, value, found := strings.Cut(line, "=")` gives the key, everything after
the first `"="`, and whether one was found. Return `"", "", false` when `found`
is false.

### Answer

```go
key, value, found := strings.Cut(line, "=")
if !found {
	return "", "", false
}
return strings.TrimSpace(key), strings.TrimSpace(value), true
```

`Cut` keeps everything after the first `"="`, including any later `"="`.

## Task 3: Read `/etc/hosts` lines with any spacing

Look at the failing input printed with `%q`: `Split(line, " ")` turns extra
spaces into empty `""` words and doesn't split on tabs at all. Which function
splits on any run of whitespace?

### Closer

Change only the line that builds `words`: use `strings.Fields(line)` instead
of `strings.Split(line, " ")`.

### Answer

```go
line, _, _ = strings.Cut(line, "#")
words := strings.Fields(line)
if len(words) < 2 {
	return "", nil, false
}
return words[0], words[1:], true
```

`Fields` splits on runs of spaces, tabs and newlines and never returns empty
pieces.
