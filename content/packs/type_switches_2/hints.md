## Task 1: Fix GetBool

The assertion already returns a second value that says whether `cfg[key]` really held a `bool`. The starter throws it away and returns `true` instead.

### Closer

Replace the `_` with a name such as `ok`, and return it instead of the hard-coded `true`.

### Answer

```go
b, ok := cfg[key].(bool)
return b, ok
```

A missing key gives `nil`, and `nil` is not a `bool`, so `ok` is `false` there too.

## Task 2: Finish GetInt

Add a `case float64:` and a `case string:` to the existing switch. Each one has to check that the value really is a whole number before returning it.

### Closer

Inside `case float64:`, `v` is a `float64`. Convert it with `n := int(v)`, then check `float64(n) == v`. Inside `case string:`, `strconv.Atoi(v)` returns the number and an error; a non-nil error means "not a number". Remember to add `import "strconv"`.

### Answer

```go
case float64:
	n := int(v)
	if float64(n) != v {
		return 0, false
	}
	return n, true
case string:
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, false
	}
	return n, true
```

Converting back to `float64` and comparing reveals any fractional part that `int(v)` dropped.

## Task 3: Finish GetStrings

Add a `case []any:` that builds a new `[]string`. Each element is an `any`, so you have to check that each one is a string.

### Closer

Loop over `v` and use a comma-ok assertion on each element: `s, ok := item.(string)`. Return `nil, false` as soon as one is not a string, otherwise `append` it to your list.

### Answer

```go
case []any:
	list := make([]string, 0, len(v))
	for _, item := range v {
		s, ok := item.(string)
		if !ok {
			return nil, false
		}
		list = append(list, s)
	}
	return list, true
```

Starting from an empty (non-nil) slice means an empty `[]any{}` gives an empty list and `true`.
