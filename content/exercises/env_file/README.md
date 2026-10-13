# Practice: .env file parser

Almost every service you deploy reads settings from environment variables, and
many teams keep them in a `.env` file next to the code. Docker Compose,
systemd's `EnvironmentFile=` and countless CI tools read this format. In this
practice you write a small, strict parser for it, using what you learned in
this module and in *Errors*.

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
| `KEY=VALUE` | sets `KEY`; spaces around the key and the value are trimmed |
| `export KEY=VALUE` | same as above; the `export ` prefix is allowed so the file also works with the shell's `source` command |
| `KEY="a value"` or `KEY='a value'` | the surrounding quotes are removed, the inside is kept exactly |

Only the **first** `=` splits the key from the value, so
`DATABASE_URL=postgres://...?sslmode=require` keeps the `=` in its value.
Everything after the first `=` is the value: there are no inline comments, so
`PORT=8080 # web` has the value `8080 # web`.

If the same key appears twice, the **last one wins**, just like assigning a
variable twice in a shell script.

### Recap of the tools you need

- **Reading lines** (*Reading input line by line*): `bufio.NewScanner(r)`, the
  `for sc.Scan()` loop, a line counter, and `sc.Err()` after the loop.
- **Splitting** (*The strings package*): `key, value, found := strings.Cut(line, "=")`.
  `strings.CutPrefix(s, "export ")` returns `s` without the prefix and
  whether it was there.
- **Patterns** (*Regular expressions*): the starter already contains
  `validKey`, a compiled pattern for a valid key: a letter or `_`, then
  letters, digits or `_`. Call `validKey.MatchString(key)`.
- **Errors** (*Wrapping errors*, *Reporting many errors*): wrap with
  `fmt.Errorf("...: %w", err)` so `errors.Is` still finds the cause, and
  combine many errors with `errors.Join(errs...)`. Joined errors print one per
  line, and `errors.Is` checks every one of them.

### Error values

The starter declares three *sentinel errors*. Your code returns them (wrapped
where the task says so), and the tests check them with `errors.Is`:

```go
var (
	ErrUnterminatedQuote = errors.New("unterminated quote")
	ErrMissingEquals     = errors.New(`missing "="`)
	ErrInvalidKey        = errors.New("invalid key")
)
```

## Your task

The checkout team wants a parser that refuses broken `.env` files instead of
silently starting the service with half its settings. Work through
`exercise.go` from top to bottom.

### 1. Remove quotes from a value

Implement `Unquote(value string) (string, error)`. `value` has already been
trimmed.

- If `value` is at least 2 characters long and starts and ends with the same
  quote character (`"` or `'`), return what is between the quotes, unchanged.
- If `value` starts with a quote but doesn't end with the same one, return
  `ErrUnterminatedQuote` (unwrapped).
- Otherwise return `value` unchanged.

```go
Unquote(`"Hello, ops team"`) // => "Hello, ops team", nil
Unquote(`'json'`)            // => "json", nil
Unquote(`""`)                // => "", nil
Unquote(`8080`)              // => "8080", nil
Unquote(`"oops`)             // => "", ErrUnterminatedQuote
Unquote(`'mixed"`)           // => "", ErrUnterminatedQuote
```

### 2. Parse one line

Finish `ParseLine(line string) (key, value string, err error)`. The line is
never blank and never a comment; `Parse` filters those out first. The starter
already checks `key` with `validKey`; the steps around that check are yours.

1. Trim spaces from `line`, then remove an `export ` prefix if there is one.
2. Split at the first `=`. No `=` at all: return `ErrMissingEquals` (unwrapped).
3. Trim spaces from the key. If `validKey` doesn't match it, return
   `fmt.Errorf("%w %q", ErrInvalidKey, key)`, which prints like
   `invalid key "MY-VAR"`.
4. Trim spaces from the value and pass it through `Unquote`. Return its error
   unchanged.

```go
ParseLine("PORT=8080")                // => "PORT", "8080", nil
ParseLine("  export APP_ENV = prod ") // => "APP_ENV", "prod", nil
ParseLine(`GREETING="Hi = hello"`)    // => "GREETING", "Hi = hello", nil
ParseLine("EMPTY=")                   // => "EMPTY", "", nil
ParseLine("PORT 8080")                // error: missing "="
ParseLine("MY-VAR=1")                 // error: invalid key "MY-VAR"
ParseLine("=value")                   // error: invalid key ""
ParseLine(`NAME="checkout`)           // error: unterminated quote
```

On error, return empty strings for `key` and `value`.

### 3. Parse a whole file

Implement `Parse(r io.Reader) (map[string]string, error)`.

- Read `r` line by line. Count every line, starting at 1, including blank lines
  and comments.
- Skip lines that are blank or start with `#` after trimming spaces.
- Pass every other line to `ParseLine`. When it fails, **don't stop**: wrap the
  error as `fmt.Errorf("line %d: %w", n, err)`, remember it, and carry on.
- If any line failed, return `nil` and `errors.Join` of all the line errors, in
  file order.
- If reading itself fails (`sc.Err()`), return `nil` and that error.
- Otherwise return the map. A repeated key keeps its last value.

```go
input := `# checkout
PORT=8080
PORT=9090
export APP_ENV=production
`
Parse(strings.NewReader(input))
// => map[APP_ENV:production PORT:9090], nil
```

A broken file reports every bad line at once:

```go
input := `PORT=8080
LOG_LEVEL debug

MY-VAR=1
NAME="checkout
`
_, err := Parse(strings.NewReader(input))
fmt.Println(err)
```

```text
line 2: missing "="
line 4: invalid key "MY-VAR"
line 5: unterminated quote
```

### 4. Build an environment list

Programs you start from Go (in *Running commands*, later in the course) take
their environment as a `[]string` of `KEY=VALUE` entries. Implement
`Environ(vars map[string]string) []string`, which returns one `KEY=VALUE`
string per entry, **sorted by key**. Values are written as they are, without
quotes. An empty map gives an empty (or nil) slice.

```go
Environ(map[string]string{"PORT": "8080", "APP_ENV": "production"})
// => []string{"APP_ENV=production", "PORT=8080"}
```

Map iteration order is random in Go, so collect the keys, sort them with
`slices.Sort`, and then build the result.

## Learn more

- [Docker docs: env file syntax](https://docs.docker.com/compose/how-tos/environment-variables/variable-interpolation/#env-file-syntax)
- [pkg.go.dev: strings.Cut](https://pkg.go.dev/strings#Cut)
- [pkg.go.dev: errors.Join](https://pkg.go.dev/errors#Join)
- [Go blog: Working with errors](https://go.dev/blog/go1.13-errors)
