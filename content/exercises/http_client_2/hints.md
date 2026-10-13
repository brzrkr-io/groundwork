## Task 1: close, then check

Right after the `err` check for `client.Get`, add `defer resp.Body.Close()`.
Then, before reading the body, check `resp.StatusCode != http.StatusOK` and
return the error with `fmt.Errorf`.

## Task 2: marshal first

Call `json.Marshal(ev)` at the top and return early on error. Pass
`bytes.NewReader(body)` instead of `nil` as the last argument of
`http.NewRequestWithContext`, then call
`req.Header.Set("Content-Type", "application/json")`.

## Task 3: a range check

"Not 2xx" means "below 200 or above 299":

```go
if resp.StatusCode < 200 || resp.StatusCode > 299 {
	return fmt.Errorf("notify: unexpected status %d", resp.StatusCode)
}
```
