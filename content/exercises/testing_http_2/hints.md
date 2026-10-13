## Red, then green

Write the tests first (tasks 1 and 2), then fix the handler (task 3). A
failing regression test before the fix is expected. If `TestGrader` names a
mutant your tests missed, ask which check would fail if the code did that.

## Task 1: header and body

```go
if got := rec.Header().Get("Content-Type"); got != "application/json" {
	t.Errorf("POST /deploys: Content-Type = %q, want %q", got, "application/json")
}
```

Then copy the `json.Unmarshal` example from Learn: it decodes into a
`DeployResponse` and compares it with `want`.

## Task 2: a small table

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
		// check rec.Code, then json.Unmarshal into an ErrorResponse
		// and check its Error field
	})
}
```

## Task 3: the fix

```go
if req.Service == "" {
	writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "service is required"})
	return
}
```
