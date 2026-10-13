# HTTP clients: fix the deploy notifier

After every deploy, a small tool checks which version is live and posts an
event to the team's webhook. The tool has bugs, and the part that sends the
event isn't finished.

## Learn

### Recap

Send requests with your own `*http.Client` that has a `Timeout`, never
`http.DefaultClient`. After `client.Do` succeeds, `defer resp.Body.Close()`
and check `resp.StatusCode` yourself, because a `500` still comes back with a
`nil` error. Read bodies through `io.LimitReader`. `client.Get(url)` is a
shortcut for building a `GET` request and calling `client.Do`, handy when you
need no headers.

### Sending JSON with POST

To send data, encode it first, then hand the bytes to the request as its body.
`json.Marshal` turns a value into JSON bytes, using its struct tags.
`bytes.NewReader` wraps a byte slice in an `io.Reader`, which is what a
request body must be:

```go
body, err := json.Marshal(ev) // []byte(`{"service":"payments",...}`)
if err != nil {
	return nil, err
}
req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
```

Then set the `Content-Type` header to `application/json`, so the server knows
how to read the body.

### Success is any 2xx

Status codes come in families: 2xx means success, 4xx means the client made a
mistake, 5xx means the server failed. A webhook may answer `200 OK`,
`201 Created`, `202 Accepted` or `204 No Content`, and all of them mean "got
it". So a client that sends data usually accepts the whole 2xx range.

## Your task

### 1. Fix `GetVersion`

`GetVersion(client, url)` reads the live version from a plain-text body such
as `"v1.4.2\n"` and trims the spaces and newline. It has two bugs:

- it never closes the response body
- it treats an error page as a version: a `500` with body `internal error`
  comes back as the version `"internal error"`

Close the body on every path, and when the status is not `200`, return `""`
and the error `version: unexpected status 500` (with the real code).

```go
GetVersion(client, url)
// 200 with "v1.4.2\n" => "v1.4.2", nil
// 500                 => "", error "version: unexpected status 500"
```

### 2. Finish `NewDeployRequest`

`NewDeployRequest(ctx, url, ev)` must return a `POST` request to `url` whose
body is `ev` encoded as JSON, with the header
`Content-Type: application/json`. The request is already built, but without a
body or the header. Add the `bytes` and `encoding/json` imports.

```go
NewDeployRequest(ctx, "http://hooks.internal/deploys",
	DeployEvent{Service: "payments", Version: "v1.4.2", Env: "production"})
// => POST http://hooks.internal/deploys
//    Content-Type: application/json
//
//    {"service":"payments","version":"v1.4.2","env":"production"}
```

### 3. Accept every 2xx in `NotifyDeploy`

`NotifyDeploy(client, req)` only accepts `200`, so a webhook that answers
`201` or `204` looks like a failure. Return `nil` for every status from 200 to
299, and for anything else the error `notify: unexpected status 500` (with the
real code).

```go
NotifyDeploy(client, req)
// server answers 202 => nil
// server answers 503 => error "notify: unexpected status 503"
```

## Learn more

- [json.Marshal](https://pkg.go.dev/encoding/json#Marshal)
- [http.NewRequestWithContext](https://pkg.go.dev/net/http#NewRequestWithContext)
- [net/http status code constants](https://pkg.go.dev/net/http#pkg-constants)
