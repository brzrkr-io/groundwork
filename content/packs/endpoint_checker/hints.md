## Task 1: Build a checker

Fill in both fields of the `Checker`: a client that gives up after `timeout`,
and the real clock.

### Closer

The client is `&http.Client{Timeout: timeout}`. For the clock, assign the
function `time.Now` itself, without the `()`.

### Answer

```go
return &Checker{
	Client: &http.Client{Timeout: timeout},
	Now:    time.Now,
}
```

## Task 2: Probe one URL

Send one GET request. If it fails there is no body, so return right away;
otherwise close the body and return the status code.

### Closer

`resp, err := c.Client.Get(url)`. On error return `0, err`. Otherwise
`defer resp.Body.Close()` and return `resp.StatusCode, nil`.

### Answer

```go
resp, err := c.Client.Get(url)
if err != nil {
	return 0, err
}
defer resp.Body.Close()
return resp.StatusCode, nil
```

## Task 3: Check one endpoint

Read the clock once before probing and once after, no matter how many
attempts. Retry only when the first probe worked and returned a 5xx.

### Closer

Keep `attempts := 1`, and inside `if err == nil && status >= 500` probe again
and set `attempts = 2`. `Latency` is `c.Now().Sub(start)`, and `OK` is
`err == nil && status == ep.Expect`.

### Answer

```go
start := c.Now()
status, err := c.Probe(ep.URL)
attempts := 1
if err == nil && status >= 500 {
	status, err = c.Probe(ep.URL)
	attempts = 2
}
return Result{
	Name:     ep.Name,
	Expect:   ep.Expect,
	Status:   status,
	Latency:  c.Now().Sub(start),
	Attempts: attempts,
	Err:      err,
	OK:       err == nil && status == ep.Expect,
}
```

## Task 4: Render the report

Loop over the results, work out the status text and the result text for each
row, and count the OK ones for the summary line.

### Closer

Start each row with `status := strconv.Itoa(r.Status)` and pick the result
text with `switch { case r.Err != nil: ... case r.OK: ... default: ... }`; set
`status = "-"` in the error case. Finish with
`fmt.Fprintf(&b, "%d of %d endpoints ok\n", ok, len(results))`. Add the
`"strconv"` import.

### Answer

```go
ok := 0
for _, r := range results {
	status := strconv.Itoa(r.Status)
	var verdict string
	switch {
	case r.Err != nil:
		status = "-"
		verdict = "error: " + r.Err.Error()
	case r.OK:
		verdict = "ok"
		ok++
	default:
		verdict = fmt.Sprintf("want %d", r.Expect)
	}
	fmt.Fprintf(&b, rowFormat, r.Name, status, r.Latency, verdict)
}
fmt.Fprintf(&b, "%d of %d endpoints ok\n", ok, len(results))
```

## Task 5: Run everything

Check every endpoint, collect the results, and remember whether any of them
was not OK.

### Closer

Start with `code := 0` and set it to `1` when `!r.OK`. After the loop,
`fmt.Fprint(out, Report(results))` and return `code`.

### Answer

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
