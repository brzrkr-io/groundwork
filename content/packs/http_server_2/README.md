# HTTP servers: fix the log viewer

When a service misbehaves at 3 a.m., the quickest look at its logs is often a
plain HTTP endpoint. Here you fix the routes and the query handling of a small
log viewer.

**New here:** `{$}` and method patterns, slicing from the end, query parameters, 400 Bad Request.

## Learn

### Recap

- A handler is `func(w http.ResponseWriter, r *http.Request)`: headers, then
  status (default 200), then body.
- `http.Error(w, msg, code)` sends an error status with `msg` plus a newline.
- A `ServeMux` routes by patterns like `"GET /services/{name}"`;
  `r.PathValue("name")` reads the wildcard.

### Patterns that end in a slash

A pattern ending in `/` matches that path **and everything below it**. So
`"/"` alone matches every request. To match only the exact path, end it with
`{$}`, meaning "the path ends here":

| Pattern | Matches `/admin/` | Matches `/admin/x` |
|---|---|---|
| `"/admin/"` | yes | yes |
| `"/admin/{$}"` | yes | no, the mux answers 404 |

A pattern with no method, like `"/users/{id}"`, accepts **every** method, so
`POST` and `DELETE` reach a handler meant only for reading. Put the method in
front, followed by one space.

> **Why is it written like this?** The pattern syntax `METHOD /path` with
> `{name}` and `{$}` is required by `ServeMux`; it's all inside one string.
> Uppercase method names are required too: `"get /x"` won't match GET.

### Slicing from the end

`s[a:b]` takes items from index `a` up to (not including) `b`. Leave one side
out to mean "from the start" or "to the end":

```go
nums := []int{10, 20, 30, 40, 50}
nums[:2]           // [10 20]       first 2
nums[1:]           // [20 30 40 50] skip the first
nums[len(nums)-1:] // [50]          from 1 before the end to the end
```

### Query parameters

The part of a URL after `?` is the *query string*: `/users?page=2`.
`r.URL.Query().Get("page")` returns the value as a string: `"2"`, or `""`
when it's missing. Convert with `strconv.Atoi`.

### 400 Bad Request

When the client sends something unusable, like `page=abc`, reply with status
400 (`http.StatusBadRequest`) and a message saying what's wrong. The mistake
is the client's to fix.

## Your task

`LogStore` maps service names to log lines, oldest first. The viewer answers
`GET /` with a usage message and `GET /logs/{service}` with that service's
newest lines, one per line. Its routes and `ServeLogs` have bugs.

### 1. Fix the routes

**Goal:** in `NewMux`, make `Index` answer only `GET /` and `ServeLogs` answer only `GET` on `/logs/{service}`.

Now `"/"` answers every path, and the logs pattern accepts any method.

| Request | Status |
|---|---|
| `GET /` | 200 |
| `GET /logs/api` | 200 |
| `GET /nope` | 404 |
| `GET /logs` | 404 |
| `POST /logs/api` | 405 |
| `DELETE /logs/worker` | 405 |

**Think it through:**
- Which pattern needs "the path ends here"?
- Which pieces go in front of each path, and is there a space?

### 2. Show the newest lines

**Goal:** `ServeLogs` keeps the **last** `limit` lines instead of the first.

`limit` is 10 by default. A service with fewer lines than `limit` shows all
of them. Only the slicing line is wrong.

```text
GET /logs/api     ("api" has 12 lines: "api line 1" ... "api line 12")
=> 200, body "api line 3\napi line 4\n...api line 12\n"

GET /logs/worker  ("worker" has 3 lines)
=> 200, all 3 lines
```

**Think it through:**
- With 12 lines and a limit of 10, at which index does the slice start?
- How do you write that index using `len(lines)` and `limit`?

### 3. Read the limit parameter

**Goal:** when the request has a `limit` query parameter, use it instead of 10, or reply 400 if it's bad.

- No `limit` parameter (`""`): keep 10.
- Not a whole number from 1 to 100: reply with `http.Error`, status 400, and
  the message `limit must be a number from 1 to 100`. Then stop.
- Add the `"strconv"` import.

```text
GET /logs/api?limit=3
=> 200, body "api line 10\napi line 11\napi line 12\n"

GET /logs/api?limit=100
=> 200, all 12 lines

GET /logs/api?limit=abc   (also ?limit=0 and ?limit=101)
=> 400, body "limit must be a number from 1 to 100\n"
```

**Think it through:**
- Three ways to be bad: not a number, too small, too big. One `if` can check all three.
- Change the existing `limit` variable; don't declare a new one.

## Learn more

- [Package net/http: ServeMux patterns](https://pkg.go.dev/net/http#hdr-Patterns-ServeMux)
- [Routing enhancements for Go 1.22 (Go blog)](https://go.dev/blog/routing-enhancements)
- [Package net/url: Values.Get](https://pkg.go.dev/net/url#Values.Get)
