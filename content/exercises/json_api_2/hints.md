## Task 1: order, then format

In `writeJSON`, swap two lines so the header is set before `WriteHeader`.
For `writeError`, replace `http.Error` with a call to `writeJSON`, passing a
`map[string]string` with one `"error"` key.

## Task 2: one added line, one changed constant

After creating the decoder, call `dec.DisallowUnknownFields()`. At the end of
the handler, `http.StatusOK` should be `http.StatusCreated`.

## Task 3: nil check, then sort

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
