# Type switches: read decoded config values

Config files in JSON or YAML often get decoded into a `map[string]any` before
anyone knows exactly what is in them. To read a setting safely you check its
type first, because `replicas: 3` can arrive as an `int`, a `float64` or even
a `string`.

**New here:** whole-number checks on `float64`, `strconv.Atoi`; otherwise practice with type assertions and switches.

## Learn

### Recap

- `v, ok := x.(T)` gives the `T` inside `x` and `true`, or the zero `T` and `false`. It never panics.
- `switch v := x.(type)` picks a case by the type `x` holds; inside each case `v` has that type.
- A missing key in a `map[string]any` gives `nil`, and `nil` matches no type. So both forms treat a missing key as "wrong type".

### What decoded config looks like

Different decoders produce different Go types for the same setting:

| Setting in the file | Go's JSON decoder gives | A typical YAML decoder gives |
|---|---|---|
| `replicas: 3` | `float64(3)` | `int(3)` |
| `replicas: "3"` | `string("3")` | `string("3")` |
| `debug: true` | `bool(true)` | `bool(true)` |
| `regions: [a, b]` | `[]any{"a", "b"}` | `[]any{"a", "b"}` |

JSON has only one kind of number, so Go's JSON decoder turns every number into
a `float64`.

### Is a float64 a whole number?

Converting a `float64` to `int` drops everything after the decimal point:
`int(2.5)` is `2`. Converting back gives `2.0`, which is not `2.5`. So a
round trip tells you whether anything was lost:

```go
price := 9.99
cents := int(price)              // 9
fmt.Println(float64(cents) == price) // false: the .99 was dropped
```

> **Why is it written like this?** Go requires the explicit conversion
> `int(...)` / `float64(...)`; it never converts numbers for you, and you
> can't compare an `int` with a `float64` directly.

### Turning text into a number: `strconv.Atoi`

`strconv.Atoi` ("ASCII to int") returns the number and an error:

```go
n, err := strconv.Atoi("42")   // 42, nil
n, err = strconv.Atoi("lots")  // 0, an error
```

The error is not `nil` when the text is not a whole number. You need
`import "strconv"` at the top of the file.

> **Why is it written like this?** Returning `(value, error)` is a Go
> convention for anything that can fail; checking `err != nil` right after is
> the expected pattern.

## Your task

Your deploy tool reads its settings from a decoded `map[string]any`. The
three getters in `exercise.go` are half-written or buggy.

### 1. Fix GetBool

**Goal:** make `GetBool(cfg, key)` return `true` as its second result only when `cfg[key]` holds a `bool`.

When it holds a `bool`, return that value and `true`. When the key is missing
or holds another type, return `false, false`. Right now it claims success
every time.

```go
cfg := map[string]any{"debug": true, "dry_run": "yes"}

GetBool(cfg, "debug")   // => true, true
GetBool(cfg, "dry_run") // => false, false  (a string, not a bool)
GetBool(cfg, "verbose") // => false, false  (missing)
```

**Think it through:**
- The assertion already produces a second value. The starter throws it away with `_`.
- What is the function returning as its second result instead?

### 2. Finish GetInt

**Goal:** add a `float64` case and a `string` case to the type switch in `GetInt`.

The `int` case is already there. The new cases:

- `float64`: accept it only if it is a whole number. `3.0` gives `3, true`; `2.5` gives `0, false`.
- `string`: convert it with `strconv.Atoi`. `"8080"` gives `8080, true`; text that is not a number, such as `"many"`, gives `0, false`. Add `"strconv"` to the imports.

Every other type, and a missing key, gives `0, false` (the line after the
switch already does that).

```go
cfg := map[string]any{"replicas": 3, "port": float64(8080), "workers": "4", "ratio": 2.5}

GetInt(cfg, "replicas") // => 3, true
GetInt(cfg, "port")     // => 8080, true
GetInt(cfg, "workers")  // => 4, true
GetInt(cfg, "ratio")    // => 0, false
```

**Think it through:**
- Inside each case, `v` already has that case's type. Which conversion does a `float64` need?
- For the string case, what tells you that `Atoi` failed?

### 3. Finish GetStrings

**Goal:** add a `[]any` case to `GetStrings` that builds a new `[]string` from the elements.

The `[]string` case is already there. For `[]any`, every element must be a
`string`; if **any** element is not, return `nil, false`. An empty `[]any{}`
is fine: it gives an empty list and `true`. Every other type, and a missing
key, gives `nil, false`.

```go
cfg := map[string]any{
	"regions": []any{"eu-west-1", "us-east-1"},
	"ports":   []any{"80", 443},
}

GetStrings(cfg, "regions") // => []string{"eu-west-1", "us-east-1"}, true
GetStrings(cfg, "ports")   // => nil, false  (443 is not a string)
```

**Think it through:**
- You need a loop over the elements, and a check on each one. Which form of assertion lets you check without panicking?
- Start with an empty `[]string` and add to it, as in *Slices*.

## Learn more

- [A Tour of Go: Type switches](https://go.dev/tour/methods/16)
- [encoding/json: Unmarshal into an interface value](https://pkg.go.dev/encoding/json#Unmarshal)
- [strconv.Atoi](https://pkg.go.dev/strconv#Atoi)
