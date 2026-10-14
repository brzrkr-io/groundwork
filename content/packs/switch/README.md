# Switch: size, method, latency

A long `if`/`else if` chain that compares one value against many options is
hard to read. A `switch` lists the options neatly. Ops code is full of
these: environments, HTTP methods, severity levels.

**New here:** `switch`, `case`, `default`, tagless switch, `strings.ToUpper`.

## Learn

### Expression switch

```go
switch region {
case "eu-west-1":
	return "Ireland"
case "us-east-1":
	return "Virginia"
default:
	return "elsewhere"
}
```

Go compares `region` with each `case` value, top to bottom, and runs the
first match. `default` runs when nothing matches; it is optional.

> **Why is it written like this?** `switch`, `case`, `default`, the `{ }`
> and the `:` after each case are required. gofmt puts `case` at the same
> indent as `switch`; that looks odd at first but is standard.

### Several values in one case

Separate values with commas. The case matches if any of them is equal:

```go
case "mon", "tue", "wed", "thu", "fri":
	return "weekday"
```

### No fallthrough

Only one case runs, then the whole `switch` is done. You don't need `break`
like in C or Java. An empty case does *nothing*; it does not "fall into" the
next case. Use commas instead.

### Switching on a function call

The value after `switch` can be any expression:

```go
switch strings.ToLower(level) {
case "warn":
	// matches "warn", "WARN", "Warn"...
}
```

`strings.ToLower` and `strings.ToUpper` live in the `strings` package, so add
`import "strings"` below the `package` line (Go requires it).

### Tagless switch

Leave out the value and each `case` becomes a true/false condition. It's a
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

With `cpu` at 95 the result is `"hot"`, even though `cpu > 50` is also true:
the first true case wins. Order matters.

## Your task

You're building helpers for a deployment dashboard.

### 1. Pick an instance size

**Goal:** `InstanceSize(env)` returns an instance size using a `switch` on `env`.

| `env` | Result |
|---|---|
| `"production"` | `"large"` |
| `"staging"` | `"medium"` |
| `"dev"` or `"test"` | `"small"` |
| anything else (e.g. `"prod"`, `""`) | `"unknown"` |

```go
InstanceSize("test") // => "small"
InstanceSize("prod") // => "unknown"
```

**Think it through:**
- Each table row is one `case`. Which row becomes `default`?
- One row has two values. How do two values share one case?

### 2. Classify an HTTP method

**Goal:** `MethodKind(method)` returns `"read"`, `"write"` or `"unknown"`, for any capitalisation.

| Method | Result |
|---|---|
| `GET`, `HEAD`, `OPTIONS` | `"read"` |
| `POST`, `PUT`, `PATCH`, `DELETE` | `"write"` |
| anything else (e.g. `TRACE`, `""`) | `"unknown"` |

`"get"`, `"Get"` and `"GET"` must all give the same result. Switch on
`strings.ToUpper(method)` and add `import "strings"`.

```go
MethodKind("get")    // => "read"
MethodKind("Delete") // => "write"
MethodKind("TRACE")  // => "unknown"
```

**Think it through:**
- After `ToUpper`, which spelling do the case values need?

### 3. Rate a response time

**Goal:** `LatencyRating(ms)` rates a response time using a tagless switch.

| `ms` | Result |
|---|---|
| below 100 | `"fast"` |
| 100 to 499 | `"ok"` |
| 500 to 1999 | `"slow"` |
| 2000 and above | `"critical"` |

```go
LatencyRating(99)   // => "fast"
LatencyRating(100)  // => "ok"
LatencyRating(1999) // => "slow"
LatencyRating(2000) // => "critical"
```

**Think it through:**
- The first true case wins. Which limit should you check first?
- Once earlier cases are ruled out, each case only needs an upper limit.

## Learn more

- [A Tour of Go: Switch](https://go.dev/tour/flowcontrol/9)
- [A Tour of Go: Switch with no condition](https://go.dev/tour/flowcontrol/11)
- [Go by Example: Switch](https://gobyexample.com/switch)
