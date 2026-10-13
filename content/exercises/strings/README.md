# Strings

Host names typed by people come with stray spaces and random capitals. Image
names need checking before you trust them. Go's `strings` package has a tool
for each job, and here you learn to use it.

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

Go is strict about imports: importing a package and not using it is a compile
error (`"strings" imported and not used`). So add the import when you first
need it.

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

You can pass the result of one call straight into another. The inner call runs
first:

```go
strings.ToUpper(strings.TrimSpace("  api ")) // => "API"
```

### The length of a string

`len` is built into Go, so it needs no import. It returns the number of bytes
in a string, as an `int`:

```go
len("api")     // => 3
len("web-01")  // => 6
```

For plain English letters, digits and punctuation, one character is one byte.
(Accented letters and emoji take more. You'll meet that in *Bytes, runes and
UTF-8*.)

### true and false

`HasPrefix`, `HasSuffix` and `Contains` answer a yes/no question. Their result
has type `bool`, which has exactly two values: `true` and `false`. A function
can return a `bool` like any other value:

```go
// IsProdHost reports whether host is in the production domain.
func IsProdHost(host string) bool {
	return strings.HasSuffix(host, ".prod.example.com")
}
```

### Special characters in strings

Inside double quotes, a backslash starts an *escape sequence*: `"\n"` is a
newline (line break) and `"\t"` is a tab. `"a\nb"` prints as two lines.

## Your task

Write three helpers for your team's inventory tool. Start by adding
`import "strings"` under the `package` line.

### 1. Normalise a host name

People paste host names with extra spaces, tabs or newlines, and in any mix
of capitals. Change `NormalizeHost` so it removes the whitespace at both ends
and turns the name into lowercase:

```go
NormalizeHost("  API-01.Prod.Example.COM\n")
// => "api-01.prod.example.com"

NormalizeHost("\tDB.internal ")
// => "db.internal"

NormalizeHost("web-02.example.com")
// => "web-02.example.com"
```

### 2. Check the image registry

Production may only run images from your own registry. Change
`IsInternalImage` so it returns `true` when the image starts with
`registry.internal/` (including the slash), and `false` otherwise:

```go
IsInternalImage("registry.internal/api:v1.4.2")    // => true
IsInternalImage("docker.io/library/nginx:1.27")    // => false
IsInternalImage("registry.internal.evil.com/api:v1") // => false
```

The last example is why the slash matters: without it, a look-alike registry
would pass.

### 3. Underline a heading

Runbook sections start with a heading in capitals, a newline, then a row of
`=` signs as long as the title. Change `Heading` to build it:

```go
Heading("deploy api")
// => "DEPLOY API\n=========="

Heading("rollback")
// => "ROLLBACK\n========"
```

`"deploy api"` has 10 characters (the space counts), so it gets 10 `=` signs.

## Learn more

- [Package strings](https://pkg.go.dev/strings)
- [A Tour of Go: Packages and imports](https://go.dev/tour/basics/2)
- [Go by Example: String Functions](https://gobyexample.com/string-functions)
