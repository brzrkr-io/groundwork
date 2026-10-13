# Practice: .env file parser

Most services read settings from environment variables, and many teams keep
them in a `.env` file. Docker Compose, systemd's `EnvironmentFile=` and many
CI tools read this format. Here you write a small, strict parser for it.

**New here:** nothing; practice with strings, scanners, regexps and errors.

## Learn

### The format

```text
# Settings for the checkout service
export APP_ENV=production
PORT=8080

DATABASE_URL=postgres://app@db.internal:5432/shop?sslmode=require
GREETING="Hello, ops team"
LOG_FORMAT='json'
```

| Line | Meaning |
|---|---|
| empty, or only spaces | ignored |
| starts with `#` (after leading spaces) | a comment, ignored |
| `KEY=VALUE` | sets `KEY`; spaces around key and value are trimmed |
| `export KEY=VALUE` | same; `export ` is allowed so the shell can `source` the file |
| `KEY="a value"` or `KEY='a value'` | quotes removed, the inside kept exactly |

Three rules to remember:

- Only the **first** `=` splits key from value. `DSN=db?x=1` has value
  `db?x=1`.
- No inline comments: `PORT=8080 # web` has value `8080 # web`.
- A repeated key: the **last one wins**.

### Recap of the tools

| Need | Tool | From |
|---|---|---|
| read lines | `bufio.NewScanner(r)`, `for sc.Scan()`, `sc.Err()` | *Reading input line by line* |
| split at first `=` | `before, after, found := strings.Cut(s, "=")` | *The strings package* |
| drop a prefix | `rest, had := strings.CutPrefix(s, "export ")` | *The strings package* |
| check a key | `validKey.MatchString(key)` (already in the starter) | *Regular expressions* |
| wrap an error | `fmt.Errorf("...: %w", err)` | *Wrapping errors* |
| combine errors | `errors.Join(errs...)` | *Reporting many errors* |
| sort | `slices.Sort(s)` | *Slices* |

`validKey` matches a letter or `_`, then letters, digits or `_`.

### Reading single bytes

`s[0]` is the first byte of a string, `s[len(s)-1]` the last. Compare a byte
with a *rune literal* in single quotes:

```go
path := "/etc/"
path[0] == '/'           // true
path[len(path)-1] == '/' // true
path[1 : len(path)-1]    // "etc": from index 1 up to (not including) the last byte
```

> **Why is it written like this?** Go requires single quotes for one
> character (`'/'`) and double quotes for a string (`"/"`). A single quote
> itself is written `'\''`. Indexing an empty string crashes, so check
> `s == ""` first.

### The error values

The starter declares three *sentinel errors*. Tests check them with
`errors.Is`, so wrapping is fine:

```go
ErrUnterminatedQuote = errors.New("unterminated quote")
ErrMissingEquals     = errors.New(`missing "="`)
ErrInvalidKey        = errors.New("invalid key")
```

`%q` in a format string prints a value in double quotes:
`fmt.Sprintf("%q", "a b")` gives `"a b"`, and an empty string gives `""`.

## Your task

The checkout team wants a parser that refuses broken `.env` files instead of
starting the service with half its settings. Work through `exercise.go` top
to bottom.

### 1. Remove quotes from a value

**Goal:** `Unquote(value)` strips one pair of matching quotes, or returns
`ErrUnterminatedQuote`.

`value` is already trimmed.

- Starts **and** ends with the same quote (`"` or `'`), length at least 2:
  return what's between them, unchanged.
- Starts with a quote but doesn't end with the same one (including a lone
  `"`): return `""` and `ErrUnterminatedQuote` (not wrapped).
- Anything else (including `""` empty and `it's`): return `value` unchanged.

```go
Unquote(`"Hello, ops team"`) // => "Hello, ops team", nil
Unquote(`'json'`)            // => "json", nil
Unquote(`"  padded  "`)      // => "  padded  ", nil
Unquote(`""`)                // => "", nil
Unquote(`8080`)              // => "8080", nil
Unquote(``)                  // => "", nil
Unquote(`it's`)              // => "it's", nil
Unquote(`"oops`)             // => "", ErrUnterminatedQuote
Unquote(`"`)                 // => "", ErrUnterminatedQuote
Unquote(`'mixed"`)           // => "", ErrUnterminatedQuote
```

**Think it through:**
- Which byte decides whether there is a quote at all? Handle "no quote" first.
- Why does a lone `"` need the length check?

