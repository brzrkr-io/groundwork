# Custom types: Kubernetes memory quantities

Kubernetes writes memory as *quantities* like `512Mi` and `2Gi`. In this
exercise you give memory sizes their own type, parse quantities, and check
whether a set of pods fits on a node.

## Learn

### Recap

As in *Custom types and enums*, `type Bytes int64` declares a new type with
underlying type `int64`. You convert to and from it with the type name:
`Bytes(n)` and `int64(b)`.

### Constants built from constants

A constant of a custom type can be built from another one. Because `KiB` is
a `Bytes`, so are `MiB` and `GiB`:

```go
const (
	KiB Bytes = 1024
	MiB       = 1024 * KiB // 1,048,576 bytes
	GiB       = 1024 * MiB // 1,073,741,824 bytes
)
```

Kubernetes uses these powers of 1024: `Ki`, `Mi` and `Gi`.

### Mixing types needs a conversion

Go never mixes two different types in one calculation, even if both are
numbers underneath. `strconv.Atoi` gives you an `int`, so convert it before
multiplying by a `Bytes` constant:

```go
n := 512 // an int, for example from strconv.Atoi

size := n * MiB        // compile error: mismatched types int and Bytes
size := Bytes(n) * MiB // OK: Bytes * Bytes
```

To print a fraction, convert both sides to `float64` first, as in
*Type conversion*:

```go
b := Bytes(1536)
float64(b) / float64(KiB) // 1.5
```

### Removing a suffix with strings.CutSuffix

`strings.CutSuffix(s, suffix)` works like `strings.Cut` from
*The strings package*, but only looks at the end of `s`. It returns `s`
without the suffix, and whether the suffix was there:

```go
number, found := strings.CutSuffix("512Mi", "Mi")
// number == "512", found == true

number, found = strings.CutSuffix("2Gi", "Mi")
// number == "2Gi", found == false
```

## Your task

Your cluster capacity tool reads pod memory requests and node sizes. The
`Bytes` type and its constants are in `exercise.go`. You'll need `fmt`,
`strconv` and `strings`, so add them:

```go
import (
	"fmt"
	"strconv"
	"strings"
)
```

### 1. Print a size for humans

Complete `Human` so it formats `b` with the largest unit that `b` is at
least one of:

| When `b` is | Format | Example |
|---|---|---|
| at least `GiB` | `%.1f GiB` | `"1.5 GiB"` |
| at least `MiB` | `%.1f MiB` | `"256.0 MiB"` |
| at least `KiB` | `%.1f KiB` | `"1.5 KiB"` |
| less than `KiB` | `%d B` | `"900 B"` |

For the first three rows, divide `b` by the unit as `float64`s. `%.1f`
rounds to one decimal place. `b` is never negative.

```go
Bytes(900).Human()   // => "900 B"
Bytes(1536).Human()  // => "1.5 KiB"
(256 * MiB).Human()  // => "256.0 MiB"
(1536 * MiB).Human() // => "1.5 GiB"
```

### 2. Parse a quantity

Complete `ParseMemory` so it turns a quantity into `Bytes`. A quantity is a
whole number, optionally followed by one unit suffix:

| Suffix | Multiply by | Example | Result |
|---|---|---|---|
| `Ki` | `KiB` | `"64Ki"` | `65536` |
| `Mi` | `MiB` | `"512Mi"` | `536870912` |
| `Gi` | `GiB` | `"2Gi"` | `2147483648` |
| none | `1` | `"1000"` | `1000` |

Return `0` and an error (made with `fmt.Errorf` or `errors.New`) when:

- the number part is not a whole number (`"lotsMi"`, `"1.5Gi"`, `"Mi"`, `""`)
- the number is negative (`"-5Mi"`)
- the suffix is anything else (`"2G"`, `"512MB"`): the number part then
  isn't a number, so `strconv.Atoi` already catches this.

The tests only check that an error is returned, not its message. A helpful
one is `invalid memory quantity "lotsMi"`.

```go
ParseMemory("512Mi") // => 536870912, nil
ParseMemory("2G")    // => 0, error
```

### 3. Check whether pods fit on a node

Complete `FitsOnNode`. Parse every quantity in `requests` with
`ParseMemory`, add them up, and return `true` if the total is less than or
equal to `free`. An empty list always fits.

If any request fails to parse, return `false` and that error straight away.

```go
FitsOnNode([]string{"512Mi", "1Gi"}, 2*GiB) // => true, nil
FitsOnNode([]string{"1Gi", "1Gi"}, 2*GiB)   // => true, nil  (exactly full)
FitsOnNode([]string{"3Gi"}, 2*GiB)          // => false, nil
FitsOnNode([]string{"1Gi", "huge"}, 2*GiB)  // => false, error
```

## Learn more

- [Kubernetes: Resource units for memory](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#meaning-of-memory)
- [strings.CutSuffix](https://pkg.go.dev/strings#CutSuffix)
- [strconv.Atoi](https://pkg.go.dev/strconv#Atoi)
- [Go spec: Conversions](https://go.dev/ref/spec#Conversions)
