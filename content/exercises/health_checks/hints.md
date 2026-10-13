## Task 1: Answer the liveness probe

Liveness only proves the process can answer HTTP, so `Livez` just writes a short body. You don't need to set the status yourself.

### Closer

`fmt.Fprint(w, ...)` writes text to the response. Add `"fmt"` to the imports. A handler that never calls `WriteHeader` sends 200 automatically.

### Answer

```go
fmt.Fprint(w, "ok\n")
```

## Task 2: Run one check with a timeout

Give the check its own context that expires after `timeout`, run it, and turn the error (or no error) into a `Result`.

### Closer

`context.WithTimeout(ctx, timeout)` returns a new context and a `cancel` function; `defer cancel()` right away. Shadowing `ctx` keeps the code short: `ctx, cancel := context.WithTimeout(ctx, timeout)`.

### Answer

```go
ctx, cancel := context.WithTimeout(ctx, timeout)
defer cancel()

if err := c.Run(ctx); err != nil {
	return Result{Name: c.Name, Status: "fail", Error: err.Error()}
}
return Result{Name: c.Name, Status: "ok"}
```

A check that hangs returns `ctx.Err()` once the deadline passes, so its error text is `context deadline exceeded`.

## Task 3: Serve the readiness report

Look through `results` for any check that failed. The status code and header must be set before the body is encoded.

### Closer

Keep a `code` variable that starts as `http.StatusOK` and becomes `http.StatusServiceUnavailable` when a result's `Status` isn't `"ok"`. Then `w.Header().Set("Content-Type", "application/json")` and `w.WriteHeader(code)`.

### Answer

```go
code := http.StatusOK
for _, res := range results {
	if res.Status != "ok" {
		report.Status = "fail"
		code = http.StatusServiceUnavailable
	}
}

w.Header().Set("Content-Type", "application/json")
w.WriteHeader(code)
json.NewEncoder(w).Encode(report)
```
