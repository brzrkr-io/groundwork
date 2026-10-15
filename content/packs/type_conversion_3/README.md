# Type conversion: report disk usage

Disk sizes are counted in whole MiB, but people read them in GiB with a
decimal place. You'll build a disk usage line like
`/var: 1.5 GiB of 2.0 GiB used (75%)` out of three small functions.

**New here:** nothing; practice combining conversions and your own functions.

## Learn

### Recap

- `float64(n)` turns an `int` into a `float64`; `int(f)` chops the fraction
  off a `float64`.
- `strconv.Itoa(n)` turns an `int` into text.
- `strconv.FormatFloat(f, 'f', 1, 64)` turns a `float64` into text with one
  decimal place.
- You can call your own functions from other functions, passing values in
  `( )` without types.

One GiB is 1024 MiB.

### Building on your own functions

Each function here can call the ones before it. That way the GiB maths lives
in one place, and the text formatting in one place. If you fix a bug in
`MiBToGiB`, everything that calls it is fixed too.

> **Why is it written like this?** **Your choice:** you could repeat the
> maths in every function, and it would still compile. **Convention:** small
> functions that build on each other are easier to test and read.

## Your task

Write the three functions. Each one may call the ones before it.

Each function starts with a stand-in `return 0` or `return ""`. Replace it
with your code: a stand-in left below your own `return` makes `go vet` report
`unreachable code`.

### 1. MiB to GiB

**Goal:** make `MiBToGiB` convert a whole number of MiB into GiB, keeping the
fraction.

```go
MiBToGiB(1536) // => 1.5
MiBToGiB(2048) // => 2
MiBToGiB(512)  // => 0.5
```

**Think it through:**
- `mib` is an `int`. What must happen before dividing by 1024?

### 2. GiB as text

**Goal:** add `import "strconv"` (line 4), then make `GiBString` return the
size in GiB with exactly one decimal, a space, then `GiB`.

```go
GiBString(1536) // => "1.5 GiB"
GiBString(3072) // => "3.0 GiB"
GiBString(1000) // => "1.0 GiB"   (0.9765625 rounds to 1.0)
```

**Think it through:**
- Which of your functions gives the GiB number?
- Which `strconv` function turns it into text with one decimal?

### 3. Disk usage line

**Goal:** make `DiskLine` return this line:

```text
<mount>: <used> of <total> used (<percent>%)
```

`<used>` and `<total>` come from `GiBString`. `<percent>` is used divided by
total, times 100, with the fraction **chopped off**, not rounded: 66.67
becomes 66.

```go
DiskLine("/var", 1536, 2048)
// => "/var: 1.5 GiB of 2.0 GiB used (75%)"

DiskLine("/data", 2048, 3072)
// => "/data: 2.0 GiB of 3.0 GiB used (66%)"

DiskLine("/", 1000, 1000)
// => "/: 1.0 GiB of 1.0 GiB used (100%)"
```

`totalMiB` is never 0.

**Think it through:**
- Work out the percentage first, in a variable. Which conversions does it need?
- The fixed text: `: `, ` of `, ` used (`, `%)`.

## Learn more

- [Package strconv](https://pkg.go.dev/strconv)
- [Go spec: Conversions between numeric types](https://go.dev/ref/spec#Conversions_between_numeric_types)
- [Wikipedia: Byte multiples (MiB, GiB)](https://en.wikipedia.org/wiki/Byte#Multiple-byte_units)
