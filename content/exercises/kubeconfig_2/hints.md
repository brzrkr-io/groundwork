## Task 1: pick the context

`NewDefaultClientConfig` has no parameter for a context name, so it can only
use `current-context`. Swap it for `NewNonInteractiveClientConfig`, which
takes the context name as its second argument. The last argument can be
`nil`.

## Task 2: units and swaps

Read each line next to the README: two numbers are swapped, and `30` on its
own is 30 nanoseconds. Multiply by `time.Second`, and don't forget
`c.UserAgent = userAgent`.

## Task 3: three return values

`Namespace()` returns `(string, bool, error)`. Ignore the `bool` with `_`:

```go
clientConfig := clientcmd.NewNonInteractiveClientConfig(*cfg, contextName, &clientcmd.ConfigOverrides{}, nil)
ns, _, err := clientConfig.Namespace()
return ns, err
```
