## Task 1: replace the flag with a pool

Delete `InsecureSkipVerify: true`. Create a pool with `x509.NewCertPool()`,
fill it with `pool.AppendCertsFromPEM(caPEM)`, and return an error when that
call returns `false`. Then set `RootCAs: pool` in the `tls.Config`.

## Task 2: order matters

Check `now.After(cert.NotAfter)` first and return
`"EXPIRED on " + cert.NotAfter.Format(time.DateOnly)`. Then check
`days < warnDays` and build the WARN line with `fmt.Sprintf`, just like the
OK line.

## Task 3: copy the pattern

The unknown-authority check shows the shape. Repeat it with
`errors.AsType[x509.HostnameError](err)`, and for expiry also look at the
`Reason` field:

```go
if invalid, ok := errors.AsType[x509.CertificateInvalidError](err); ok && invalid.Reason == x509.Expired {
	return "expired"
}
```
