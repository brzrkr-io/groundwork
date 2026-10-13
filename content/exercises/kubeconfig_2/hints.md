## Task 1: Fix the context bug

`NewDefaultClientConfig` has no parameter for a context name, so it can only ever use `current-context`. Look for the other constructor in the README that takes a context name.

### Closer

`clientcmd.NewNonInteractiveClientConfig` takes the config value, the context name, the overrides and a last argument that can be `nil`:

```go
clientConfig := clientcmd.NewNonInteractiveClientConfig(*cfg, contextName, ...)
```

### Answer

```go
clientConfig := clientcmd.NewNonInteractiveClientConfig(*cfg, contextName, &clientcmd.ConfigOverrides{}, nil)
return clientConfig.ClientConfig()
```

Returning the error from `ClientConfig()` unchanged lets `clientcmd.IsContextNotFound` recognise it.

## Task 2: Fix the client settings

Compare each line with the README's list: two numbers are swapped, one number has no unit, and one setting is missing.

### Closer

A `time.Duration` counts nanoseconds, so `30` alone is 30 nanoseconds; multiply by `time.Second` (add the import `"time"`). The missing line sets `c.UserAgent`.

### Answer

```go
c.QPS = 20
c.Burst = 40
c.Timeout = 30 * time.Second
c.UserAgent = userAgent
```

`30 * time.Second` is 30 seconds as a `time.Duration`.

## Task 3: Use the real namespace lookup

Throw away the hand-written map lookup. Build the same client config as in task 1 and ask it for the namespace instead.

### Closer

`clientConfig.Namespace()` returns three values, `(string, bool, error)`. You don't need the `bool`, so use `_`:

```go
ns, _, err := clientConfig.Namespace()
```

### Answer

```go
clientConfig := clientcmd.NewNonInteractiveClientConfig(*cfg, contextName, &clientcmd.ConfigOverrides{}, nil)
ns, _, err := clientConfig.Namespace()
return ns, err
```

`Namespace()` fills in `"default"` when the context sets none and returns an error for an unknown context.
