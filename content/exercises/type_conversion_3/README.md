# Type conversion: report disk usage

Disk sizes are counted in whole MiB, but people read them in GiB with a
decimal place. You'll build a disk usage line like
`/var: 1.5 GiB of 2.0 GiB used (75%)` out of three small functions.

## Learn

### Recap

`float64(n)` turns an `int` into a `float64`, and `int(f)` chops the
fraction off a `float64`. `strconv.Itoa(n)` turns an `int` into text, and
`strconv.FormatFloat(f, 'f', 1, 64)` turns a `float64` into text with one
decimal place.

One GiB is 1024 MiB.

## Your task

Write the three functions. Each one may call the ones before it.

### 1. MiB to GiB

Write `MiBToGiB` so it converts a whole number of MiB into GiB, keeping the
fraction:

```go
MiBToGiB(1536) // => 1.5
MiBToGiB(2048) // => 2
MiBToGiB(512)  // => 0.5
```

### 2. GiB as text

Write `GiBString` so it returns the size in GiB with exactly one decimal
place, a space, then `GiB`. Add `import "strconv"` first.

```go
GiBString(1536) // => "1.5 GiB"
GiBString(3072) // => "3.0 GiB"
GiBString(1000) // => "1.0 GiB"   (0.9765625 rounds to 1.0)
```

### 3. Disk usage line

Write `DiskLine` so it returns the mount point, a colon and a space, the used
and total sizes from `GiBString`, and the used percentage as a whole number
in parentheses:

```text
<mount>: <used> of <total> used (<percent>%)
```

The percentage is used divided by total, times 100, with the fraction
**chopped off**, not rounded: 66.67 becomes 66.

```go
DiskLine("/var", 1536, 2048)
// => "/var: 1.5 GiB of 2.0 GiB used (75%)"

DiskLine("/data", 2048, 3072)
// => "/data: 2.0 GiB of 3.0 GiB used (66%)"

DiskLine("/", 1000, 1000)
// => "/: 1.0 GiB of 1.0 GiB used (100%)"
```

`totalMiB` is never 0.

## Learn more

- [Package strconv](https://pkg.go.dev/strconv)
- [Go spec: Conversions between numeric types](https://go.dev/ref/spec#Conversions_between_numeric_types)
- [Wikipedia: Byte multiples (MiB, GiB)](https://en.wikipedia.org/wiki/Byte#Multiple-byte_units)
