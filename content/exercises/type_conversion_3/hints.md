## Task 1: convert, then divide

Convert `mib` to `float64` before dividing by 1024.

## Task 2: reuse task 1

```go
return strconv.FormatFloat(MiBToGiB(mib), 'f', 1, 64) + " GiB"
```

## Task 3: the percentage

Work out the percentage as a `float64`, then chop it to an `int` and turn it
into text:

```go
pct := float64(usedMiB) / float64(totalMiB) * 100
percent := strconv.Itoa(int(pct))
```

Then join `mount`, `": "`, `GiBString(usedMiB)`, `" of "`,
`GiBString(totalMiB)`, `" used ("`, `percent` and `"%)"`.
