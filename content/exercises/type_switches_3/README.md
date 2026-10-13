# Type switches: flatten a nested config

Tools such as Helm (`--set db.port=5432`), Consul and many config loaders work
with *flat* keys like `db.port` instead of nested blocks. In this exercise you
flatten a decoded config, using type switches to tell nested blocks apart from
plain values.

## Learn

### Recap

`switch v := x.(type)` picks a case by the type `x` holds, and inside a case
`v` has that type. `default` catches everything else. `%T` prints a value's
type.

### Several types in one case

A case may list several types. Then Go can't pick one type for `v`, so `v`
keeps the type of the switched value (`any`). That is fine when you treat
them all the same way:

```go
switch v := x.(type) {
case int, float64, bool:
	return fmt.Sprintf("%v", v) // v is still an any here
}
```

### Interface cases

A case can also name an interface. It matches every type that has the
interface's methods. `fmt.Stringer` (as in *Stringers*) matches anything with
a `String() string` method, such as `time.Duration`:

```go
switch v := x.(type) {
case fmt.Stringer:
	return v.String() // time.Duration(30 * time.Second) gives "30s"
}
```

Cases are checked from top to bottom, and the first match wins.

### A function that calls itself

A nested config is a map whose values can be more maps, as deep as you like:

```go
cfg := map[string]any{
	"db": map[string]any{
		"primary": map[string]any{"port": 5432},
	},
}
```

You can't write one loop per level when you don't know how many levels there
are. Instead, a function handles one map and calls *itself* for every inner
map it finds. This is called *recursion*:

```go
// countValues counts the plain (non-map) values in a nested config.
func countValues(m map[string]any) int {
	count := 0
	for _, value := range m {
		switch v := value.(type) {
		case map[string]any:
			count += countValues(v) // the same function, one level deeper
		default:
			count++
		}
	}
	return count
}
```

It stops on its own: the innermost maps contain no more maps, so they make no
further calls.

## Your task

Your deploy tool loads a nested config and has to pass it on as flat
`key=value` pairs.

### 1. Convert a plain value to text

Write `Scalar(v any) (string, bool)`, which turns a single config value into
text:

| Type of `v` | Result |
|---|---|
| `string` | the string itself |
| `int`, `float64` or `bool` | `fmt.Sprintf("%v", v)` |
| anything with a `String() string` method (`fmt.Stringer`) | `v.String()` |
| anything else (including `nil`, `int64`, slices and maps) | `""`, `false` |

Every supported value also returns `true`.

```go
Scalar("eu-west-1")                 // => "eu-west-1", true
Scalar(5432)                        // => "5432", true
Scalar(0.75)                        // => "0.75", true
Scalar(true)                        // => "true", true
Scalar(30 * time.Second)            // => "30s", true
Scalar(net.ParseIP("10.0.0.1"))     // => "10.0.0.1", true
Scalar([]string{"a", "b"})          // => "", false
```

### 2. Flatten the config

`Flatten` is already written. It calls `flattenInto` with an empty prefix.
Finish `flattenInto(out, prefix, m)` so that for every key in `m`:

- If the value is a `map[string]any`, call `flattenInto` again with the full
  key as the new prefix. If that call returns an error, return it.
- Otherwise convert it with `Scalar` and store it in `out` under the full key.
- If `Scalar` can't convert it, return the error
  `unsupported value for "<full key>": <type>`, using `%q` for the key and
  `%T` for the type (`Flatten` then returns `nil` and that error).

The full key is the prefix, a dot, and the key; at the top level, where the
prefix is `""`, it is just the key. The starter already builds it as
`fullKey`. An empty inner map adds nothing.

```go
Flatten(map[string]any{
	"service": "api",
	"db": map[string]any{
		"port":    5432,
		"timeout": 30 * time.Second,
		"primary": map[string]any{"host": "10.0.0.5"},
	},
})
// => map[string]string{
//	"service":         "api",
//	"db.port":         "5432",
//	"db.timeout":      "30s",
//	"db.primary.host": "10.0.0.5",
// }, nil

Flatten(map[string]any{
	"db": map[string]any{"hosts": []string{"a", "b"}},
})
// => nil, error `unsupported value for "db.hosts": []string`
```

## Learn more

- [Go spec: Type switches](https://go.dev/ref/spec#Type_switches)
- [fmt.Stringer](https://pkg.go.dev/fmt#Stringer)
- [Go by Example: Recursion](https://gobyexample.com/recursion)
- [Helm: The format and limitations of --set](https://helm.sh/docs/intro/using_helm/#the-format-and-limitations-of---set)
