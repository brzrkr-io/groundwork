## Task 1: Stop skipping verification

Remove `InsecureSkipVerify` and give the config a pool of trusted CAs instead.
The pool is filled from the PEM bytes, and the fill call tells you whether it
found any certificates.

### Closer

`pool := x509.NewCertPool()`, then `pool.AppendCertsFromPEM(caPEM)` returns
`false` when it found none; return `nil` and an `errors.New(...)` then. Set
`RootCAs: pool` in the `tls.Config`.

### Answer

```go
pool := x509.NewCertPool()
if !pool.AppendCertsFromPEM(caPEM) {
	return nil, errors.New("no CA certificates found in PEM data")
}
tlsConfig := &tls.Config{
	RootCAs:    pool,
	MinVersion: tls.VersionTLS12,
}
```

## Task 2: Finish the expiry status

Order matters: check "expired" first, then "warn", and let the existing `OK`
line handle the rest.

### Closer

`now.After(cert.NotAfter)` tells you it has expired; format the date with
`cert.NotAfter.Format(time.DateOnly)`. For the warning, check
`days < warnDays` and build the line with `fmt.Sprintf`, like the `OK` line.

### Answer

```go
if now.After(cert.NotAfter) {
	return "EXPIRED on " + cert.NotAfter.Format(time.DateOnly)
}
if days < warnDays {
	return fmt.Sprintf("WARN: expires in %d days", days)
}
return fmt.Sprintf("OK: expires in %d days", days)
```

## Task 3: Diagnose TLS errors

The unknown-authority check already shows the shape. Copy it for the other two
error types; for expiry you also need to look at a field on the error.

### Closer

Use `errors.AsType[x509.HostnameError](err)` for the host name. For expiry,
keep the value from `errors.AsType[x509.CertificateInvalidError](err)` and also
check `invalid.Reason == x509.Expired`.

### Answer

```go
if _, ok := errors.AsType[x509.HostnameError](err); ok {
	return "hostname mismatch"
}
if invalid, ok := errors.AsType[x509.CertificateInvalidError](err); ok && invalid.Reason == x509.Expired {
	return "expired"
}
return "other"
```
