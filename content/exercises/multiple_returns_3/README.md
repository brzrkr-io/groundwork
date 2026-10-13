# Multiple return values: image refs and TTLs

Deploy tools are mostly parsers: they take short strings people type, such as
`nginx:1.27` or `5m`, and turn them into values a program can trust. Here
you write two of them.

## Learn

### Recap

Return several values with `return a, b, err`, receive them with
`a, b, err := f()`. On success the error is `nil`. On failure, return zero
values plus an error from `errors.New` or `fmt.Errorf`. `strings.Cut` splits
at a separator, `strings.CutSuffix` removes an ending, and `strconv.Atoi`
turns text into an `int`, as in *Multiple return values: don't drop the
error*.

## Your task

### 1. Parse a container image reference

Write `ParseImage`. An image reference is `name:tag`. Split it at the first
`:`.

- No `:` means the tag is `"latest"`.
- Return an error when the name is empty, or when there is a `:` but the
  tag after it is empty.
- On error, return `"", ""`. The messages are up to you.

The references in this exercise never contain a registry port, so you don't
need to worry about a second `:`.

```go
ParseImage("nginx:1.27")       // => "nginx", "1.27", nil
ParseImage("redis")            // => "redis", "latest", nil
ParseImage("team/api:v2.3.0")  // => "team/api", "v2.3.0", nil
ParseImage(":1.27")            // => "", "", (an error)
ParseImage("nginx:")           // => "", "", (an error)
ParseImage("")                 // => "", "", (an error)
```

### 2. Parse a cache TTL

Write `ParseTTL`. A *TTL* (time to live) is a whole number followed by one
unit: `s` (seconds), `m` (minutes) or `h` (hours). Return the TTL in
seconds.

- The number must be a whole number above 0.
- Any other unit, or no unit, is an error.
- On error, return `0`. The messages are up to you.

```go
ParseTTL("90s") // => 90, nil
ParseTTL("5m")  // => 300, nil
ParseTTL("2h")  // => 7200, nil
ParseTTL("0s")  // => 0, (an error)
ParseTTL("10")  // => 0, (an error)
ParseTTL("1d")  // => 0, (an error)
ParseTTL("xm")  // => 0, (an error)
```

## Learn more

- [Go by Example: Errors](https://gobyexample.com/errors)
- [strings.Cut](https://pkg.go.dev/strings#Cut)
- [strings.CutSuffix](https://pkg.go.dev/strings#CutSuffix)
