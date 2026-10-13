## Task 1: use the ok value

The starter throws away the second result of the assertion with `_`. Give it
a name and return it instead of the hard-coded `true`.

## Task 2: a float64 and a string case

Inside `case float64:`, `v` is a `float64`. Convert it with `int(v)`, then
check that `float64(n) == v`. Inside `case string:`, `strconv.Atoi(v)` returns
the number and an error; a non-nil error means "not a number".

## Task 3: check each element

Inside `case []any:`, loop over `v`. Each element is an `any`, so use a
comma-ok assertion on it: `s, ok := item.(string)`. Return `nil, false` as
soon as one is not a string, otherwise `append` it to your new list.

## Almost the answer

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

For `GetStrings`, start with `list := []string{}`, and after the loop
`return list, true`.
