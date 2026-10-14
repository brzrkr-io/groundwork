# Observability: trace IDs through logs and calls

When a request fails somewhere between the API, the deployer and the
registry, you need to find *which* request broke *where*, and *what* it
logged. One shared trace ID, passed from service to service and written on
every log line, makes that one search. Here you build it and pass it on.

**New here:** traces and spans, the W3C `traceparent` header,
`io.ReadFull`, `strings.Trim`, `context.WithValue` with an unexported key
type, `r.WithContext`.

## Learn

### Three signals, one ID

| Signal | Answers | Built in |
|---|---|---|
| Metrics | Is something wrong, since when? | *Prometheus-style metrics: counters and /metrics* |
| Traces | Which hop of which request failed or was slow? | this exercise |
| Logs | What exactly happened? | *Structured logging with slog* |

A *trace* is one request's whole journey through your services. Each hop
(a service handling it, or one call to another service) is a *span*. Every
span carries the same *trace ID*, and so does every log line:

```text
an alert fires on the error rate          metrics: something is wrong
 → open a failed request's trace          traces: api → deployer → registry (503)
 → search logs for trace_id="4bf92f..."   logs: every line from all three services
```

### The traceparent header

Services pass the trace on in one HTTP header, standardized by the W3C:

```text
traceparent: 00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01
             │  │                                │                └ flags: 2 hex characters, 01 = sampled
             │  │                                └ parent ID: 16 hex characters, the caller's span
             │  └ trace ID: 32 hex characters, the same on every hop
             └ version: 00
```

*Hex* uses only `0`–`9` and `a`–`f`; this header must be lowercase. A trace
ID or parent ID of only zeros is invalid. A service that receives a broken
header ignores it and starts a new trace, rather than trusting garbage.

### Random IDs from an injected reader

A 16-byte random ID printed as hex is 32 characters. Production reads random
bytes from `crypto/rand.Reader`, an `io.Reader`. A test can't predict random
bytes, so the functions here take the reader as a parameter: tests pass a
`bytes.NewReader` holding bytes they chose, like the injected clock in
*Prometheus-style metrics: histograms*.

`io.ReadFull(r, buf)` fills all of `buf` from `r`, or returns an error if
`r` runs out first:

```go
salt := make([]byte, 4)
if _, err := io.ReadFull(random, salt); err != nil {
	return "", err
}
fmt.Println(hex.EncodeToString(salt)) // 8 characters, such as "9f86d081"
```

`hex.EncodeToString` is the one from *Checksums and archives: verify and
extract safely*.

> **Why is it written like this?** A plain `random.Read(buf)` may legally
> fill only part of `buf`; `io.ReadFull` keeps reading until it's full, which
> is why it exists. It returns a count and an error; `_` drops the count,
> because a `nil` error already means "all of it". The name `salt` is the
> author's choice.

### Only zeros?

`strings.Trim(s, cutset)` removes every character in `cutset` from both ends
of `s`. If nothing is left, `s` was made of only those characters:

```go
strings.Trim("....", ".") == "" // true
strings.Trim("a..", ".") == ""  // false: "a" is left
```

### Values in a context

As in *Context and cancellation*, a context travels with a request down
through every function. `context.WithValue(parent, key, value)` returns a
child context that also carries one value. `ctx.Value(key)` gives it back as
`any`, so you take it out with a comma-ok type assertion, as in *Type
assertions and switches*:

```go
type tenantKey struct{}

ctx = context.WithValue(ctx, tenantKey{}, "team-payments")

tenant, ok := ctx.Value(tenantKey{}).(string) // "team-payments", true
```

A request carries its context: `r.Context()` reads it, and
`r.WithContext(ctx)` returns a **copy** of the request with the new context.
Middleware hands that copy to `next`, so the handler can read the value.

> **Why is it written like this?** Go accepts any comparable key, but the
> convention is a private type used for nothing else: no other package can
> make an equal key, so nobody overwrites your value by accident.
> `struct{}` is a type with no fields; `tenantKey{}` is its value. Put only
> request data (IDs, the caller) in a context, not settings a function
> should take as parameters.

### Passing the trace on

When your service calls another one, that call is a new span: same trace ID
and flags, a **new** parent ID. The next service sees your call as its
parent.

```text
client  ──00-4bf9…4736-00f067aa0ba902b7-01──▶  deploy API
deploy API  ──00-4bf9…4736-b7ad6b7169203331-01──▶  registry
              same trace ID   new parent ID
```

Real services don't hand-write this: the OpenTelemetry SDK creates spans,
reads and writes `traceparent`, and ships traces to a backend such as Jaeger
or Tempo. It speaks this same header.

## Your task

The deploy API has logs, but nothing links them across services. Make it
join traces, log the trace ID and pass the trace on. Work in `exercise.go`.

### 1. Make random IDs

**Goal:** finish `NewID(random, size)` so it reads exactly `size` bytes from
`random` and returns them as lowercase hex.

- Read with `io.ReadFull` into a new byte slice of length `size`.
- Success: return the hex text (`2*size` characters) and `nil`.
- The reader runs out: return `""` and the error.

