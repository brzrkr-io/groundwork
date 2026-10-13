# Functions: swapped arguments and helpers

Go matches arguments to parameters by position, not by name. Pass two
numbers in the wrong order and the compiler can't help you. This exercise
starts with exactly that bug, then has you pull a helper out of a function.

**New here:** argument order, `%%` in format strings, small helper functions.

## Learn

### Arguments go in by position

```go
func Divide(top, bottom float64) float64 {
	return top / bottom
}

Divide(10, 2) // top is 10, bottom is 2 => 5
Divide(2, 10) // top is 2, bottom is 10 => 0.2
```

The first value you pass becomes the first parameter, the second becomes the
second. The names at the call site don't matter: Go never matches by name.

> **Why is it written like this?** Position-only matching is a Go rule.
> Both arguments are `float64`, so a swap still compiles. Only a test (or
> reading the function line carefully) catches it.

### Printing a percent sign

In `fmt.Sprintf`, `%` starts a verb like `%d`. To print a literal `%`, write
`%%`:

```go
fmt.Sprintf("disk %d%% full", 80) // => "disk 80% full"
```

### Helpers and defaults

A *helper* is a small function that does one job, so other functions can
reuse it. A common job is "use this value, or a default if it's empty":

```go
func NameOrAnon(name string) string {
	if name == "" {
		return "anonymous"
	}
	return name
}
```

`""` is the *empty string*. Comparing with `==` tells you whether a string is empty.

### Joining strings

`+` glues strings together. Text inside quotes is fixed text; names outside
quotes are values:

```go
user + "@" + host // "ops" and "db1" => "ops@db1"
```

Every separator (`/`, `:`, a space) must be its own quoted piece.
A name inside quotes is just letters: `"user"` prints the word `user`.

## Your task

Your monitoring bot prints CPU usage per service and builds container image
references.

### 1. Fix the usage line

**Goal:** fix the call to `Percent` inside `UsageLine` so the arguments are in the right order.

`Percent(part, whole float64) float64` already works: it returns what percent
`part` is of `whole`. `UsageLine(service string, used, total float64) string`
calls it with its arguments swapped. The format string is already correct:
`<service>: <percent with one decimal>% CPU`.

```go
UsageLine("api", 3, 8)    // => "api: 37.5% CPU"
UsageLine("worker", 1, 4) // => "worker: 25.0% CPU"
UsageLine("db", 2, 2)     // => "db: 100.0% CPU"
```

**Think it through:**
- Read `Percent`'s function line: which parameter comes first?
- Which of `used` and `total` is the *part*, and which is the *whole*?

### 2. Write the tag helper

**Goal:** make `TagOrLatest(tag string) string` return `tag`, or `"latest"` when `tag` is empty.

```go
TagOrLatest("1.4.2")  // => "1.4.2"
TagOrLatest("stable") // => "stable"
TagOrLatest("")       // => "latest"
```

**Think it through:**
- How do you test whether a string is empty?
- Two outcomes means two `return` lines.

### 3. Build the image reference

**Goal:** make `ImageRef(registry, name, tag string) string` return `<registry>/<name>:<tag>`.

`registry` and `name` come from the parameters. The tag comes from calling
`TagOrLatest` with the `tag` parameter. The fixed text is `/` between registry
and name, and `:` before the tag. No spaces.

```go
ImageRef("ghcr.io/acme", "api", "1.4.2")  // => "ghcr.io/acme/api:1.4.2"
ImageRef("docker.io", "nginx", "")        // => "docker.io/nginx:latest"
ImageRef("registry.local", "worker", "v2") // => "registry.local/worker:v2"
```

**Think it through:**
- List the five pieces in order. Which are names, which are quoted text?
- A function call like `TagOrLatest(...)` is a string value, so it can sit in a `+` chain.

## Learn more

- [A Tour of Go: Functions continued](https://go.dev/tour/basics/5)
- [fmt package: Printing](https://pkg.go.dev/fmt#hdr-Printing)
- [Docker: image tags](https://docs.docker.com/reference/cli/docker/image/tag/)
