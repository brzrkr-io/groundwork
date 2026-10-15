## Task 1: Remove quotes from a value

Look at the first and the last byte of `value`. Handle "doesn't start with a
quote" first; that case returns `value` unchanged.

### Closer

`value[0]` is the first byte and `value[len(value)-1]` is the last. Compare
with `'"'` and `'\''`. Remember the starting quote, then check the length is
at least 2 and the last byte is the same quote. Slice the inside with
`value[1 : len(value)-1]`.

### Answer

```go
if value == "" || (value[0] != '"' && value[0] != '\'') {
	return value, nil
}
quote := value[0]
if len(value) < 2 || value[len(value)-1] != quote {
	return "", ErrUnterminatedQuote
}
return value[1 : len(value)-1], nil
```

The length check stops a lone `"` from counting as both the opening and the
closing quote.

## Task 2: Parse one line

Follow the four numbered steps in the README in order. The `strings` package
has functions for trimming, removing a prefix, and splitting at the first
`=`. Add `"strings"` and `"fmt"` to the imports.

### Closer

`strings.CutPrefix(strings.TrimSpace(line), "export ")` drops the prefix.
`key, value, found := strings.Cut(line, "=")` splits at the first `=`. Trim
`key` before the `validKey` check, wrap with
`fmt.Errorf("%w %q", ErrInvalidKey, key)`, and finish with
`Unquote(strings.TrimSpace(value))`.

### Answer

```go
line, _ = strings.CutPrefix(strings.TrimSpace(line), "export ")
key, value, found := strings.Cut(line, "=")
if !found {
	return "", "", ErrMissingEquals
}
key = strings.TrimSpace(key)
if !validKey.MatchString(key) {
	return "", "", fmt.Errorf("%w %q", ErrInvalidKey, key)
}
value, err = Unquote(strings.TrimSpace(value))
if err != nil {
	return "", "", err
}
return key, value, nil
```

## Task 3: Parse a whole file

Use a `bufio.Scanner` and a line counter. A bad line doesn't end the
function: save its error and keep scanning. Add `"bufio"` to the imports.

### Closer

Inside `for sc.Scan()`, do `n++`, trim `sc.Text()`, and `continue` on blank
lines or `strings.HasPrefix(line, "#")`. On a `ParseLine` error, append
`fmt.Errorf("line %d: %w", n, err)` to `errs` and `continue`. After the loop,
check `sc.Err()` first, then `errors.Join(errs...)` when `len(errs) > 0`.

### Answer

```go
vars := map[string]string{}
var errs []error
sc := bufio.NewScanner(r)
n := 0
for sc.Scan() {
	n++
	line := strings.TrimSpace(sc.Text())
	if line == "" || strings.HasPrefix(line, "#") {
		continue
	}
	key, value, err := ParseLine(line)
	if err != nil {
		errs = append(errs, fmt.Errorf("line %d: %w", n, err))
		continue
	}
	vars[key] = value
}
if err := sc.Err(); err != nil {
	return nil, err
}
if len(errs) > 0 {
	return nil, errors.Join(errs...)
}
return vars, nil
```

## Task 4: Build an environment list

You can't sort a map, but you can sort a slice of its keys. Collect the keys,
sort them, then build the strings in that order. Add `"slices"` to the
imports.

### Closer

Loop `for key := range vars` and append each key to a `[]string`. Call
`slices.Sort(keys)`. Then loop over `keys` and append `key + "=" + vars[key]`.

### Answer

```go
keys := make([]string, 0, len(vars))
for key := range vars {
	keys = append(keys, key)
}
slices.Sort(keys)

env := make([]string, 0, len(keys))
for _, key := range keys {
	env = append(env, key+"="+vars[key])
}
return env
```
