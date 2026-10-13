# Variables

A *variable* is a named box that holds a value. Ops code is full of them: the
service name, the environment, the image tag. Here you glue those values
together into labels and image references a deploy pipeline uses.

**New here:** parameters, joining strings with `+`, `var`, `:=`, `=`.

## Learn

### Parameters: a function's inputs

In *Hello, Go* your functions took no inputs. Most functions do. The inputs,
called *parameters*, go inside the parentheses, each with a name and a type:

```go
// Shout returns word followed by an exclamation mark.
func Shout(word string) string {
	return word + "!"
}
```

`word` is a parameter of type `string`. Whoever calls the function picks its
value:

```go
Shout("deploy")   // => "deploy!"
Shout("rollback") // => "rollback!"
```

Inside the body, you use the **parameter name** (`word`) to get that value.
The function's own name (`Shout`) is not the value; it's how other code calls
the function.

When several parameters in a row share a type, you can write the type once:
`func Join(a, b string)` means the same as `func Join(a string, b string)`.

> **Why is it written like this?**
> - **Go requires** the type after each name (`word string`). Go reads it
>   left to right: "word, which is a string".
> - **Go requires** the parameter names in the `func` line only. When you
>   *use* a parameter in the body, write just its name: `word`, not
>   `word string`.
> - **Your choice:** parameter names. Short lowercase names are convention.

### Joining strings with `+`

`+` glues two strings together. Go never adds spaces or separators for you:

```go
"api" + "-" + "v2" // => "api-v2"
"api" + "v2"       // => "apiv2"
```

Fixed text goes **inside** quotes. Names of variables and parameters go
**outside** them:

```go
word + "!"    // the value of word, then a !
"word" + "!"  // the text "word!", whatever word holds
```

> **Why is it written like this?** **Go requires** a `+` between every two
> pieces, and each quoted piece must open and close its own quotes. The `+`
> goes outside the quotes: inside them it's just a plus sign in the text.

### Declaring a variable

Go has two ways to create a variable:

| Form | Example | When to use it |
|---|---|---|
| `var name type` | `var env string` | You want to state the type, or set the value later. |
| `name := value` | `env := "production"` | You have a value right now. Go works out the type. |

```go
var region string    // region is "" (the empty string)
region = "eu-west-1" // now it is "eu-west-1"

zone := region + "a" // zone is "eu-west-1a", type string
```

> **Why is it written like this?**
> - **Go requires** `var` to put the name before the type, like parameters.
> - **Go requires** `:=` to be inside a function. Outside functions, every
>   line must start with a keyword like `var` or `func`, so Go knows what
>   it's reading.
> - **Convention:** inside functions, most Go code uses `:=`.

### Changing a variable

Once a variable exists, give it a new value with plain `=`. The new value can
use the old one:

```go
path := "/v1"
path = path + "/deploys" // path is now "/v1/deploys"
```

Read `path = path + "/deploys"` right to left: take the current `path`, add
`"/deploys"`, store the result back in `path`.

Two rules the compiler enforces:

- `:=` **creates**, `=` **changes**. Writing `path := ...` twice for the same
  name is an error.
- Every variable you create must be used, or you get `declared and not used`.

### Reading the tests

The tests check several inputs per function. Each check calls your function
and compares the result with the value after `!=`.

## Your task

Your team's deploy pipeline needs a few strings built from the service name,
environment and image details.

### 1. Build the service label

**Goal:** make `ServiceLabel` return the environment, a slash, then the
service name.

```go
ServiceLabel("api", "production")
// => "production/api"

ServiceLabel("worker", "staging")
// => "staging/worker"
```

The pieces: the `env` parameter, the fixed text `/`, the `name` parameter.

**Think it through:**
- Which parameter comes first in the result? (Not the first one in the `func` line.)
- Which piece is fixed text and needs quotes?

### 2. Build the image reference

**Goal:** in `ImageRef`, create a variable `ref` with `:=` holding the image
reference, then return `ref`.

```go
ImageRef("ghcr.io/acme", "api", "v1.4.2")
// => "ghcr.io/acme/api:v1.4.2"

ImageRef("registry.internal", "worker", "latest")
// => "registry.internal/worker:latest"
```

An *image reference* says where an image lives and which version to run. The
pieces: `registry`, the fixed text `/`, `name`, the fixed text `:`, `tag`.

**Think it through:**
- `ref` doesn't exist yet. Which of `:=` and `=` creates a variable?
- How many `+` do five pieces need?

### 3. Finish the deploy API path

**Goal:** in `DeployPath`, add `/services/` and the service name to `path`,
by changing `path` with `=`.

```go
DeployPath("production", "api")
// => "/v1/environments/production/services/api"

DeployPath("staging", "billing")
// => "/v1/environments/staging/services/billing"
```

The line above the `TODO(3)` already adds `/environments/` and `env`. Yours
adds the fixed text `/services/` and the `service` parameter.

**Think it through:**
- Look at the line that adds `/environments/`. What changes for services?
- `path` already exists. Create or change?

## Learn more

- [A Tour of Go: Functions](https://go.dev/tour/basics/4)
- [A Tour of Go: Variables](https://go.dev/tour/basics/8)
- [A Tour of Go: Short variable declarations](https://go.dev/tour/basics/10)
- [Go by Example: Variables](https://gobyexample.com/variables)
