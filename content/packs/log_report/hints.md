## Task 1: Parse one line

The starter's `group` helper gives you each captured piece as a string:
`group("status")` is `"200"`, `group("path")` is `"/api/orders?page=2"`.
Copy the shape of the time step (call, check `err`, keep the value) for each
of the other conversions.

### Closer

Use `fmt.Errorf("%w: %q", ErrMalformed, line)` for the no-match case.
Convert with `strconv.Atoi(group("status"))`, `strconv.Atoi(group("bytes"))`
and `time.ParseDuration(group("duration") + "s")`. Drop the query string with
`path, _, _ := strings.Cut(group("path"), "?")`. Add `"fmt"`, `"strconv"` and
`"strings"` to the imports.

### Answer

```go
status, err := strconv.Atoi(group("status"))
if err != nil {
	return Entry{}, fmt.Errorf("bad status %q: %w", group("status"), err)
}
bytes, err := strconv.Atoi(group("bytes"))
if err != nil {
	return Entry{}, fmt.Errorf("bad bytes %q: %w", group("bytes"), err)
}
duration, err := time.ParseDuration(group("duration") + "s")
if err != nil {
	return Entry{}, fmt.Errorf("bad duration %q: %w", group("duration"), err)
}
path, _, _ := strings.Cut(group("path"), "?")

return Entry{
	IP:       group("ip"),
	Time:     t,
	Method:   group("method"),
	Path:     path,
	Status:   status,
	Bytes:    bytes,
	Duration: duration,
}, nil
```

Also wrap the time error, e.g. `fmt.Errorf("bad time %q: %w", group("time"), err)`.

## Task 2: Read a whole log

This is the loop from *Reading input line by line*, with a counter. A bad
line is recorded and skipped, not returned. Add `"bufio"` to the imports.

### Closer

Inside `for sc.Scan()`, do `n++` first, trim `sc.Text()`, and `continue` on
`""`. On a `ParseLine` error, append `fmt.Errorf("line %d: %w", n, err)` to
`log.Problems` and `continue`. After the loop, check `sc.Err()`.

### Answer

```go
var log Log
sc := bufio.NewScanner(r)
n := 0
for sc.Scan() {
	n++
	line := strings.TrimSpace(sc.Text())
	if line == "" {
		continue
	}
	entry, err := ParseLine(line)
	if err != nil {
		log.Problems = append(log.Problems, fmt.Errorf("line %d: %w", n, err))
		continue
	}
	log.Entries = append(log.Entries, entry)
}
if err := sc.Err(); err != nil {
	return Log{}, err
}
return log, nil
```

## Task 3: Count status classes

Integer division drops the remainder, so dividing the status by 100 leaves
just its first digit. Turn that into a string key and count it in a map.

### Closer

`fmt.Sprintf("%dxx", e.Status/100)` turns 404 into `"4xx"`. Start with
`map[string]int{}` (not `nil`) and do `counts[key]++`. Add `"fmt"` to the
imports.

### Answer

```go
counts := map[string]int{}
for _, e := range entries {
	counts[fmt.Sprintf("%dxx", e.Status/100)]++
}
return counts
```

## Task 4: Find the busiest paths

Count into a map first, then copy the map into a `[]PathCount` so you can
sort it. The sort compares counts first and paths only when the counts are
equal.

### Closer

In `slices.SortFunc(top, func(a, b PathCount) int { ... })`, return
`cmp.Compare(b.Count, a.Count)` when the counts differ (b first means biggest
first), otherwise `cmp.Compare(a.Path, b.Path)`. Then cut to `top[:n]` if it's
longer. Add `"cmp"` and `"slices"` to the imports.

### Answer

