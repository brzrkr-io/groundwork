# Wrapping errors: fix the secrets fetcher

A deploy tool reads passwords and tokens from a secrets vault. Its messages
print correctly, but three checks silently give the wrong answer. Bugs like
these hide until the bad day you need them.

**New here:** nothing; practice finding broken error chains.

## Learn

### Recap

From *Wrapping errors*:

| Tool | What it does |
|---|---|
| `fmt.Errorf("context: %w", err)` | wraps `err`, adding to the chain |
| `errors.Is(err, ErrX)` | searches the chain for a sentinel value |
| `errors.AsType[*T](err)` | searches the chain for an error of type `*T` |
| `==` and `err.(*T)` | look at the outermost error only |

### The message lies

Every bug in this exercise prints exactly the right *message*. `%v` and `%w`
print the same text, so reading the output won't show a broken chain. Only
the checks will.

### One %v cuts the chain

A single `%v` anywhere cuts the chain. Everything below it is gone, even if
every other layer uses `%w`:

```text
renew cert for edge: ...            (%w)  chain intact
  read cert "edge": ...             (%v)  chain cut here
    expired                               unreachable for errors.Is
```

### Same text is not the same error

`errors.Is` compares error *values*, not text:

```go
a := errors.New("disk full")
b := errors.New("disk full")
errors.Is(a, b) // false: two different values
```

> **Why is it written like this?** Each `errors.New` call makes a new, unique
> value (Go requires this behaviour). That is why sentinels are stored once in
> a package-level `var` and everyone compares against that variable.

## Your task

The vault code is in `exercise.go`. `read`, `Vault`, the sentinel errors and
`AccessDeniedError` are correct. Fix the three spots marked with TODOs.

### 1. Fix the broken chain in FetchSecret

**Goal:** make `FetchSecret` wrap the vault's error instead of copying its text.

`FetchSecret(v, role, path)` adds the path as context. Right now callers
(including `DatabasePassword`) can't find `ErrSecretNotFound`, `ErrTimeout`
or `*AccessDeniedError` in its error. The messages must stay the same.

```go
FetchSecret(vault, "deployer", "dev/api/token")
// => "", error `fetch secret "dev/api/token": secret not found`
// errors.Is(err, ErrSecretNotFound) == true

DatabasePassword(vault, "intern", "prod")
// => "", error `database password for prod: fetch secret "prod/db/password": role "intern" may not read "prod/db/password"`
// errors.AsType[*AccessDeniedError](err) finds it
```

**Think it through:**
- Look at the format string on the line under the TODO. Which verb copies text?
- This is a one-character fix.

### 2. Fix ShouldRetry

**Goal:** make `ShouldRetry(err)` return `true` when `ErrTimeout` is anywhere in the chain.

Return `false` otherwise, including for `nil`. Right now it compares with
`==`, which misses wrapped timeouts.

```go
ShouldRetry(fmt.Errorf("fetch secret %q: %w", "prod/api/token", ErrTimeout))
// => true

ShouldRetry(errors.New("vault timed out"))
// => false: same text, but not ErrTimeout
```

**Think it through:**
- Which tool in the recap table searches the chain for a *value*?

### 3. Fix Explain

**Goal:** make `Explain(err)` find an `*AccessDeniedError` anywhere in the chain.

| Error | Return |
|---|---|
| `nil` | `"ok"` |
| chain contains an `*AccessDeniedError` | `ask an admin to let role "ROLE" read "PATH"`, using its `Role` and `Path` fields |
| anything else | `"failed: "` followed by the error message |

Only the type check is wrong: it uses a type assertion, which sees just the
outermost error.

```go
Explain(fmt.Errorf("fetch secret %q: %w", "prod/api/token", &AccessDeniedError{Role: "ci", Path: "prod/api/token"}))
// => `ask an admin to let role "ci" read "prod/api/token"`
```

**Think it through:**
- Which tool in the recap table searches the chain for a *type*?
- It returns the same two things as the type assertion, so only the `if` line changes.

## Learn more

- [Package errors](https://pkg.go.dev/errors)
- [Go blog: Working with Errors in Go 1.13](https://go.dev/blog/go1.13-errors)
- [fmt.Errorf](https://pkg.go.dev/fmt#Errorf)
