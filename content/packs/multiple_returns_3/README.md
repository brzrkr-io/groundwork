# Multiple return values: image refs and TTLs

Deploy tools are mostly parsers: they take short strings people type, such as
`nginx:1.27` or `5m`, and turn them into values a program can trust. Here
you write two of them.

**New here:** nothing; practice. (Named results in a signature are explained below.)

## Learn

### Recap

You met all of this in *Multiple return values: don't drop the error*:

| Tool | What it gives you |
| --- | --- |
| `strings.Cut(s, sep)` | text before `sep`, text after it, and a `bool`: was `sep` found? |
| `strings.CutSuffix(s, end)` | `s` without `end`, and a `bool`: did `s` end with it? |
| `strconv.Atoi(s)` | the `int` in `s`, and an error if `s` isn't a whole number |
| `errors.New("...")` / `fmt.Errorf("... %q", x)` | a new error value |

On success the error is `nil`. On failure, return *zero values* (`""`, `0`)
plus an error.

```go
host, port, found := strings.Cut("db:5432", ":")
// host == "db", port == "5432", found == true

base, ok := strings.CutSuffix("report.txt", ".txt")
// base == "report", ok == true
```

> **Why is it written like this?** Go requires one variable on the left of
> `:=` for each value a function returns, so `strings.Cut` needs three names.
> The names themselves (`host`, `found`) are your choice.

### Named results

Your `ParseImage` line names its results, like this example:

```go
func SplitHost(addr string) (host, port string, err error)
```

The part after the parameters is the list of *results*. Here they have
names. `host, port string` means both are strings. You can still finish
with an ordinary `return a, b, err`.

> **Why is it written like this?** Naming results is the author's choice;
> it documents what each value means. Don't change the signature: the
> tests call it as written.

### Checking cases in order

A parser is a list of `if` checks, each ending in `return`. The first one
that matches wins, so put the bad cases first and the success case last.

## Your task

You're writing the input checks for a deploy tool.

### 1. Parse a container image reference

**Goal:** make `ParseImage` split `name:tag` into a name and a tag, or return an error.

Split at the first `:`.

- No `:` at all means the tag is the fixed text `"latest"`.
- Return an error when the name is empty, or when there is a `:` but the
  tag after it is empty.
- On error, return `"", ""` plus the error. The messages are up to you.

References here never contain a registry port, so you don't need to worry
about a second `:`.

```go
ParseImage("nginx:1.27")       // => "nginx", "1.27", nil
ParseImage("redis")            // => "redis", "latest", nil
ParseImage("team/api:v2.3.0")  // => "team/api", "v2.3.0", nil
ParseImage(":1.27")            // => "", "", (an error)
ParseImage("nginx:")           // => "", "", (an error)
ParseImage("")                 // => "", "", (an error)
```

**Think it through:**
- Which of the three values from `strings.Cut` tells you "no `:` at all"?
- `"redis"` and `"nginx:"` both have an empty tag. What tells them apart?
- The starter has no `import` yet. Add an `import ( ... )` block under the
  `package` line with the packages you use (`strings`, and `errors` or `fmt`).

### 2. Parse a cache TTL

**Goal:** make `ParseTTL` turn a TTL like `"5m"` into a number of seconds, or return an error.

A *TTL* (time to live) is a whole number followed by one unit:

| Unit | Means | Seconds per unit |
| --- | --- | --- |
| `s` | seconds | 1 |
| `m` | minutes | 60 |
| `h` | hours | 3600 |

- The number must be a whole number above 0 (`0s` and `-5m` are errors).
- Any other unit, no unit, or an empty string is an error.
- On error, return `0` plus the error. The messages are up to you.

```go
ParseTTL("90s") // => 90, nil
ParseTTL("5m")  // => 300, nil
ParseTTL("2h")  // => 7200, nil
ParseTTL("1s")  // => 1, nil
ParseTTL("0s")  // => 0, (an error)
ParseTTL("-5m") // => 0, (an error)
ParseTTL("10")  // => 0, (an error)
ParseTTL("1d")  // => 0, (an error)
ParseTTL("xm")  // => 0, (an error)
ParseTTL("")    // => 0, (an error)
```

**Think it through:**
- Find the unit first: which tool tells you whether a string ends with `"m"`?
- Once the unit is gone, what's left is the number text. Which tool turns
  it into an `int`, and what does it return for `"x"`?
- Add `strconv` to your `import` block.

## Learn more

- [Go by Example: Errors](https://gobyexample.com/errors)
- [strings.Cut](https://pkg.go.dev/strings#Cut)
- [strings.CutSuffix](https://pkg.go.dev/strings#CutSuffix)
- [strconv.Atoi](https://pkg.go.dev/strconv#Atoi)
