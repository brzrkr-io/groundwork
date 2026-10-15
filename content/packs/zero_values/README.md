# Zero values: default the config

In Go, a variable never holds garbage. If you don't give it a value, it gets
its type's *zero value*. Config code leans on this all the time: "the user
left this field empty, so use the default".

**New here:** the zero value of each basic type.

## Learn

### Declaring without a value

As in *Variables*, `:=` always needs a value, but `var` lets you write the
name and the *type* and leave the value out. Go then fills in the type's
zero value:

```go
var replicas int     // replicas is 0
var region string    // region is ""
var enabled bool     // enabled is false
var load float64     // load is 0
```

> **Why is it written like this?** The order `var name type` is what Go
> requires (not `var type name`). The names are the author's choice. Lining
> up the comments is gofmt's job.

| Type | Zero value |
|---|---|
| `int` | `0` |
| `float64` | `0` |
| `string` | `""` (the empty string) |
| `bool` | `false` |

### Applying defaults

"Not set" means "zero", so check for the zero value and use a default:

```go
func RegionOrDefault(region string) string {
	if region == "" {
		return "us-east-1"
	}
	return region
}
```

When a function has several defaults, overwrite each parameter and carry on:

```go
if retries == 0 {
	retries = 3
}
```

> **Why is it written like this?** `==` compares; a single `=` assigns.
> Inside the `if` you use `=`, not `:=` and not `const`: the variable already
> exists, you are just giving it a new value.

### A limitation

This only works when zero is never a real setting. For a `bool`, `false`
means both "unset" and "off". In this exercise, zero always means "use the
default".

## Your task

You're writing the config loader for a small service. Missing settings
arrive as zero values.

### 1. Declare variables with their zero values

**Goal:** in `ZeroSummary`, change the four `:=` lines to `var` lines with no value.

The types are: `port` is an `int`, `host` a `string`, `debug` a `bool`,
`ratio` a `float64`. Keep the names. Leave the `fmt.Sprintf` line alone.

```go
ZeroSummary()
// => `port=0 host="" debug=false ratio=0`
```

**Think it through:**
- Each new line has three parts: the keyword, the name, the type. Where does the value go? (Nowhere.)

### 2. Default the port

**Goal:** `PortOrDefault(port)` returns `8080` when `port` is `0`, otherwise `port`.

```go
PortOrDefault(0)    // => 8080
PortOrDefault(9090) // => 9090
PortOrDefault(443)  // => 443
```

**Think it through:**
- What is the zero value of an `int`? That's what you compare `port` with.
- The starter's `return port` is already the "otherwise" case.

### 3. Build a connection string

**Goal:** `ConnectionString(host, port, timeoutSeconds)` fills in defaults, then returns `host:port timeout=Ns`.

- an empty `host` becomes `"localhost"`
- a `port` of `0` becomes `8080` (call `PortOrDefault`)
- a `timeoutSeconds` of `0` becomes `30`

The `fmt.Sprintf` line is already correct.

```go
ConnectionString("", 0, 0)               // => "localhost:8080 timeout=30s"
ConnectionString("db.internal", 5432, 5) // => "db.internal:5432 timeout=5s"
ConnectionString("cache", 0, 10)         // => "cache:8080 timeout=10s"
ConnectionString("", 9000, 0)            // => "localhost:9000 timeout=30s"
```

**Think it through:**
- Fix each parameter in turn, *above* the `return` line.
- For the port, pass the parameter `port` into your task 2 function and store what it returns.

## Learn more

- [A Tour of Go: Zero values](https://go.dev/tour/basics/12)
- [A Tour of Go: Variables](https://go.dev/tour/basics/8)
- [Go spec: The zero value](https://go.dev/ref/spec#The_zero_value)
