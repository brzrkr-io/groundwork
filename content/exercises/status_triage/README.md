# Practice: HTTP status triage

When a health check or a deploy smoke test gets an HTTP response, the first
question is "what kind of answer was that, and should we try again?" In this
practice exercise you build a tiny triage helper out of everything from this
module: `switch`, booleans, multiple return values and functions that call
each other.

## Learn

### Recap

- A *tagless switch* (`switch { case code >= 500: ... }`) checks conditions
  top to bottom and runs the first one that is true, as in *switch*.
- One switch case can list several values: `case 429, 502:`.
- `strconv.Atoi(text)` returns `(int, error)`. Check `if err != nil` before
  using the number, as in *Multiple return values*.
- `fmt.Errorf("...")` creates an error with a message.

### HTTP status codes

Every HTTP response carries a three-digit *status code*. The first digit tells
you the family:

| Codes | Class | Meaning |
|---|---|---|
| 100–199 | `informational` | "Keep going", rarely seen by tools |
| 200–299 | `success` | It worked |
| 300–399 | `redirect` | Look somewhere else |
| 400–499 | `client error` | The request was wrong |
| 500–599 | `server error` | The server failed |

Some failures are temporary, so retrying later can help:

| Code | Name | Why a retry can help |
|---|---|---|
| 429 | Too Many Requests | You were rate-limited; wait and try again |
| 502 | Bad Gateway | A proxy could not reach the backend, maybe mid-deploy |
| 503 | Service Unavailable | The service is overloaded or restarting |
| 504 | Gateway Timeout | The backend was too slow this time |

Other errors, like `404 Not Found` or `500 Internal Server Error`, usually fail
the same way again, so this tool does not retry them.

## Your task

Your on-call tooling receives status codes as text from a log and needs a one-line
triage message for each one.

### 1. Classify a status code

Change `StatusClass` to return the class name from the first table above. Any
code outside 100–599 returns `"unknown"`.

```go
StatusClass(200) // => "success"
StatusClass(301) // => "redirect"
StatusClass(404) // => "client error"
StatusClass(503) // => "server error"
StatusClass(101) // => "informational"
StatusClass(700) // => "unknown"
```

### 2. Decide whether to retry

Change `ShouldRetry` to return `true` for exactly 429, 502, 503 and 504, and
`false` for every other code.

```go
ShouldRetry(503) // => true
ShouldRetry(500) // => false
ShouldRetry(200) // => false
```

### 3. Parse a status code from text

Change `ParseStatus` to turn text into a status code:

- Remove spaces around the text first (`strings.TrimSpace`).
- If the text is not a whole number, return `0` and an error.
- If the number is outside 100–599, return `0` and an error.
- Otherwise return the number and a `nil` error.

The tests only check whether the error is `nil`; the message is up to you.

```go
ParseStatus("503")   // => 503, nil
ParseStatus(" 200 ") // => 200, nil
ParseStatus("abc")   // => 0, error
ParseStatus("42")    // => 0, error
```

You need the `strconv` and `strings` imports; the TODO reminds you.

### 4. Write the triage line

Change `Triage` to build the message for one service. Use your three
functions:

1. Parse `statusText` with `ParseStatus`. If that fails, return
   `<service>: invalid status "<text>"`, with the original text in double
   quotes (`%q` does this).
2. Otherwise return `<service>: <code> <class>, retry` when the code should be
   retried, or `<service>: <code> <class>, no retry` when it should not.

```go
Triage("checkout", "503")
// => "checkout: 503 server error, retry"

Triage("search", "404")
// => "search: 404 client error, no retry"

Triage("api", "200")
// => "api: 200 success, no retry"

Triage("billing", "oops")
// => `billing: invalid status "oops"`
```

## Learn more

- [MDN: HTTP response status codes](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status)
- [Go by Example: Switch](https://gobyexample.com/switch)
- [pkg.go.dev: strconv.Atoi](https://pkg.go.dev/strconv#Atoi)
