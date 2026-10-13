# Multiple return values: don't drop the error

Ignoring an error is the classic way a config typo turns into an outage:
`replicas: three` silently becomes zero replicas. Here you fix a parser that
does exactly that, and write two more that check every result.

## Learn

### Recap

A function can return several values, as in *Multiple return values*. Errors
come last, `nil` means success, and you check them right away with
`if err != nil { return ..., err }`. `_` throws a value away, which is
handy for values you truly don't need, and dangerous for errors.

### Named results

You can give results names in the signature. They work like variables
declared at the top of the function, starting at their zero values, and they
document what each value means:

```go
// ParseLabel returns the key, the value, and whether the label was valid.
func ParseLabel(label string) (key, value string, ok bool) {
	...
	return k, v, true
}
```

You still write `return` with the values, just as before.

### strings.CutSuffix

`strings.CutSuffix` removes an ending from a string, if it is there, and
tells you whether it was:

```go
rest, found := strings.CutSuffix("512Mi", "Mi") // "512", true
rest, found = strings.CutSuffix("2Gi", "Mi")    // "2Gi", false
```

## Your task

You're writing the parsers for a small deploy tool that reads Kubernetes-style
settings from text.

### 1. Fix ParseReplicas

`ParseReplicas` throws away the error from `strconv.Atoi`, so `"three"`
becomes 0 replicas. Change it to return a non-nil error when the text is not
a whole number, and also when the number is negative. Zero is allowed (it
scales a service down). On error, return `0`. The messages are up to you.

```go
ParseReplicas("3")     // => 3, nil
ParseReplicas("0")     // => 0, nil
ParseReplicas("three") // => 0, (an error)
ParseReplicas("-2")    // => 0, (an error)
```

### 2. Finish ParseLabel

A label looks like `key=value`. Split it at the first `=` with `strings.Cut`
and trim spaces from both parts with `strings.TrimSpace`. Return
`"", "", false` when there is no `=` or when the key is empty after
trimming. An empty value is fine.

```go
ParseLabel("env=prod")      // => "env", "prod", true
ParseLabel(" team = sre ")  // => "team", "sre", true
ParseLabel("tier=")         // => "tier", "", true
ParseLabel("env")           // => "", "", false
ParseLabel("=prod")         // => "", "", false
```

### 3. Write ParseMemoryMiB

Memory sizes end in a unit: `Mi` (mebibytes) or `Gi` (gibibytes; 1 Gi is
1024 Mi). Return the size in Mi.

- `"512Mi"` → 512, `"2Gi"` → 2048.
- Any other ending, or no ending, is an error: `"512"`, `"1G"`.
- The number in front must be a whole number: `"lotsMi"` is an error.

On error, return `0`. The messages are up to you. Use `strings.CutSuffix`
to try `Mi`, then `Gi`.

## Learn more

- [Effective Go: Named result parameters](https://go.dev/doc/effective_go#named-results)
- [strings.CutSuffix](https://pkg.go.dev/strings#CutSuffix)
- [strconv.Atoi](https://pkg.go.dev/strconv#Atoi)
