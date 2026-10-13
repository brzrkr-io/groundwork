# HTTP servers: fix the log viewer

When a service misbehaves at 3 a.m., the quickest look at its logs is often
a plain HTTP endpoint. This exercise fixes the routes and the query handling
of a small log viewer.

## Learn

### Recap

A handler has the shape `func(w http.ResponseWriter, r *http.Request)`. It
sets headers, then the status (default 200), then writes the body.
`http.Error(w, msg, code)` sends an error status with `msg` plus a newline as
the body. A `ServeMux` routes requests by patterns such as
`"GET /services/{name}"`, and `r.PathValue("name")` reads the wildcard.

### Patterns that end in a slash

A pattern ending in `/` matches that path **and everything below it**. So
`"/"` on its own matches every request: `/`, `/nope`, `/favicon.ico`. To
match only the root, end the pattern with `{$}`, which means "the path ends
here":

| Pattern | Matches `/` | Matches `/nope` |
|---|---|---|
| `"/"` | yes | yes |
| `"GET /{$}"` | yes | no, the mux answers 404 |

A pattern without a method, like `"/logs/{service}"`, accepts every method,
so `POST` and `DELETE` reach a handler that was only meant for reading. Put
the method in front.

### Query parameters

The part of a URL after `?` is the *query string*:
`/logs/api?limit=5`. `r.URL.Query().Get("limit")` returns the value as a
string: `"5"` here, or `""` when the parameter is missing. Convert it with
`strconv.Atoi`, as you've done before.

### 400 Bad Request

When the client sends something you can't use, such as `limit=abc`, reply
with status 400, `http.StatusBadRequest`, and a message saying what's wrong.
That way the mistake is the client's to fix, and it shows up clearly in their
logs instead of yours.

## Your task

`LogStore` maps service names to log lines, oldest first. The log viewer
answers `GET /` with a usage message and `GET /logs/{service}` with that
service's newest lines, one per line. Its routes and `ServeLogs` have bugs.

### 1. Fix the routes

In `NewMux`, change the patterns to `GET /{$}` for `Index` and
`GET /logs/{service}` for `ServeLogs`.

| Request | Status |
|---|---|
| `GET /` | 200 |
| `GET /logs/api` | 200 |
| `GET /nope` | 404 |
| `GET /logs` | 404 |
| `POST /logs/api` | 405 |
| `DELETE /logs/worker` | 405 |

### 2. Show the newest lines

`ServeLogs` shows at most `limit` lines (10 by default), but it keeps the
oldest ones. Fix the slicing so it keeps the **last** `limit` lines. A
service with fewer lines shows all of them.

```text
GET /logs/api     ("api" has 12 lines: "api line 1" ... "api line 12")
=> 200, body "api line 3\napi line 4\n...api line 12\n"
```

### 3. Read the limit parameter

When the request has a `limit` query parameter, use it instead of 10. If it
isn't a whole number from 1 to 100, reply with `http.Error`, status 400 and
the message `limit must be a number from 1 to 100`.

```text
GET /logs/api?limit=3
=> 200, body "api line 10\napi line 11\napi line 12\n"

GET /logs/api?limit=abc   (also ?limit=0 and ?limit=101)
=> 400, body "limit must be a number from 1 to 100\n"
```

## Learn more

- [Package net/http: ServeMux patterns](https://pkg.go.dev/net/http#hdr-Patterns-ServeMux)
- [Routing enhancements for Go 1.22 (Go blog)](https://go.dev/blog/routing-enhancements)
- [Package net/url: Values.Get](https://pkg.go.dev/net/url#Values.Get)
