# Middleware: chains, status codes and auth

Your deploy API's middleware has two bugs and one missing piece: the stack
runs in the wrong order, the access log thinks every request succeeded, and
anyone can trigger a deploy.

**New here:** `slices.Backward`, wrapping a `ResponseWriter` with embedding,
`crypto/subtle`.

## Learn

### Recap

A middleware has the shape `func(http.Handler) http.Handler`. It does its
extra work and calls `next.ServeHTTP(w, r)` to run the wrapped handler. Not
calling `next` stops the request there.

When you wrap by hand, the last wrapper applied is the outermost, and the
outermost runs first.

### Looping backwards

`slices.Backward(s)` gives the same index and value pairs as `range s`, but
last first:

```go
for i, name := range slices.Backward([]string{"a", "b", "c"}) {
	fmt.Println(i, name) // 2 c, then 1 b, then 0 a
}
```

You'll also see the older loop `for i := len(s) - 1; i >= 0; i--` in
existing code.

### Remembering the status code

The status goes out through `w.WriteHeader(code)`, and the middleware never
sees it. To log it, hand the handler a different `ResponseWriter` that notes
the code and passes it on:

```go
type statusRecorder struct {
	http.ResponseWriter // embedded
	status int
}
```

As in *Struct embedding*, an embedded field's methods are *promoted*:
`statusRecorder` gets `Header`, `Write` and `WriteHeader` from the real
writer, so it is a `ResponseWriter` itself.

Declare your own method with the same name and it replaces the promoted one.
Inside it, you can still reach the real one through the field. Here is a
`Write` that counts bytes:

```go
func (s *countingWriter) Write(b []byte) (int, error) {
	s.bytes += len(b)
	return s.ResponseWriter.Write(b)
}
```

A handler that never calls `WriteHeader` gets 200 automatically, which is why
the starter creates the recorder with `status: http.StatusOK`.

> **Why is it written like this?** An embedded field has a type but no name
> (Go syntax); its field name is the type's name, `ResponseWriter`. The
> receiver must be a pointer (`*countingWriter`) so the change to `s.bytes`
> sticks. The receiver name `s` is the author's choice.

### Comparing secrets

A plain `==` stops at the first different byte, so it takes slightly longer
the more of a secret was guessed right. Attackers can measure that.
`subtle.ConstantTimeCompare(a, b)` from `crypto/subtle` always takes the same
time. It compares two `[]byte` and returns `1` when they're equal:

```go
if subtle.ConstantTimeCompare([]byte(typed), []byte(password)) == 1 {
	fmt.Println("welcome")
}
```

`[]byte(s)` converts a `string` to bytes.

## Your task

The deploy API is wrapped as
`Chain(mux, Recover(os.Stderr), AccessLog(os.Stdout), RequireToken(token))`.
Fix and finish the pieces.

### 1. Fix the chain order

**Goal:** `Chain(h, mws...)` must make the **first** middleware listed the
outermost, so it runs first.

With no middlewares, it returns `h` unchanged. Right now the order comes out
backwards. Wrap starting from the last middleware instead of the first, and
add the `slices` import.

```go
Chain(handler, recover, log, auth)
// a request runs: recover, then log, then auth, then handler
```

**Think it through:**
- Each `h = mw(h)` adds a layer outside the others. Which one should be added
  last?
- Only the loop header needs to change.

### 2. Log the real status code

**Goal:** add a `WriteHeader(code int)` method to `*statusRecorder` so
`AccessLog` logs the status the handler really sent.

`AccessLog(out)` already writes `METHOD PATH STATUS` and a newline after each
request, but always logs 200. Your method saves `code` in the `status` field,
then calls the real `WriteHeader`, so the client still gets that status.

```text
handler writes only a body      logs "GET /healthz 200\n"
handler calls http.NotFound     logs "GET /missing 404\n"
handler calls WriteHeader(201)  logs "POST /deploys 201\n"
```

**Think it through:**
- Your method replaces the promoted one. How do you reach the real one?
- Which receiver type lets the saved status survive?

### 3. Require a token

**Goal:** `RequireToken(token)` lets a request through only with the header
`Authorization: Bearer <token>`; otherwise it answers 401.

The header must be exactly the fixed text `Bearer ` (with one space) followed
by the `token` parameter. Compare with `subtle.ConstantTimeCompare`, and add
the `crypto/subtle` import.

On a mismatch, **don't** call `next`. Instead:

- set `Content-Type` to `application/json`
- write status 401 (`http.StatusUnauthorized`)
- write the body `{"error":"unauthorized"}` followed by a newline

```text
token "s3cret", header "Authorization: Bearer s3cret"   // next runs
token "s3cret", header "Authorization: Bearer hunter2"  // 401
token "s3cret", no Authorization header                 // 401
token "s3cret", header "Authorization: s3cret"          // 401
```

**Think it through:**
- Which part of the expected value is fixed text, and which is the parameter?
- After writing the 401, what stops the handler from reaching `next`?

## Learn more

- [Effective Go: Embedding](https://go.dev/doc/effective_go#embedding)
- [net/http: ResponseWriter](https://pkg.go.dev/net/http#ResponseWriter)
- [crypto/subtle: ConstantTimeCompare](https://pkg.go.dev/crypto/subtle#ConstantTimeCompare)
