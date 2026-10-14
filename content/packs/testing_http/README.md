# Testing HTTP: recorders and fake servers

HTTP code breaks in quiet ways: a wrong status code, a missing header, a
client that treats a 404 page as real data. The `net/http/httptest` package
tests both sides, servers and clients, in milliseconds and without a real
network.

**New here:** `httptest.NewRequest`, `httptest.NewRecorder`, `httptest.NewServer`.

## Learn

### Testing a handler with a ResponseRecorder

A handler only needs a `ResponseWriter` and a `*Request`, so a test can call
it directly with fakes:

```go
req := httptest.NewRequest("GET", "/users/42", nil)
rec := httptest.NewRecorder()
h.ServeHTTP(rec, req)
```

- `httptest.NewRequest(method, path, body)` builds a request; `nil` = no body.
- `httptest.NewRecorder()` returns a *ResponseRecorder*: a `ResponseWriter`
  that remembers everything written to it.
- `h.ServeHTTP(rec, req)` runs the handler. A `ServeMux` has this method
  too, so the request goes through the routes.

Then read the response from the recorder:

| Field | What it holds |
|---|---|
| `rec.Code` | the status code, such as `200` |
| `rec.Header().Get("X-Request-Id")` | one response header |
| `rec.Body.String()` | the whole body as a string |

A check looks like this (here for a header, with fixed text in quotes):

```go
if got := rec.Header().Get("Cache-Control"); got != "no-store" {
	t.Errorf("Cache-Control = %q, want %q", got, "no-store")
}
```

> **Why is it written like this?** `if got := ...; got != want` is Go syntax
> for "declare, then test" in one line; `got` only exists inside the `if`.
> The names `got`/`want` and the `"X = got, want Y"` message are convention.

### Table cases

A table test is a slice of structs; each `{...}` line is one case. Values go
in the same order as the struct's fields, separated by commas:

```go
{"admin user", "/users/1", http.StatusOK, "ana\n"},
```

### Testing a client with a fake server

A client needs a real URL. `httptest.NewServer(handler)` starts a real server
on a random port on your own machine (`127.0.0.1`):

```go
mux := http.NewServeMux()
mux.HandleFunc("GET /users/42", func(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "bob\n")
})
srv := httptest.NewServer(mux)
defer srv.Close()

name, err := FetchName(srv.Client(), srv.URL, "42")
```

- `srv.URL` is its address, like `http://127.0.0.1:41235`.
- `srv.Client()` is an `*http.Client` set up to talk to it.
- `defer srv.Close()` shuts it down when the test ends.

This fake knows one path. The mux answers 404 for every other path, so a
client asking for the wrong URL or an unknown user gets a 404. That's how you
test the client's error handling.

> **Why is it written like this?** The `func(w, r) {...}` passed to
> `HandleFunc` is an *anonymous function*: a handler with no name, written
> inline. Go requires the same parameter types as a named handler.

### How Groundwork grades your tests

`grader_test.go` holds `TestGrader`. It runs your `TestHostHandler` and
`TestFetchIP` against the real code (they must pass), then against *mutants*:
copies with one deliberate bug each (your tests must fail). Don't touch
`grader_test.go` or the `newHandler`/`fetchIP` variables in `exercise.go`.

Bugs your tests must catch:

| Mutant | The bug |
|---|---|
| `unknown_host_200` | the handler answers 200 OK for an unknown host |
| `vague_not_found` | the 404 body is `not found\n` instead of `unknown host: <name>\n` |
| `same_ip_for_every_host` | the handler answers `10.0.0.5` for every known host |
| `json_content_type` | the handler sends `Content-Type: application/json` |
| `client_ignores_status` | `FetchIP` returns a 404 body instead of an error |
| `client_wrong_path` | `FetchIP` requests `/host/<name>` instead of `/hosts/<name>` |
| `client_keeps_newline` | `FetchIP` returns `"10.0.0.5\n"` instead of `"10.0.0.5"` |
| `client_always_fails` | `FetchIP` returns an error for every host |

## Your task

`exercise.go` holds a host lookup API and a client for it. Both are correct.
You write the tests in `exercise_test.go`.

The handler answers `GET /hosts/{name}` from a `Directory`:

```text
Directory{"web-1": "10.0.0.5", "db-1": "10.0.1.7"}

GET /hosts/web-1  => 200, Content-Type: text/plain; charset=utf-8, body "10.0.0.5\n"
GET /hosts/db-9   => 404, Content-Type: text/plain; charset=utf-8, body "unknown host: db-9\n"
```

`FetchIP(client, baseURL, host)` requests `baseURL + "/hosts/" + host`,
returns the body without the trailing newline, and returns an error for any
status other than 200:

```go
FetchIP(srv.Client(), srv.URL, "web-1") // => "10.0.0.5", nil
FetchIP(srv.Client(), srv.URL, "db-9")  // => "", error (server said 404)
```

### 1. Add the missing cases

**Goal:** add two cases to the `TestHostHandler` table: `db-1`, and a host that isn't in the directory.

- `db-1`: path `/hosts/db-1`, status 200, body `10.0.1.7\n`
- unknown host, for example `db-9`: path `/hosts/db-9`, status 404, body
  `unknown host: db-9\n`

**Think it through:**
- Copy the shape of the existing case: name, path, code, body, in that order.
- Use the `http.Status...` constants for codes.

### 2. Check the response

**Goal:** inside the loop, check the status, the `Content-Type` header, and the body, with `t.Errorf` for each.

- `rec.Code` must equal `tt.wantCode`
- `rec.Body.String()` must equal `tt.wantBody`
- the `Content-Type` header must be `text/plain; charset=utf-8` for every case

**Think it through:**
- Which values come from the case (`tt.`), and which is fixed text?
- Three separate `if` checks, so one failure doesn't hide another.

### 3. Test the client

**Goal:** write `TestFetchIP`: start a fake server, then check `FetchIP` for a known and an unknown host.

- The fake answers `GET /hosts/web-1` with body `10.0.0.5\n`. Add the `"fmt"` import.
- `FetchIP(srv.Client(), srv.URL, "web-1")` returns `"10.0.0.5"` and no
  error. On an error, stop with `t.Fatalf`.
- `FetchIP(srv.Client(), srv.URL, "db-9")` returns an error.

Done when your tests pass and `TestGrader` reports every mutant caught.

**Think it through:**
- Which path and body must your fake serve? Compare with the Learn example.
- For `db-9`, you only care whether `err` is `nil`.

## Learn more

- [Package net/http/httptest](https://pkg.go.dev/net/http/httptest)
- [httptest.ResponseRecorder](https://pkg.go.dev/net/http/httptest#ResponseRecorder)
- [httptest.NewServer](https://pkg.go.dev/net/http/httptest#NewServer)
