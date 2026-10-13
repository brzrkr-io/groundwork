## Task 1: return early, then lock

Check `v < 0` first and `return` if so. Then lock the mutex the same way
`Value` does (`c.mu.Lock()` and `defer c.mu.Unlock()`) before `c.value += v`.

## Task 2: loop over sorted keys

`slices.Sorted(maps.Keys(v.counters))` gives you the label values in order.
For each one, `v.counters[value].Value()` is its count. Print it with
`strconv.FormatFloat(count, 'f', -1, 64)`, not `%v`. To print a literal `"` inside a Go string, write `\"`.

## Task 3: header first, then the body

Headers must be set before anything is written to the body. Use
`w.Header().Set("Content-Type", "...")`, then loop over `families` and call
`f.Write(w)`: an `http.ResponseWriter` is an `io.Writer`.

## Near-solution

```go
// CounterVec.Write
for _, value := range slices.Sorted(maps.Keys(v.counters)) {
	count := v.counters[value].Value()
	fmt.Fprintf(w, "%s{%s=\"%s\"} %s\n", v.name, v.label, value, strconv.FormatFloat(count, 'f', -1, 64))
}

// Handler
w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
for _, f := range families {
	f.Write(w)
}
```
