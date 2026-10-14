# kubeconfig: building a rest.Config

Knowing the server URL isn't enough to make a request. Here you turn a
kubeconfig context into the connection settings client-go actually uses, and
tune them the way a well-behaved ops tool should.

**New here:** `rest.Config`, `clientcmd.NewNonInteractiveClientConfig`.

## Learn

### Recap

A kubeconfig lists clusters (server URL and CA), users (credentials) and
contexts (cluster + user + optional namespace), plus a `current-context`.
`clientcmd.Load` / `clientcmd.LoadFromFile` parse it into a
`*clientcmdapi.Config` with the maps `Contexts`, `Clusters`, `AuthInfos`.

### From a context to a `*rest.Config`

Every client-go client is built from a `*rest.Config` (package
`k8s.io/client-go/rest`). It holds everything needed to reach one cluster:
`Host` (server URL), `BearerToken` or certificates, the CA, and client
settings.

`clientcmd` builds one in two steps. First a *client config* for a
context, then `ClientConfig()` on it:

| Constructor | Which context it uses |
|---|---|
| `clientcmd.NewDefaultClientConfig(config, overrides)` | always `current-context` |
| `clientcmd.NewNonInteractiveClientConfig(config, contextName, overrides, configAccess)` | the one you name |

```go
cc := clientcmd.NewNonInteractiveClientConfig(
	*loaded,                      // the config value, not the pointer
	"ci-runner",                  // the context to use
	&clientcmd.ConfigOverrides{}, // empty: no command-line overrides
	nil,                          // only needed to write refreshed tokens back
)
rc, err := cc.ClientConfig()
```

This **does not contact the cluster**. It only joins the context, cluster
and user entries, so it works offline and in tests.

> **Why is it written like this?**
> - **Go requires** `*loaded`: the function wants a `clientcmdapi.Config`
>   value, and `*` turns your pointer into the value it points to.
> - **Go requires** all four arguments in this order; there are no optional
>   arguments. `nil` and an empty `&ConfigOverrides{}` mean "nothing".
> - **Author's choice:** one argument per line with a comment is just for
>   reading; a single line compiles the same.

If the context doesn't exist, `ClientConfig()` returns an error for which
`clientcmd.IsContextNotFound(err)` is `true`. Its text looks like:

```text
invalid configuration: [context was not found for specified context: staging, cluster has no server defined]
```

### Client settings worth setting

| Field | Type | Why |
|---|---|---|
| `QPS` | `float32` | client-go limits *itself* to this many requests per second, so one busy tool can't overload the shared API server. Default 5. |
| `Burst` | `int` | Extra requests allowed in a short spike above `QPS`. Default 10. |
| `UserAgent` | `string` | Shown in the API server's audit logs, so the platform team sees which tool made a request. |
| `Timeout` | `time.Duration` | The longest one request may take. Zero means no limit. |

A `time.Duration` counts **nanoseconds**. A bare number like `5` means 5
nanoseconds. Multiply by a unit from package `time`:

```go
wait := 2 * time.Minute   // 2m0s
tick := 500 * time.Millisecond
```

### The default namespace

A client config also has a `Namespace()` method with **three** results:

| Result | Type | Meaning |
|---|---|---|
| 1st | `string` | the context's namespace, or `"default"` if it sets none |
| 2nd | `bool` | whether an override set the namespace |
| 3rd | `error` | e.g. context not found |

Use `_` for a result you don't need:

```go
_, port, err := net.SplitHostPort("example.com:443") // skip the host
```

> **Why is it written like this?** **Go requires** you to receive every
> result of a multi-result call. `_` is the blank identifier: it takes a
> value and throws it away.

## Your task

The release team's deploy bot connects to whichever cluster the operator
picks. The tests use `testdata/kubeconfig.yaml`, with contexts `dev`, `ci`
and `prod` (current context: `dev`).

### 1. Fix the context bug

**Goal:** make `RESTConfig(cfg, contextName)` use the context named `contextName`, not the current context.

- Replace `NewDefaultClientConfig` with `NewNonInteractiveClientConfig`.
- Return the result of `ClientConfig()` unchanged, including its error
  (tests check it with `clientcmd.IsContextNotFound`).

```go
c, _ := RESTConfig(cfg, "prod")
// c.Host        => "https://prod-us-east-1.k8s.example.com:6443"
// c.BearerToken => "fake-token-release-bot"
RESTConfig(cfg, "staging")
// => nil, a context-not-found error
```

**Think it through:**
- Compare the two constructors in the table: which argument is new?
- Which parameter of `RESTConfig` holds the context name?

### 2. Fix the client settings

**Goal:** make `Tune(c, userAgent)` set exactly `QPS` 20, `Burst` 40, `Timeout` 30 seconds, `UserAgent` = `userAgent`.

`Tune` has three wrong values and one missing line. Add the import `"time"`.

```go
Tune(c, "deploy-bot/1.4.0")
// c.QPS => 20, c.Burst => 40, c.Timeout => 30s, c.UserAgent => "deploy-bot/1.4.0"
```

**Think it through:**
- Two numbers are swapped. Which?
- `Timeout = 30` is 30 nanoseconds. Which unit from `time` do you need?
- `UserAgent` comes from the parameter, so no quotes.

### 3. Use the real namespace lookup

**Goal:** make `Namespace(cfg, contextName)` return the namespace and error from a client config's `Namespace()` method.

The hand-written lookup is wrong twice: it returns `""` instead of
`"default"`, and it hides unknown contexts. Replace it: build a client
config for `contextName` (as in task 1), call `Namespace()`, return the
first and third results.

```go
Namespace(cfg, "prod")    // => "payments", nil
Namespace(cfg, "ci")      // => "default", nil
Namespace(cfg, "staging") // => "", a context-not-found error
```

**Think it through:**
- `Namespace()` has three results; which one don't you need?
- Delete the old map lookup: the client config handles both "default" and the missing-context error.

## Learn more

- [pkg.go.dev: clientcmd.NewNonInteractiveClientConfig](https://pkg.go.dev/k8s.io/client-go/tools/clientcmd#NewNonInteractiveClientConfig)
- [pkg.go.dev: rest.Config](https://pkg.go.dev/k8s.io/client-go/rest#Config)
- [pkg.go.dev: time.Duration](https://pkg.go.dev/time#Duration)
- [Kubernetes: Auditing](https://kubernetes.io/docs/tasks/debug/debug-cluster/audit/)
