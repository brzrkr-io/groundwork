# Switch: the fallthrough bug

Someone ported this log-level code from C and expected cases to "fall
through". In Go they don't. Fix the bug, then write a switch of your own.

## Learn

### Recap

A `switch` runs the first matching `case` and then stops. List several values
in one case with commas (`case "a", "b":`). A tagless `switch { ... }` checks
boolean conditions top to bottom. `default` handles everything else.

### The C habit that breaks in Go

In C, an empty case runs into the next one. In Go, an empty case matches and
then does nothing:

```go
switch s {
case "warn": // matches "warn", runs nothing, the switch ends
case "warning":
	return 2
}
```

The Go way is one case with both values: `case "warn", "warning":`.

### Checking prefixes

`strings.HasPrefix(s, prefix)` reports whether `s` starts with `prefix`. It
fits nicely in a tagless switch:

```go
switch {
case strings.HasPrefix(name, "web-"):
	return "frontend"
}
```

## Your task

Your log pipeline turns level names into numbers and groups cloud regions by
area.

### 1. Fix the level numbers

`LevelNumber(level)` returns:

| `level` | Result |
|---|---|
| `"debug"` | `0` |
| `"info"` | `1` |
| `"warn"` or `"warning"` | `2` |
| `"error"` | `3` |
| `"fatal"` or `"panic"` | `4` |
| anything else | `-1` |

The starter gets the two-name levels wrong because it relies on fallthrough.

```go
LevelNumber("warn")  // => 2
LevelNumber("panic") // => 4
LevelNumber("trace") // => -1
```

### 2. Group regions by area

`RegionArea(region)` uses a tagless switch with `strings.HasPrefix`:

| Region starts with | Result |
|---|---|
| `"us-"` | `"North America"` |
| `"eu-"` | `"Europe"` |
| `"ap-"` | `"Asia Pacific"` |
| anything else | `"other"` |

The starter already imports `strings`.

```go
RegionArea("eu-west-1")      // => "Europe"
RegionArea("ap-southeast-2") // => "Asia Pacific"
RegionArea("sa-east-1")      // => "other"
```

## Learn more

- [Effective Go: Switch](https://go.dev/doc/effective_go#switch)
- [strings.HasPrefix](https://pkg.go.dev/strings#HasPrefix)
