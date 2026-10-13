# Functions: a human-readable byte formatter

Disk usage, memory limits, log sizes: ops tools convert bytes all the time.
You'll build a byte formatter out of small functions that call each other.

## Learn

You've written functions since *Hello, Go*. Here is the full picture.

### Parameters and results

```go
func Area(width int, height int) int {
	return width * height
}
```

`width` and `height` are *parameters*: inputs, each with a type. The `int`
after the parentheses is the result type. When neighbouring parameters share
a type, write the type once:

```go
func Area(width, height int) int { ... }
```

Parameters can have different types: `func Label(name string, port int) string`.

### Calling your own functions

A function can call any other function in the package. Small helpers keep
each function short and easy to test:

```go
func Double(n int) int {
	return n * 2
}

func Quadruple(n int) int {
	return Double(Double(n))
}
```

Good functions do one job and have a name that says what it is.

### Exported or not

A name that starts with a capital letter (`Double`) is *exported*: other
packages can use it. A lowercase name (`double`) is only visible inside its
own package.

### Formatting floats

Dividing two `int`s throws away the fraction (`1536 / 1024` is `1`). Convert
to `float64` first, then format with one decimal place using `%.1f`:

```go
fmt.Sprintf("%.1f %s", float64(1536)/float64(1024), "KiB")
// => "1.5 KiB"
```

## Your task

The starter defines the constants `KiB` (1024), `MiB` (1024 × 1024) and
`GiB` (1024 × 1024 × 1024) for you.

### 1. Convert MiB to bytes

Change `MiBToBytes(mib int) int` so it returns the number of bytes in `mib`
mebibytes.

```go
MiBToBytes(1)   // => 1048576
MiBToBytes(512) // => 536870912
```

### 2. Format a value in a unit

Change `FormatUnit(bytes, unitSize int, unit string) string` so it divides
`bytes` by `unitSize` as floats and returns the result with one decimal
place, a space and the unit name.

```go
FormatUnit(1536, KiB, "KiB")     // => "1.5 KiB"
FormatUnit(3*MiB, MiB, "MiB")    // => "3.0 MiB"
```

### 3. Pick the right unit

Change `HumanBytes(bytes int) string` so it picks the biggest unit that fits
and calls `FormatUnit`:

| bytes | result |
|---|---|
| at least `GiB` | `FormatUnit(bytes, GiB, "GiB")` |
| at least `MiB` | `FormatUnit(bytes, MiB, "MiB")` |
| at least `KiB` | `FormatUnit(bytes, KiB, "KiB")` |
| anything smaller | the number and `" B"`, like `"500 B"` |

```go
HumanBytes(500)              // => "500 B"
HumanBytes(1024)             // => "1.0 KiB"
HumanBytes(MiBToBytes(512))  // => "512.0 MiB"
HumanBytes(3 * GiB / 2)      // => "1.5 GiB"
```

## Learn more

- [A Tour of Go: Functions](https://go.dev/tour/basics/4)
- [A Tour of Go: Exported names](https://go.dev/tour/basics/3)
- [Go by Example: Functions](https://gobyexample.com/functions)
