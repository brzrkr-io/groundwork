# Custom types: Kubernetes memory quantities

Kubernetes writes memory as *quantities* like `512Mi` and `2Gi`. In this
exercise you give memory sizes their own type, parse quantities, and check
whether a set of pods fits on a node.

**New here:** constants built from constants, `strings.CutSuffix`, mixing a custom type with `int` and `float64`.

## Learn

### Recap

As in *Custom types and enums*, `type Bytes int64` declares a new type with
underlying type `int64`. You convert with the type name: `Bytes(n)` turns a
number into `Bytes`, `int64(b)` turns it back.

### Constants built from constants

A constant can be built from another one. If the first has a custom type,
the ones built from it get that type too:

```go
type Duration int64

const (
	Second Duration = 1
	Minute          = 60 * Second // also a Duration
	Hour            = 60 * Minute // also a Duration
)
```

`exercise.go` builds `KiB`, `MiB` and `GiB` the same way, with powers of
1024 (the units Kubernetes calls `Ki`, `Mi`, `Gi`).

> **Why is it written like this?** Go requires a type on the first constant
> to make it a `Duration`; the rest pick it up from the expression. Lining up
> the `=` signs is gofmt's doing, not yours.

### Mixing types needs a conversion

Go never mixes two different types in one calculation, even if both are
numbers underneath. `strconv.Atoi` gives you an `int`, so convert it before
using it with a custom type:

```go
n := 3 // an int, for example from strconv.Atoi

d := n * Hour           // compile error: mismatched types int and Duration
d := Duration(n) * Hour // OK: Duration * Duration
```

Dividing two integers drops the fraction (`3 / 2` is `1`). To get `1.5`,
convert both sides to `float64` first, as in *Type conversion*:

```go
total, parts := 3, 2
float64(total) / float64(parts) // 1.5
```

> **Why is it written like this?** The conversion is required by Go: it has
> no automatic number conversion. Where you put it (on the number, not on
> the result) decides whether the fraction survives.

### Formatting with one decimal

In `fmt.Sprintf`, `%.1f` prints a `float64` with exactly one decimal place,
rounding if needed. `%d` prints a whole number.

```go
fmt.Sprintf("%.1f s", 2.0)  // "2.0 s"
fmt.Sprintf("%.1f s", 2.25) // "2.2 s" or "2.3 s" (rounded)
fmt.Sprintf("%d items", 7)  // "7 items"
```

### Removing a suffix with strings.CutSuffix

`strings.CutSuffix(s, suffix)` works like `strings.Cut` from
*The strings package*, but only looks at the end of `s`. It returns `s`
without the suffix, and whether the suffix was there:

```go
name, found := strings.CutSuffix("report.csv", ".csv")
// name == "report", found == true

name, found = strings.CutSuffix("report.txt", ".csv")
// name == "report.txt", found == false
```

## Your task

Your cluster capacity tool reads pod memory requests and node sizes. The
`Bytes` type and its constants `KiB`, `MiB`, `GiB` are in `exercise.go`.
You'll need `fmt`, `strconv` and `strings`, so add them below the `package`
line:

```go
import (
	"fmt"
	"strconv"
	"strings"
)
```

### 1. Print a size for humans

**Goal:** make `Human` return `b` as text, using the largest unit that `b` is at least one of.

| When `b` is | Format | Example |
|---|---|---|
| at least `GiB` | `%.1f GiB` | `"1.5 GiB"` |
| at least `MiB` | `%.1f MiB` | `"256.0 MiB"` |
| at least `KiB` | `%.1f KiB` | `"1.5 KiB"` |
| less than `KiB` | `%d B` | `"900 B"` |

For the first three rows, the number shown is `b` divided by that unit, with
the fraction kept. Note the single space before the unit. `b` is never
negative.

```go
Bytes(0).Human()     // => "0 B"
Bytes(1023).Human()  // => "1023 B"
Bytes(1024).Human()  // => "1.0 KiB"
Bytes(1536).Human()  // => "1.5 KiB"
(256 * MiB).Human()  // => "256.0 MiB"
(1536 * MiB).Human() // => "1.5 GiB"
```

**Think it through:**
- In which order must you check the units so the largest one wins?
- Which two values do you convert to `float64` before dividing?

### 2. Parse a quantity

**Goal:** make `ParseMemory` turn a quantity like `"512Mi"` into `Bytes` and a `nil` error.

A quantity is a whole number, optionally followed by one unit suffix:

| Suffix | Multiply by | Example | Result |
|---|---|---|---|
| `Ki` | `KiB` | `"64Ki"` | `65536` |
| `Mi` | `MiB` | `"512Mi"` | `536870912` |
| `Gi` | `GiB` | `"2Gi"` | `2147483648` |
| none | `1` | `"1000"` | `1000` |

`"0"` is valid and gives `0`. Return `0` and a non-`nil` error when:

- the number part is not a whole number: `"lotsMi"`, `"1.5Gi"`, `"Mi"`, `""`
- the number is negative: `"-5Mi"`
- the suffix is anything else: `"2G"`, `"512MB"`. The number part then
  isn't a number, so `strconv.Atoi` already catches this.

Make the error with `fmt.Errorf` or `errors.New`. The tests only check that
there is one, not its text; `invalid memory quantity "lotsMi"` is helpful.

```go
ParseMemory("512Mi") // => 536870912, nil
ParseMemory("2G")    // => 0, error
```

**Think it through:**
- Split it in two steps: find and remove the suffix (remembering its unit), then turn the rest into a number.
- If no suffix matches, what is the number part and what is the unit?
- `strconv.Atoi` returns an `int`. What must happen before you multiply it by a unit?

### 3. Check whether pods fit on a node

**Goal:** make `FitsOnNode` return `true, nil` if all the `requests` added together are less than or equal to `free`.

- Parse each string in `requests` with your `ParseMemory`.
- If any request fails to parse, return `false` and that error straight away.
- An empty list always fits.

```go
FitsOnNode([]string{"512Mi", "1Gi"}, 2*GiB) // => true, nil
FitsOnNode([]string{"1Gi", "1Gi"}, 2*GiB)   // => true, nil  (exactly full)
FitsOnNode([]string{"3Gi"}, 2*GiB)          // => false, nil
FitsOnNode([]string{"1Gi", "huge"}, 2*GiB)  // => false, error
```

**Think it through:**
- You need a running total. What type should it be, and where does it start?
- Compare with `free` only after the loop. Which operator includes "exactly equal"?

## Learn more

- [Kubernetes: Resource units for memory](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#meaning-of-memory)
- [strings.CutSuffix](https://pkg.go.dev/strings#CutSuffix)
- [strconv.Atoi](https://pkg.go.dev/strconv#Atoi)
- [Go spec: Conversions](https://go.dev/ref/spec#Conversions)
