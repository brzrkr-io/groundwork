# Zero values: default the config

In Go, a variable never holds garbage. If you don't give it a value, it gets
its type's *zero value*. Config code leans on this all the time: "the user
left this field empty, so use the default".

## Learn

### Declaring without a value

Up to now you created variables with `:=`, which always needs a value. With
`var` you can name a variable and its type and leave the value out:

```go
var port int      // port is 0
var host string   // host is ""
var debug bool    // debug is false
var ratio float64 // ratio is 0
```

Each type has its own zero value:

| Type | Zero value |
|---|---|
| `int` | `0` |
| `float64` | `0` |
| `string` | `""` (the empty string) |
| `bool` | `false` |

### Applying defaults

Because "not set" means "zero", you can check for the zero value and swap in
a default:

```go
func RegionOrDefault(region string) string {
	if region == "" {
		return "us-east-1"
	}
	return region
}
```

You can also overwrite a parameter and carry on, which is handy when there are
several defaults in one function:

```go
if retries == 0 {
	retries = 3
}
```

### A limitation

This trick only works when the zero value is never a real setting. For a
`bool`, `false` is both "unset" and "switched off", so you can't tell them
apart. The same goes for a port of `0` if 0 ever meant something. For the
settings in this exercise, zero always means "use the default".

## Your task

You're writing the config loader for a small service. Missing settings arrive
as zero values and need sensible defaults.

### 1. Declare variables with their zero values

`ZeroSummary` creates four variables with `:=` and real values. Change the
four declarations to use `var` with no value, so each one holds its zero
value: `port` is an `int`, `host` a `string`, `debug` a `bool` and `ratio` a
`float64`. Leave the `fmt.Sprintf` line alone. The result must be:

```text
port=0 host="" debug=false ratio=0
```

### 2. Default the port

`PortOrDefault` returns `8080` when `port` is `0`, and `port` otherwise.

```go
PortOrDefault(0)    // => 8080
PortOrDefault(9090) // => 9090
```

### 3. Build a connection string

`ConnectionString(host, port, timeoutSeconds)` fills in defaults and then
formats the result as `host:port timeout=Ns`:

- an empty `host` becomes `"localhost"`
- a `port` of `0` becomes `8080` (call `PortOrDefault`)
- a `timeoutSeconds` of `0` becomes `30`

```go
ConnectionString("", 0, 0)            // => "localhost:8080 timeout=30s"
ConnectionString("db.internal", 5432, 5) // => "db.internal:5432 timeout=5s"
ConnectionString("cache", 0, 10)      // => "cache:8080 timeout=10s"
```

## Learn more

- [A Tour of Go: Zero values](https://go.dev/tour/basics/12)
- [A Tour of Go: Variables](https://go.dev/tour/basics/8)
- [Go spec: The zero value](https://go.dev/ref/spec#The_zero_value)