```go
counts := map[string]int{}
for _, e := range entries {
	counts[e.Path]++
}

top := make([]PathCount, 0, len(counts))
for path, count := range counts {
	top = append(top, PathCount{Path: path, Count: count})
}
slices.SortFunc(top, func(a, b PathCount) int {
	if a.Count != b.Count {
		return cmp.Compare(b.Count, a.Count)
	}
	return cmp.Compare(a.Path, b.Path)
})

if len(top) > n {
	top = top[:n]
}
return top
```

## Task 5: Compute a percentile

Sort a copy, then work out the rank: `p` percent of the length, rounded up.
Ranks start at 1, slice indexes at 0.

### Closer

`sorted := slices.Clone(durations)` then `slices.Sort(sorted)`. You can round
up with whole numbers: `(p*len(sorted) + 99) / 100`. Return
`sorted[rank-1]`, and handle the empty slice first.

### Answer

```go
if len(durations) == 0 {
	return 0
}
sorted := slices.Clone(durations)
slices.Sort(sorted)

rank := (p*len(sorted) + 99) / 100
rank = max(rank, 1)
return sorted[rank-1]
```

## Task 6: Render the reports

For the text report, write each line with `fmt.Fprintf`; a map's keys must be
sorted first. For JSON, keep the encoder in a variable so you can set the
indent before encoding.

### Closer

`slices.Sorted(maps.Keys(r.ByClass))` gives the class names in order. You can
build everything in a `strings.Builder` and write it once with
`io.WriteString(w, b.String())`, so there's one error to return. For JSON,
call `enc.SetIndent("", "  ")` and fill every `jsonReport` field. Add
`"fmt"`, `"maps"`, `"slices"` and `"strings"` to the imports.

### Answer

```go
var b strings.Builder
fmt.Fprintf(&b, "requests: %d\n", r.Requests)
fmt.Fprintf(&b, "skipped: %d\n", r.Skipped)
b.WriteString("status classes:\n")
for _, class := range slices.Sorted(maps.Keys(r.ByClass)) {
	fmt.Fprintf(&b, "  %s: %d\n", class, r.ByClass[class])
}
b.WriteString("top paths:\n")
for _, pc := range r.TopPaths {
	fmt.Fprintf(&b, "  %s: %d\n", pc.Path, pc.Count)
}
fmt.Fprintf(&b, "latency: p50 %v, p95 %v\n", r.P50, r.P95)

_, err := io.WriteString(w, b.String())
return err
```

```go
enc := json.NewEncoder(w)
enc.SetIndent("", "  ")
return enc.Encode(jsonReport{
	Requests: r.Requests,
	Skipped:  r.Skipped,
	ByClass:  r.ByClass,
	TopPaths: r.TopPaths,
	P50Ms:    r.P50.Milliseconds(),
	P95Ms:    r.P95.Milliseconds(),
})
```

## Task 7: Test the whole pipeline with a golden file

The test has three parts: produce the real report for `testdata/access.log`
(that's `got`), read `testdata/report.golden` (that's `want`), and compare
the two as text.

### Closer

`os.ReadFile(filepath.Join("testdata", "access.log"))` gives bytes;
`ReadLog(strings.NewReader(string(data)))` parses them. Render with
`WriteText(&got, Summarize(parsed))` into a `var got strings.Builder`, then
compare `got.String()` with `string(want)`.

### Answer

```go
data, err := os.ReadFile(filepath.Join("testdata", "access.log"))
if err != nil {
	t.Fatal(err)
}
parsed, err := ReadLog(strings.NewReader(string(data)))
if err != nil {
	t.Fatalf("ReadLog(access.log) returned error %v", err)
}
var got strings.Builder
if err := WriteText(&got, Summarize(parsed)); err != nil {
	t.Fatalf("WriteText returned error %v", err)
}

want, err := os.ReadFile(filepath.Join("testdata", "report.golden"))
if err != nil {
	t.Fatal(err)
}
if got.String() != string(want) {
	t.Errorf("report for access.log =\n%s\nwant (report.golden):\n%s", got.String(), want)
}
```
