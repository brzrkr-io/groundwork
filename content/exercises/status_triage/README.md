# Practice: HTTP status triage

When a health check or a deploy smoke test gets an HTTP response, the first
question is "what kind of answer was that, and should we try again?" In this
practice exercise you build a tiny triage helper out of everything from this
module: `switch`, booleans, multiple return values and functions that call
each other.

**New here:** nothing; practice.

## Learn

### How the pieces fit

Everything is in one file, `exercise.go`. Three small helpers do one job
each, and `Triage` combines them:

| Function | Takes | Gives back | Task |
|---|---|---|---|
| `StatusClass` | a code (`int`) | its class name | 1 |
| `ShouldRetry` | a code (`int`) | `true` / `false` | 2 |
| `ParseStatus` | text, e.g. `"503"` | the code, and an error | 3 |
| `Triage` | a service name and text | one summary line | 4 |

```
"503" ──ParseStatus──▶ 503 ──StatusClass──▶ "server error"
                           └─ShouldRetry──▶ true
                    Triage joins these into one line
```

> **Why is it written like this?** Small functions that each do one thing is
> a Go convention: each is easy to test on its own. Splitting it into these
> four is the author's choice.

### Recap

- A *tagless switch* checks conditions top to bottom and runs the first true
  one, as in *Switch: size, method, latency*:

  ```go
  switch {
  case temp < 0:
  	return "freezing"
  case temp < 20:
  	return "cool"
  default:
  	return "warm"
  }
  ```

- A switch *on a value* can list several values in one case:
  `switch day { case "sat", "sun": ... }`.
- `strconv.Atoi(text)` returns `(int, error)`. Check `if err != nil` before
  using the number, as in *Multiple return values*.
- `strings.TrimSpace(" x ")` gives `"x"`.
- `fmt.Sprintf` builds a string: fixed text goes *inside* the quoted
  format; values go after it. `%s` is a string, `%d` an `int`, and `%q` a
  string wrapped in double quotes:

  ```go
  fmt.Sprintf("%s has %d pods, label %q", "web", 3, "blue")
  // => `web has 3 pods, label "blue"`
  ```

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

Your on-call tooling receives status codes as text from a log and needs a
one-line triage message for each one.

### 1. Classify a status code

**Goal:** make `StatusClass` return the class name from the first table above.

Return the class text exactly as in the table. Any code outside 100–599
(like `0`, `99` or `600`) returns `"unknown"`. Both ends of each range count:
`199` is informational, `200` is success.

```go
StatusClass(200) // => "success"
StatusClass(301) // => "redirect"
StatusClass(404) // => "client error"
StatusClass(503) // => "server error"
StatusClass(101) // => "informational"
StatusClass(700) // => "unknown"
```

**Think it through:**
- Each class is a range. How do you say "at least this *and* at most that"?
- Which part of a switch catches everything the cases didn't?

### 2. Decide whether to retry

**Goal:** make `ShouldRetry` return `true` for exactly 429, 502, 503 and 504, and `false` for every other code.

```go
ShouldRetry(503) // => true
ShouldRetry(429) // => true
ShouldRetry(500) // => false
ShouldRetry(505) // => false
ShouldRetry(200) // => false
```

**Think it through:**
- These are four exact values, not a range. Which kind of switch fits that?

### 3. Parse a status code from text

**Goal:** make `ParseStatus` turn text like `"503"` into the number 503, or return an error.

- Remove spaces around the text first.
- If the text is not a whole number (`"abc"`, `""`, `"5o3"`), return `0` and an error.
- If the number is outside 100–599 (`"42"`, `"600"`, `"-200"`), return `0` and an error.
- Otherwise return the number and a `nil` error.

The tests only check whether the error is `nil`; the message is up to you.

```go
ParseStatus("503")   // => 503, nil
ParseStatus(" 200 ") // => 200, nil
ParseStatus("abc")   // => 0, (an error)
ParseStatus("42")    // => 0, (an error)
```

**Think it through:**
- Three steps in order: trim, convert, check the range. Where can each one fail?
- You need the `strconv` and `strings` packages. Add them to the `import`
  line at the top (it becomes a list in `( )`).

### 4. Write the triage line

**Goal:** make `Triage` return a one-line summary for one service, using your three functions.

1. Parse `statusText` with `ParseStatus`. If that fails, return
   `<service>: invalid status "<text>"`, with the original `statusText` in
   double quotes.
2. Otherwise return `<service>: <code> <class>, retry` when the code should
   be retried, or `<service>: <code> <class>, no retry` when it should not.

The pieces: `service` is a parameter; the code and class come from your
functions; `: `, the space, `, ` and `retry` / `no retry` are fixed text.

```go
Triage("checkout", "503") // => "checkout: 503 server error, retry"
Triage("gateway", "429")  // => "gateway: 429 client error, retry"
Triage("search", "404")   // => "search: 404 client error, no retry"
Triage("api", "200")      // => "api: 200 success, no retry"
Triage("web", " 301")     // => "web: 301 redirect, no retry"
Triage("billing", "oops") // => `billing: invalid status "oops"`
Triage("auth", "700")     // => `auth: invalid status "700"`
```

**Think it through:**
- Use the parameter `service`, not a function name, for the first piece.
- Which format verb puts quotes around text for you?
- The last word is either `retry` or `no retry`. Which function decides?

## Learn more

- [MDN: HTTP response status codes](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status)
- [Go by Example: Switch](https://gobyexample.com/switch)
- [pkg.go.dev: strconv.Atoi](https://pkg.go.dev/strconv#Atoi)
