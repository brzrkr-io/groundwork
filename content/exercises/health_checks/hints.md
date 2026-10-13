## Task 1: one line

`Livez` only has to write `"ok\n"`. Use `fmt.Fprint(w, "ok\n")` and add
`"fmt"` to the import list. The status is 200 automatically.

## Task 2: the timeout context

Shadowing `ctx` keeps the code short:

```go
ctx, cancel := context.WithTimeout(ctx, timeout)
defer cancel()
err := c.Run(ctx)
```

Then return an `"ok"` or `"fail"` `Result` depending on `err`.

## Task 3: find any failure

Loop over `results`. If one has `Status != "ok"`, change `report.Status`
and remember a 503 in a `code` variable that starts as `http.StatusOK`.

## Task 3: almost the answer

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
