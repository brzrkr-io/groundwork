# Switch: the fallthrough bug

Someone copied this log-level code from C and expected cases to "fall
through" into the next one. In Go they don't. Fix the bug, then write a switch
of your own.

**New here:** empty cases (and why Go has no fallthrough by default), `strings.HasPrefix`.

## Learn

### Recap

- A `switch` runs the **first** matching `case`, then stops.
- One case can list several values, separated by commas.
- A *tagless* switch (`switch {`) checks true/false conditions from top to bottom.
- `default` handles everything no case matched.

### The C habit that breaks in Go

In C, an empty case "falls through" and runs the next case's code. In Go, an
empty case matches and then does **nothing**. The switch ends there.

```go
switch fruit {
case "apple": // matches "apple", runs nothing, switch ends
case "pear":
	return "green"
}
```

Here `"apple"` never reaches `return "green"`. To share one body between
several values, list them in a single case with commas:

```go
switch fruit {
case "lemon", "banana":
	return "yellow"
}
```

> **Why is it written like this?** Go requires that each case stops by itself.
> This removes a very common C bug (a forgotten `break`). Go does have a
> `fallthrough` keyword, but it's rare. The comma list is the normal style.

### Checking prefixes

`strings.HasPrefix(s, prefix)` returns `true` if the string `s` starts with
`prefix`. Because it returns true or false, it fits into a tagless switch:

```go
switch {
case strings.HasPrefix(host, "db-"):
	return "database"
default:
	return "unknown"
}
```

> **Why is it written like this?** `strings.` is required: `HasPrefix` lives
> in the `strings` package, which the file must import. The first argument is
> the value you check (often a parameter, so no quotes). The second is fixed
> text, so it goes in quotes.

## Your task

Your log pipeline turns level names into numbers and groups cloud regions by
area.

### 1. Fix the level numbers

**Goal:** make `LevelNumber(level)` return the right number for every level, including the two-name ones.

| `level` | Result |
|---|---|
| `"debug"` | `0` |
| `"info"` | `1` |
| `"warn"` or `"warning"` | `2` |
| `"error"` | `3` |
| `"fatal"` or `"panic"` | `4` |
| anything else (including `""`) | `-1` |

```go
LevelNumber("warn")  // => 2   (the starter returns -1)
LevelNumber("panic") // => 4
LevelNumber("trace") // => -1
```

**Think it through:**
- Which cases in the starter have nothing written under them?
- What happens in Go when one of those cases matches?
- Which two cases should become one?

### 2. Group regions by area

**Goal:** make `RegionArea(region)` return the area name, using a tagless switch with `strings.HasPrefix`.

| `region` starts with | Result |
|---|---|
| `"us-"` | `"North America"` |
| `"eu-"` | `"Europe"` |
| `"ap-"` | `"Asia Pacific"` |
| anything else | `"other"` |

The prefixes include the dash, so `"europe"` is `"other"`. The starter
already imports `strings`; replace its `if` with the switch.

```go
RegionArea("eu-west-1")      // => "Europe"
RegionArea("ap-southeast-2") // => "Asia Pacific"
RegionArea("sa-east-1")      // => "other"
```

**Think it through:**
- Which part of each `HasPrefix` call is the parameter, and which is fixed text in quotes?
- Which result belongs under `default`?

## Learn more

- [Effective Go: Switch](https://go.dev/doc/effective_go#switch)
- [Go spec: Switch statements](https://go.dev/ref/spec#Switch_statements)
- [strings.HasPrefix](https://pkg.go.dev/strings#HasPrefix)
