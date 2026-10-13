## Task 1: three lines, in order

`writeJSON` is three lines: set the header, call `w.WriteHeader(status)`, then
`json.NewEncoder(w).Encode(v)`. Remember to add `"encoding/json"` to the
imports. `writeError` is one line that calls `writeJSON`.

```go
writeJSON(w, status, map[string]string{"error": msg})
```

## Task 2: look up, then pick a response

```go
name := r.PathValue("name")
svc, err := a.store.Get(name)
if err != nil {
	writeError(w, http.StatusNotFound, "service not found: "+name)
	return
}
```

Then send `svc` with `writeJSON` and `http.StatusOK`.

## Task 3: decode first

Start with the snippet from the README's *Reading JSON safely* section. When
`Decode` fails, call `writeError` with `http.StatusBadRequest` and
`"invalid request body"`, then `return`.

## Task 3: the remaining checks

Each check has the same shape: if something failed, write an error and
`return`.

```go
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
