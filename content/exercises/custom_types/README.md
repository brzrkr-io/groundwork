# Custom types and enums

Ops code passes environment names and alert severities around all day. Give
them their own types, and Go catches typos and mix-ups before they page the
wrong person at 3 a.m.

## Learn

### Your own types

So far an environment name has been a plain `string`. Any string fits,
including typos like `"prod"`. You can declare a new type instead:

```go
type Environment string
```

`Environment` is a *custom type*. Its *underlying type* is `string`, so it
stores text and works with `==` and `len`, but Go treats it as a different
type from `string`.

| Declaration | Underlying type | Example value |
|---|---|---|
| `type Environment string` | `string` | `"staging"` |
| `type Severity int` | `int` | `3` |
| `type Port int` | `int` | `8080` |

### Typed constants

List the allowed values as constants of the new type:

```go
const (
	Dev        Environment = "dev"
	Staging    Environment = "staging"
	Production Environment = "production"
)
```

Now code says `Production` instead of retyping `"production"`. A typo like
`Prodution` is a compile error instead of a silent bug.

### Converting between types

A literal like `"dev"` fits into an `Environment` directly. A `string`
*variable* does not. Convert it with the type name, just like `float64(n)` in
*Type conversion*:

```go
name := "staging"        // a string, maybe read from a flag
env := Environment(name) // string → Environment
s := string(env)         // Environment → string
```

A conversion only changes the type, never the value. `Environment("banana")`
compiles fine, which is why you'll write `IsValid` below.

### Enums with iota

Go has no `enum` keyword. Instead you use a custom integer type and a
constant block with *`iota`*. Inside a `const ( ... )` block, `iota` is `0`
on the first line, `1` on the next line, and so on. A line with only a name
repeats the line above it, with `iota` one higher:

```go
type Severity int

const (
	Debug    Severity = iota // 0
	Info                     // 1
	Warning                  // 2
	Error                    // 3
	Critical                 // 4
)
```

Because the underlying type is `int`, severities compare with `<` and `>=`:
`Warning < Critical` is `true`.

The zero value of an `int` is `0`, so a `Severity` you never set
(`var s Severity`) is `Debug`, the first constant. Put the safest value
first.

### Methods on your types

In *Methods* you gave structs methods. Any type you declare in your package
can have methods, including `Environment` and `Severity`:

```go
// IsDev reports whether e is the dev environment.
func (e Environment) IsDev() bool {
	return e == Dev
}

Dev.IsDev()                 // true
Environment("prod").IsDev() // false
```

### Why bother?

With plain values, `Page(3, "production")` and `Page("production", 3)` look
alike to a tired human. With `func Page(s Severity, env Environment)`, the
swapped call doesn't compile, and code like `env.IsValid()` reads like
English.

## Your task

Your alerting service decides who gets woken up. The types and constants are
already in `exercise.go`. The methods are not.

### 1. Validate an environment

Complete `IsValid` so it returns `true` when `e` is exactly `Dev`, `Staging`
or `Production`, and `false` for anything else. Letter case matters.

```go
Production.IsValid()             // => true
Environment("prod").IsValid()    // => false
Environment("Staging").IsValid() // => false
```

### 2. Decide whether to page

Complete `ShouldPage` so it follows these rules:

| Severity | Pages the on-call engineer |
|---|---|
| `Critical` | always, in every environment |
| `Error` | only when `env` is `Production` |
| `Debug`, `Info`, `Warning` | never |

```go
Critical.ShouldPage(Dev)       // => true
Error.ShouldPage(Production)   // => true
Error.ShouldPage(Staging)      // => false
Warning.ShouldPage(Production) // => false
```

### 3. Parse a severity

Alert payloads carry severities as text. Complete `ParseSeverity` so it turns
one of the names `debug`, `info`, `warning`, `error` or `critical` into the
matching `Severity` and `true`. Accept any mix of upper and lower case. For
any other text, return `Debug` (the zero value, `0`) and `false`.

`strings.ToLower` from *The strings package* helps here. Add the import below
the `package` line:

```go
import "strings"
```

```go
ParseSeverity("critical") // => Critical, true
ParseSeverity("WARNING")  // => Warning, true
ParseSeverity("fatal")    // => Debug, false
```

The tests print severities as numbers, so `Critical` shows up as `4` in a
failure message.

## Learn more

- [A Tour of Go: Methods on non-struct types](https://go.dev/tour/methods/3)
- [Go by Example: Enums](https://gobyexample.com/enums)
- [Go spec: Type definitions](https://go.dev/ref/spec#Type_definitions)
- [Go spec: Iota](https://go.dev/ref/spec#Iota)
