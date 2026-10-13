## Task 1: Fix the production host check

Look at the function the code calls. `HasPrefix` checks the *start* of a string, but the production domain is at the *end* of a host name.

### Closer

The `strings` package has a twin of `HasPrefix` that checks the end. Its name has "suffix" in it.

### Answer

```go
return strings.HasSuffix(host, ".prod.example.com")
```

## Task 2: Make a valid metric name

Two steps: change every `-` into `_`, then add the fixed ending. Which `strings` function replaces text?

### Closer

`strings.ReplaceAll(service, "-", "_")` swaps every dash. The ending is fixed text, so it goes in quotes: `+ "_requests_total"`.

### Answer

```go
return strings.ReplaceAll(service, "-", "_") + "_requests_total"
```

## Task 3: Catch every timeout

`Contains` is exact, so `Timeout` doesn't match `timeout`. The word you search for is already lowercase. What if the line were lowercase too?

### Closer

Lowercase the line before you search it: pass `strings.ToLower(line)` to `strings.Contains` instead of `line`.

### Answer

```go
return strings.Contains(strings.ToLower(line), "timeout")
```
