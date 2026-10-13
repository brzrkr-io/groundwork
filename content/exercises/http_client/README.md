# HTTP clients: fetch a service status

Most ops tooling talks HTTP: health checks, status pages, cloud APIs, chat
webhooks. Here you write a small client that asks a status service how a
service is doing, the way a production tool should.

## Learn

### Requests and responses

An HTTP *request* has a *method* (`GET` to read, `POST` to send), a URL, and
*headers*: `Name: value` pairs with extra information such as who you are. The
server sends back a *response* with a *status code* (`200` OK, `404` Not
Found, `503` Service Unavailable, ...), its own headers, and a *body*.

### Your own client, with a timeout

In the `net/http` package an `*http.Client` sends requests. It has a `Timeout`
field that limits the whole request, from connecting to reading the body:

```go
client := &http.Client{Timeout: 10 * time.Second}
```

You will see `http.Get(url)` in tutorials. It uses the shared
`http.DefaultClient`, which has **no timeout**. One stuck server and your
health check hangs forever. In real tools, always build your own client with a
timeout.

### Building a request

`http.NewRequestWithContext` builds a request without sending it, so you can
add headers first:

```go
req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://status.internal/status/api", nil)
if err != nil {
	return nil, err // for example, the URL couldn't be parsed
}
req.Header.Set("Authorization", "Bearer s3cret")
req.Header.Set("Accept", "application/json")
```

- `ctx` is a *context*. A `context.Context` carries a deadline and a cancel
  signal, and `context.Background()` is the empty starting context. The
  Concurrency module covers contexts properly; here you pass `ctx` along.
- `http.MethodGet` is the constant `"GET"`.
- The last argument is the request body. A `GET` has none, so it is `nil`.
- `Authorization: Bearer <token>` is the usual way to send an API token.
  `Accept` says which format you want back.

### Sending it and reading the response

```go
resp, err := client.Do(req)
if err != nil {
	return err // no response at all: refused, timed out, ...
}
defer resp.Body.Close()

if resp.StatusCode != http.StatusOK {
	return fmt.Errorf("unexpected status %d", resp.StatusCode)
}
```

Three rules for every response:

1. **Close the body.** `resp.Body` holds the network connection open until you
   close it. `defer resp.Body.Close()` right after the error check does it on
   every path, including early returns.
2. **Check the status code.** `client.Do` only returns an error when there is
   no response at all. A `503` is still a response, so `err` is `nil`.
   `http.StatusOK` is the constant `200`.
3. **Limit what you read.** A broken server can send gigabytes.
   `io.LimitReader(r, n)` wraps a reader and stops after `n` bytes.

The body is an `io.Reader`, so you decode JSON from it with `json.NewDecoder`,
as in the JSON exercises:

```go
var s ServiceStatus
err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&s)
```

`1<<20` is 1,048,576 bytes: 1 MiB. If the JSON is cut off by the limit,
`Decode` returns an error.

## Your task

Your team's status service answers `GET /status/{service}` with JSON like:

```json
{"name": "payments", "version": "1.4.2", "healthy": true}
```

### 1. Build a client with a timeout

`NewClient()` returns a new `*http.Client` whose `Timeout` is 5 seconds. Add
the `time` import.

### 2. Build the status request

`NewStatusRequest(ctx, baseURL, service, token)` returns a `GET` request for
`baseURL + "/status/" + service`, made with `http.NewRequestWithContext` and
`ctx`, with two headers:

| Header | Value |
|---|---|
| `Authorization` | `Bearer ` followed by `token` |
| `Accept` | `application/json` |

```go
NewStatusRequest(ctx, "http://status.internal", "payments", "s3cret")
// => GET http://status.internal/status/payments
//    Authorization: Bearer s3cret
//    Accept: application/json
```

If `http.NewRequestWithContext` returns an error (for example for the base URL
`"://status.internal"`), return `nil` and that error.

### 3. Finish `FetchStatus`

`FetchStatus(client, req)` already sends the request and decodes the body.
Finish it:

- close the response body when the function returns, on every path
- when the status code is not `200`, return an empty `ServiceStatus` and the
  error `unexpected status 503` (with the real code); add the `fmt` import
- decode at most 1 MiB of the body; add the `io` import

```go
FetchStatus(client, req)
// 200 with {"name":"payments","version":"1.4.2","healthy":true}
//   => ServiceStatus{Name: "payments", Version: "1.4.2", Healthy: true}, nil
// 503
//   => ServiceStatus{}, error "unexpected status 503"
```

The tests use `httptest.NewServer`, a real HTTP server on your machine that
lives only during the test. They also check that the body was closed, and send
a body over 1 MiB that must produce an error.

## Learn more

- [net/http package](https://pkg.go.dev/net/http)
- [http.Client](https://pkg.go.dev/net/http#Client)
- [io.LimitReader](https://pkg.go.dev/io#LimitReader)
- [Go by Example: HTTP Client](https://gobyexample.com/http-client)
