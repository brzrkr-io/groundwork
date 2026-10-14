# Reading Go code: the parts of a file

Before you write more code, it helps to know what each part of a Go file is
for. You'll also learn which parts Go insists on, which parts are just the way
everyone writes them, and which parts are up to you.

**New here:** how a Go file is laid out; writing a whole function yourself;
calling a function vs naming it; constants at the top level.

## Learn

### The parts of a file, top to bottom

```go
// Package hello says hello.          ← comment describing the package
package hello                          ← 1. which package this file belongs to

import "strings"                       ← 2. other packages this file uses

const Owner = "platform team"          ← 3. top-level declarations:
                                       ←    constants, variables, functions
// Greet returns a greeting for name.
func Greet(name string) string {       ← 4. a function: its "signature" line...
	return "hello " + strings.ToUpper(name)  ← ...and its "body" in { }
}
```

Every Go file follows this order: `package`, then `import`, then
declarations.

| Part | Go requires it? | |
|---|---|---|
| `package hello` as the first code line | **Required** | The name is your choice. By convention it's short and lowercase. |
| `import` right after `package` | **Required** (if you use other packages) | gofmt sorts the import lines for you. |
| Only declarations (`const`, `var`, `func`) at the top level | **Required** | Statements like `return` or `x := 1` only work *inside* a function. |
| A `//` comment above each function | Convention | Go doesn't need it, but every Go project does it. It starts with the function's name, and editors warn when it doesn't. |
| Tabs for indentation | Convention | gofmt (your editor's format-on-save) does it for you. |

> **Why is it written like this?** Go keeps one fixed layout so every Go file
> reads the same. Opening any project, you know where to look: imports at the
> top, then declarations.

### The signature line of a function

The first line of a function says everything about how to use it:

```text
func   Greet   ( name   string )   string   {
 │       │        │       │         │       └ the body starts here, on the same line
 │       │        │       │         └ the type of the result it gives back
 │       │        │       └ the parameter's type
 │       │        └ the parameter's name
 │       └ the function's name: your choice
 └ the keyword that starts every function
```

- **Parameters** go inside `( )`. Each one is a *name, then a type*: `name string`.
  More than one? Separate them with commas: `(env string, name string)`. When
  they share a type, you can write it once: `(env, name string)`.
- **The result type** comes after the `)`. `string` means "this function
  gives back text". `int` means "gives back a whole number".
- **The body** is between `{` and `}`. `return` hands the result back.

> **Why is it written like this?** Go puts the *name first and the type
> second* so a signature reads left to right: "Greet takes a name, which is a
> string, and returns a string". The `{` must sit on the same line as `func`.
> If you put it on the next line, Go ends the statement too early and reports
> a syntax error.

| In a signature | Go requires it? |
|---|---|
| `func`, the name, `( )`, the result type, `{` in that order | **Required** |
| A name and a type for every parameter | **Required** |
| The names of the function and its parameters | **Your choice.** Pick names that say what they hold. |
| A capital first letter (`Greet`) | Convention for functions other packages may use. You'll learn more in *Packages*. |

### Inside the body: values, names and quotes

Inside a function you work with *values*. There are three ways to get one:

```go
"hello "            // fixed text: whatever is in the quotes, letter for letter
name                // a name: the value stored under it (here, a parameter)
Greet("sam")        // a call: run Greet with "sam" and use what it returns
```

Two mix-ups catch almost everyone at first:

- **Quotes or no quotes.** `"name"` is the four letters n, a, m, e. `name` is
  whatever value the parameter holds. If you want the value, don't use quotes.
- **A function's name or a call.** `Greet` on its own is the function itself,
  like the name of a recipe. `Greet("sam")` *uses* the recipe and gives you
  its result. When you want the result, add the brackets and the inputs.

> **Why is it written like this?** Go never guesses. A name in quotes is text.
> A bare name must have been declared somewhere: a parameter, a constant or a
> function. If Go can't find it, you get `undefined: x`.

### Which names a function can use

Inside a function you can use:
1. its own **parameters** (the names in its `( )`),
2. variables you create inside it with `:=`,
3. **top-level** names in the file: constants like `Owner` and other functions
   like `Greet`.

That's all. A parameter of one function doesn't exist inside another.

## Your task

A few pieces of a deploy bot are missing. This exercise starts with a
**compile error**, because the tests use a function that doesn't exist yet.
Until you write it, every test shows as failing. That's expected.

### 1. Write the Shout function

**Goal:** add a function named `Shout` that takes one parameter, `word`, of
type `string`, and returns a `string`: the word followed by `!`.

There's a `TODO(1)` where it goes, at the top level of the file (not inside
another function).

```go
Shout("deploy")   // => "deploy!"
Shout("rollback") // => "rollback!"
```

**Think it through:**
- Write the signature line first, piece by piece, using the diagram above:
  keyword, name, `(` parameter name and type `)`, result type, `{`.
- In the body, which part is the parameter and which is fixed text?

### 2. Announce a deploy

**Goal:** make `Announce` return the result of calling `Shout` on the text
`deploy ` followed by the service name.

Right now it returns the letters `deploy service`, whatever the input is.

```go
Announce("api")    // => "deploy api!"
Announce("worker") // => "deploy worker!"
```

**Think it through:**
- `service` is inside the quotes right now. Is that the parameter's value, or
  just the word "service"?
- `Announce` shouldn't add the `!` itself. Which function already does that?
  Do you want that function itself, or its result?

### 3. Add a version to the banner

**Goal:** declare a constant `Version` with the value `"v1.0.0"` at the top
level, then make `Banner` return `Tool`, a space, and `Version`.

```go
Banner() // => "groundwork v1.0.0"
```

**Think it through:**
- `Tool` is declared at the top of the file. Copy its shape for `Version`.
- The space between the two words has to come from somewhere. Where does
  fixed text go?

## Learn more

- [A Tour of Go: Functions](https://go.dev/tour/basics/4)
- [Go spec: Source file organization](https://go.dev/ref/spec#Source_file_organization)
- [Effective Go: Formatting](https://go.dev/doc/effective_go#formatting)
