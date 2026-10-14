## Task 1: Build a client with a timeout

`&http.Client{...}` is a struct literal, and the client has a field for the
timeout. Fill it in instead of returning an empty client.

### Closer

Set the `Timeout` field inside the braces: `&http.Client{Timeout: ...}`. A
duration of five seconds is `5 * time.Second`.

### Answer

```go
return &http.Client{Timeout: 5 * time.Second}
```

## Task 2: Build the status request

Build the request first and return early if that fails, then add the two
headers. The URL is plain string joining.

### Closer

Use `http.NewRequestWithContext(ctx, http.MethodGet, url, nil)` with
`baseURL + "/status/" + service` as the URL. Then call
`req.Header.Set(name, value)` once per header.

### Answer

```go
req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/status/"+service, nil)
if err != nil {
	return nil, err
}
req.Header.Set("Authorization", "Bearer "+token)
req.Header.Set("Accept", "application/json")
return req, nil
```

## Task 3: Finish `FetchStatus`

Each TODO is one or two lines: close the body when the function returns,
check the status code yourself, and stop the decoder from reading more than
1 MiB.

### Closer

Use `defer resp.Body.Close()`, compare `resp.StatusCode` with
`http.StatusOK`, and give the decoder `io.LimitReader(resp.Body, 1<<20)`
instead of `resp.Body`.

### Answer

```go
defer resp.Body.Close()

if resp.StatusCode != http.StatusOK {
	return ServiceStatus{}, fmt.Errorf("unexpected status %d", resp.StatusCode)
}

var status ServiceStatus
if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&status); err != nil {
	return ServiceStatus{}, err
}
```
