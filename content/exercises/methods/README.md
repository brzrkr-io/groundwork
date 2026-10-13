# Methods

A *method* is a function that belongs to a type. Instead of passing a host to
a helper function, you ask the host itself: `h.Address()`. Ops tooling is full
of these: a host knows its address, a deployment knows its image tag.

**New here:** method declarations, receivers, calling methods with a dot, `strings.Cut`.

## Learn

### Declaring a method

A method is a function with one extra part in front of its name: the
*receiver*.

```go
type Disk struct {
	Device string
	SizeGB int
}

func (d Disk) Describe() string {
	return d.Device + " disk"
}
```

Here are the parts of that first line:

| Part | What it is | Required? |
|---|---|---|
| `func` | starts any function or method | Go requires it |
| `(d Disk)` | the receiver: "belongs to `Disk`; inside, call it `d`" | Go requires the type; the name `d` is your choice |
| `Describe` | the method name | your choice |
| `()` | the parameter list (empty here) | Go requires the brackets, even when empty |
| `string` | the result type | Go requires it when the method returns something |

Inside the body, use `d` exactly like a parameter: `d.Device`, `d.SizeGB`.

> **Why is it written like this?** Go requires the receiver in brackets
> **before** the name. A short receiver name (often the type's first letter,
> like `d` for `Disk`) is convention. The method names in this exercise are
> already written for you: you only fill in the bodies.

### Calling a method

Call a method on a value with a dot, like reading a field, plus `()`:

```go
root := Disk{Device: "/dev/sda", SizeGB: 100}
fmt.Println(root.Describe()) // /dev/sda disk
```

Go passes `root` in as the receiver `d` for you.

### Methods vs functions

| | Function | Method |
|---|---|---|
| Declaration | `func Describe(d Disk) string` | `func (d Disk) Describe() string` |
| Call | `Describe(root)` | `root.Describe()` |
| Belongs to | the package | the type `Disk` |

Both do the same job. Methods keep behaviour next to the data it uses. In
*Interfaces* you'll see methods are also how a type proves it can do
something.

### Methods calling methods

Inside a method you can call another method on the same receiver, so one
builds on another instead of repeating work:

```go
func (d Disk) Size() string {
	return fmt.Sprintf("%dGB", d.SizeGB)
}

func (d Disk) Summary() string {
	return "[" + d.Size() + "] " + d.Device
}
```

Notice: `d.Size()` with brackets **calls** the method and gives you its string.
`d.SizeGB` without brackets reads a field.

### A value receiver works on a copy

`(d Disk)` is a *value receiver*: the method gets its own copy. Reading fields
works perfectly. Changing a field would only change the copy. Every method
here only reads, so that's fine. *Pointer receivers* come a few exercises
later.

### Splitting a string once with `strings.Cut`

`strings.Cut(s, sep)` splits `s` around the **first** `sep`. It returns three
things: the part before, the part after, and a `bool` saying whether `sep`
was found:

```go
before, after, found := strings.Cut("app=web", "=")
// before == "app", after == "web", found == true

before, after, found = strings.Cut("web", "=")
// before == "web", after == "", found == false
```

If you don't need one of the results, write `_` in its place. Go requires you
to use every variable you declare, and `_` means "throw this away".

## Your task

Your inventory tool describes hosts and deployments. Give the types some
methods so the rest of the tool can ask them questions.

The starter file has no imports yet. You'll need `fmt` and `strings`, so add
this under the `package` line:

```go
import (
	"fmt"
	"strings"
)
```

### 1. Build the address

**Goal:** make `Address()` return the host's IP and port joined by a colon.

The pieces: the `IP` field (a `string`), the fixed text `:`, and the `Port`
field (an `int`). No spaces.

```go
Host{Name: "web-1", IP: "10.0.0.5", Port: 8080}.Address()
// => "10.0.0.5:8080"
```

**Think it through:**
- Inside the method, what is the host called? Look at the receiver in brackets.
- `Port` is an `int`, so `+` can't join it to a string. Which `fmt` function builds a string from a pattern with `%s` and `%d`?

### 2. Build a URL

**Goal:** make `URL(path)` return an `http://` URL for `path` on the host.

The pieces, in order: the fixed text `http://`, the host's address, then the
parameter `path`. `path` always starts with `/`, so add nothing between the
address and `path`. Call `Address()` from Task 1 instead of formatting the IP
and port again.

```go
Host{Name: "web-1", IP: "10.0.0.5", Port: 8080}.URL("/healthz")
// => "http://10.0.0.5:8080/healthz"
```

**Think it through:**
- How do you call another method on the same receiver? (Brackets matter.)
- Which piece is fixed text in quotes, and which are values you already have?

### 3. Read the image tag

**Goal:** make `ImageTag()` return the part of the deployment's `Image` after the colon, or `"latest"` if there is no colon.

A container image is written `name:tag`, like `nginx:1.27`. With no tag,
container runtimes use `latest`. Images here never contain a registry port,
so there is at most one colon (a registry like `ghcr.io/acme/` before the
name is fine).

```go
Deployment{Name: "web", Image: "nginx:1.27"}.ImageTag()
// => "1.27"

Deployment{Name: "cache", Image: "redis"}.ImageTag()
// => "latest"
```

**Think it through:**
- This receiver is called `d`, not `h`. Which field holds the image?
- Of the three results `strings.Cut` gives you, which two do you need? What goes in place of the third?
- Which case returns early?

## Learn more

- [A Tour of Go: Methods](https://go.dev/tour/methods/1)
- [Go by Example: Methods](https://gobyexample.com/methods)
- [Go spec: Method declarations](https://go.dev/ref/spec#Method_declarations)
- [strings.Cut](https://pkg.go.dev/strings#Cut)
