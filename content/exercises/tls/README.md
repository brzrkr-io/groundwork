# TLS: trust a CA and read certificates

Every `https://` URL your tools call uses *TLS*. Knowing how a Go client
decides whom to trust lets you call internal services signed by your
company's own certificate authority, and spot certificates about to expire.

**New here:** `crypto/x509` certificates and `CertPool`, `tls.Config`,
`http.Transport`, `resp.TLS`, dividing durations.

## Learn

### What TLS gives you

TLS wraps a TCP connection and adds two things:

| Guarantee | Meaning |
|---|---|
| Encryption | Nobody in between can read or change the bytes. |
| Identity | The client knows it really is talking to `payments.internal`. |

Identity comes from a *certificate*: a signed document the server sends when
the connection starts. Go represents one as `*x509.Certificate` (package
`crypto/x509`):

| Field | Example | Meaning |
|---|---|---|
| `DNSNames` | `[]string{"example.com"}` | Host names it is valid for. |
| `NotBefore`, `NotAfter` | `time.Time` values | Valid only between these times. |
| `Issuer` | `O=Acme Co` | Who signed it. |

### Who do you trust?

A *certificate authority* (CA) signs certificates. The client accepts a
server only if a CA it trusts signed the server's certificate.

By default Go trusts the operating system's list of public CAs. Internal
services often use a private company CA, so you give the client its own list,
an `x509.CertPool`:

```go
pool := x509.NewCertPool()
pool.AddCert(rootCert) // trust this one certificate
```

### Plugging the pool into a client

Settings nest, from the inside out:

| Piece | Holds |
|---|---|
| `tls.Config` (package `crypto/tls`) | `RootCAs` (the pool), `MinVersion` |
| `http.Transport` | `TLSClientConfig` (the config). The *transport* opens connections. |
| `http.Client` | `Transport`, `Timeout` |

The same nesting pattern, with a connection-pool setting instead of TLS:

```go
client := &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns: 20,
	},
}
```

`tls.VersionTLS12` is the constant for TLS 1.2. Setting it as `MinVersion`
refuses older, broken versions.

> **Why is it written like this?** `&http.Transport{...}` makes a value and
> takes its address (`&`), because the field wants a pointer: Go requires the
> types to match. `Name: value` pairs need a comma after each one when the
> closing `}` is on its own line (Go requires that too). Lining up the values
> is gofmt's convention.

You'll see `InsecureSkipVerify: true` in many examples online. It switches off
the identity check, so anyone can pretend to be your server. Never ship it.

### Reading the server's certificate

After a request over TLS, `resp.TLS` describes the connection. Its
`PeerCertificates` slice holds the certificates the server sent; the server's
own one is first (index `0`). For plain `http://`, `resp.TLS` is `nil`.

`client.Get(url)` is a shorthand for building a GET request and calling
`client.Do`.

### Days until expiry

`t1.Sub(t2)` returns a `time.Duration`. Dividing a duration by another gives
a whole number, with the remainder dropped:

```go
uptime := 150 * time.Minute
hours := int(uptime / time.Hour) // 2
```

The remainder is dropped toward zero, also for negative durations: -36 hours
divided by a day gives `-1`.

> **Why is it written like this?** The division gives a `time.Duration`; Go
> requires the `int(...)` conversion to return it as an `int`. Parentheses
> around `24 * time.Hour` make sure the multiplication happens first.

### Testing TLS

`httptest.NewTLSServer` starts a local HTTPS server with a test certificate.
`srv.Certificate()` returns it, so the tests can pass it to your code as the
CA to trust.

## Your task

You're writing a certificate checker for your platform team's internal
services.

### 1. Count the days until expiry

**Goal:** `DaysUntilExpiry(cert, now)` returns the whole days from `now`
until `cert.NotAfter`.

Drop partial days. Expired certificates give zero or a negative number.

```go
// cert.NotAfter is 30 days and 5 hours after now
DaysUntilExpiry(cert, now)  // => 30

// cert.NotAfter is 12 hours after now
DaysUntilExpiry(cert, now)  // => 0

// cert.NotAfter was 36 hours before now
DaysUntilExpiry(cert, now)  // => -1
```

**Think it through:**
- Which time do you subtract from which, so that time left is positive?
- How long is one day as a `time.Duration`?

### 2. Build a client that trusts your CA

**Goal:** `NewClient(ca)` returns a client that trusts **only** `ca` and
requires TLS 1.2 or newer.

- Put `ca` in a new `x509.CertPool` and set that pool as `RootCAs`.
- Set `MinVersion` to `tls.VersionTLS12`.
- Keep the 5-second `Timeout`.
- Don't set `InsecureSkipVerify`.

The tests check `client.Transport` is an `*http.Transport` with a
`TLSClientConfig`. A request to a server signed by `ca` must succeed; one
signed by another CA must fail. Add the `"crypto/tls"` import.

**Think it through:**
- Build inside out: pool, then config, then transport, then client.
- Which name holds the CA: the parameter `ca`, used inside the function.

### 3. Fetch the server's certificate

**Goal:** `PeerCertificate(client, url)` sends a GET and returns the server's
own certificate.

- Send the GET with `client`. If it fails, return the error.
- Close the response body.
- If `resp.TLS` is `nil` (a plain `http://` URL), return an error, such as
  `http://127.0.0.1:41234 is not using TLS`.
- Otherwise return the first certificate in `resp.TLS.PeerCertificates`.

Add the `"fmt"` import for `fmt.Errorf`.

**Think it through:**
- What must you check before you read anything inside `resp.TLS`?
- Which index is the server's own certificate?

## Learn more

- [crypto/tls package](https://pkg.go.dev/crypto/tls)
- [crypto/x509: CertPool](https://pkg.go.dev/crypto/x509#CertPool)
- [net/http/httptest: NewTLSServer](https://pkg.go.dev/net/http/httptest#NewTLSServer)
