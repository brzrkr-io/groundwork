# Switch: size, method, latency

A long `if`/`else if` chain that compares one value against many options is
hard to read. A `switch` lists the options neatly, one `case` per line. Ops
code is full of these: environments, HTTP methods, severity levels.

## Learn

### Expression switch

```go
switch env {
case "production":
	return "large"
case "staging":
	return "medium"
default:
	return "unknown"
}
```

Go compares `env` with each `case` value from top to bottom and runs the
first one that matches. `default` runs when nothing matches; it is optional.

### Several values in one case

Separate values with commas. The case matches if any of them is equal:

```go
case "dev", "test":
	return "small"
```

### No fallthrough

Only one case runs. When its code finishes, the whole `switch` is done. You
don't need `break` like in C or Java. (An empty case does *nothing*, it does
not "fall into" the next case. Use commas instead.)

### Switching on an expression

The value after `switch` can be any expression, such as a function call:

```go
switch strings.ToLower(level) {
case "warn":
	// matches "warn", "WARN", "Warn"...
}
```

### Tagless switch

Leave out the value and each `case` becomes a boolean condition. This is a
tidy replacement for an `if`/`else if` chain:

```go
switch {
case cpu > 90:
	return "hot"
case cpu > 50:
	return "warm"
default:
	return "cool"
}
```

Cases are checked from top to bottom, so with `cpu` at 95 the result is
`"hot"`, even though `cpu > 50` is also true. Put the strictest check first.

## Your task

You're building helpers for a deployment dashboard.

### 1. Pick an instance size

`InstanceSize(env)` returns:

| `env` | Result |
|---|---|
| `"production"` | `"large"` |
| `"staging"` | `"medium"` |
| `"dev"` or `"test"` | `"small"` |
| anything else | `"unknown"` |

Use a `switch` with a comma-separated case for `"dev"` and `"test"`.

```go
InstanceSize("test")  // => "small"
InstanceSize("prod")  // => "unknown"
```

### 2. Classify an HTTP method

`MethodKind(method)` works for any capitalisation (`"get"`, `"Get"`, `"GET"`).
Switch on `strings.ToUpper(method)`; you need to add `import "strings"` at the
top of the file.

| Method | Result |
|---|---|
| `GET`, `HEAD`, `OPTIONS` | `"read"` |
| `POST`, `PUT`, `PATCH`, `DELETE` | `"write"` |
| anything else | `"unknown"` |

```go
MethodKind("get")    // => "read"
MethodKind("Delete") // => "write"
MethodKind("TRACE")  // => "unknown"
```

### 3. Rate a response time

`LatencyRating(ms)` uses a tagless switch:

| `ms` | Result |
|---|---|
| below 100 | `"fast"` |
| 100 to 499 | `"ok"` |
| 500 to 1999 | `"slow"` |
| 2000 and above | `"critical"` |

```go
LatencyRating(99)   // => "fast"
LatencyRating(100)  // => "ok"
LatencyRating(2000) // => "critical"
```

## Learn more

- [A Tour of Go: Switch](https://go.dev/tour/flowcontrol/9)
- [A Tour of Go: Switch with no condition](https://go.dev/tour/flowcontrol/11)
- [Go by Example: Switch](https://gobyexample.com/switch)
