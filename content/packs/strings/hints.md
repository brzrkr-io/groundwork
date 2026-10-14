## Task 1: Normalise a host name

First add `import "strings"` below the `package` line. Then you need two things: remove the spaces at both ends, and make it lowercase. The `strings` package has a function for each.

### Closer

`strings.TrimSpace(host)` removes the whitespace and `strings.ToLower(...)` lowercases. You can put one call inside the other: `strings.ToLower(strings.TrimSpace(...))`.

### Answer

```go
import "strings"

// in NormalizeHost:
return strings.ToLower(strings.TrimSpace(host))
```

The inner call runs first, and its result goes into the outer call.

## Task 2: Check the image registry

You need to know if `image` *starts with* some text. Look for a `strings` function with "prefix" in its name.

### Closer

`strings.HasPrefix(image, "...")` already gives `true` or `false`, so return it directly. The text to check is fixed, so it goes in quotes: `"registry.internal/"`.

### Answer

```go
return strings.HasPrefix(image, "registry.internal/")
```

## Task 3: Underline a heading

The result has three parts joined together: the title in capitals, a newline, and a row of `=` signs. How many `=` signs? As many as `len(title)`.

### Closer

Fixed text goes in quotes (`"\n"`, `"="`); the parameter `title` doesn't. Use `strings.ToUpper(title)`, then `strings.Repeat("=", len(title))`, and join the three pieces with `+`.

### Answer

```go
return strings.ToUpper(title) + "\n" + strings.Repeat("=", len(title))
```
