## Task 1: Count the days until expiry

Subtracting two times gives a `time.Duration`. Dividing a duration by the
length of one day gives whole days, with the remainder dropped.

### Closer

`cert.NotAfter.Sub(now)` is the time left. Divide it by `(24 * time.Hour)` and
convert the result with `int(...)`.

### Answer

```go
return int(cert.NotAfter.Sub(now) / (24 * time.Hour))
```

## Task 2: Build a client that trusts your CA

Build it from the inside out: a pool holding `ca`, a TLS config using that
pool, a transport using that config, and the client using that transport.

### Closer

`pool := x509.NewCertPool()` then `pool.AddCert(ca)`. The config is
`&tls.Config{RootCAs: pool, MinVersion: tls.VersionTLS12}`, and it goes in
`&http.Transport{TLSClientConfig: ...}`. Add the `"crypto/tls"` import.

### Answer

```go
pool := x509.NewCertPool()
pool.AddCert(ca)
return &http.Client{
	Timeout: 5 * time.Second,
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:    pool,
			MinVersion: tls.VersionTLS12,
		},
	},
}
```

## Task 3: Fetch the server's certificate

Send the request, handle its error, and close the body. Then `resp.TLS` tells
you about the TLS connection, or is `nil` for plain `http://`.

### Closer

After `client.Get(url)` and `defer resp.Body.Close()`, check
`if resp.TLS == nil` and return a `fmt.Errorf("%s is not using TLS", url)`.
The server's own certificate is `resp.TLS.PeerCertificates[0]`.

### Answer

```go
resp, err := client.Get(url)
if err != nil {
	return nil, err
}
defer resp.Body.Close()
if resp.TLS == nil {
	return nil, fmt.Errorf("%s is not using TLS", url)
}
return resp.TLS.PeerCertificates[0], nil
```