Add the `"encoding/hex"` import. `NewTrace` is already written: it calls
your `NewID` for a 16-byte trace ID, then an 8-byte parent ID.

```go
random := bytes.NewReader([]byte{0x00, 0xf0, 0x67, 0xaa, 0x0b, 0xa9, 0x02, 0xb7, 0xff})
NewID(random, 8) // => "00f067aa0ba902b7", nil (0xff stays in the reader)
NewID(random, 8) // => "", error: only 1 byte left
```

**Think it through:**
- Which built-in function makes a byte slice of a given length?
- What do you return on the error path, and what on success?

### 2. Validate a traceparent header

**Goal:** finish `ParseTraceparent` so it also rejects bad trace IDs and
parent IDs.

The starter already splits the header on `-` and checks the number of parts,
the version and the flags. Add the same kind of check for the two IDs:

| Field | Length | Also invalid |
|---|---|---|
| `tc.TraceID` | 32 | characters other than `0`–`9`, `a`–`f`; only zeros |
| `tc.ParentID` | 16 | characters other than `0`–`9`, `a`–`f`; only zeros |

On any problem, return `TraceContext{}` and an error. The message is your
choice; say what's wrong. A valid header returns its `TraceContext` and `nil`.

```go
ParseTraceparent("00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")
// => TraceContext{TraceID: "4bf92f3577b34da6a3ce929d0e0e4736", ParentID: "00f067aa0ba902b7", Flags: "01"}, nil

ParseTraceparent("00-00000000000000000000000000000000-00f067aa0ba902b7-01") // => TraceContext{}, error: only zeros
ParseTraceparent("00-4BF92F3577B34DA6A3CE929D0E0E4736-00f067aa0ba902b7-01") // => TraceContext{}, error: uppercase
ParseTraceparent("00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902-01")   // => TraceContext{}, error: 12 characters
```

**Think it through:**
- Look at how the flags are checked. Which helper tests the characters?
- Which `strings` function tells you a string is only zeros?

### 3. Trace every request

**Goal:** finish the middleware `Trace(logger, random)` so every request
joins a trace, carries it in its context, returns its ID and logs it.

In order:

1. Parse the `traceparent` request header with `ParseTraceparent`. Missing
   or invalid: start a new trace with `NewTrace(random)`. If that fails,
   reply with `http.Error`, status 500, message `internal server error`, and
   return without calling `next`.
2. Set the response header `X-Trace-ID` to the trace ID.
3. Store the `TraceContext` in the request's context under the key
   `traceKey{}`, and call `next` with the request that carries it.
   `TraceFrom`, already written, reads it back.
4. After `next` returns, log one line with `logger.Info`: message `request`,
   then the attributes `method`, `path` (`r.URL.Path`) and `trace_id`.

```text
GET /deploys with traceparent: 00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01
// next handler: TraceFrom(r.Context()) => that TraceContext, true
// response header X-Trace-ID: 4bf92f3577b34da6a3ce929d0e0e4736
// log: {"time":"...","level":"INFO","msg":"request","method":"GET","path":"/deploys","trace_id":"4bf92f3577b34da6a3ce929d0e0e4736"}

POST /deploys without traceparent
// a new trace from NewTrace(random), flags "01", used the same way
```

A valid header must not read from `random` at all.

**Think it through:**
- `ParseTraceparent("")` returns an error. How many checks do you need for
  "missing or invalid"?
- Which request does `next` get: `r`, or the copy with the new context?

### 4. Pass the trace on

**Goal:** finish `Inject(req, random)` so an outgoing request carries the
same trace with a new parent ID.

- Read the `TraceContext` from `req.Context()` with `TraceFrom`. None:
  return an error (message your choice).
- Make a new 8-byte parent ID with `NewID`. On error, return it.
- Set the request header `traceparent`: same trace ID and flags, the new
  parent ID. `TraceContext`'s `String` method formats the header.
- Return `nil`.

If you use `errors.New`, add the `"errors"` import.

```go
// req's context holds TraceContext{"4bf92f3577b34da6a3ce929d0e0e4736", "00f067aa0ba902b7", "01"}
// random holds the bytes b7 ad 6b 71 69 20 33 31
Inject(req, random) // => nil
// req header traceparent: 00-4bf92f3577b34da6a3ce929d0e0e4736-b7ad6b7169203331-01
```

A handler behind `Trace` calls it before each request to another service:
`req, err := http.NewRequestWithContext(r.Context(), ...)`, then
`Inject(req, rand.Reader)`. The test sends your request to a real local
server (`httptest.NewServer`) and checks what arrived.

**Think it through:**
- `TraceFrom` returns a copy. Is it safe to change its `ParentID`?
- Which header name, and which method turns the struct into header text?

## Learn more

- [W3C Trace Context](https://www.w3.org/TR/trace-context/)
- [context.WithValue](https://pkg.go.dev/context#WithValue)
- [io.ReadFull](https://pkg.go.dev/io#ReadFull)
- [OpenTelemetry: Go](https://opentelemetry.io/docs/languages/go/)
