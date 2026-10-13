## Task 1: Redact a secret

First add `import "strings"`. You need two things: a row of stars as long as the secret, and a way to replace every copy of the secret with it.

### Closer

`strings.Repeat("*", len(secret))` makes the stars. Give that to `strings.ReplaceAll(line, secret, ...)` as the replacement. `"*"` is fixed text in quotes; `secret` is a parameter, so no quotes.

### Answer

```go
return strings.ReplaceAll(line, secret, strings.Repeat("*", len(secret)))
```

## Task 2: Get the host from a URL

Do it one step at a time: remove `https://` from the start, remove `/` from the end, then lowercase. The README section *Trimming one prefix or suffix* names the functions.

### Closer

Use a variable and update it each step:

```go
host := strings.TrimPrefix(url, "https://")
host = strings.TrimSuffix(host, ...)
```

Then return it lowercased.

### Answer

```go
host := strings.TrimPrefix(url, "https://")
host = strings.TrimSuffix(host, "/")
return strings.ToLower(host)
```

If the prefix or suffix isn't there, `TrimPrefix` and `TrimSuffix` change nothing.

## Task 3: Name an environment variable

Build the name in steps: join service and setting with an underscore, turn dashes and spaces into underscores, then make it all capitals.

### Closer

Join with `service + "_" + setting` (the underscore is fixed text, so quotes; the parameters get none). Then call `strings.ReplaceAll` once for `"-"` and once for `" "`, and finish with `strings.ToUpper`.

### Answer

```go
name := service + "_" + setting
name = strings.ReplaceAll(name, "-", "_")
name = strings.ReplaceAll(name, " ", "_")
return strings.ToUpper(name)
```
