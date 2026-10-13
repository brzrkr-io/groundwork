# Functions: a human-readable byte formatter

Disk usage, memory limits, log sizes: ops tools convert bytes all the time.
You'll build a byte formatter out of small functions that call each other.

**New here:** grouped parameter types, functions calling functions, exported names, `float64` division, `%.1f`.

## Learn

You've written functions since *Hello, Go*. Here is the full picture.

### The parts of a function line

```go
func Area(width int, height int) int {
	return width * height
}
```

| Part | What it is | Can you change it? |
|---|---|---|
| `func` | keyword that starts a function | No, Go requires it |
| `Area` | the function's name | Yes, your choice |
| `(width int, height int)` | *parameters*: input names, each with a type | Names yes, the brackets no |
| `int` after `)` | the *result type* | Must match what you return |
| `{ ... }` | the *body*: the code that runs | Yes, that's your work |

Inside the body you use the **parameter names** (`width`), never the
function's name. A `return` line holds only a value, like `return width * height`.
Types and the word `func` never appear after `return`.

> **Why is it written like this?** Go requires `func`, the brackets and the
> `{` on the same line as `func`. The name goes before its type
> (`width int`) — that is Go's rule, the opposite of C or Java.
> The names `width` and `height` are the author's choice.

### Sharing a type

When neighbouring parameters share a type, you may write the type once:

```go
func Area(width, height int) int { ... }
```

Both lines mean the same thing. The short form is common style, not a rule.

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

To call a function, write its name and put the values in brackets.
`Double(n)` *is* a number, so you can use it anywhere a number fits.

### Exported or not

A name that starts with a capital letter (`Double`) is *exported*: other
packages (and the tests) can use it. A lowercase name (`double`) is only
visible inside its own package.

> **Why is it written like this?** The capital letter is a Go rule, not
> style: the compiler uses it to decide what is public.

### Constants as values

A constant like `KiB` is just a named number. Use it in maths the same way
you'd use `1024`:

```go
seconds := minutes * Minute // Minute is a constant
```

### Formatting floats

Dividing two `int`s throws away the fraction (`7 / 2` is `3`). Convert
each side to `float64` first, then format with one decimal place using `%.1f`:

```go
fmt.Sprintf("%.1f%%", float64(7)/float64(2))
// => "3.5%"
```

`%d` prints a whole number; `%.1f` prints a float with one digit after the point.

## Your task

The starter defines the constants `KiB` (1024), `MiB` (1024 × 1024) and
`GiB` (1024 × 1024 × 1024) for you.

### 1. Convert MiB to bytes

**Goal:** make `MiBToBytes(mib int) int` return the number of bytes in `mib` mebibytes.

```go
MiBToBytes(1)   // => 1048576
MiBToBytes(512) // => 536870912
MiBToBytes(0)   // => 0
```

**Think it through:**
- One MiB is how many bytes? Which constant already holds that?
- Use the parameter `mib`, not the function name.

### 2. Format a value in a unit

**Goal:** make `FormatUnit(bytes, unitSize int, unit string) string` return
`bytes ÷ unitSize` with one decimal place, a space, then `unit`.

All three pieces come from the parameters. The only fixed text is the single space.

```go
FormatUnit(1536, KiB, "KiB")    // => "1.5 KiB"
FormatUnit(3*MiB, MiB, "MiB")   // => "3.0 MiB"
FormatUnit(5*GiB/2, GiB, "GiB") // => "2.5 GiB"
```

**Think it through:**
- Which two parameters do you divide? Convert both to `float64` first.
- Which verb gives one decimal place, and which prints a string?

### 3. Pick the right unit

**Goal:** make `HumanBytes(bytes int) string` choose the biggest unit that
fits and let `FormatUnit` do the formatting.

Check from biggest to smallest:

| If `bytes` is… | Format in | Unit text |
|---|---|---|
| at least `GiB` | `GiB` | `"GiB"` |
| at least `MiB` | `MiB` | `"MiB"` |
| at least `KiB` | `KiB` | `"KiB"` |
| anything smaller | plain whole number, then `" B"` | like `"500 B"` |

```go
HumanBytes(0)               // => "0 B"
HumanBytes(500)             // => "500 B"
HumanBytes(1024)            // => "1.0 KiB"
HumanBytes(2560)            // => "2.5 KiB"
HumanBytes(MiBToBytes(512)) // => "512.0 MiB"
HumanBytes(3 * GiB / 2)     // => "1.5 GiB"
```

**Think it through:**
- Why must you check `GiB` before `KiB`? (Every GiB value is also at least a KiB.)
- Each branch can `return` a call to `FormatUnit` with the right constant and unit text.
- The starter's last line already handles the "smaller" case.

## Learn more

- [A Tour of Go: Functions](https://go.dev/tour/basics/4)
- [A Tour of Go: Exported names](https://go.dev/tour/basics/3)
- [Go by Example: Functions](https://gobyexample.com/functions)
