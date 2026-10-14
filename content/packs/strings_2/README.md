# Strings: fix the log and host checks

The on-call tooling has three string helpers. One uses the wrong check, one
produces metric names Prometheus rejects, and one misses half the timeouts in
the logs.

**New here:** case-insensitive checks.

## Learn

### Recap

- `import "strings"` gives you the `strings` package. This file already has it.
- `HasPrefix`, `HasSuffix` and `Contains` return a `bool`.
- `ToLower`, `ToUpper`, `TrimSpace`, `ReplaceAll` and `Repeat` return a new
  string. Calls can be nested; the inner one runs first.

### Strings don't change in place

`strings` functions never change the string you pass in. They **return** a
new one, which you use or store:

```go
name := "Web-01"
strings.ToLower(name)        // the result is thrown away
lower := strings.ToLower(name) // lower is "web-01"; name is still "Web-01"
```

> **Why is it written like this?** **Go requires** it: strings in Go can't be
> changed once made. Every "change" makes a new string. So use the result
> directly in your expression, or store it in a variable.

### Comparisons are exact

Every `strings` check compares characters exactly, so capitals matter:

```go
strings.Contains("Disk FULL", "full") // => false
```

To ignore capitals, make the text lowercase first, then search for a
lowercase word:

```go
strings.Contains(strings.ToLower("Disk FULL"), "full") // => true
```

## Your task

Fix the three helpers. Each fix is in the `return` line under its `TODO`.

### 1. Fix the production host check

**Goal:** make `IsProdHost` return `true` when `host` **ends with**
`.prod.example.com`.

```go
IsProdHost("web-01.prod.example.com")    // => true
IsProdHost("db-02.prod.example.com")     // => true
IsProdHost("web-01.staging.example.com") // => false
```

Right now it returns `false` for every host.

**Think it through:**
- What does the starter's line check: start or end?

### 2. Make a valid metric name

**Goal:** make `MetricName` replace every `-` in `service` with `_`, then add
`_requests_total`.

```go
MetricName("checkout-api")       // => "checkout_api_requests_total"
MetricName("payment-gateway-v2") // => "payment_gateway_v2_requests_total"
MetricName("auth")               // => "auth_requests_total"
```

Prometheus metric names may contain letters, digits and underscores, but no
dashes. The fixed `_requests_total` part already has no dashes.

**Think it through:**
- Which function swaps one piece of text for another everywhere?
- In the starter's line, which part should be replaced by that call?

### 3. Catch every timeout

**Goal:** make `MentionsTimeout` find the word `timeout` in `line` whatever
the capitals.

```go
MentionsTimeout("upstream timeout after 30s")    // => true
MentionsTimeout("ERROR: Timeout waiting for db") // => true
MentionsTimeout("CONNECTION TIMEOUT")            // => true
MentionsTimeout("request ok in 20ms")            // => false
```

**Think it through:**
- The word you search for is lowercase. What should happen to `line` first?

## Learn more

- [Package strings](https://pkg.go.dev/strings)
- [Prometheus: Metric and label naming](https://prometheus.io/docs/practices/naming/)
