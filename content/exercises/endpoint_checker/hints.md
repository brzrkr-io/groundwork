## Tasks 1 and 2: client and probe

`NewChecker` returns `&Checker{Client: &http.Client{Timeout: timeout}, Now: time.Now}`.

In `Probe`, call `c.Client.Get(url)`. On error return `0, err`. Otherwise
`defer resp.Body.Close()` and return `resp.StatusCode, nil`.

## Task 3: the retry

```go
start := c.Now()
status, err := c.Probe(ep.URL)
attempts := 1
if err == nil && status >= 500 {
	status, err = c.Probe(ep.URL)
	attempts = 2
}
```

Then build the `Result`, with `Latency: c.Now().Sub(start)` and
`OK: err == nil && status == ep.Expect`.

## Task 4: one row

Inside `for _, r := range results`, start with
`status := strconv.Itoa(r.Status)` and pick the result text with a
`switch { case r.Err != nil: ... case r.OK: ... default: ... }`. In the error
case, set `status = "-"`. Count the OK results for the summary line:

```go
fmt.Fprintf(&b, "%d of %d endpoints ok\n", ok, len(results))
```

## Task 5: the exit code

```go
code := 0
var results []Result
for _, ep := range endpoints {
	r := c.Check(ep)
	if !r.OK {
		code = 1
	}
	results = append(results, r)
}
fmt.Fprint(out, Report(results))
return code
```
