## Read TestGrader's message

Each failing `TestGrader` subtest names a bug your tests missed, such as
"your tests didn't notice when the handler answered 200 OK for an unknown
host". Ask yourself which check would fail if the code did that, then add it.

## Task 1: two more rows

Each row follows the same order as the struct: name, path, status, body.

```go
{"another known host", "/hosts/db-1", http.StatusOK, "10.0.1.7\n"},
{"unknown host", "/hosts/db-9", http.StatusNotFound, "unknown host: db-9\n"},
```

## Task 2: three checks

```go
if rec.Code != tt.wantCode {
	t.Errorf("GET %s: status = %d, want %d", tt.path, rec.Code, tt.wantCode)
}
if got := rec.Body.String(); got != tt.wantBody {
	t.Errorf("GET %s: body = %q, want %q", tt.path, got, tt.wantBody)
}
```

The `Content-Type` check looks the same, using
`rec.Header().Get("Content-Type")` and `"text/plain; charset=utf-8"`.

## Task 3: the client test

Copy the fake server from Learn into `TestFetchIP`, then:

```go
ip, err := FetchIP(srv.Client(), srv.URL, "web-1")
if err != nil {
	t.Fatalf("FetchIP(%q) returned error: %v", "web-1", err)
}
if ip != "10.0.0.5" {
	t.Errorf("FetchIP(%q) = %q, want %q", "web-1", ip, "10.0.0.5")
}
```

Finish with an `err == nil` check for `"db-9"`.
