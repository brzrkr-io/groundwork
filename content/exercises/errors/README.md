# Errors are values

Things go wrong all the time in ops work: a config has a typo, a port is not a
number, a hostname is missing. In Go, a function tells its caller "something
went wrong" by returning an *error* value.

**New here:** the `error` type, `nil`, `errors.New`, `fmt.Errorf`, `strings.Cut`, returning early.

## Learn

### The error type

`error` is a built-in type. Any value that has an `Error() string` method is
an error. `err.Error()` gives you its message as a string.

```go
var err error = errors.New("disk is full")
fmt.Println(err.Error()) // disk is full
```

### nil means "no error"

A function that can fail returns an `error` as its **last** result. If it
worked, the error is `nil`. If it failed, the error is not `nil`.

You met this with `strconv.Atoi`:

```go
n, err := strconv.Atoi("8080")
if err != nil {
	// the text was not a number
}
```

Read `if err != nil` as "if something went wrong".

> **Why is it written like this?** Go requires nothing special here: `error`
> last is a *convention*, but every Go programmer expects it. The name `err`
> is also convention. `if err != nil` is just a normal `if`.

### Making your own errors

| Tool | Package | Use it when |
|---|---|---|
| `errors.New("...")` | `errors` | the message is always the same |
| `fmt.Errorf("...", values)` | `fmt` | the message includes a value |

`fmt.Errorf` works like `fmt.Sprintf`, but gives you an error instead of a
string. The verbs you need:

| Verb | Prints | Example |
|---|---|---|
| `%q` | a string, with double quotes added | `fmt.Errorf("user %q not found", "bob")` → `user "bob" not found` |
| `%d` | an integer | `fmt.Errorf("%d disks failed", 3)` → `3 disks failed` |
| `%v` | any value, including another error's message | `fmt.Errorf("backup: %v", err)` → `backup: disk is full` |

The quotes in the message come from `%q`. You never type `\"` yourself.

### Adding context

When a function you called fails, you can put a short note in front of its
error. That note is called *context*: it tells the reader *where* it failed.

```go
size, err := ParseSize(text)
if err != nil {
	return fmt.Errorf("volume %q: %v", name, err)
}
// "volume \"logs\": size \"big\" is not a number"
```

The caller's message goes first, then `: `, then the original error.

### Return early, with zero values

When something fails, return straight away. The other results get their
*zero value*: `0` for numbers, `""` for strings.

```go
func ParseRetries(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("retries %q is not a number", s)
	}
	if n > 10 {
		return 0, fmt.Errorf("retries %d is more than 10", n)
	}
	return n, nil
}
```

Each problem gets its own `if` that returns. The success line `return n, nil`
sits at the bottom. The code stays flat.

> **Why is it written like this?** Go requires a `return` to list one value
> per result, in order. Checking each problem and returning early is
> convention. The order of the checks is the author's choice, but it changes
> which error the caller sees first.

### Splitting text with strings.Cut

`strings.Cut(s, sep)` splits `s` at the first `sep`. It returns three values:
the text before, the text after, and `found` (`false` if `sep` was not there).

```go
user, domain, found := strings.Cut("ops@example.com", "@")
// user = "ops", domain = "example.com", found = true
```

### Named results

`ParseEndpoint` has *named results*: `(host string, port int, err error)`.
These are ordinary variables, already declared. You can assign to them with
`=` instead of `:=`, and you still write full `return` statements.

> **Why is it written like this?** Naming results is the author's choice; it
> documents what each result means. Go then requires `=` for them (`:=` on the
> same names only works if a *new* name also appears on the left).

### Message style

| Good | Bad | Why |
|---|---|---|
| `service name is empty` | `Service name is empty` | start lowercase |
| `port "x" is not a number` | `port "x" is not a number.` | no full stop at the end |
| `port "x" is not a number` | `invalid input` | say what was wrong, show the value |

## Your task

A deploy tool reads service names and endpoints like `db:5432` from its
settings. Make it report clear errors when the input is bad.

Add imports to `exercise.go` as you need them (`errors`, `fmt`, `strconv`,
`strings`). Go refuses to compile a file that imports a package it doesn't use.

### 1. Reject an empty service name

**Goal:** `RequireName` returns an error when `name` is `""`, and `nil` otherwise.

Make the error with `errors.New`. The message is exactly:

```text
service name is empty
```

```go
RequireName("api") // => nil
RequireName("")    // => error "service name is empty"
```

**Think it through:**
- Which value of the parameter `name` is the problem?
- The message never changes. Which of the two error tools fits that?

### 2. Parse a port number

**Goal:** `ParsePort` turns text like `"8080"` into the number `8080`, or returns `0` and an error.

A valid port is from 1 to 65535, including both ends.

| Input | Problem | Error message |
|---|---|---|
| `"http"` | `strconv.Atoi` fails | `port "http" is not a number` |
| `""` | `strconv.Atoi` fails | `port "" is not a number` |
| `"70000"` | above 65535 | `port 70000 is out of range 1-65535` |
| `"0"` | below 1 | `port 0 is out of range 1-65535` |
| `"-1"` | below 1 | `port -1 is out of range 1-65535` |

```go
ParsePort("8080") // => 8080, nil
ParsePort("http") // => 0, error `port "http" is not a number`
```

**Think it through:**
- "Not a number" shows the *text* `s` in quotes. "Out of range" shows the *number*. Which verb fits each?
- The fixed text is `port `, ` is not a number` and ` is out of range 1-65535`.
- Do you need one range check or two? Can one `if` cover both ends?

### 3. Parse an endpoint

**Goal:** `ParseEndpoint` splits `"db:5432"` into host `"db"` and port `5432`, or returns `""`, `0` and an error.

Split `s` with `strings.Cut` at `":"`. Check these in this order:

| Input | Problem | Error message |
|---|---|---|
| `"db"` | no `:` | `endpoint "db" has no port` |
| `":5432"` | host is empty | `endpoint ":5432" has no host` |
| `"db:http"` | `ParsePort` fails | `endpoint "db:http": port "http" is not a number` |
| `"db:"` | `ParsePort` fails | `endpoint "db:": port "" is not a number` |
| `"db:0"` | `ParsePort` fails | `endpoint "db:0": port 0 is out of range 1-65535` |

```go
ParseEndpoint("db:5432") // => "db", 5432, nil
ParseEndpoint("db")      // => "", 0, error `endpoint "db" has no port`
```

**Think it through:**
- Which of the three values from `strings.Cut` tells you there was no `:`?
- Don't re-check the port yourself: call your `ParsePort` with the text after the `:`.
- When `ParsePort` fails, the message is the *whole* input `s` in quotes, then `: `, then its error. See "Adding context".

## Learn more

- [A Tour of Go: Errors](https://go.dev/tour/methods/19)
- [Go by Example: Errors](https://gobyexample.com/errors)
- [Go blog: Error handling and Go](https://go.dev/blog/error-handling-and-go)
- [Go Code Review Comments: Error strings](https://go.dev/wiki/CodeReviewComments#error-strings)
