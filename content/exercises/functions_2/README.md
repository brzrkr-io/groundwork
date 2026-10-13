# Functions: swapped arguments and helpers

Go matches arguments to parameters by position, not by name. Pass two
numbers in the wrong order and the compiler can't help you. This exercise
starts with exactly that bug, then has you pull a helper out of a function.

## Learn

### Recap

A function lists typed parameters, `func Percent(part, whole float64) float64`,
and callers pass values in the same order. Functions can call each other, so
one job can live in one small helper that several functions reuse.

### Printing a percent sign

In `fmt.Sprintf`, `%` starts a verb like `%d`. To print a literal `%`, write
`%%`:

```go
fmt.Sprintf("%.1f%%", 37.5) // => "37.5%"
```

## Your task

Your monitoring bot prints CPU usage per service and builds container image
references.

### 1. Fix the usage line

`Percent(part, whole float64) float64` already works: it returns what percent
`part` is of `whole`. `UsageLine(service string, used, total float64) string`
calls it with the arguments swapped. Fix the call. The format is
`<service>: <percent with one decimal>% CPU`.

```go
UsageLine("api", 3, 8)     // => "api: 37.5% CPU"
UsageLine("worker", 1, 4)  // => "worker: 25.0% CPU"
```

### 2. Write the tag helper

Write `TagOrLatest(tag string) string`: return `tag`, or `"latest"` when `tag`
is empty.

```go
TagOrLatest("1.4.2") // => "1.4.2"
TagOrLatest("")      // => "latest"
```

### 3. Build the image reference

Change `ImageRef(registry, name, tag string) string` so it returns
`<registry>/<name>:<tag>`, using `TagOrLatest` for the tag.

```go
ImageRef("ghcr.io/acme", "api", "1.4.2") // => "ghcr.io/acme/api:1.4.2"
ImageRef("docker.io", "nginx", "")       // => "docker.io/nginx:latest"
```

## Learn more

- [A Tour of Go: Functions continued](https://go.dev/tour/basics/5)
- [fmt package: Printing](https://pkg.go.dev/fmt#hdr-Printing)
- [Docker: image tags](https://docs.docker.com/reference/cli/docker/image/tag/)
