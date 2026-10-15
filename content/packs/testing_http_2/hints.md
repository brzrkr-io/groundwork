## Task 1: Finish the happy-path test

Two more checks: one header and the decoded body. The `json.Unmarshal`
example in Learn shows the pattern; decode into a `DeployResponse` instead of
a `User`.

### Closer

Read the header with `rec.Header().Get("Content-Type")`. Decode with
`json.Unmarshal(rec.Body.Bytes(), &got)`, then compare `got` with a `want`
struct using `!=`.

### Answer

```go
if got := rec.Header().Get("Content-Type"); got != "application/json" {
	t.Errorf("POST /deploys: Content-Type = %q, want %q", got, "application/json")
}
var got DeployResponse
if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
	t.Fatalf("POST /deploys: body %q is not valid JSON: %v", rec.Body.String(), err)
}
want := DeployResponse{Service: "api", Version: "v1.2.0", Status: "queued"}
if got != want {
	t.Errorf("POST /deploys: response = %+v, want %+v", got, want)
}
```

## Task 2: Write the regression test

Two bodies need the same three checks, so a small table with a loop keeps it
short. This test is supposed to fail until you fix the bug in task 3.

### Closer

Make a table of `struct{ name, body string }`, loop with `t.Run`, and call
`rec := postDeploy(tt.body)`. Check `rec.Code` against
`http.StatusBadRequest`, then unmarshal into an `ErrorResponse` and check its
`Error` field.

### Answer

```go
tests := []struct {
	name string
	body string
}{
	{"empty service", `{"service": "", "version": "v1.2.0"}`},
	{"missing service", `{"version": "v1.2.0"}`},
}
for _, tt := range tests {
	t.Run(tt.name, func(t *testing.T) {
		rec := postDeploy(tt.body)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("POST /deploys %s: status = %d, want %d", tt.body, rec.Code, http.StatusBadRequest)
		}
		var got ErrorResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("POST /deploys %s: body %q is not valid JSON: %v", tt.body, rec.Body.String(), err)
		}
		if want := "service is required"; got.Error != want {
			t.Errorf("POST /deploys %s: error = %q, want %q", tt.body, got.Error, want)
		}
	})
}
```

## Task 3: Fix the bug

The version check right below the TODO already shows the pattern. Do the same
for the service.

### Closer

Check `req.Service == ""`, answer with `writeJSON(w, http.StatusBadRequest, ...)`
and an `ErrorResponse`, then `return`.

### Answer

```go
if req.Service == "" {
	writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "service is required"})
	return
}
```
