# Wrapping errors: fix the secrets fetcher

A deploy tool reads passwords and tokens from a secrets vault. Its error
handling looks fine, and the messages print correctly, but three checks
silently give the wrong answer. Bugs like these hide until the bad day when
you need them.

## Learn

Recap from the previous exercise: `fmt.Errorf("context: %w", err)` wraps
`err`, building an error chain. `errors.Is(err, ErrX)` searches the chain for
a sentinel value, and `errors.AsType[*T](err)` searches it for an error of
type `*T`. Plain `==` and type assertions only see the outermost error.

The tricky part: every bug in this exercise produces exactly the right
*message*. `%v` and `%w` print the same text, so reading the output does not
reveal a broken chain. Only the checks do.

A single `%v` anywhere in the chain cuts it. Everything below that point is
gone, even if every other layer uses `%w`:

```text
database password for dev: ...         (%w)  chain intact
  fetch secret "dev/db/password": ...  (%v)  chain cut here
    secret not found                         unreachable for errors.Is
```

Also remember that `errors.Is` compares error *values*, not text. A second
`errors.New("vault timed out")` is a different error from `ErrTimeout`, even
though the messages match.

## Your task

The vault code is in `exercise.go`. `read`, `Vault`, the sentinel errors and
`AccessDeniedError` are correct. Fix the three buggy spots marked with TODOs.

### 1. Fix the broken chain in FetchSecret

`FetchSecret(v, role, path)` adds the path as context, but uses `%v`, so
`DatabasePassword`'s callers can no longer find `ErrSecretNotFound`,
`ErrTimeout` or `*AccessDeniedError`. Make it wrap instead. The messages stay
the same:

```go
FetchSecret(vault, "deployer", "dev/api/token")
// => "", error `fetch secret "dev/api/token": secret not found`
// errors.Is(err, ErrSecretNotFound) == true

DatabasePassword(vault, "intern", "prod")
// => "", error `database password for prod: fetch secret "prod/db/password": role "intern" may not read "prod/db/password"`
// errors.AsType[*AccessDeniedError](err) finds it
```

### 2. Fix ShouldRetry

`ShouldRetry(err)` should return `true` when `ErrTimeout` is anywhere in the
chain, and `false` otherwise (including for `nil`). It currently compares with
`==`, which misses wrapped timeouts:

```go
ShouldRetry(fmt.Errorf("fetch secret %q: %w", "prod/api/token", ErrTimeout))
// => true

ShouldRetry(errors.New("vault timed out"))
// => false: same text, but not ErrTimeout
```

### 3. Fix Explain

`Explain(err)` should return:

| Error | Return |
|---|---|
| `nil` | `"ok"` |
| chain contains an `*AccessDeniedError` | `ask an admin to let role "ROLE" read "PATH"`, using its fields |
| anything else | `"failed: "` followed by the error message |

It uses a type assertion, which only sees the outermost error. Make it search
the whole chain:

```go
Explain(fmt.Errorf("fetch secret %q: %w", "prod/api/token", &AccessDeniedError{Role: "ci", Path: "prod/api/token"}))
// => `ask an admin to let role "ci" read "prod/api/token"`
```

## Learn more

- [Package errors](https://pkg.go.dev/errors)
- [Go blog: Working with Errors in Go 1.13](https://go.dev/blog/go1.13-errors)
- [fmt.Errorf](https://pkg.go.dev/fmt#Errorf)
