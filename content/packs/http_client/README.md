# HTTP clients: fetch a service status

Most ops tooling talks HTTP: health checks, status pages, cloud APIs, chat
webhooks. Here you write a small client that asks a status service how a
service is doing, the way a production tool should.

**New here:** `net/http` clients, requests, headers, status codes, `io.LimitReader`.

## Learn

### Requests and responses

An HTTP *request* has:

- a *method*: `GET` to read, `POST` to send
- a URL
- *headers*: `Name: value` pairs with extra information, such as who you are

The server sends back a *response* with a *status code* (`200` OK, `404` Not
Found, `503` Service Unavailable), its own headers, and a *body*.

### Your own client, with a timeout

In `net/http`, an `*http.Client` sends requests. Its `Timeout` field limits
the whole request, from connecting to reading the body:

```go
client := &http.Client{Timeout: 30 * time.Second}
```

Tutorials often use `http.Get(url)`. That uses the shared
`http.DefaultClient`, which has **no timeout**: one stuck server and your
health check hangs forever. Real tools build their own client.

> **Why is it written like this?** `&http.Client{...}` is a *struct literal*
> with `&` in front to get a pointer; Go requires `Field: value` inside the
> braces. `30 * time.Second` works because `time.Second` is a `Duration`, and
> Go needs a `Duration` there. Using a pointer to a client is convention.

### Building a request

`http.NewRequestWithContext` builds a request without sending it, so you can
add headers first:

```go
req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://chat.internal/hooks/"+room, body)
if err != nil {
	return nil, err // for example, the URL couldn't be parsed
}
req.Header.Set("X-Api-Key", key)
req.Header.Set("Content-Type", "text/plain")
```

- `ctx` is a *context*: it carries a deadline and a cancel signal. Here you just pass it along.
- `http.MethodPost` is the constant `"POST"`; `http.MethodGet` is `"GET"`.
- The last argument is the request body. A `GET` has none, so it is `nil`.
- `Header.Set(name, value)`: both are strings. The fixed text goes in quotes;
  a variable like `key` does not.

Two common headers:

| Header | Meaning |
|---|---|
| `Authorization: Bearer <token>` | the usual way to send an API token (note the space after `Bearer`) |
| `Accept: application/json` | "please answer in JSON" |

> **Why is it written like this?** Go requires you to receive both return
> values. Checking `err` right away and returning it is convention. The name
> `req` is the author's choice.

### Three rules for every response

`client.Do(req)` sends the request and returns `resp, err`.

1. **Close the body.** `resp.Body` holds the connection open until closed.
   `defer resp.Body.Close()` right after the error check closes it on every
   path, including early returns.
2. **Check the status code.** `client.Do` only returns an error when there is
   no response at all. A `503` is still a response, so `err` is `nil`. Look at
   `resp.StatusCode`; `http.StatusOK` is the constant `200`.
3. **Limit what you read.** A broken server can send gigabytes.
   `io.LimitReader(r, n)` wraps a reader and stops after `n` bytes.

```go
data, err := io.ReadAll(io.LimitReader(file, 512)) // reads at most 512 bytes
```

`1<<20` is 1,048,576 bytes: 1 MiB. If JSON is cut off by the limit, `Decode`
returns an error.

> **Why is it written like this?** `defer` is Go syntax: the call runs when
> the function returns. Putting it right after the error check is convention,
> because before that `resp` may be `nil`.

## Your task

Your team's status service answers `GET /status/{service}` with JSON like:

```json
{"name": "payments", "version": "1.4.2", "healthy": true}
```

### 1. Build a client with a timeout

**Goal:** `NewClient()` returns a new `*http.Client` whose `Timeout` is 5 seconds.

Do not return `http.DefaultClient`. Add the `time` import.

**Think it through:**
- Which field of the client sets the limit?
- How do you write "5 seconds" as a `time.Duration`?

### 2. Build the status request

**Goal:** `NewStatusRequest(ctx, baseURL, service, token)` returns a `GET` request with two headers.

- URL: `baseURL`, then the fixed text `/status/`, then `service`
- build it with `http.NewRequestWithContext`, passing `ctx`
- headers:

| Header | Value |
|---|---|
| `Authorization` | fixed text `Bearer ` (with the space), then `token` |
| `Accept` | `application/json` |

```go
NewStatusRequest(ctx, "http://status.internal", "payments", "s3cret")
// => GET http://status.internal/status/payments
//    Authorization: Bearer s3cret
//    Accept: application/json
```

If `http.NewRequestWithContext` returns an error (for example for the base
URL `"://status.internal"`), return `nil` and that error.

**Think it through:**
- Which parts are parameters and which are fixed text in quotes?
- Return the request only after both headers are set.

### 3. Finish `FetchStatus`

**Goal:** close the body, reject non-200 responses, and decode at most 1 MiB.

`FetchStatus(client, req)` already sends the request and decodes the body.
Fill in the three `TODO(3)` spots:

- close the response body when the function returns, on every path
- status code is not `200`: return an empty `ServiceStatus` and the error
  `unexpected status 503` (with the real code). Add the `fmt` import.
- decode at most 1 MiB (`1<<20` bytes) of the body. Add the `io` import.

```go
FetchStatus(client, req)
// 200 with {"name":"payments","version":"1.4.2","healthy":true}
//   => ServiceStatus{Name: "payments", Version: "1.4.2", Healthy: true}, nil
// 503 => ServiceStatus{}, error "unexpected status 503"
// 404 => ServiceStatus{}, error "unexpected status 404"
```

The tests run a real local server with `httptest.NewServer`. They check the
body was closed, and send a body over 1 MiB that must give an error.

**Think it through:**
- Which field of `resp` holds the code, and which `fmt` verb prints an `int`?
- What does the decoder read from now, and how do you wrap it?

## Learn more

- [net/http package](https://pkg.go.dev/net/http)
- [http.Client](https://pkg.go.dev/net/http#Client)
- [io.LimitReader](https://pkg.go.dev/io#LimitReader)
- [Go by Example: HTTP Client](https://gobyexample.com/http-client)
