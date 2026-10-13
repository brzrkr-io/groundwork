# Testing HTTP: recorders and fake servers

HTTP code breaks in quiet ways: a wrong status code, a missing header, a
client that treats a 404 page as real data. The `net/http/httptest` package
lets you test both sides, servers and clients, in milliseconds and without
touching a real network.

## Learn

### Testing a handler with a ResponseRecorder

A handler only needs a `ResponseWriter` and a `*Request`, so a test can call
it directly with fakes:

```go
req := httptest.NewRequest("GET", "/hosts/web-1", nil)
rec := httptest.NewRecorder()
h.ServeHTTP(rec, req)
```

- `httptest.NewRequest(method, path, body)` builds a request. `nil` means no
  body.
- `httptest.NewRecorder()` returns a *ResponseRecorder*, a `ResponseWriter`
  that remembers everything written to it.
- `h.ServeHTTP(rec, req)` runs the handler. Every `http.Handler` has this
  method, a `ServeMux` included, so the request goes through the routes too.

Afterwards, read the response from the recorder:

| Field | What it holds |
|---|---|
| `rec.Code` | the status code, such as `200` |
| `rec.Header().Get("Content-Type")` | one response header |
| `rec.Body.String()` | the whole body as a string |

```go
if rec.Code != http.StatusOK {
	t.Errorf("GET %s: status = %d, want %d", path, rec.Code, http.StatusOK)
}
```

### Testing a client with a fake server

A client needs a real URL to call. `httptest.NewServer` starts a real HTTP
server on a random port of `127.0.0.1` (your own machine) and serves it with
any handler you give it:

```go
mux := http.NewServeMux()
mux.HandleFunc("GET /hosts/web-1", func(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "10.0.0.5\n")
})
srv := httptest.NewServer(mux)
defer srv.Close()

ip, err := FetchIP(srv.Client(), srv.URL, "web-1")
```

- `srv.URL` is the server's address, such as `http://127.0.0.1:41235`.
- `srv.Client()` returns an `*http.Client` set up to talk to it.
- `defer srv.Close()` shuts the server down when the test ends.

This fake knows exactly one path. The mux answers 404 for every other path,
so a client that asks for the wrong URL, or a host the fake doesn't know,
gets a 404. That is how you test the client's error handling.

### How Gopherlings grades your tests

As in *Your first test*, `grader_test.go` holds `TestGrader`. It re-runs your
`TestHostHandler` and `TestFetchIP` against the real code, where they must
pass, and then against *mutants*: copies of the code with one deliberate bug
each. Your tests must fail against every mutant. The `newHandler` and
`fetchIP` variables in `exercise.go` are how the grader swaps the mutants in.
You don't need to touch them or `grader_test.go`. These are the bugs your
tests must catch:

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
FetchIP(srv.Client(), srv.URL, "web-1")  // => "10.0.0.5", nil
FetchIP(srv.Client(), srv.URL, "db-9")   // => "", error (server said 404)
```

### 1. Add the missing cases

`TestHostHandler` has a table with one case. Add a case for `db-1`
(`/hosts/db-1`, 200, body `10.0.1.7\n`) and a case for a host that isn't in
the directory (for example `/hosts/db-9`, 404, body `unknown host: db-9\n`).

### 2. Check the response

Inside the loop, compare `rec.Code` with `tt.wantCode` and
`rec.Body.String()` with `tt.wantBody`. Also check that the `Content-Type`
header is `text/plain; charset=utf-8`, which is right for every case. Use
`t.Errorf` for each check.

### 3. Test the client

Write `TestFetchIP`. Start a fake server that answers `GET /hosts/web-1` with
the body `10.0.0.5\n` (the example in Learn does exactly this). Add the
`"fmt"` import for it. Then check that:

- `FetchIP(srv.Client(), srv.URL, "web-1")` returns `"10.0.0.5"` and no
  error (stop with `t.Fatalf` on an error),
- `FetchIP(srv.Client(), srv.URL, "db-9")` returns an error.

The exercise is done when your tests pass and `TestGrader` reports that every
mutant was caught.

## Learn more

- [Package net/http/httptest](https://pkg.go.dev/net/http/httptest)
- [httptest.ResponseRecorder](https://pkg.go.dev/net/http/httptest#ResponseRecorder)
- [httptest.NewServer](https://pkg.go.dev/net/http/httptest#NewServer)
