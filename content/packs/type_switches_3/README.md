# Type switches: flatten a nested config

Tools such as Helm (`--set db.port=5432`) and Consul work with *flat* keys
like `db.port` instead of nested blocks. Here you flatten a decoded config,
using type switches to tell nested blocks apart from plain values.

**New here:** several types in one `case`, interface cases, recursion (a function calling itself).

## Learn

### Recap

`switch v := x.(type)` picks a case by the type `x` holds; inside a case `v`
has that type. `default` catches everything else. `%T` prints a value's type.

### Several types in one case

A case may list several types, separated by commas. Then Go can't pick one
type for `v`, so `v` keeps the type of the switched value (`any`). That is
fine when you treat them all the same way:

```go
switch v := x.(type) {
case int8, int16, int32:
	return fmt.Sprintf("small number %d", v) // v is still an any here
}
```

### Interface cases

A case can name an interface. It matches every type that has the interface's
methods. `fmt.Stringer` (as in *Stringers*) matches anything with a
`String() string` method, such as `time.Duration`:

```go
var d any = 30 * time.Second
switch v := d.(type) {
case fmt.Stringer:
	fmt.Println(v.String()) // "30s"
}
```

Cases are checked **top to bottom**, and the first match wins.

> **Why is it written like this?** The commas in a multi-type case are
> required syntax. The order of cases is your choice, but it matters: put the
> specific types before a broad interface case.

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
map it finds. This is *recursion*:

```go
// depth reports how many levels of maps are nested inside m.
func depth(m map[string]any) int {
	deepest := 0
	for _, value := range m {
		if inner, ok := value.(map[string]any); ok {
			deepest = max(deepest, depth(inner)) // same function, one level down
		}
	}
	return deepest + 1
}
```

It stops on its own: the innermost maps contain no more maps, so they make no
further calls.

> **Why is it written like this?** Go needs nothing special for recursion;
> a function may simply call itself by name. Passing the shared result (here,
> a map) in as a parameter is a common choice so every level writes to the
> same place.

## Your task

Your deploy tool loads a nested config and has to pass it on as flat
`key=value` pairs.

### 1. Convert a plain value to text

**Goal:** make `Scalar(v)` turn one config value into text and `true`, or return `"", false` if it can't.

Use a type switch:

| Type of `v` | Text to return |
|---|---|
| `string` | the string itself |
| `int`, `float64` or `bool` | the value printed with `%v` |
| anything with a `String() string` method (`fmt.Stringer`) | the result of its `String()` method |
| anything else (including `nil`, `int64`, slices and maps) | `""`, and `false` |

Supported values also return `true`.

```go
Scalar("eu-west-1")             // => "eu-west-1", true
Scalar(5432)                    // => "5432", true
Scalar(0.75)                    // => "0.75", true
Scalar(true)                    // => "true", true
Scalar(30 * time.Second)        // => "30s", true
Scalar(net.ParseIP("10.0.0.1")) // => "10.0.0.1", true
Scalar([]string{"a", "b"})      // => "", false
```

**Think it through:**
- Which three types can share one case, because they're all handled the same way?
- Which case does `time.Duration` reach? It isn't `int`, even though it looks like a number.

### 2. Flatten the config

**Goal:** finish `flattenInto(out, prefix, m)` so every plain value ends up in `out` under its full dotted key.

`Flatten` is already written; it calls `flattenInto` with an empty prefix.
The starter builds `fullKey` for you: the prefix, a dot, and the key (just
the key at the top level). For every key in `m`:

- If the value is a `map[string]any`, call `flattenInto` again with `fullKey` as the new prefix. If that call returns an error, return it.
- Otherwise convert it with `Scalar` and store the text in `out` under `fullKey`.
- If `Scalar` can't convert it, return the error `unsupported value for "<full key>": <type>`, using `%q` for the key and `%T` for the type. (`Flatten` then returns `nil` and that error.)

An empty inner map adds nothing.

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

**Think it through:**
- Replace the starter's `out[fullKey] = fmt.Sprintf(...)` line; it skips both the nested check and `Scalar`.
- Which function makes an error from a format string, as in earlier error exercises?

## Learn more

- [Go spec: Type switches](https://go.dev/ref/spec#Type_switches)
- [fmt.Stringer](https://pkg.go.dev/fmt#Stringer)
- [Go by Example: Recursion](https://gobyexample.com/recursion)
- [Helm: The format and limitations of --set](https://helm.sh/docs/intro/using_helm/#the-format-and-limitations-of---set)
