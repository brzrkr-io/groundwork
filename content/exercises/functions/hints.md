## Task 1: use the constant

`MiB` already holds the number of bytes in one mebibyte. Multiply.

## Task 2: convert, divide, format

```go
return fmt.Sprintf("%.1f %s", float64(bytes)/float64(unitSize), unit)
```

## Task 3: biggest unit first

Check the units from biggest to smallest, so a large number doesn't stop at
`KiB`:

```go
if bytes >= GiB {
	return FormatUnit(bytes, GiB, "GiB")
}
```

Repeat for `MiB` and `KiB`. For the last case, use
`strconv.Itoa(bytes) + " B"` or `fmt.Sprintf("%d B", bytes)`.