### 2. Parse one line

**Goal:** `ParseLine(line)` returns the key and value of one `KEY=VALUE`
line, or an error.

The line is never blank or a comment; `Parse` filters those. The starter
already has the `validKey` check; you fill in around it.

1. Trim spaces from `line`, then drop an `export ` prefix if present.
2. Split at the first `=`. No `=`: return `ErrMissingEquals` (not wrapped).
3. Trim spaces from the key. If `validKey` doesn't match, return an error
   that wraps `ErrInvalidKey` and prints as `invalid key`, a space, then the
   key in double quotes: `invalid key "MY-VAR"`.
4. Trim spaces from the value and pass it to `Unquote`. Return its error
   unchanged.

On any error, `key` and `value` are both `""`.

```go
ParseLine("PORT=8080")                // => "PORT", "8080", nil
ParseLine("  export APP_ENV = prod ") // => "APP_ENV", "prod", nil
ParseLine(`GREETING="Hi = hello"`)    // => "GREETING", "Hi = hello", nil
ParseLine("EMPTY=")                   // => "EMPTY", "", nil
ParseLine("_INTERNAL=1")              // => "_INTERNAL", "1", nil
ParseLine("PORT 8080")                // error: missing "="
ParseLine("MY-VAR=1")                 // error: invalid key "MY-VAR"
ParseLine("1PORT=1")                  // error: invalid key "1PORT"
ParseLine("=value")                   // error: invalid key ""
ParseLine(`NAME="checkout`)           // error: unterminated quote
```

**Think it through:**
- Which `strings` function splits at only the *first* `=`?
- The error message has a fixed part (from `ErrInvalidKey`) and a part from
  the key. Which format verb wraps, and which one adds the quotes?
- You need `"strings"` and `"fmt"` in the imports.

### 3. Parse a whole file

**Goal:** `Parse(r)` reads every line and returns a map of all variables, or
every bad line's error at once.

- Count lines from 1, **including** blank lines and comments.
- Skip lines that are blank or start with `#` after trimming spaces.
- Send other lines to `ParseLine`. On an error, **don't stop**: wrap it so it
  prints as `line N: <error>` (wrapped with `%w`), save it, and keep going.
- After the loop, check in this order:
  1. Reading failed (`sc.Err()`): return `nil` and that error.
  2. Any line failed: return `nil` and `errors.Join` of the line errors, in
     file order.
  3. Otherwise return the map (empty map if the file has no variables). A
     repeated key keeps its last value.

```go
input := "# checkout\nPORT=8080\nPORT=9090\nexport APP_ENV=production\n"
Parse(strings.NewReader(input))
// => map[APP_ENV:production PORT:9090], nil
```

A broken file reports every bad line:

```go
input := "PORT=8080\nLOG_LEVEL debug\n\nMY-VAR=1\nNAME=\"checkout\n"
_, err := Parse(strings.NewReader(input))
fmt.Println(err)
```

```text
line 2: missing "="
line 4: invalid key "MY-VAR"
line 5: unterminated quote
```

**Think it through:**
- Where does the counter go up so blank lines still count?
- What kind of variable collects many errors before you join them?
- You need `"bufio"` in the imports.

### 4. Build an environment list

**Goal:** `Environ(vars)` returns one `KEY=VALUE` string per entry, sorted
by key.

Programs started from Go (*Running commands*, later) take their environment
as a `[]string` like this. Write values as they are, no quotes added. An
empty map gives an empty or nil slice.

```go
Environ(map[string]string{"PORT": "8080", "APP_ENV": "production"})
// => []string{"APP_ENV=production", "PORT=8080"}

Environ(map[string]string{"EMPTY": "", "DSN": "db?x=1"})
// => []string{"DSN=db?x=1", "EMPTY="}
```

**Think it through:**
- Map order is random. What can you sort instead of the map?
- Each string has three pieces: key, fixed `=`, value.
- You need `"slices"` in the imports.

## Learn more

- [Docker docs: env file syntax](https://docs.docker.com/compose/how-tos/environment-variables/variable-interpolation/#env-file-syntax)
- [pkg.go.dev: strings.Cut](https://pkg.go.dev/strings#Cut)
- [pkg.go.dev: errors.Join](https://pkg.go.dev/errors#Join)
- [Go blog: Working with errors](https://go.dev/blog/go1.13-errors)
