# Type switches: read decoded config values

Config files written in JSON or YAML often get decoded into a
`map[string]any` before anyone knows exactly what is in them. Reading a
setting back out safely means checking its type first, because the same
`replicas: 3` can arrive as an `int`, a `float64` or even a `string`.

## Learn

### Recap

A comma-ok assertion `v, ok := x.(T)` gives you the `T` inside `x` and
`true`, or the zero `T` and `false`, without ever panicking. A type switch
`switch v := x.(type)` picks a case by the type `x` holds, and inside each
case `v` has that case's type. Looking up a missing key in a `map[string]any`
gives `nil`, and `nil` matches no type, so both forms treat a missing key as
"wrong type".

### What decoded config looks like

Different decoders produce different Go types for the same setting:

| Setting in the file | Go's JSON decoder gives | A typical YAML decoder gives |
|---|---|---|
| `replicas: 3` | `float64(3)` | `int(3)` |
| `replicas: "3"` | `string("3")` | `string("3")` |
| `debug: true` | `bool(true)` | `bool(true)` |
| `regions: [a, b]` | `[]any{"a", "b"}` | `[]any{"a", "b"}` |

JSON has only one kind of number, so Go's JSON decoder turns every number
into a `float64`. Your getters have to cope with all of these.

### Is a float64 a whole number?

Converting a `float64` to `int` drops everything after the decimal point:
`int(2.5)` is `2`. To check that nothing was lost, convert back and compare:

```go
n := int(f)
if float64(n) != f {
	// f had a fractional part, such as 2.5
}
```

## Your task

Your deploy tool reads its settings from a decoded `map[string]any`. The
three getters in `exercise.go` are half-written or buggy.

### 1. Fix GetBool

`GetBool(cfg, key)` should return the value and `true` only when `cfg[key]`
holds a `bool`. When the key is missing or holds another type, return `false,
false`. Right now it claims success every time.

```go
cfg := map[string]any{"debug": true, "dry_run": "yes"}

GetBool(cfg, "debug")   // => true, true
GetBool(cfg, "dry_run") // => false, false  (a string, not a bool)
GetBool(cfg, "verbose") // => false, false  (missing)
```

### 2. Finish GetInt

`GetInt(cfg, key)` already handles `int`. Add two cases to its type switch:

- `float64`: accept it only if it is a whole number. `3.0` gives `3, true`;
  `2.5` gives `0, false`.
- `string`: convert it with `strconv.Atoi`. `"8080"` gives `8080, true`; a
  string that is not a number, such as `"many"`, gives `0, false`. Add
  `"strconv"` to the imports.

Every other type, and a missing key, gives `0, false`.

```go
cfg := map[string]any{"replicas": 3, "port": float64(8080), "workers": "4", "ratio": 2.5}

GetInt(cfg, "replicas") // => 3, true
GetInt(cfg, "port")     // => 8080, true
GetInt(cfg, "workers")  // => 4, true
GetInt(cfg, "ratio")    // => 0, false
```

### 3. Finish GetStrings

`GetStrings(cfg, key)` returns a list of strings, such as the regions to
deploy to. It already handles a `[]string`. Add a case for `[]any`, which is
what a decoder produces: build a new `[]string` from the elements. If **any**
element is not a string, return `nil, false`. Every other type, and a missing
key, gives `nil, false`.

```go
cfg := map[string]any{
	"regions": []any{"eu-west-1", "us-east-1"},
	"ports":   []any{"80", 443},
}

GetStrings(cfg, "regions") // => []string{"eu-west-1", "us-east-1"}, true
GetStrings(cfg, "ports")   // => nil, false  (443 is not a string)
```

An empty `[]any{}` is fine: it gives an empty list and `true`.

## Learn more

- [A Tour of Go: Type switches](https://go.dev/tour/methods/16)
- [encoding/json: Unmarshal into an interface value](https://pkg.go.dev/encoding/json#Unmarshal)
- [strconv.Atoi](https://pkg.go.dev/strconv#Atoi)
