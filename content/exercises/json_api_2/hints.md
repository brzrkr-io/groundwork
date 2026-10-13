## Task 1: Fix the response helpers

`WriteHeader` sends the headers, so anything set after it is lost. Look at the
order of the lines in `writeJSON`. In `writeError`, `http.Error` always sends
plain text; reuse `writeJSON` instead.

### Closer

Move `w.Header().Set("Content-Type", "application/json")` above
`w.WriteHeader(status)`. Replace `http.Error(...)` with a `writeJSON` call that
passes a `map[string]string` with one `"error"` key.

### Answer

```go
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
```

## Task 2: Fix creating an alert

One bug is a missing line on the decoder; the other is a wrong status constant
at the very end of the handler.

### Closer

The decoder needs `dec.DisallowUnknownFields()` before `Decode`. A created
thing gets `http.StatusCreated`, not `http.StatusOK`.

### Answer

```go
dec := json.NewDecoder(r.Body)
dec.DisallowUnknownFields()
// ... decode, validate and create stay the same ...
writeJSON(w, http.StatusCreated, alert)
```

## Task 3: Finish listing alerts

Two things: turn a nil slice into an empty one so it encodes as `[]`, and sort
the slice by name before you send it.

### Closer

Check `if alerts == nil` and assign `[]Alert{}`. Sort with
`slices.SortFunc(alerts, func(x, y Alert) int { ... })` and compare with
`cmp.Compare(x.Name, y.Name)`. Add the `cmp` and `slices` imports.

### Answer

```go
alerts := a.store.All()
if alerts == nil {
	alerts = []Alert{}
}
slices.SortFunc(alerts, func(x, y Alert) int {
	return cmp.Compare(x.Name, y.Name)
})
writeJSON(w, http.StatusOK, alerts)
```
