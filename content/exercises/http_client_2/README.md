# HTTP clients: fix the deploy notifier

After every deploy, a small tool checks which version is live and posts an
event to the team's webhook. The tool has bugs, and the part that sends the
event isn't finished.

**New here:** `bytes.NewReader`, POST bodies, the 2xx range. (`json.Marshal` is from *Encoding JSON: a deployment status report*.)

## Learn

### Recap

- Use your own `*http.Client` with a `Timeout`.
- After `client.Do` succeeds: `defer resp.Body.Close()`, then check
  `resp.StatusCode` yourself. A `500` still comes back with `err == nil`.
- Read bodies through `io.LimitReader`.
- `client.Get(url)` is a shortcut: build a `GET` request and send it, no headers.

### Sending JSON with POST

To send data, turn it into bytes first, then give those bytes to the request
as its body.

- `json.Marshal(v)` turns a value into JSON bytes (`[]byte`), using its struct tags.
- A request body must be an `io.Reader`. `bytes.NewReader(b)` wraps a byte
  slice in one.

```go
data, err := json.Marshal(user) // []byte(`{"name":"ana"}`)
if err != nil {
	return err
}
r := bytes.NewReader(data) // r can now be read like a file
```

Then set `Content-Type: application/json`, so the server knows how to read the body.

> **Why is it written like this?** Go requires the body type to be an
> `io.Reader`, so a plain `[]byte` won't compile there. Checking `err` from
> `Marshal` is convention: it rarely fails, but it can.

### Success is any 2xx

Status codes come in families:

| Range | Meaning |
|---|---|
| 2xx | success (`200 OK`, `201 Created`, `202 Accepted`, `204 No Content`) |
| 4xx | the client made a mistake |
| 5xx | the server failed |

A client that sends data usually accepts the whole 2xx range. To test a
range, combine two comparisons:

```go
if age < 18 || age > 65 { // outside 18–65
```

## Your task

### 1. Fix `GetVersion`

**Goal:** `GetVersion(client, url)` must close the body and return an error for any status that isn't `200`.

It reads the live version from a plain-text body such as `"v1.4.2\n"` and
trims spaces and newlines. Two bugs:

- it never closes the response body
- it treats an error page as a version: a `500` with body `internal error`
  comes back as the version `"internal error"`

Close the body on every path. When the status is not `200`, return `""` and
the error `version: unexpected status 500` (with the real code).

```go
GetVersion(client, url)
// 200 with "v1.4.2\n" => "v1.4.2", nil
// 500                 => "", error "version: unexpected status 500"
// 404                 => "", error "version: unexpected status 404"
```

**Think it through:**
- Where must the close go so that the early return closes too?
- Check the status before reading the body.

### 2. Finish `NewDeployRequest`

**Goal:** `NewDeployRequest(ctx, url, ev)` returns a `POST` request whose body is `ev` as JSON, with header `Content-Type: application/json`.

The request is already built, but with no body (`nil`) and no header. Add the
`bytes` and `encoding/json` imports.

```go
NewDeployRequest(ctx, "http://hooks.internal/deploys",
	DeployEvent{Service: "payments", Version: "v1.4.2", Env: "production"})
// => POST http://hooks.internal/deploys
//    Content-Type: application/json
//
//    {"service":"payments","version":"v1.4.2","env":"production"}
```

**Think it through:**
- Encode first, so you have the bytes before building the request.
- What replaces the `nil` body argument?

### 3. Accept every 2xx in `NotifyDeploy`

**Goal:** `NotifyDeploy(client, req)` returns `nil` for every status from 200 to 299.

Right now it only accepts `200`, so a `201` or `204` looks like a failure.
For anything outside 200–299, return the error
`notify: unexpected status 500` (with the real code). Only the `if`
condition needs to change.

```go
NotifyDeploy(client, req)
// server answers 202 => nil
// server answers 400 => error "notify: unexpected status 400"
// server answers 503 => error "notify: unexpected status 503"
```

**Think it through:**
- Which two numbers bound the success range?
- The `if` describes the *failure* case: below one, or above the other.

## Learn more

- [json.Marshal](https://pkg.go.dev/encoding/json#Marshal)
- [http.NewRequestWithContext](https://pkg.go.dev/net/http#NewRequestWithContext)
- [net/http status code constants](https://pkg.go.dev/net/http#pkg-constants)
