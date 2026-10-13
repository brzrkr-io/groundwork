# Strings: fix the log and host checks

The on-call tooling has three string helpers. One uses the wrong check, one
produces metric names Prometheus rejects, and one misses half the timeouts in
the logs.

## Learn

### Recap

`import "strings"` gives you the `strings` package. `HasPrefix`, `HasSuffix`
and `Contains` return a `bool` (`true` or `false`). `ToLower`, `ToUpper`,
`TrimSpace`, `ReplaceAll` and `Repeat` return a new string. Calls can be
nested, and the inner one runs first.

### Comparisons are exact

Every `strings` check compares characters exactly, so capitals matter:

```go
strings.Contains("Disk FULL", "full") // => false
```

To ignore capitals, lower-case the text first, then search for a lowercase
word:

```go
strings.Contains(strings.ToLower("Disk FULL"), "full") // => true
```

## Your task

Fix the three helpers.

### 1. Fix the production host check

`IsProdHost` should return `true` when a host name **ends with**
`.prod.example.com`. It returns `false` for every host. Fix it:

```go
IsProdHost("web-01.prod.example.com")    // => true
IsProdHost("db-02.prod.example.com")     // => true
IsProdHost("web-01.staging.example.com") // => false
```

### 2. Make a valid metric name

Prometheus metric names may contain letters, digits and underscores, but no
dashes. `MetricName` should replace every `-` in the service name with `_`,
then add `_requests_total`:

```go
MetricName("checkout-api")       // => "checkout_api_requests_total"
MetricName("payment-gateway-v2") // => "payment_gateway_v2_requests_total"
MetricName("auth")               // => "auth_requests_total"
```

### 3. Catch every timeout

`MentionsTimeout` scans log lines for the word `timeout`. It misses lines
that write it with capitals. Fix it so capitals don't matter:

```go
MentionsTimeout("upstream timeout after 30s")      // => true
MentionsTimeout("ERROR: Timeout waiting for db")   // => true
MentionsTimeout("CONNECTION TIMEOUT")              // => true
MentionsTimeout("request ok in 20ms")              // => false
```

## Learn more

- [Package strings](https://pkg.go.dev/strings)
- [Prometheus: Metric and label naming](https://prometheus.io/docs/practices/naming/)
