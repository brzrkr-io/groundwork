# TLS: trust a CA and read certificates

Every `https://` URL your tools call uses *TLS*. Knowing how a Go client
decides whom to trust lets you talk to internal services signed by your
company's own certificate authority, and lets you spot certificates that are
about to expire before they take production down.

## Learn

### What TLS gives you

TLS wraps a TCP connection and adds two things:

| Guarantee | Meaning |
|---|---|
| Encryption | Nobody between client and server can read or change the bytes. |
| Identity | The client knows it really is talking to `payments.internal`, not an impostor. |

Identity comes from a *certificate*: a small signed document the server sends
at the start of the connection. Go represents one as `*x509.Certificate`
(package `crypto/x509`). The fields you will use most:

| Field | Example | Meaning |
|---|---|---|
| `DNSNames` | `[]string{"example.com"}` | Host names the certificate is valid for. |
| `NotBefore`, `NotAfter` | `time.Time` values | The certificate is only valid between these times. |
| `Issuer` | `O=Acme Co` | Who signed it. |

### Who do you trust?

A certificate is signed by a *certificate authority* (CA). The client keeps a
list of CAs it trusts, and accepts a server only if a trusted CA signed its
certificate. By default Go uses the operating system's list, which contains the
public CAs. Internal services are often signed by a private company CA instead,
so you give the client its own list, an `x509.CertPool`:

```go
pool := x509.NewCertPool()
pool.AddCert(caCert) // trust this CA, and only this CA
```

The pool goes into a `tls.Config` (package `crypto/tls`), which goes into an
`http.Transport`, which goes into the client:

```go
client := &http.Client{
	Timeout: 5 * time.Second,
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:    pool,             // the CAs to trust
			MinVersion: tls.VersionTLS12, // refuse old, broken TLS versions
		},
	},
}
```

The *transport* is the part of the client that opens connections. The
`Timeout` works as in *HTTP clients: fetch a service status*.

You will find `InsecureSkipVerify: true` in many examples online. It turns off
the identity check, so anyone can pretend to be your server. Never ship it.
Trust the right CA instead.

### Reading the server's certificate

After a request over TLS, `resp.TLS` describes the connection. Its
`PeerCertificates` slice holds the certificates the server sent, and the
server's own certificate comes first. For a plain `http://` request,
`resp.TLS` is `nil`.

```go
resp, err := client.Get("https://example.com") // shorthand for a GET request + client.Do
if err != nil {
	return nil, err
}
defer resp.Body.Close()
cert := resp.TLS.PeerCertificates[0]
```

### Days until expiry

`t1.Sub(t2)` returns a `time.Duration`. Dividing one duration by another gives
a whole number with the remainder dropped, so this counts whole days:

```go
left := cert.NotAfter.Sub(now)    // e.g. 30 days and 5 hours
days := int(left / (24 * time.Hour)) // 30
```

The remainder is dropped toward zero, also for negative durations: a
certificate that expired 36 hours ago gives `-1`.

### Testing TLS

`httptest.NewTLSServer` starts a local HTTPS server with a built-in test
certificate. `srv.Certificate()` returns that certificate, so the tests can
pass it to your code as the CA to trust.

## Your task

You are writing a certificate checker for your platform team's internal
services.

### 1. Count the days until expiry

Complete `DaysUntilExpiry` so it returns the whole days from `now` until
`cert.NotAfter`. Drop partial days. Expired certificates give zero or a
negative number.

```go
// cert.NotAfter is 30 days and 5 hours after now
DaysUntilExpiry(cert, now)
// => 30

// cert.NotAfter was 3 days before now
DaysUntilExpiry(cert, now)
// => -3
```

### 2. Build a client that trusts your CA

Complete `NewClient(ca)` so the client it returns:

- trusts **only** `ca`: put it in a new `x509.CertPool` and set that pool as
  `RootCAs`,
- sets `MinVersion` to `tls.VersionTLS12`,
- keeps the 5-second `Timeout`.

A request to a server signed by `ca` succeeds. A request to a server signed by
any other CA fails with an error. Add the `"crypto/tls"` import.

### 3. Fetch the server's certificate

Complete `PeerCertificate(client, url)`. Send a GET request to `url` with
`client`, close the response body, and return the server's own certificate,
`resp.TLS.PeerCertificates[0]`.

Return the error if the request fails. If `resp.TLS` is `nil` (a plain
`http://` URL), return an error such as:

```text
http://127.0.0.1:41234 is not using TLS
```

Add the `"fmt"` import for `fmt.Errorf`.

## Learn more

- [crypto/tls package](https://pkg.go.dev/crypto/tls)
- [crypto/x509: CertPool](https://pkg.go.dev/crypto/x509#CertPool)
- [net/http/httptest: NewTLSServer](https://pkg.go.dev/net/http/httptest#NewTLSServer)
