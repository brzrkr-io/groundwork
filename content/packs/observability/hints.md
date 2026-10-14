## Task 1: Make random IDs

Make a byte slice of length `size`, fill it from `random`, and turn it into
hex. The only way this can fail is the reader running out.

### Closer

`b := make([]byte, size)`, then `io.ReadFull(random, b)` returns a count and
an error; you only need the error. `hex.EncodeToString(b)` gives the text.
Add `"encoding/hex"` to the imports.

### Answer

```go
b := make([]byte, size)
if _, err := io.ReadFull(random, b); err != nil {
	return "", err
}
return hex.EncodeToString(b), nil
```

## Task 2: Validate a traceparent header

Copy the flags check for `tc.TraceID` (length 32) and `tc.ParentID` (length
16), and add one more condition to each: the ID is not only zeros.

### Closer

Join the three conditions with `||`: wrong length, `!isLowerHex(...)`, or
`strings.Trim(id, "0") == ""`. When any is true, return `TraceContext{}` and
a `fmt.Errorf(...)` error.

### Answer

```go
if len(tc.TraceID) != 32 || !isLowerHex(tc.TraceID) || strings.Trim(tc.TraceID, "0") == "" {
	return TraceContext{}, fmt.Errorf("traceparent %q: trace ID must be 32 lowercase hex characters, not all zeros", header)
}
if len(tc.ParentID) != 16 || !isLowerHex(tc.ParentID) || strings.Trim(tc.ParentID, "0") == "" {
	return TraceContext{}, fmt.Errorf("traceparent %q: parent ID must be 16 lowercase hex characters, not all zeros", header)
}
```

## Task 3: Trace every request

Try the header first; only when `ParseTraceparent` fails do you call
`NewTrace`. Then set the response header, put the trace in a new context,
hand `next` a request carrying that context, and log once `next` is done.

### Closer

`tc, err := ParseTraceparent(r.Header.Get("traceparent"))`; on error,
`tc, err = NewTrace(random)`, and on a second error `http.Error(...)` and
`return`. Then `ctx := context.WithValue(r.Context(), traceKey{}, tc)` and
`next.ServeHTTP(w, r.WithContext(ctx))`. Log with
`logger.Info("request", "method", r.Method, ...)`.

### Answer

```go
tc, err := ParseTraceparent(r.Header.Get("traceparent"))
if err != nil {
	tc, err = NewTrace(random)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}
w.Header().Set("X-Trace-ID", tc.TraceID)
ctx := context.WithValue(r.Context(), traceKey{}, tc)
next.ServeHTTP(w, r.WithContext(ctx))
logger.Info("request", "method", r.Method, "path", r.URL.Path, "trace_id", tc.TraceID)
```

## Task 4: Pass the trace on

Get the trace with `TraceFrom`, swap its parent ID for a fresh one, and set
the header. Every step that can fail returns its error.

### Closer

`tc, ok := TraceFrom(req.Context())`; when `!ok`, return
`errors.New(...)`. A new parent ID is `NewID(random, 8)`. Assign it to
`tc.ParentID`, then `req.Header.Set("traceparent", tc.String())`.

### Answer

```go
tc, ok := TraceFrom(req.Context())
if !ok {
	return errors.New("no trace in the request's context")
}
parentID, err := NewID(random, 8)
if err != nil {
	return err
}
tc.ParentID = parentID
req.Header.Set("traceparent", tc.String())
return nil
```
