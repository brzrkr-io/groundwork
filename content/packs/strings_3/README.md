# Strings: clean up config and logs

Three small jobs that come up constantly: hiding a secret in a log line,
pulling the host out of a URL, and turning a setting name into an environment
variable name.

**New here:** `strings.TrimPrefix`, `strings.TrimSuffix`.

## Learn

### Recap

- `import "strings"` goes under the `package` line.
- Calls can be nested: `strings.ToLower(strings.TrimSpace(s))`.
- `len(s)` counts the characters in plain-English text; `strings.Repeat`
  builds a run of one character.
- `strings` functions return a new string; they don't change the old one.

### Trimming one prefix or suffix

`strings.TrimPrefix` removes a prefix if the string starts with it, and
`strings.TrimSuffix` removes a suffix if the string ends with it. If it isn't
there, you get the string back unchanged:

```go
strings.TrimPrefix("s3://backups", "s3://") // => "backups"
strings.TrimPrefix("backups", "s3://")      // => "backups"
strings.TrimSuffix("app.log", ".log")       // => "app"
```

### Building a result in steps

When a result needs several calls, store each step in a variable:

```go
key := strings.TrimPrefix(path, "/")
key = strings.ReplaceAll(key, "/", ".")
return key
```

The first line creates `key` with `:=`; the next changes it with `=`.

> **Why is it written like this?** **Your choice:** nest everything in one
> line or use steps. Steps are easier to read and to check one at a time.
> **Go requires** `:=` only on the first line; after that, `key` exists.

## Your task

Write the three helpers. The file has no imports yet, so start by adding
`import "strings"` on line 4, where `TODO(1)` says.

### 1. Redact a secret

**Goal:** make `Redact` replace every copy of `secret` in `line` with `*`
characters, one per character of the secret.

```go
Redact("password=hunter2 user=bob", "hunter2")
// => "password=******* user=bob"

Redact("token=abc123 retry token=abc123", "abc123")
// => "token=****** retry token=******"
```

**Think it through:**
- Which function replaces every copy?
- What do you replace it with, and which two functions build that?

### 2. Get the host from a URL

**Goal:** make `HostFromURL` remove `https://` from the start and a `/` from
the end (if they're there), and return the result in lowercase.

```go
HostFromURL("https://API.Example.com/")
// => "api.example.com"

HostFromURL("https://grafana.internal")
// => "grafana.internal"

HostFromURL("vault.internal/")
// => "vault.internal"
```

**Think it through:**
- Three steps, three functions. Which ones?
- A variable makes each step easy to follow.

### 3. Name an environment variable

**Goal:** make `EnvVarName` return the service, an underscore, then the
setting, all in capitals, with every dash and space turned into an
underscore.

```go
EnvVarName("checkout-api", "db url")
// => "CHECKOUT_API_DB_URL"

EnvVarName("auth", "log-level")
// => "AUTH_LOG_LEVEL"

EnvVarName("payment-gateway", "max open conns")
// => "PAYMENT_GATEWAY_MAX_OPEN_CONNS"
```

Services read settings from environment variables named like this.

**Think it through:**
- Join the two parameters first. What fixed text goes between them?
- Dashes and spaces are two different replacements.

## Learn more

- [Package strings](https://pkg.go.dev/strings)
- [The Twelve-Factor App: Config](https://12factor.net/config)
- [Go by Example: String Functions](https://gobyexample.com/string-functions)
