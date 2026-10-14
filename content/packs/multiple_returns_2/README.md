# Multiple return values: don't drop the error

Ignoring an error is the classic way a config typo turns into an outage:
`replicas: three` silently becomes zero replicas. Here you fix a parser that
does exactly that, and write two more that check every result.

**New here:** named results, `strings.TrimSpace`, `strings.CutSuffix`.

## Learn

### Recap

As in *Multiple return values*: errors come last, `nil` means success, and
you check right away with `if err != nil`. `_` throws a value away. That's
fine for values you truly don't need, and dangerous for errors.

### Named results

You can give results names in the function line. They document what each
value means, and work like variables declared at the top of the function:

```go
// Bounds returns the lowest and highest value, and whether nums had any.
func Bounds(nums []int) (low, high int, ok bool) {
	...
	return 1, 9, true
}
```

You still write `return` with the values, just as before.

> **Why is it written like this?** Names are optional (author's choice);
> types are required. As with parameters, `low, high int` shares one type.

### Trimming spaces

`strings.TrimSpace` removes spaces from both ends, not the middle:

```go
strings.TrimSpace("  web 1  ") // => "web 1"
```

It returns a *new* string. To keep it, assign it: `name = strings.TrimSpace(name)`.

### strings.CutSuffix

`strings.CutSuffix` removes an ending from a string, if it's there, and
tells you whether it was:

```go
rest, found := strings.CutSuffix("30s", "s") // "30", true
rest, found = strings.CutSuffix("5m", "s")   // "5m", false
```

### `if` with a short statement

An `if` can run one statement before its condition, separated by `;`. The
variables it creates only exist inside that `if`:

```go
if rest, found := strings.CutSuffix(d, "s"); found {
	// use rest here
}
```

This keeps tries like "is it seconds? is it minutes?" from clashing names.

## Your task

You're writing the parsers for a small deploy tool that reads Kubernetes-style
settings from text.

### 1. Fix ParseReplicas

**Goal:** make `ParseReplicas(s string) (int, error)` return an error instead of silently turning bad text into 0.

Return `0` and a non-nil error when:

- the text is not a whole number (including `""`),
- the number is negative.

Zero is allowed (it scales a service down). The messages are up to you.

```go
ParseReplicas("3")     // => 3, nil
ParseReplicas("12")    // => 12, nil
ParseReplicas("0")     // => 0, nil
ParseReplicas("three") // => 0, (an error)
ParseReplicas("")      // => 0, (an error)
ParseReplicas("-2")    // => 0, (an error)
```

**Think it through:**
- Which value does the starter throw away with `_`? Give it a name.
- `strconv.Atoi("-2")` succeeds, so the negative check is a separate step.
- You'll need `errors` or `fmt` in the imports for your own error.

### 2. Finish ParseLabel

**Goal:** make `ParseLabel(label string) (key, value string, ok bool)` return the trimmed key and value of a `key=value` label.

- Split at the **first** `=` (the starter already does this with `strings.Cut`).
- Trim spaces from both the key and the value.
- Return `"", "", false` when there is no `=`, or when the key is empty *after* trimming.
- An empty value is fine.

```go
ParseLabel("env=prod")     // => "env", "prod", true
ParseLabel(" team = sre ") // => "team", "sre", true
ParseLabel("tier=")        // => "tier", "", true
ParseLabel("query=a=b")    // => "query", "a=b", true
ParseLabel("env")          // => "", "", false
ParseLabel("=prod")        // => "", "", false
ParseLabel("  =prod")      // => "", "", false
```

**Think it through:**
- Why must you trim the key *before* checking whether it's empty? (Look at `"  =prod"`.)
- Two reasons to fail; can one `if` check both?

### 3. Write ParseMemoryMiB

**Goal:** make `ParseMemoryMiB(size string) (int, error)` return a memory size in Mi.

Sizes end in a unit: `Mi` (mebibytes) or `Gi` (gibibytes; 1 Gi is 1024 Mi).
Return `0` and a non-nil error when:

- the ending is anything else, or missing (`"512"`, `"1G"`, `""`),
- the number in front isn't a whole number (`"lotsMi"`).

The messages are up to you.

```go
ParseMemoryMiB("512Mi")  // => 512, nil
ParseMemoryMiB("64Mi")   // => 64, nil
ParseMemoryMiB("2Gi")    // => 2048, nil
ParseMemoryMiB("1Gi")    // => 1024, nil
ParseMemoryMiB("512")    // => 0, (an error)
ParseMemoryMiB("1G")     // => 0, (an error)
ParseMemoryMiB("lotsMi") // => 0, (an error)
ParseMemoryMiB("")       // => 0, (an error)
```

**Think it through:**
- Try `strings.CutSuffix` with `"Mi"`, then with `"Gi"`. What's left after cutting?
- The leftover text still needs `strconv.Atoi`, and its error still needs checking.
- If neither suffix matched, what do you return? Add the imports you need.

## Learn more

- [Effective Go: Named result parameters](https://go.dev/doc/effective_go#named-results)
- [strings.CutSuffix](https://pkg.go.dev/strings#CutSuffix)
- [strconv.Atoi](https://pkg.go.dev/strconv#Atoi)
