# TLS: stop skipping verification

A teammate wrote a nightly certificate audit for your internal services. It
"works", but only because it skips the security check. Fix it, then teach it to
explain the TLS errors that wake people up at night.

## Learn

### Recap

A TLS client trusts a server when a CA in its `x509.CertPool` signed the
server's certificate. You put the pool in `tls.Config.RootCAs`, the config in
`http.Transport.TLSClientConfig`, and the transport in `http.Client`.
`InsecureSkipVerify: true` switches the check off, and with it the identity
guarantee. A certificate is valid until `cert.NotAfter`.

### PEM files

CA certificates usually arrive as a `ca.crt` file using the *PEM* format: base64 text
between marker lines.

```text
-----BEGIN CERTIFICATE-----
MIIBlzCCAT2gAwIBAgIBKjAKBggqhkjOPQQDAjAyMRMwEQYDVQQKEwpPdGhlciBD
...
-----END CERTIFICATE-----
```

`pool.AppendCertsFromPEM(data)` adds every certificate it finds in the bytes.
It returns `false` if it found none, which usually means the wrong file:

```go
pool := x509.NewCertPool()
if !pool.AppendCertsFromPEM(caPEM) {
	return nil, errors.New("no CA certificates found in PEM data")
}
```

### Telling TLS errors apart

When verification fails, `client.Get` returns an error that wraps an error
from `crypto/x509`. `errors.AsType` (from *Wrapping errors*) digs through the
wrapping. These three error types are plain struct values, so the type in the
brackets is the struct itself, not a pointer:

| Error type | Meaning | Typical fix |
|---|---|---|
| `x509.UnknownAuthorityError` | Signed by a CA you don't trust. | Load the right CA. |
| `x509.HostnameError` | Certificate is for a different host name. | Fix the URL or reissue the certificate. |
| `x509.CertificateInvalidError` | Certificate is invalid. Its `Reason` field says why; `x509.Expired` means expired. | Renew it. |

```go
if invalid, ok := errors.AsType[x509.CertificateInvalidError](err); ok && invalid.Reason == x509.Expired {
	// the certificate has expired
}
```

## Your task

### 1. Stop skipping verification

`NewClient(caPEM)` sets `InsecureSkipVerify: true`. Remove that, and make the
client trust only the certificates in `caPEM` instead:

- build a pool with `AppendCertsFromPEM` and set it as `RootCAs`,
- if `caPEM` holds no certificates, return a `nil` client and an error such as
  `no CA certificates found in PEM data`,
- keep `MinVersion` and the 5-second `Timeout`.

### 2. Finish the expiry status

`ExpiryStatus(cert, now, warnDays)` already counts the whole days left and
returns the `OK` line. Add the other two cases, checked in this order:

| Condition | Return |
|---|---|
| `now` is after `cert.NotAfter` | `EXPIRED on <NotAfter as YYYY-MM-DD>` |
| fewer than `warnDays` days left | `WARN: expires in <days> days` |
| otherwise | `OK: expires in <days> days` |

Format the date with `cert.NotAfter.Format(time.DateOnly)`; `time.DateOnly`
is the layout `"2006-01-02"`.

```go
// NotAfter is 2026-10-20, now is 2026-10-11, warnDays is 14
ExpiryStatus(cert, now, 14)
// => "WARN: expires in 9 days"

// NotAfter is 2026-03-01, now is 2026-10-11
ExpiryStatus(cert, now, 14)
// => "EXPIRED on 2026-03-01"
```

### 3. Diagnose TLS errors

`Diagnose(err)` turns a request error into a short label for the audit
report. It already handles `nil` and unknown authorities. Add the two missing
cases before the final `return "other"`:

| Error | Return |
|---|---|
| `nil` | `ok` (done) |
| wraps `x509.UnknownAuthorityError` | `unknown authority` (done) |
| wraps `x509.HostnameError` | `hostname mismatch` |
| wraps `x509.CertificateInvalidError` with `Reason == x509.Expired` | `expired` |
| anything else | `other` (done) |

The tests produce real errors by connecting to a local HTTPS server with a
client that trusts the wrong CA, expects a different host name, or pretends
it is the year 2090.

## Learn more

- [crypto/x509: CertPool.AppendCertsFromPEM](https://pkg.go.dev/crypto/x509#CertPool.AppendCertsFromPEM)
- [crypto/x509: HostnameError](https://pkg.go.dev/crypto/x509#HostnameError)
- [Go blog: Working with errors](https://go.dev/blog/go1.13-errors)
