## Task 1: one field

`&http.Client{...}` is a struct literal. Set its `Timeout` field to
`5 * time.Second`.

## Task 2: build, then set headers

Call `http.NewRequestWithContext(ctx, http.MethodGet, url, nil)`, return early
on error, then call `req.Header.Set` once per header. Build the URL with `+`:
`baseURL + "/status/" + service`.

## Task 3: three small additions

Each TODO is one or two lines:

```go
defer resp.Body.Close()

if resp.StatusCode != http.StatusOK {
	return ServiceStatus{}, fmt.Errorf("unexpected status %d", resp.StatusCode)
}
```

For the limit, give the decoder `io.LimitReader(resp.Body, 1<<20)` instead of
`resp.Body`.

## Almost the answer for task 2

```go
req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/status/"+service, nil)
if err != nil {
	return nil, err
}
req.Header.Set("Authorization", "Bearer "+token)
req.Header.Set("Accept", "application/json")
return req, nil
```
