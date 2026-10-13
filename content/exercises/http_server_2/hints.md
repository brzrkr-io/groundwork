## Task 1: Fix the routes

Both patterns need a method in front. The root pattern also needs to stop
matching every other path.

### Closer

Put `GET ` before each path, and end the root pattern with `{$}`, which means
"the path ends here": `"GET /{$}"`.

### Answer

```go
mux.HandleFunc("GET /{$}", Index)
mux.HandleFunc("GET /logs/{service}", s.ServeLogs)
```

## Task 2: Show the newest lines

The newest lines are at the end of the slice. For 12 lines and a limit of 10,
the newest 10 start at index 2.

### Closer

The start index is `len(lines)-limit`, and `lines[start:]` keeps everything
from `start` to the end.

### Answer

```go
if len(lines) > limit {
	lines = lines[len(lines)-limit:]
}
```

## Task 3: Read the limit parameter

Read the parameter as a string, and only change `limit` when it is not empty.
Convert it to a number, and reply 400 if the conversion fails or the number
is out of range.

### Closer

`r.URL.Query().Get("limit")` returns `""` when it is missing. Convert with
`n, err := strconv.Atoi(text)`, then check `err != nil || n < 1 || n > 100`
and call `http.Error` with `http.StatusBadRequest`. Add `"strconv"` to the
imports.

### Answer

```go
if text := r.URL.Query().Get("limit"); text != "" {
	n, err := strconv.Atoi(text)
	if err != nil || n < 1 || n > 100 {
		http.Error(w, "limit must be a number from 1 to 100", http.StatusBadRequest)
		return
	}
	limit = n
}
```
