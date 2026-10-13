# Custom types and enums

Ops code passes environment names and alert severities around all day. Give
them their own types, and Go catches typos and mix-ups before they page the
wrong person at 3 a.m.

**New here:** `type X string`, typed constants, `iota`, methods on non-struct types.

## Learn

### Your own types

So far an environment name has been a plain `string`. Any string fits,
including typos like `"prod"`. You can declare a new type instead:

```go
type Region string
```

`Region` is a *custom type*. Its *underlying type* is `string`: it stores
text and works with `==` and `len`. But Go treats it as a **different** type
from `string`.

| Declaration | Underlying type | Example value |
|---|---|---|
| `type Region string` | `string` | `"eu-west-1"` |
| `type Port int` | `int` | `8080` |

> **Why is it written like this?** `type`, then the new name, then the
> underlying type: Go requires that order. The name `Region` is the author's
> choice; starting it with a capital letter (so other packages can use it)
> is a Go rule for exported names.

### Typed constants

List the allowed values as constants of the new type:

```go
const (
	EUWest Region = "eu-west-1"
	USEast Region = "us-east-1"
)
```

Now code says `EUWest` instead of retyping `"eu-west-1"`. A typo like
`EUWst` is a compile error instead of a silent bug.

> **Why is it written like this?** `const ( ... )` groups several constants:
> Go allows it, and grouping related ones is convention. Writing `Region`
> after the name gives the constant that type; without it, it would be a
> plain string constant.

### Converting between types

A literal like `"eu-west-1"` fits into a `Region` directly. A `string`
*variable* does not. Convert it with the type name, like `float64(n)` in
*Type conversion*:

```go
name := "us-east-1"  // a string, maybe read from a flag
r := Region(name)    // string → Region
s := string(r)       // Region → string
```

A conversion only changes the type, never the value. `Region("banana")`
compiles fine. That's why a validity check (task 1) is useful.

### Enums with iota

Go has no `enum` keyword. You use a custom integer type plus a `const`
block with *`iota`*. Inside `const ( ... )`, `iota` is `0` on the first
line, `1` on the next, and so on. A line with only a name repeats the line
above, with `iota` one higher:

```go
type Priority int

const (
	Low    Priority = iota // 0
	Medium                 // 1
	High                   // 2
)
```

Because the underlying type is `int`, these compare with `<` and `>=`:
`Low < High` is `true`.

A `Priority` you never set (`var p Priority`) is `0`, which is `Low`. Put
the safest value first.

> **Why is it written like this?** `iota` and the "repeat the line above"
> rule are Go language features. The `// 0`, `// 1` comments are the
> author's choice, to show you the values.

### Methods on your types

In *Methods* you gave structs methods. Any type you declare in your package
can have methods, including string and int types:

```go
// IsUS reports whether r is a US region.
func (r Region) IsUS() bool {
	return r == USEast
}

USEast.IsUS()              // true
Region("ap-south-1").IsUS() // false
```

> **Why is it written like this?** `(r Region)` before the name is the
> *receiver*: Go requires it to make a method. Inside, `r` is the value the
> method was called on. Naming it with one short letter is convention.

### Why bother?

With plain values, `Page(3, "production")` and `Page("production", 3)` look
alike to a tired human. With `func Page(s Severity, env Environment)`, the
swapped call doesn't compile.

## Your task

Your alerting service decides who gets woken up. The types `Environment`
and `Severity` and all their constants are already in `exercise.go`. You
fill in the method bodies only; don't change the `func` lines.

| `Severity` constant | Value |
|---|---|
| `Debug` | 0 |
| `Info` | 1 |
| `Warning` | 2 |
| `Error` | 3 |
| `Critical` | 4 |

### 1. Validate an environment

**Goal:** make `IsValid` return `true` only for `Dev`, `Staging` or `Production`.

Anything else returns `false`. Letter case matters, and so does the exact
spelling.

```go
Production.IsValid()             // => true
Environment("prod").IsValid()    // => false
Environment("Staging").IsValid() // => false
Environment("").IsValid()        // => false
```

**Think it through:**
- Inside the method, which name holds the environment being checked? (Look at the receiver.)
- You need three comparisons. Which operator means "or"?

### 2. Decide whether to page

**Goal:** make `ShouldPage` return whether this alert wakes the on-call engineer.

The receiver `s` is the severity; the parameter `env` is the environment.

| Severity | Pages? |
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

**Think it through:**
- Which rows of the table need to look at `env`, and which don't?
- For the `Error` row, a comparison is already a `bool`; you can return it.

### 3. Parse a severity

**Goal:** make `ParseSeverity` turn a name like `"warning"` into the matching `Severity` and `true`.

- The five names are `debug`, `info`, `warning`, `error`, `critical`.
- Accept any mix of upper and lower case (`"WARNING"`, `"Error"`).
- For any other text, including `""`, return `Debug` and `false`.

```go
ParseSeverity("critical") // => Critical, true
ParseSeverity("WARNING")  // => Warning, true
ParseSeverity("fatal")    // => Debug, false
```

`strings.ToLower` from *The strings package* helps. It needs an import
below the `package` line, the same shape as `import "fmt"`.

The tests print severities as numbers, so `Critical` shows up as `4` in a
failure message.

**Think it through:**
- If you lower-case the input first, how many spellings of each name do you need to check?
- Each match returns two values. Which constant goes with each name?

## Learn more

- [A Tour of Go: Methods on non-struct types](https://go.dev/tour/methods/3)
- [Go by Example: Enums](https://gobyexample.com/enums)
- [Go spec: Type definitions](https://go.dev/ref/spec#Type_definitions)
- [Go spec: Iota](https://go.dev/ref/spec#Iota)
