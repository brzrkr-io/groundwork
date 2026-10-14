## Task 1: Make Counter safe

Handle the negative case first and return early. For the real update, lock the
mutex the same way `Value` does.

### Closer

`if v < 0 { return }`, then `c.mu.Lock()` and `defer c.mu.Unlock()` before
`c.value += v`.

### Answer

```go
func (c *Counter) Add(v float64) {
	if v < 0 {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value += v
}
```

## Task 2: Write a counter family

Loop over the label values in sorted order and print one line each. A map's
own order is random, so sort the keys first.

### Closer

`slices.Sorted(maps.Keys(v.counters))` gives the label values in order, and
`v.counters[value].Value()` is the count. Print the count with
`strconv.FormatFloat(count, 'f', -1, 64)`. Write a literal `"` inside a Go
string as `\"`: `"%s{%s=\"%s\"} %s\n"`. Import `"maps"`, `"slices"` and
`"strconv"`.

### Answer

```go
for _, value := range slices.Sorted(maps.Keys(v.counters)) {
	count := v.counters[value].Value()
	fmt.Fprintf(w, "%s{%s=\"%s\"} %s\n", v.name, v.label, value, strconv.FormatFloat(count, 'f', -1, 64))
}
```

## Task 3: Serve /metrics

Set the header before writing any body, then write each family in order. An
`http.ResponseWriter` is an `io.Writer`, so it can be passed to `Write`.

### Closer

`w.Header().Set("Content-Type", "...")`, then
`for _, f := range families { ... }`.

### Answer

```go
w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
for _, f := range families {
	f.Write(w)
}
```
