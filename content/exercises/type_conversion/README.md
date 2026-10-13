# Type conversion

Real numbers arrive in mixed types: pod counts are `int`s, percentages are
`float64`s, and ports end up inside strings like `0.0.0.0:8080`. Go never
converts between types behind your back, so you learn to do it yourself.

**New here:** `float64(x)`, `int(x)`, `strconv.Itoa`, `strconv.FormatFloat`, calling your own function.

## Learn

### Converting between number types

Go refuses to mix `int` and `float64` in one calculation. To convert a value,
write the type you want, then the value in parentheses:

```go
replicas := 3          // int
f := float64(replicas) // 3.0, a float64

cpu := 2.9    // float64
n := int(cpu) // 2, an int
```

Converting a `float64` to an `int` *truncates*: it chops off the fraction. It
never rounds, so `int(2.9)` is 2 and `int(0.99)` is 0.

> **Why is it written like this?**
> - **Go requires** you to convert by hand. Mixing types silently loses
>   data in other languages, so Go makes you say it.
> - **Go requires** the type name first, then the value in `( )`. It looks
>   like a function call, and works like one: it returns a new value. The
>   original variable keeps its type.

### Convert before you divide

As in *Numbers and arithmetic*, `int / int` throws the fraction away.
Convert **before** dividing, or the fraction is already gone:

```go
done, jobs := 1, 2
float64(done / jobs)          // => 0    (1 / 2 is 0 as ints)
float64(done) / float64(jobs) // => 0.5
```

(`done, jobs := 1, 2` declares two variables in one line.)

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

`fmt.Sprintf` from *Formatting with fmt* also works. `strconv` is the direct
tool when you want one number as text.

### Calling your own function

A function you wrote can be called from another one, just like
`strings.ToUpper`. Pass values for its parameters in `( )`, and use the
result:

```go
func Double(n int) int {
	return n * 2
}

func Quad(n int) int {
	twice := Double(n) // call Double, keep its result
	return Double(twice)
}
```

No package name is needed, since it's in the same package. And no types in
the call: `Double(n)`, not `Double(n int)`. Types belong only in the `func`
line.

### The `string(65)` trap

It looks like `string(port)` should turn 8080 into `"8080"`. It doesn't. For
an integer, `string(...)` gives the character with that code number, so
`string(65)` is `"A"`, not `"65"`. `go vet` reports it:

```text
conversion from int to string yields a string of one rune, not a string of digits
```

When you see that, use `strconv.Itoa` instead.

## Your task

Write the health and networking helpers for a small status page.

### 1. Healthy percentage

**Goal:** make `HealthyPercent` return healthy divided by total, times 100,
as a `float64`.

```go
HealthyPercent(3, 4) // => 75
HealthyPercent(1, 8) // => 12.5
HealthyPercent(4, 4) // => 100
HealthyPercent(0, 5) // => 0
```

`healthy` and `total` are `int`s. `total` is never 0.

**Think it through:**
- Which values need converting, and must that happen before or after `/`?

### 2. Listen address

**Goal:** add `import "strconv"` and make `ListenAddr` return the host, a
colon, then the port.

```go
ListenAddr("0.0.0.0", 8080)   // => "0.0.0.0:8080"
ListenAddr("localhost", 9090) // => "localhost:9090"
```

The import goes under the `package` line (line 2), not inside the function.

**Think it through:**
- `port` is an `int`. What turns it into text you can join with `+`?

### 3. Health label

**Goal:** make `HealthLabel` return the counts, ` pods healthy (`, the
percentage from `HealthyPercent` with one decimal, then `%)`.

```go
HealthLabel(3, 4) // => "3/4 pods healthy (75.0%)"
HealthLabel(2, 3) // => "2/3 pods healthy (66.7%)"
HealthLabel(0, 5) // => "0/5 pods healthy (0.0%)"
```

The pieces: `healthy`, `/`, `total`, ` pods healthy (`, the percentage, `%)`.

**Think it through:**
- Call `HealthyPercent` with this function's parameters. Store the result?
- Which pieces are numbers that need turning into text, and with which function?

## Learn more

- [A Tour of Go: Type conversions](https://go.dev/tour/basics/13)
- [Package strconv](https://pkg.go.dev/strconv)
- [Go spec: Conversions](https://go.dev/ref/spec#Conversions)
