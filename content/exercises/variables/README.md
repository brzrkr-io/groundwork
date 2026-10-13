# Variables

A *variable* is a named box that holds a value. Ops code is full of them: the
service name, the environment, the image tag. In this exercise you glue those
values together into the labels and image references a deploy pipeline uses.

## Learn

### Parameters: a function's inputs

In *Hello, Go* your functions took no inputs. Most functions do. You list the
inputs, called *parameters*, inside the parentheses, each with a name and a
type:

```go
// Shout returns word followed by an exclamation mark.
func Shout(word string) string {
	return word + "!"
}
```

`word` is a parameter of type `string`. Whoever calls the function chooses its
value:

```go
Shout("deploy") // => "deploy!"
Shout("rollback") // => "rollback!"
```

When several parameters in a row share a type, you can write the type once:
`func Join(a, b string)` means the same as `func Join(a string, b string)`.

### Joining strings with `+`

`+` glues two strings together. Go never adds spaces or separators for you, so
put in the ones you need:

```go
"api" + "-" + "v2" // => "api-v2"
"api" + "v2"       // => "apiv2"
```

### Declaring a variable

Go has two ways to declare a variable.

| Form | Example | When to use it |
|---|---|---|
| `var name type` | `var env string` | You want to state the type, or set the value later. |
| `name := value` | `env := "production"` | You have a value right now. Go works out the type from it. |

A `string` variable declared with `var` and no value starts out as the empty
string `""`.

```go
var region string     // region is ""
region = "eu-west-1"  // now it is "eu-west-1"

zone := region + "a"  // zone is "eu-west-1a", type string
```

`:=` is called the *short variable declaration*. It only works inside a
function.

### Changing a variable

Once a variable exists, you give it a new value with plain `=`. The new value
can use the old one:

```go
path := "/v1"
path = path + "/deploys" // path is now "/v1/deploys"
```

Read `path = path + "/deploys"` from right to left: take the current `path`,
add `"/deploys"`, and store the result back in `path`.

Two rules the compiler enforces:

- Use `:=` to **create** a variable and `=` to **change** it. Writing
  `path := ...` twice for the same name is an error.
- Every variable you declare must be used. An unused variable stops the
  program from compiling with `declared and not used`.

### Reading the tests

From now on the tests check several inputs for each function. Each check calls
your function with some inputs and compares the result with the value it
wants. Look for the call and the value after `!=` to see what is expected.

## Your task

Your team's deploy pipeline needs a few strings built from the service name,
environment and image details.

### 1. Build the service label

Change `ServiceLabel` so it returns the environment, a slash, then the service
name:

```go
ServiceLabel("api", "production")
// => "production/api"

ServiceLabel("worker", "staging")
// => "staging/worker"
```

### 2. Build the image reference

A container *image reference* says where an image lives and which version to
run: the registry, a slash, the image name, a colon, then the tag.

In `ImageRef`, create a variable called `ref` with `:=`, holding the image
reference built from the three parameters. Then return `ref`:

```go
ImageRef("ghcr.io/acme", "api", "v1.4.2")
// => "ghcr.io/acme/api:v1.4.2"

ImageRef("registry.internal", "worker", "latest")
// => "registry.internal/worker:latest"
```

### 3. Finish the deploy API path

`DeployPath` builds the path of a deploy API, one piece at a time. It already
adds `/environments/` and the environment. Add `/services/` and the service
name to `path` in the same way, by reassigning `path` with `=`:

```go
DeployPath("production", "api")
// => "/v1/environments/production/services/api"

DeployPath("staging", "billing")
// => "/v1/environments/staging/services/billing"
```

## Learn more

- [A Tour of Go: Functions](https://go.dev/tour/basics/4)
- [A Tour of Go: Variables](https://go.dev/tour/basics/8)
- [A Tour of Go: Short variable declarations](https://go.dev/tour/basics/10)
- [Go by Example: Variables](https://gobyexample.com/variables)
