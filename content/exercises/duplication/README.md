# Removing duplication

Copy and paste is the fastest way to write code and the slowest way to
change it. When one rule lives in three copies, a fix lands in one copy
and the other two keep the bug. This exercise finds two kinds of
duplication in a config validator, folds each into one place, and then
shows the payoff: a new setting takes one line.

**New here:** DRY, duplicated logic vs duplicated knowledge,
`strings.Join`, package-level `var`.

## Learn

### Duplicated logic

Here are two copies of the same rule:

```go
if c.CPU < 1 || c.CPU > 64 {
	problems = append(problems, fmt.Sprintf("cpu %d is out of range 1-64", c.CPU))
}
if c.MaxCPU < 1 || c.MaxCPU > 64 {
	problems = append(problems, fmt.Sprintf("cpu %d is out of range 1-64", c.MaxCPU))
}
```

Spot the bug? The second message says `cpu`, not `max_cpu`. Someone copied
the block and missed one edit. Copies drift, and each copy is a place for
a bug to hide.

Write the rule once, with the parts that differ as parameters:

```go
func checkCPU(name string, value int) error {
	if value < 1 || value > 64 {
		return fmt.Errorf("%s %d is out of range 1-64", name, value)
	}
	return nil
}
```

The principle is called *DRY*: **Don't Repeat Yourself**. Every rule
should have one home.

### Duplicated knowledge

Duplication isn't only copied lines. A **fact** written in two places is
duplicated too, even if the code around it looks different:

```go
case "us-east-1", "eu-west-1":   // in one function
regions := "us-east-1, eu-west-1" // in an error message elsewhere
```

When a region is added, someone updates one and forgets the other. Put the
fact in one place, a package-level variable, and let every function read
it:

```go
// regions lists every region we deploy to.
var regions = []string{"us-east-1", "eu-west-1"}
```

`slices.Contains(regions, r)` asks whether `r` is in the list, and
`strings.Join(regions, ", ")` turns the list into
`"us-east-1, eu-west-1"` for a message.

> **Why is it written like this?** **Go requires** `var`, not `:=`, outside
> a function, and a slice can't be a `const` (constants are only numbers,
> strings and booleans). A comment on a package-level variable is
> **convention**, like on a function.

### Similar is not the same

Three calls to one helper are not duplication:

```go
if err := checkCPU("cpu", c.CPU); err != nil {
	problems = append(problems, err.Error())
}
```

Repeated once per field, these lines look alike, but the rule itself lives
in one place. Don't chase every similar-looking line into an abstraction.
Merge code that must **change together**, because it encodes the same
rule. Code that only looks alike today may need to change differently
tomorrow. The *Abstraction* guide calls merging that kind "the wrong
abstraction".

### How the tests check

`structure_test.go` reads `exercise.go` with Go's `go/ast` package. It
checks that `Validate` calls `checkPort` instead of holding port numbers
itself, and that each environment name is written only once in the file.
The behaviour checks run first.

## Your task

`Validate` checks a service's config before deploy. The rules:

| Field | Rule | Message |
|---|---|---|
| `Environment` | one of `development`, `staging`, `production` | `environment "<env>" is not one of development, staging, production` |
| each port | 1 to 65535 | `<name> <port> is out of range 1-65535` |
| each port | 1024 or above | `<name> <port> needs root; use 1024 or above` |

The port names in messages are `http_port`, `metrics_port`, `admin_port`
and `grpc_port`. Problems come in the order of the table and the fields;
a port gets at most one message (out of range wins).

### 1. Write the port rule once

**Goal:** write `checkPort(name string, port int) error` and call it from
`Validate` for each of the three ports it checks now.

`checkPort` returns the problem as an error, or `nil`. `Validate` adds the
error's text to `problems`. When you're done, `Validate` has no numbers in
it. The admin port's messages are fixed on the way:

```go
checkPort("admin_port", 443)
// => error "admin_port 443 needs root; use 1024 or above"
```

**Think it through:**
- Lay the three blocks side by side. What differs between them? Those become parameters.
- How do you get an error's message as a string? See *Errors are values*.

### 2. Write the environments once

**Goal:** declare `var environments` listing `development`, `staging` and
`production`, and use it in `Validate` (the check and the message) and in
`AlertChannel`.

Right now the two functions disagree: `Validate` rejects `staging`.

```go
Validate(Config{Environment: "prod", ...})
// => [`environment "prod" is not one of development, staging, production`]
AlertChannel("staging") // => "#alerts-staging"
AlertChannel("qa")      // => "#alerts-unknown"
```

Each environment name may appear only once in `exercise.go`. Declare the
list where `TODO(2)` is.

**Think it through:**
- Which `slices` function asks "is this value in the list"?
- Which `strings` function joins a list with `", "` between the items?

### 3. Check the gRPC port

**Goal:** make `Validate` check `GRPCPort` with the same rule, named
`grpc_port`, after the admin port.

```go
// GRPCPort: 50
// => "grpc_port 50 needs root; use 1024 or above"
```

**Think it through:**
- How many lines does this take now? How many would it have taken before task 1?

## Learn more

- [Don't repeat yourself (Wikipedia)](https://en.wikipedia.org/wiki/Don%27t_repeat_yourself)
- [The Wrong Abstraction (Sandi Metz)](https://sandimetz.com/blog/2016/1/20/the-wrong-abstraction)
- [slices.Contains](https://pkg.go.dev/slices#Contains) and [strings.Join](https://pkg.go.dev/strings#Join)
