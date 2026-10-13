# Formatting with fmt

Gluing strings with `+` works until numbers turn up. Deploy bots, CLIs and
reports all print lines that mix text and numbers, often lined up in columns.
`fmt.Sprintf` does that in one call.

**New here:** `fmt.Sprintf`, verbs (`%s %d %q %v %.1f %%`), width, raw strings.

## Learn

### fmt.Sprintf

The `fmt` package formats values. `fmt.Sprintf` takes a *format string* with
placeholders called *verbs*, then one value per verb. It fills each verb with
the next value and returns the result as a string:

```go
import "fmt"

fmt.Sprintf("%s has %d jobs queued", "backup", 7)
// => "backup has 7 jobs queued"
```

The values fill the verbs in order: `"backup"` goes into `%s`, `7` into `%d`.
Everything else in the format string is fixed text, copied as it is.

> **Why is it written like this?**
> - **Go requires** the format string to be a string, so it's in quotes. The
>   verbs live **inside** the quotes; the values come **after** it, outside,
>   separated by commas.
> - **Convention:** use `Sprintf` when a line mixes text and numbers. `+`
>   only joins strings; it can't join a string and an `int`.
> - `fmt.Println` *prints* instead of returning a string. Here you only need
>   `Sprintf` (the S is for "string").

### Verbs

| Verb | Use it for | Example | Result |
|---|---|---|---|
| `%s` | a string | `Sprintf("%s", "api")` | `api` |
| `%d` | an `int` | `Sprintf("%d", 42)` | `42` |
| `%q` | a string, in double quotes | `Sprintf("%q", "api")` | `"api"` |
| `%v` | any value, in its default format | `Sprintf("%v", 2.5)` | `2.5` |
| `%.1f` | a `float64`, 1 digit after the point | `Sprintf("%.1f", 42.456)` | `42.5` |
| `%%` | a literal `%` sign | `Sprintf("%d%%", 90)` | `90%` |

`%.2f` gives two digits, `%.0f` none, and so on. The number is rounded, not
cut off: `%.1f` turns `0.96` into `1.0`.

### Width and alignment

A number between `%` and the verb sets a minimum *width*. Shorter values are
padded with spaces. Values are right-aligned by default; a `-` aligns left:

```go
fmt.Sprintf("[%5d]", 42)     // => "[   42]"
fmt.Sprintf("[%-6s]", "api") // => "[api   ]"
fmt.Sprintf("[%6s]", "api")  // => "[   api]"
```

A value longer than the width is never cut. It just takes more room. This is
how tools like `kubectl get pods` line up their columns.

### Raw strings

Text between backticks is a *raw string*: Go takes every character literally,
so `` `say "hi"` `` is the same as `"say \"hi\""`. Some results below contain
double quotes, so they're written in backticks. The tests do the same.

### When the verbs don't match

If a verb doesn't fit its value (say `%d` with a string), or there are more
verbs than values, `go vet` reports it. Gopherlings runs `go vet` with the
tests, so you'll see a message such as:

```text
fmt.Sprintf format %d has arg name of wrong type string
```

## Your task

Format the messages for your team's deploy bot. Start by adding
`import "fmt"` on line 4, where `TODO(1)` says.

### 1. Write the status line

**Goal:** make `StatusLine` return the service, `: `, the replica count, then
` replicas ready`.

```go
StatusLine("api", 3)
// => "api: 3 replicas ready"

StatusLine("billing", 12)
// => "billing: 12 replicas ready"
```

**Think it through:**
- Which two pieces change? Those become verbs; the rest is fixed text.
- `service` is a string and `replicas` an `int`. Which verb fits each?

### 2. Report rollout progress

**Goal:** make `RolloutProgress` return `rollout `, the service in double
quotes, the percentage with exactly one decimal and a `%`, then ` complete`.

```go
RolloutProgress("api", 42.5)
// => `rollout "api" 42.5% complete`

RolloutProgress("billing", 100)
// => `rollout "billing" 100.0% complete`

RolloutProgress("auth", 33.333)
// => `rollout "auth" 33.3% complete`
```

**Think it through:**
- Which verb adds the double quotes for you?
- Which verb gives one decimal, and how do you print a `%` after it?

### 3. Line up a table row

**Goal:** make `TableRow` return these five pieces in order:

1. the service name, left-aligned in 10 characters,
2. ` | ` (space, bar, space),
3. the replica count, right-aligned in 5 characters,
4. ` | `,
5. the status.

```go
TableRow("api", 3, "Running")
// => "api        |     3 | Running"

TableRow("checkout", 12, "Pending")
// => "checkout   |    12 | Pending"

TableRow("notification-svc", 120, "Running")
// => "notification-svc |   120 | Running"
```

The last name is longer than 10 characters, so it isn't padded at all.

**Think it through:**
- Where does the width number go in a verb? Where does the `-` go?

## Learn more

- [Package fmt](https://pkg.go.dev/fmt)
- [Go by Example: String Formatting](https://gobyexample.com/string-formatting)
- [kubectl get](https://kubernetes.io/docs/reference/kubectl/generated/kubectl_get/)
