## Task 1: Healthy percentage

Dividing two ints throws away the fraction. Turn both into `float64` before you divide, then times 100.

### Closer

Convert each one separately: `float64(healthy) / float64(total)`. Then multiply the result by `100`.

### Answer

```go
return float64(healthy) / float64(total) * 100
```

## Task 2: Listen address

First add `import "strconv"`. You can only join strings with `+`, so turn the port into a string first.

### Closer

`strconv.Itoa(port)` turns an int into a string. The colon is fixed text, so it goes in quotes: `host + ":" + ...`.

### Answer

```go
return host + ":" + strconv.Itoa(port)
```

## Task 3: Health label

Reuse `HealthyPercent` from task 1 to get the percentage. Then turn every number into a string and join the pieces.

### Closer

Use `strconv.Itoa` for the two counts and `strconv.FormatFloat(pct, 'f', 1, 64)` for one decimal. Fixed text like `"/"` and `" pods healthy ("` goes in quotes; the values don't.

### Answer

```go
pct := HealthyPercent(healthy, total)
return strconv.Itoa(healthy) + "/" + strconv.Itoa(total) +
	" pods healthy (" + strconv.FormatFloat(pct, 'f', 1, 64) + "%)"
```
