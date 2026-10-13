## Task 1: Fix `GetVersion`

Two lines are missing right after the `err` check for `client.Get`: one that
closes the body whenever the function returns, and a status check before the
body is read.

### Closer

Add `defer resp.Body.Close()`, then compare `resp.StatusCode` with
`http.StatusOK` and return `""` with an error from
`fmt.Errorf("version: unexpected status %d", ...)`.

### Answer

```go
defer resp.Body.Close()

if resp.StatusCode != http.StatusOK {
	return "", fmt.Errorf("version: unexpected status %d", resp.StatusCode)
}
```

## Task 2: Finish `NewDeployRequest`

Turn `ev` into JSON bytes first, then pass those bytes as the request body
instead of `nil`. Finally tell the server what kind of body it is.

### Closer

`body, err := json.Marshal(ev)` gives the bytes. Wrap them with
`bytes.NewReader(body)` as the last argument of `http.NewRequestWithContext`,
then call `req.Header.Set("Content-Type", ...)`.

### Answer

```go
body, err := json.Marshal(ev)
if err != nil {
	return nil, err
}
req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
if err != nil {
	return nil, err
}
req.Header.Set("Content-Type", "application/json")
return req, nil
```

## Task 3: Accept every 2xx in `NotifyDeploy`

Instead of comparing with one exact code, check whether the code falls
outside the range 200 to 299.

### Closer

"Not 2xx" means "below 200 or above 299": combine two comparisons on
`resp.StatusCode` with `||`.

### Answer

```go
if resp.StatusCode < 200 || resp.StatusCode > 299 {
	return fmt.Errorf("notify: unexpected status %d", resp.StatusCode)
}
```
