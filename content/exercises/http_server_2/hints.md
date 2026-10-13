## Task 1: method and end-of-path

Each pattern needs `GET ` in front. The root pattern also needs `{$}` at the
end, so it stops matching every other path:

```go
mux.HandleFunc("GET /{$}", Index)
```

## Task 2: slice from the end

For a slice of 12 lines and a limit of 10, the newest 10 start at index
`12 - 10 = 2`. In general that's `len(lines)-limit`, and `lines[start:]`
keeps everything from `start` to the end.

## Task 3: read and check the parameter

```go
if text := r.URL.Query().Get("limit"); text != "" {
	n, err := strconv.Atoi(text)
	// reply 400 and return if err != nil, n < 1 or n > 100
	limit = n
}
```

## Task 3: the 400 reply

```go
if err != nil || n < 1 || n > 100 {
	http.Error(w, "limit must be a number from 1 to 100", http.StatusBadRequest)
	return
}
```

Remember to add `"strconv"` to the imports.
