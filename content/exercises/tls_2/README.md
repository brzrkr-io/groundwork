# TLS: stop skipping verification

A teammate wrote a nightly certificate audit for your internal services. It
"works" only because it skips the security check. Fix it, then teach it to
explain the TLS errors that wake people up at night.

**New here:** PEM files and `AppendCertsFromPEM`, `time.DateOnly`, telling
`crypto/x509` errors apart.

## Learn

### Recap

A TLS client trusts a server when a CA in its `x509.CertPool` signed the
server's certificate. The pool goes in `tls.Config.RootCAs`, the config in
`http.Transport.TLSClientConfig`, the transport in `http.Client`.

`InsecureSkipVerify: true` switches the check off, and with it the identity
guarantee. A certificate is valid until `cert.NotAfter`.

### PEM files

CA certificates usually arrive as a `ca.crt` file in the *PEM* format: base64
text between marker lines.

```text
-----BEGIN CERTIFICATE-----
MIIBlzCCAT2gAwIBAgIBKjAKBggqhkjOPQQDAjAyMRMwEQYDVQQKEwpPdGhlciBD
...
-----END CERTIFICATE-----
```

`pool.AppendCertsFromPEM(data)` adds every certificate it finds in the bytes
and returns a `bool`: `false` means it found none (usually the wrong file).

```go
ok := pool.AppendCertsFromPEM(data)
fmt.Println(ok) // false for "" or "hello"
```

> **Why is it written like this?** `!` flips a `bool`, so `if !ok` means "if
> not ok". Returning `nil` plus an error when something fails is Go
> convention for functions that return `(value, error)`.

### Dates as text

`t.Format(layout)` turns a time into text. `time.DateOnly` is the layout
`"2006-01-02"`, so it gives `YYYY-MM-DD`:

```go
deployed := time.Date(2025, 7, 4, 9, 30, 0, 0, time.UTC)
deployed.Format(time.DateOnly) // "2025-07-04"
```

`a.After(b)` reports whether time `a` is later than `b`.

### Telling TLS errors apart

When verification fails, `client.Get` returns an error wrapping a
`crypto/x509` error. `errors.AsType` (from *Wrapping errors*) digs through
the wrapping. These types are plain structs, so you write the struct type in
the brackets, not a pointer:

| Error type | Meaning | Typical fix |
|---|---|---|
| `x509.UnknownAuthorityError` | Signed by a CA you don't trust. | Load the right CA. |
| `x509.HostnameError` | Certificate is for a different host. | Fix the URL or reissue. |
| `x509.CertificateInvalidError` | Invalid. Its `Reason` field says why; `x509.Expired` means expired. | Renew it. |

You can check a field on the found error in the same `if`:

```go
if pathErr, ok := errors.AsType[*fs.PathError](err); ok && pathErr.Op == "open" {
	fmt.Println("could not open", pathErr.Path)
}
```

> **Why is it written like this?** `errors.AsType[T](err)` returns two values:
> the error as type `T`, and `ok`. The `[T]` brackets are required Go syntax
> for choosing the type. `&&` only checks the right side when the left is
> true, so you never read a field of an error that wasn't found.

## Your task

### 1. Stop skipping verification

**Goal:** `NewClient(caPEM)` trusts only the certificates in `caPEM`, instead
of skipping the check.

- Remove `InsecureSkipVerify: true`.
- Build a pool with `AppendCertsFromPEM` and set it as `RootCAs`.
- If `caPEM` holds no certificates (for example `""` or
  `"not a certificate"`), return a `nil` client and an error, such as
  `no CA certificates found in PEM data`.
- Keep `MinVersion` and the 5-second `Timeout`.

A request to a server signed by a CA in `caPEM` must succeed; any other must
fail.

**Think it through:**
- What does `AppendCertsFromPEM` return when the bytes hold no certificate?
- Which field in the `tls.Config` holds the pool?

### 2. Finish the expiry status

**Goal:** `ExpiryStatus(cert, now, warnDays)` also returns the `EXPIRED` and
`WARN` lines.

It already counts whole days left (`days`) and returns the `OK` line. Check
the cases in this order:

| Condition | Return |
|---|---|
| `now` is after `cert.NotAfter` | `EXPIRED on <NotAfter as YYYY-MM-DD>` |
| fewer than `warnDays` days left | `WARN: expires in <days> days` |
| otherwise | `OK: expires in <days> days` |

```go
// NotAfter is 2026-10-20, now is 2026-10-11, warnDays is 14
ExpiryStatus(cert, now, 14)  // => "WARN: expires in 9 days"

// NotAfter is exactly 14 days after now, warnDays is 14
ExpiryStatus(cert, now, 14)  // => "OK: expires in 14 days"

// NotAfter is 3 hours after now
ExpiryStatus(cert, now, 14)  // => "WARN: expires in 0 days"

// NotAfter is 2026-03-01, now is 2026-10-11
ExpiryStatus(cert, now, 14)  // => "EXPIRED on 2026-03-01"
```

**Think it through:**
- Why must "expired" be checked before "warn"? (An hour past expiry, `days`
  is still `0`.)
- "Fewer than" is `<`, not `<=`. Copy the `OK` line's format for `WARN`.

### 3. Diagnose TLS errors

**Goal:** `Diagnose(err)` also returns `hostname mismatch` and `expired` for
the matching x509 errors.

Add the two cases before the final `return "other"`:

| Error | Return |
|---|---|
| `nil` | `ok` (done) |
| wraps `x509.UnknownAuthorityError` | `unknown authority` (done) |
| wraps `x509.HostnameError` | `hostname mismatch` |
| wraps `x509.CertificateInvalidError` with `Reason == x509.Expired` | `expired` |
| anything else | `other` (done) |

The tests make real errors: they connect to a local HTTPS server with a
client that trusts the wrong CA, expects a different host name, or pretends
it is the year 2090.

**Think it through:**
- The unknown-authority check already shows the shape. What changes?
- For expiry, which field of the error do you also check?

## Learn more

- [crypto/x509: CertPool.AppendCertsFromPEM](https://pkg.go.dev/crypto/x509#CertPool.AppendCertsFromPEM)
- [crypto/x509: HostnameError](https://pkg.go.dev/crypto/x509#HostnameError)
- [Go blog: Working with errors](https://go.dev/blog/go1.13-errors)
