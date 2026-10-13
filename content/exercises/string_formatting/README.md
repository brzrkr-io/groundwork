# Formatting with fmt

Gluing strings with `+` works until numbers turn up. Deploy bots, CLIs and
reports all print lines that mix text and numbers, often lined up in columns.
`fmt.Sprintf` does that in one call.

## Learn

### fmt.Sprintf

The `fmt` package formats values. `fmt.Sprintf` takes a *format string* with
placeholders called *verbs*, followed by one value per verb. It fills each
verb with the next value and returns the result as a string:

```go
import "fmt"

fmt.Sprintf("%s has %d replicas", "api", 3)
// => "api has 3 replicas"
```

The values fill the verbs in order: `"api"` goes into `%s`, `3` goes into
`%d`. (`fmt.Println`, which you may have seen elsewhere, *prints* instead of
returning a string. Here you only need `Sprintf`.)

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
padded with spaces. Numbers are right-aligned by default, and a `-` aligns to
the left instead:

```go
fmt.Sprintf("[%5d]", 42)    // => "[   42]"
fmt.Sprintf("[%-6s]", "api") // => "[api   ]"
fmt.Sprintf("[%6s]", "api")  // => "[   api]"
```

A value longer than the width is never cut. It just takes up more room. This
is how tools like `kubectl get pods` line up their columns.

### When the verbs don't match

If a verb doesn't fit its value, for example `%d` with a string, or there are
more verbs than values, `go vet` reports it. Gopherlings runs `go vet` along
with the tests, so you'll see a message such as:

```text
fmt.Sprintf format %d has arg name of wrong type string
```

## Your task

Format the messages for your team's deploy bot. Start by adding
`import "fmt"` under the `package` line.

### 1. Write the status line

Change `StatusLine` to return the service, a colon and a space, then the
replica count and `replicas ready`:

```go
StatusLine("api", 3)
// => "api: 3 replicas ready"

StatusLine("billing", 12)
// => "billing: 12 replicas ready"
```

### 2. Report rollout progress

Change `RolloutProgress` to return the service name in double quotes and the
percentage with exactly one digit after the point, followed by `%`:

```go
RolloutProgress("api", 42.5)
// => `rollout "api" 42.5% complete`

RolloutProgress("billing", 100)
// => `rollout "billing" 100.0% complete`

RolloutProgress("auth", 33.333)
// => `rollout "auth" 33.3% complete`
```

These results contain double quotes, so they are written between backticks.
Text in backticks is a *raw string*: Go takes every character literally,
so `` `say "hi"` `` is the same as `"say \"hi\""`. The tests use raw strings
for these cases too.

### 3. Line up a table row

Change `TableRow` to return:

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

## Learn more

- [Package fmt](https://pkg.go.dev/fmt)
- [Go by Example: String Formatting](https://gobyexample.com/string-formatting)
- [kubectl get](https://kubernetes.io/docs/reference/kubectl/generated/kubectl_get/)
