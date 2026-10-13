## Task 1: MiB to GiB

There are 1024 MiB in one GiB, so you divide by 1024. But `mib` is an `int`,
and dividing two `int`s throws the fraction away. Turn it into a `float64`
first.

### Closer

`float64(mib)` gives you the same number as a `float64`. Divide that by
`1024` and return the result.

### Answer

```go
return float64(mib) / 1024
```

Converting before dividing keeps the fraction, so 1536 gives 1.5.

## Task 2: GiB as text

Reuse `MiBToGiB` from task 1 to get the number, then turn it into text with
one decimal place. Don't forget to add `import "strconv"` at the top.

### Closer

`strconv.FormatFloat(f, 'f', 1, 64)` turns a `float64` into text with one
decimal. Then add the fixed text `" GiB"` with `+`. Fixed text goes inside
quotes; the function call doesn't.

### Answer

```go
return strconv.FormatFloat(MiBToGiB(mib), 'f', 1, 64) + " GiB"
```

## Task 3: Disk usage line

Two jobs: work out the percentage, then glue the line together. For the
sizes, call `GiBString` from task 2 instead of formatting them again.

### Closer

Compute the percentage as a `float64`, then chop it with `int(...)` and turn
it into text with `strconv.Itoa`:

```go
pct := float64(usedMiB) / float64(totalMiB) * 100
```

Then join the pieces with `+`. Fixed text goes inside quotes (`": "`,
`" of "`, `" used ("`, `"%)"`); parameters and function calls don't
(`mount`, `GiBString(usedMiB)`).

### Answer

```go
pct := float64(usedMiB) / float64(totalMiB) * 100
return mount + ": " + GiBString(usedMiB) + " of " + GiBString(totalMiB) +
	" used (" + strconv.Itoa(int(pct)) + "%)"
```

`int(pct)` drops the fraction, so 66.67 becomes 66.
