# Middleware: chains, status codes and auth

Your deploy API's middleware has two bugs and one missing piece: the stack runs
in the wrong order, the access log thinks every request succeeded, and anyone
can trigger a deploy.

## Learn

### Recap

A middleware has the shape `func(http.Handler) http.Handler`. It returns an
`http.HandlerFunc` that does its extra work and calls `next.ServeHTTP(w, r)`
to run the wrapped handler. Not calling `next` stops the request right there.
When you wrap by hand, the last wrapper applied is the outermost, and the
outermost runs first.

### New: remembering the status code

The status code goes out through `w.WriteHeader(code)`, and the middleware
never sees it. To log it, give the handler a different `ResponseWriter`: one
that notes the code, then passes it on.

```go
type statusRecorder struct {
	http.ResponseWriter // embedded: Header and Write come along for free
	status int
}
```

As in the struct embedding exercises, the methods of an embedded field are
*promoted*: `statusRecorder` gets `Header`, `Write` and `WriteHeader` from the
real `ResponseWriter`, so it is a `ResponseWriter` itself. You can embed an
interface type like `http.ResponseWriter` exactly the same way.

If you declare your own method with the same name, it replaces the promoted
one. Inside it, `s.ResponseWriter.WriteHeader(code)` still reaches the real
one:

```go
func (s *statusRecorder) Write(b []byte) (int, error) {
	// ...do something extra...
	return s.ResponseWriter.Write(b)
}
```

A handler that never calls `WriteHeader` gets status 200 automatically. That's
why the recorder starts with `status: http.StatusOK`.

### New: comparing secrets

A plain `==` stops at the first byte that differs, so it takes slightly
longer the more of a secret you guessed right. Attackers can measure that.
`subtle.ConstantTimeCompare(a, b)` from `crypto/subtle` always takes the same
time. It compares two `[]byte` and returns `1` when they're equal:

```go
if subtle.ConstantTimeCompare([]byte(got), []byte(want)) != 1 {
	// not equal
}
```

A `string` converts to `[]byte` with `[]byte(s)`.

## Your task

The deploy API is wrapped as
`Chain(mux, Recover(os.Stderr), AccessLog(os.Stdout), RequireToken(token))`.
Fix and finish the pieces.

### 1. Fix the chain order

`Chain(h, mws...)` wraps `h` with every middleware, and **the first one listed
must be the outermost**, so it runs first. With no middlewares, it returns `h`
unchanged. Right now the order comes out backwards: wrap starting from the
last middleware instead of the first.

To loop over a slice from the end, range over `slices.Backward(s)` from the
`slices` package. It gives the same index and value pairs as `range s`, last
first:

```go
for i, name := range slices.Backward([]string{"a", "b", "c"}) {
	fmt.Println(i, name) // 2 c, then 1 b, then 0 a
}
```

You'll also see the older counting loop, `for i := len(s) - 1; i >= 0; i--`,
in existing code.

```go
Chain(handler, recover, log, auth)
// a request runs: recover, then log, then auth, then handler
```

### 2. Log the real status code

`AccessLog(out)` writes `METHOD PATH STATUS` and a newline to `out` after the
request, but it always logs 200. Add a `WriteHeader(code int)` method to
`*statusRecorder` that saves `code` in `status` and then calls the real
`WriteHeader`, so the client still gets the right status.

```text
handler writes only a body      logs "GET /healthz 200\n"
handler calls http.NotFound     logs "GET /missing 404\n"
handler calls WriteHeader(201)  logs "POST /deploys 201\n"
```

### 3. Require a token

`RequireToken(token)` lets a request through only when its `Authorization`
header is exactly `Bearer ` followed by `token`. Compare the two with
`subtle.ConstantTimeCompare`. Otherwise, without calling `next`:

- set `Content-Type` to `application/json`
- write status 401 (`http.StatusUnauthorized`)
- write the body `{"error":"unauthorized"}` followed by a newline

```text
token "s3cret", header "Authorization: Bearer s3cret"   // next runs
token "s3cret", header "Authorization: Bearer hunter2"  // 401
token "s3cret", no Authorization header                 // 401
token "s3cret", header "Authorization: s3cret"          // 401
```

## Learn more

- [Effective Go: Embedding](https://go.dev/doc/effective_go#embedding)
- [net/http: ResponseWriter](https://pkg.go.dev/net/http#ResponseWriter)
- [crypto/subtle: ConstantTimeCompare](https://pkg.go.dev/crypto/subtle#ConstantTimeCompare)
