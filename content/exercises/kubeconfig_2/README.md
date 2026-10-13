# kubeconfig: building a rest.Config

Knowing the server URL isn't enough to make a request. Here you turn a
kubeconfig context into the connection settings client-go actually uses, and
tune them the way a well-behaved ops tool should.

## Learn

### Recap

A kubeconfig lists clusters (server URL and CA), users (credentials) and
contexts (a cluster + user + optional namespace), plus a `current-context`.
`clientcmd.Load` and `clientcmd.LoadFromFile` parse one into a
`*clientcmdapi.Config`, whose `Contexts`, `Clusters` and `AuthInfos` are maps.

### From a context to a `*rest.Config`

Every client-go client is built from a `*rest.Config` (package
`k8s.io/client-go/rest`). It gathers everything needed to reach one cluster:
`Host` (the server URL), `BearerToken` or certificates, the CA, and some
client settings. `clientcmd` builds one from a context for you:

```go
clientConfig := clientcmd.NewNonInteractiveClientConfig(
	*cfg,                         // the config value, not the pointer
	"prod",                       // the context to use
	&clientcmd.ConfigOverrides{}, // empty: no command-line overrides
	nil,                          // only needed to write refreshed tokens back
)
restConfig, err := clientConfig.ClientConfig()
```

This **does not contact the cluster**. It only joins the context, cluster
and user entries together, so it works offline and in tests. You find out
whether the server is reachable on your first real request.

`clientcmd.NewDefaultClientConfig(*cfg, overrides)` looks similar, but it
always uses `current-context`, which is what `kubectl` does when you don't
pass `--context`.

If the context doesn't exist, `ClientConfig()` returns an error, and
`clientcmd.IsContextNotFound(err)` reports `true` for it. Its text looks like
this:

```text
invalid configuration: [context was not found for specified context: staging, cluster has no server defined]
```

### Client settings worth setting

| Field | Type | Why |
|---|---|---|
| `QPS` | `float32` | client-go limits *itself* to this many requests per second, so one busy tool can't overload the API server everyone shares. The default is 5. |
| `Burst` | `int` | Extra requests allowed in a short spike above `QPS`. The default is 10. |
| `UserAgent` | `string` | Recorded in the API server's audit logs, so the platform team can see which tool made a request. |
| `Timeout` | `time.Duration` | The longest a single request may take. Zero means no limit, so a request to a stuck server could hang forever. |

Watch out with durations: a `time.Duration` counts nanoseconds, so
`c.Timeout = 30` means 30 *nanoseconds*. Write `30 * time.Second`.

### The default namespace

The client config can also tell you a context's namespace:

```go
ns, _, err := clientConfig.Namespace()
```

It returns `"default"` when the context doesn't set a namespace (Kubernetes
puts objects there when you don't choose one). The middle `bool` says whether
an override set the namespace. You don't need it here, so use `_`.

## Your task

The release team's deploy bot needs to connect to whichever cluster the
operator picks. The tests use `testdata/kubeconfig.yaml`, with the contexts
`dev`, `ci` and `prod` (current context: `dev`).

### 1. Fix the context bug

`RESTConfig` always connects to the current context, whatever you ask for.
Build the client config with `clientcmd.NewNonInteractiveClientConfig` so it
uses `contextName`. Return the error from `ClientConfig()` unchanged. The
tests check it with `clientcmd.IsContextNotFound(err)`.

```go
c, _ := RESTConfig(cfg, "prod")
// c.Host        => "https://prod-us-east-1.k8s.example.com:6443"
// c.BearerToken => "fake-token-release-bot"
RESTConfig(cfg, "staging")
// => nil, a context-not-found error
```

### 2. Fix the client settings

`Tune` has three wrong values and one missing line. It must set exactly:
`QPS` 20, `Burst` 40, `Timeout` 30 seconds, and `UserAgent` set to
`userAgent`. Add the import `"time"`.

```go
Tune(c, "deploy-bot/1.4.0")
// c.QPS => 20, c.Burst => 40, c.Timeout => 30s, c.UserAgent => "deploy-bot/1.4.0"
```

### 3. Use the real namespace lookup

`Namespace` reads the map by hand and gets two things wrong: it returns `""`
instead of `"default"`, and it hides unknown contexts. Rewrite it to build a
client config for `contextName` (as in task 1) and return the namespace and
error from its `Namespace()` method.

```go
Namespace(cfg, "prod")    // => "payments", nil
Namespace(cfg, "ci")      // => "default", nil
Namespace(cfg, "staging") // => "", a context-not-found error
```

## Learn more

- [pkg.go.dev: clientcmd.NewNonInteractiveClientConfig](https://pkg.go.dev/k8s.io/client-go/tools/clientcmd#NewNonInteractiveClientConfig)
- [pkg.go.dev: rest.Config](https://pkg.go.dev/k8s.io/client-go/rest#Config)
- [Kubernetes: Auditing](https://kubernetes.io/docs/tasks/debug/debug-cluster/audit/)
- [Kubernetes: API Priority and Fairness](https://kubernetes.io/docs/concepts/cluster-administration/flow-control/)
