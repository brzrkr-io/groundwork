# Type conversion

Real numbers arrive in mixed types: pod counts are `int`s, percentages are
`float64`s, and ports end up inside strings like `0.0.0.0:8080`. Go never
converts between types behind your back, so you learn to do it yourself.

## Learn

### Converting between number types

Go refuses to mix `int` and `float64` in one calculation. To convert a value,
write the type you want, then the value in parentheses:

```go
replicas := 3          // int
f := float64(replicas) // 3.0, a float64

cpu := 2.9             // float64
n := int(cpu)          // 2, an int
```

Converting a `float64` to an `int` *truncates*: it chops off the fraction. It
never rounds, so `int(2.9)` is 2 and `int(0.99)` is 0.

### Convert before you divide

Remember from *Numbers and arithmetic* that `int / int` throws the fraction
away. Convert **before** dividing, or the fraction is already gone:

```go
healthy, total := 3, 4
float64(healthy / total)          // => 0    (3 / 4 is 0 as ints)
float64(healthy) / float64(total) // => 0.75
```

(`healthy, total := 3, 4` declares two variables in one line.)

### Numbers into strings: strconv

You can't join a number to a string with `+`. The `strconv` package ("string
conversion") turns numbers into text:

```go
import "strconv"

strconv.Itoa(8080) // => "8080"
```

`Itoa` stands for "integer to ASCII". For a `float64`, use `FormatFloat`. It
takes four arguments:

```go
strconv.FormatFloat(66.666, 'f', 1, 64) // => "66.7"
```

| Argument | Meaning |
|---|---|
| `66.666` | the number |
| `'f'` | plain decimal format, like `66.7` (not scientific like `6.67e+01`). Single quotes hold one character. |
| `1` | digits after the point, rounded |
| `64` | the value is a `float64` |

`fmt.Sprintf("%d", n)` from *Formatting with fmt* would also work.
`strconv` is the direct tool when all you want is one number as text.

### The `string(65)` trap

It looks like `string(port)` should turn 8080 into `"8080"`. It doesn't. For
an integer, `string(...)` gives the character with that code number, so
`string(65)` is `"A"`, not `"65"`. This mistake is so common that `go vet`
reports it, and Gopherlings will show you:

```text
conversion from int to string yields a string of one rune, not a string of digits
```

When you see that, use `strconv.Itoa` instead.

## Your task

Write the health and networking helpers for a small status page.

### 1. Healthy percentage

Change `HealthyPercent` so it returns the percentage of pods that are healthy:
healthy divided by total, times 100, as a `float64`.

```go
HealthyPercent(3, 4) // => 75
HealthyPercent(1, 8) // => 12.5
HealthyPercent(4, 4) // => 100
HealthyPercent(0, 5) // => 0
```

`total` is never 0.

### 2. Listen address

A server listens on an *address*: a host, a colon, then the port. Add
`import "strconv"` and change `ListenAddr` to build it:

```go
ListenAddr("0.0.0.0", 8080)  // => "0.0.0.0:8080"
ListenAddr("localhost", 9090) // => "localhost:9090"
```

### 3. Health label

Change `HealthLabel` so it returns the counts, then the percentage from
`HealthyPercent` with one decimal place:

```go
HealthLabel(3, 4) // => "3/4 pods healthy (75.0%)"
HealthLabel(2, 3) // => "2/3 pods healthy (66.7%)"
HealthLabel(0, 5) // => "0/5 pods healthy (0.0%)"
```

## Learn more

- [A Tour of Go: Type conversions](https://go.dev/tour/basics/13)
- [Package strconv](https://pkg.go.dev/strconv)
- [Go spec: Conversions](https://go.dev/ref/spec#Conversions)
