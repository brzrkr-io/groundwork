## Task 1: Print a size for humans

Check the biggest unit first, then smaller ones, and fall back to plain bytes. Don't forget to add the `fmt`, `strconv` and `strings` imports.

### Closer

A `switch` with no value runs the first `case` that is true. Divide as `float64`s:

```go
switch {
case b >= GiB:
	return fmt.Sprintf("%.1f GiB", float64(b)/float64(GiB))
// ... MiB, KiB, then default
}
```

### Answer

```go
switch {
case b >= GiB:
	return fmt.Sprintf("%.1f GiB", float64(b)/float64(GiB))
case b >= MiB:
	return fmt.Sprintf("%.1f MiB", float64(b)/float64(MiB))
case b >= KiB:
	return fmt.Sprintf("%.1f KiB", float64(b)/float64(KiB))
default:
	return fmt.Sprintf("%d B", b)
}
```

Checking `GiB` first makes sure the largest fitting unit wins.

## Task 2: Parse a quantity

Split the job in two: strip a known suffix (remembering its unit), then turn what's left into a number. Start by assuming there is no suffix and the unit is `1`.

### Closer

Try each suffix with `strings.CutSuffix`, then parse with `strconv.Atoi`:

```go
number, unit := s, Bytes(1)
if n, found := strings.CutSuffix(s, "Ki"); found {
	number, unit = n, KiB
} // else if ... "Mi", "Gi"
```

Return an error if `Atoi` fails or the number is below zero, and convert with `Bytes(n)` before multiplying.

### Answer

```go
number, unit := s, Bytes(1)
if n, found := strings.CutSuffix(s, "Ki"); found {
	number, unit = n, KiB
} else if n, found := strings.CutSuffix(s, "Mi"); found {
	number, unit = n, MiB
} else if n, found := strings.CutSuffix(s, "Gi"); found {
	number, unit = n, GiB
}

n, err := strconv.Atoi(number)
if err != nil {
	return 0, fmt.Errorf("invalid memory quantity %q", s)
}
if n < 0 {
	return 0, fmt.Errorf("negative memory quantity %q", s)
}
return Bytes(n) * unit, nil
```

An unknown suffix like `"2G"` stays in `number`, so `Atoi` fails and you return an error.

## Task 3: Check whether pods fit on a node

Keep a running total. Parse each request with `ParseMemory`, stop at the first error, and compare the total with `free` at the end.

### Closer

Start with `var total Bytes`, loop with `for _, r := range requests`, and `total += size` after each successful parse. On an error, `return false, err`.

### Answer

```go
var total Bytes
for _, r := range requests {
	size, err := ParseMemory(r)
	if err != nil {
		return false, err
	}
	total += size
}
return total <= free, nil
```

An empty list leaves `total` at `0`, which always fits.
