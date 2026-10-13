# Strings

Host names typed by people come with stray spaces and random capitals. Image
names need checking before you trust them. Go's `strings` package has a tool
for each job, and here you learn to use it.

**New here:** `import`, the `strings` package, `len`, `bool`, `\n`.

## Learn

### Importing a package

Go's *standard library* is a set of packages that come with Go. To use one,
*import* it at the top of the file, under the `package` line:

```go
package text

import "strings"
```

Then call its functions as `packagename.Function(...)`:

```go
strings.ToUpper("api") // => "API"
```

> **Why is it written like this?**
> - **Go requires** `import` after `package` and before any functions.
> - **Go requires** the package path in quotes: `"strings"`. It's a path,
>   like a file name, not a name in your code.
> - **Go requires** you to use every import. An unused one stops the build
>   with `"strings" imported and not used`, so add it when you first need it.
> - **Go requires** the capital letter in `strings.ToUpper`: only names
>   starting with a capital are *exported*, usable from outside the package.

### Useful functions in `strings`

| Call | Result |
|---|---|
| `strings.ToLower("Web-01")` | `"web-01"` |
| `strings.ToUpper("eu-west-1")` | `"EU-WEST-1"` |
| `strings.TrimSpace("  api \n")` | `"api"` (spaces, tabs and newlines removed from both ends) |
| `strings.HasPrefix("ghcr.io/api", "ghcr.io/")` | `true` (does it start with...?) |
| `strings.HasSuffix("web-01.prod", ".prod")` | `true` (does it end with...?) |
| `strings.Contains("disk full", "full")` | `true` (is it anywhere inside?) |
| `strings.ReplaceAll("a-b-c", "-", "_")` | `"a_b_c"` |
| `strings.Repeat("-", 5)` | `"-----"` |

The values you pass in the `( )` are called *arguments*. They go in the order
the function expects: for `HasPrefix`, the text to check first, then the
prefix. Arguments can be parameters, constants or quoted text.

### Nesting calls

You can pass the result of one call straight into another. The inner call runs
first:

```go
strings.ReplaceAll(strings.ToUpper("eu-west"), "-", "_") // => "EU_WEST"
```

Read it inside out: `ToUpper` makes `"EU-WEST"`, then `ReplaceAll` swaps the
dash.

### The length of a string

`len` is built into Go, so it needs no import. It returns the number of bytes
in a string, as an `int`:

```go
len("api")    // => 3
len("web-01") // => 6
```

For plain English letters, digits and punctuation, one character is one byte.
(Accented letters and emoji take more; see *Bytes, runes and UTF-8*.)

### true and false

`HasPrefix`, `HasSuffix` and `Contains` answer a yes/no question. Their result
has type `bool`, which has two values: `true` and `false`. A function can
return a `bool` like any other value:

```go
// IsTicket reports whether msg mentions a ticket.
func IsTicket(msg string) bool {
	return strings.Contains(msg, "OPS-")
}
```

> **Why is it written like this?** **Go requires** `bool` after the `)` when
> the function returns `true`/`false`. **Convention:** yes/no functions are
> named `IsX` or `HasX`, and their comment says "reports whether".

### Special characters in strings

Inside double quotes, a backslash starts an *escape sequence*: `"\n"` is a
newline (line break) and `"\t"` is a tab. `"a\nb"` prints as two lines.

## Your task

Write three helpers for your team's inventory tool. Start by adding
`import "strings"` where `TODO(1)` says, on line 4.

### 1. Normalise a host name

**Goal:** make `NormalizeHost` remove the whitespace at both ends of `host`
and turn it into lowercase.

```go
NormalizeHost("  API-01.Prod.Example.COM\n")
// => "api-01.prod.example.com"

NormalizeHost("\tDB.internal ")
// => "db.internal"

NormalizeHost("web-02.example.com")
// => "web-02.example.com"
```

**Think it through:**
- Which two functions in the table do this?
- One call's result goes into the other. Which is inside?

### 2. Check the image registry

**Goal:** make `IsInternalImage` return `true` when `image` starts with
`registry.internal/` (slash included), and `false` otherwise.

```go
IsInternalImage("registry.internal/api:v1.4.2")      // => true
IsInternalImage("docker.io/library/nginx:1.27")      // => false
IsInternalImage("registry.internal.evil.com/api:v1") // => false
```

The last example is why the slash matters: without it, a look-alike registry
would pass.

**Think it through:**
- "Starts with": which function asks that?
- Which argument is the parameter and which is fixed text in quotes?

### 3. Underline a heading

**Goal:** make `Heading` return the title in capitals, a newline, then one
`=` per character of the title.

```go
Heading("deploy api")
// => "DEPLOY API\n=========="

Heading("rollback")
// => "ROLLBACK\n========"
```

`"deploy api"` has 10 characters (the space counts), so it gets 10 `=` signs.
The pieces: the title in capitals, the fixed text `\n`, the row of `=`.

**Think it through:**
- Which function repeats a character, and what tells it how many times?
- How do you join the three pieces?

## Learn more

- [Package strings](https://pkg.go.dev/strings)
- [A Tour of Go: Packages and imports](https://go.dev/tour/basics/2)
- [Go by Example: String Functions](https://gobyexample.com/string-functions)
