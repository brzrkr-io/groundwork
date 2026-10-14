## Task 1: Find the slowest response

Inside `Max`, `values` is an ordinary `[]int`. Handle the empty case first, then keep track of the largest value you've seen. Since values can be negative, start from the first value, not from `0`.

### Closer

Return `0` when `len(values) == 0`. Then start with `largest := values[0]` and loop over the rest with `range values[1:]`, replacing `largest` whenever a value is bigger.

### Answer

```go
if len(values) == 0 {
	return 0
}
largest := values[0]
for _, v := range values[1:] {
	if v > largest {
		largest = v
	}
}
return largest
```

Starting from `values[0]` means an all-negative list still gives the right answer.

## Task 2: Join a path

`parts` is a `[]string`. Build the result in a string with `+=`, skip empty parts, and only put `"/"` between parts, never at the start.

### Closer

Use `continue` to skip a part that is `""`. Before adding a part, add `"/"` only if the result so far is not empty:

```go
if path != "" {
	path += "/"
}
path += p
```

### Answer

```go
path := ""
for _, p := range parts {
	if p == "" {
		continue
	}
	if path != "" {
		path += "/"
	}
	path += p
}
return path
```

The `"/"` is only added when something is already in `path`, so there's never one at the start or end.

## Task 3: Collect resource tags

You need a new slice that starts with `base` and then holds everything in `extra`. Remember that `append` is variadic too.

### Closer

Make a one-element slice with `[]string{base}`, then append the whole of `extra` by spreading it with `...`: `append(..., extra...)`.

### Answer

```go
return append([]string{base}, extra...)
```

`extra...` spreads the slice into separate values for `append`; with no extras it adds nothing.
