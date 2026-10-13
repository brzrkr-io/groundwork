## Task 1: Write the response helpers

Order matters: headers first, then the status code, then the body. `writeError`
doesn't need its own logic; it builds a small value and hands it to `writeJSON`.

### Closer

`writeJSON` is three lines: `w.Header().Set(...)`, `w.WriteHeader(status)`, then
`json.NewEncoder(w).Encode(v)`. Add `"encoding/json"` to the imports. For
`writeError`, a `map[string]string{"error": msg}` encodes as `{"error":"..."}`.

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

## Task 2: Get one service

Read the name from the path, ask the store for it, and pick a response. Any
error from `Get` means "not found".

### Closer

`r.PathValue("name")` gives the name, and `a.store.Get(name)` returns
`(Service, error)`. On error, call `writeError` with `http.StatusNotFound` and
`return`.

### Answer

```go
name := r.PathValue("name")
svc, err := a.store.Get(name)
if err != nil {
	writeError(w, http.StatusNotFound, "service not found: "+name)
	return
}
writeJSON(w, http.StatusOK, svc)
```

## Task 3: Create a service

Start with the snippet from the README's *Reading JSON safely* section. Then
each check has the same shape: if something failed, write an error and
`return`.

### Closer

Wrap the body with `http.MaxBytesReader(w, r.Body, 1<<20)`, make a decoder,
call `dec.DisallowUnknownFields()`, then `dec.Decode(&svc)`. After that come
`svc.Validate()` (400 with `err.Error()`) and `a.store.Create(svc)` (409).

### Answer

```go
r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
dec := json.NewDecoder(r.Body)
dec.DisallowUnknownFields()

var svc Service
if err := dec.Decode(&svc); err != nil {
	writeError(w, http.StatusBadRequest, "invalid request body")
	return
}
if err := svc.Validate(); err != nil {
	writeError(w, http.StatusBadRequest, err.Error())
	return
}
if err := a.store.Create(svc); err != nil {
	writeError(w, http.StatusConflict, "service already exists: "+svc.Name)
	return
}
writeJSON(w, http.StatusCreated, svc)
```
