## Task 1: divide durations

`cert.NotAfter.Sub(now)` gives a `time.Duration`. Divide it by
`24 * time.Hour` and convert the result with `int(...)`.

## Task 2: pool, config, transport

Build it from the inside out: an `x509.NewCertPool()` with `pool.AddCert(ca)`,
then a `&tls.Config{RootCAs: ..., MinVersion: ...}`, then an
`&http.Transport{TLSClientConfig: ...}` set as the client's `Transport`.

## Task 3: resp.TLS

```go
resp, err := client.Get(url)
if err != nil {
	return nil, err
}
defer resp.Body.Close()
if resp.TLS == nil {
	return nil, fmt.Errorf("%s is not using TLS", url)
}
```

Then return `resp.TLS.PeerCertificates[0]` with a `nil` error.

## Task 2: almost the answer

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
