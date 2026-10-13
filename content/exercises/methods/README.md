# Methods

A *method* is a function that belongs to a type. Instead of passing a host to
a helper function, you ask the host itself: `h.Address()`. Ops tooling is full
of these: a host knows its address, a deployment knows its image tag.

## Learn

### Declaring a method

A method looks like a function with one extra part in front of its name: the
*receiver*.

```go
type Host struct {
	Name string
	IP   string
	Port int
}

func (h Host) Greeting() string {
	return "hello from " + h.Name
}
```

`(h Host)` is the receiver. It says "this method belongs to `Host`, and inside
the body the host is called `h`". You use `h` exactly like a normal parameter:
`h.Name`, `h.IP`, `h.Port`. By convention the receiver name is short, often
the first letter of the type.

### Calling a method

You call a method on a value with a dot, just like reading a field:

```go
web := Host{Name: "web-1", IP: "10.0.0.5", Port: 8080}
fmt.Println(web.Greeting()) // hello from web-1
```

Go passes `web` in as the receiver `h` for you.

### Methods vs functions

| | Function | Method |
|---|---|---|
| Declaration | `func Greeting(h Host) string` | `func (h Host) Greeting() string` |
| Call | `Greeting(web)` | `web.Greeting()` |
| Belongs to | the package | the type `Host` |

Both do the same job. Methods keep behaviour next to the data it works on, and
they read naturally: `web.Greeting()`. Later, in *Interfaces*, you'll see that
methods are also how a type proves it can do something.

### Methods calling methods

Inside a method you can call other methods on the same receiver. This lets
one method build on another instead of repeating its work:

```go
func (h Host) Label() string {
	return h.Name + " (" + h.IP + ")"
}

func (h Host) Banner() string {
	return "*** " + h.Label() + " ***"
}
```

### A value receiver works on a copy

`(h Host)` is a *value receiver*: the method gets its own copy of the host.
Reading fields works perfectly. If the method changed `h.Port`, only the copy
would change, and the caller's host would stay the same. Every method in this
exercise only reads, so that's fine. *Pointer receivers*, which let a method
change the original, come a few exercises later.

### Splitting a string once with `strings.Cut`

`strings.Cut(s, sep)` splits `s` around the first `sep`. It returns the part
before, the part after, and a `bool` that says whether `sep` was found:

```go
before, after, found := strings.Cut("app=web", "=")
// before == "app", after == "web", found == true

before, after, found = strings.Cut("web", "=")
// before == "web", after == "", found == false
```

If you don't need one of the results, use `_` in its place.

## Your task

Your team's inventory tool describes hosts and deployments. Give the types
some methods so the rest of the tool can ask them questions.

You'll need `fmt` and `strings`, so add them to the file:

```go
import (
	"fmt"
	"strings"
)
```

### 1. Build the address

Complete `Address` so it returns the host's IP and port joined by a colon:

```go
Host{Name: "web-1", IP: "10.0.0.5", Port: 8080}.Address()
// => "10.0.0.5:8080"
```

### 2. Build a URL

Complete `URL` so it returns an `http://` URL for `path` on the host. Call
`h.Address()` rather than formatting the IP and port again. `path` always
starts with `/`, so put it straight after the address:

```go
Host{Name: "web-1", IP: "10.0.0.5", Port: 8080}.URL("/healthz")
// => "http://10.0.0.5:8080/healthz"
```

### 3. Read the image tag

A container image is written `name:tag`, for example `nginx:1.27`. Complete
`ImageTag` so it returns the part after the colon. If the image has no colon,
there is no tag, and container runtimes use `latest`, so return `"latest"`.
Images in this exercise never contain a registry port, so there is at most one
colon.

```go
Deployment{Name: "web", Image: "nginx:1.27"}.ImageTag()
// => "1.27"

Deployment{Name: "cache", Image: "redis"}.ImageTag()
// => "latest"
```

## Learn more

- [A Tour of Go: Methods](https://go.dev/tour/methods/1)
- [Go by Example: Methods](https://gobyexample.com/methods)
- [Go spec: Method declarations](https://go.dev/ref/spec#Method_declarations)
- [strings.Cut](https://pkg.go.dev/strings#Cut)
