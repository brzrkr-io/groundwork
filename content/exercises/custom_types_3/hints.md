## Task 1: biggest unit first

A `switch` with no value after it checks each `case` in order and runs the
first one that is true. Check `b >= GiB` first, then `MiB`, then `KiB`, and
fall back to plain bytes:

```go
case b >= GiB:
	return fmt.Sprintf("%.1f GiB", float64(b)/float64(GiB))
```

## Task 2: strip the suffix, then parse

Start by assuming no suffix: the whole string is the number and the unit is
`Bytes(1)`. Then try each suffix with `strings.CutSuffix`. When one is
found, keep the shorter string and remember the unit:

```go
number, unit := s, Bytes(1)
if n, found := strings.CutSuffix(s, "Mi"); found {
	number, unit = n, MiB
}
```

Add `else if` branches for `Ki` and `Gi`.

## Task 2: from text to Bytes

`strconv.Atoi(number)` returns an `int` and an error. Return an error if it
fails or if the number is below zero. Otherwise convert before multiplying:

```go
return Bytes(n) * unit, nil
```

## Task 3: a running total

Start with `var total Bytes` (zero), loop over `requests`, and call
`ParseMemory` on each one. If it returns an error, return `false, err`
straight away. Otherwise `total += size`. After the loop,
`return total <= free, nil`.
