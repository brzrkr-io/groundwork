# Strings: clean up config and logs

Three small jobs that come up constantly: hiding a secret in a log line,
pulling the host out of a URL, and turning a setting name into an environment
variable name.

## Learn

### Recap

`import "strings"`, then nest calls such as
`strings.ToLower(strings.TrimSpace(s))`. `len(s)` counts the characters in
plain-English text, and `strings.Repeat` builds a run of the same character.

### Trimming one prefix or suffix

`strings.TrimPrefix` removes a prefix if the string starts with it, and
`strings.TrimSuffix` removes a suffix if the string ends with it. If the
string doesn't start (or end) with it, you get the string back unchanged:

```go
strings.TrimPrefix("https://api.example.com", "https://") // => "api.example.com"
strings.TrimPrefix("api.example.com", "https://")         // => "api.example.com"
strings.TrimSuffix("api.example.com/", "/")               // => "api.example.com"
```

## Your task

Write the three helpers. The file has no imports yet, so start by adding
`import "strings"`.

### 1. Redact a secret

Before a log line is shipped, every copy of the secret in it must be replaced
by `*` characters, one per character of the secret. Write `Redact`:

```go
Redact("password=hunter2 user=bob", "hunter2")
// => "password=******* user=bob"

Redact("token=abc123 retry token=abc123", "abc123")
// => "token=****** retry token=******"
```

### 2. Get the host from a URL

Write `HostFromURL`. It removes `https://` from the start and a `/` from the
end if they are there, and returns the result in lowercase:

```go
HostFromURL("https://API.Example.com/")
// => "api.example.com"

HostFromURL("https://grafana.internal")
// => "grafana.internal"

HostFromURL("vault.internal/")
// => "vault.internal"
```

### 3. Name an environment variable

Services read settings from environment variables named like
`CHECKOUT_API_DB_URL`: the service name, an underscore, then the setting,
all in capitals, with every dash and every space turned into an underscore.
Write `EnvVarName`:

```go
EnvVarName("checkout-api", "db url")
// => "CHECKOUT_API_DB_URL"

EnvVarName("auth", "log-level")
// => "AUTH_LOG_LEVEL"

EnvVarName("payment-gateway", "max open conns")
// => "PAYMENT_GATEWAY_MAX_OPEN_CONNS"
```

## Learn more

- [Package strings](https://pkg.go.dev/strings)
- [The Twelve-Factor App: Config](https://12factor.net/config)
- [Go by Example: String Functions](https://gobyexample.com/string-functions)
