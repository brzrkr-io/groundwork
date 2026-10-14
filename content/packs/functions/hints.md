## Task 1: Convert MiB to bytes

The starter already gives you the constant `MiB`, which holds the number of
bytes in one mebibyte. How do you get the bytes in `mib` of them?

### Closer

Multiply the parameter by the constant: `mib * ...`.

### Answer

```go
return mib * MiB
```

## Task 2: Format a value in a unit

Dividing two `int`s drops the fraction, so convert both to `float64` before
you divide. Then format the result with `fmt.Sprintf`.

### Closer

Divide with `float64(bytes)/float64(unitSize)`. In the format, `%.1f` shows
one decimal place and `%s` stands for the unit. The space between them is
fixed text, so it goes inside the quotes: `"%.1f %s"`. The values go after
the format.

### Answer

```go
return fmt.Sprintf("%.1f %s", float64(bytes)/float64(unitSize), unit)
```

## Task 3: Pick the right unit

Check the units from biggest to smallest, so a large number doesn't stop at
`KiB`. For each unit, call your `FormatUnit` from task 2.

### Closer

Write one `if` per unit, starting with GiB:

```go
if bytes >= GiB {
	return FormatUnit(bytes, GiB, "GiB")
}
```

Repeat for `MiB` and `KiB`. If none matched, return the number followed by
the fixed text `" B"`: `fmt.Sprintf("%d B", bytes)`.

### Answer

```go
if bytes >= GiB {
	return FormatUnit(bytes, GiB, "GiB")
}
if bytes >= MiB {
	return FormatUnit(bytes, MiB, "MiB")
}
if bytes >= KiB {
	return FormatUnit(bytes, KiB, "KiB")
}
return fmt.Sprintf("%d B", bytes)
```
