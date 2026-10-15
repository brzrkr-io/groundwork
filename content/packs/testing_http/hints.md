## Task 1: Add the missing cases

Each row in the table follows the order of the struct fields: name, path,
status code, body. Remember that the handler's body ends with a newline.

### Closer

A row looks like `{"name", "/hosts/...", http.StatusOK, "...\n"}`. Use
`http.StatusNotFound` for the unknown host, whose body is
`"unknown host: <name>\n"`.

### Answer

```go
{"another known host", "/hosts/db-1", http.StatusOK, "10.0.1.7\n"},
{"unknown host", "/hosts/db-9", http.StatusNotFound, "unknown host: db-9\n"},
```

## Task 2: Check the response

After `h.ServeHTTP(rec, req)`, the recorder holds the status, headers and
body. Write one `if` per check, each ending in `t.Errorf`.

### Closer

Compare `rec.Code` with `tt.wantCode`, `rec.Body.String()` with
`tt.wantBody`, and `rec.Header().Get("Content-Type")` with
`"text/plain; charset=utf-8"`:

```go
if rec.Code != tt.wantCode {
	t.Errorf("GET %s: status = %d, want %d", tt.path, rec.Code, tt.wantCode)
}
```

### Answer

```go
if rec.Code != tt.wantCode {
	t.Errorf("GET %s: status = %d, want %d", tt.path, rec.Code, tt.wantCode)
}
wantType := "text/plain; charset=utf-8"
if got := rec.Header().Get("Content-Type"); got != wantType {
	t.Errorf("GET %s: Content-Type = %q, want %q", tt.path, got, wantType)
}
if got := rec.Body.String(); got != tt.wantBody {
	t.Errorf("GET %s: body = %q, want %q", tt.path, got, tt.wantBody)
}
```

## Task 3: Test the client

Start with the fake server from Learn, then call `FetchIP` twice: once for a
host the fake knows and once for one it doesn't. If `TestGrader` names a bug
your tests missed, ask which check would catch that bug.

### Closer

After `srv := httptest.NewServer(mux)` and `defer srv.Close()`, call
`FetchIP(srv.Client(), srv.URL, "web-1")`. Stop with `t.Fatalf` on an error,
then compare the IP with `"10.0.0.5"`. For `"db-9"`, check `err == nil`.

### Answer

```go
func TestFetchIP(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /hosts/web-1", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "10.0.0.5\n")
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	ip, err := FetchIP(srv.Client(), srv.URL, "web-1")
	if err != nil {
		t.Fatalf("FetchIP(%q) returned error: %v", "web-1", err)
	}
	if ip != "10.0.0.5" {
		t.Errorf("FetchIP(%q) = %q, want %q", "web-1", ip, "10.0.0.5")
	}

	if _, err := FetchIP(srv.Client(), srv.URL, "db-9"); err == nil {
		t.Errorf("FetchIP(%q) returned no error, want an error", "db-9")
	}
}
```
