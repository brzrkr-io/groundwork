# kubeconfig: clusters, users and contexts

Before a tool can talk to Kubernetes it has to know *which* cluster to talk to
and who you are. That information lives in a kubeconfig file. In this
exercise you read one with client-go, the same way `kubectl` does.

## Learn

### What a kubeconfig is

As you saw in *Kubernetes objects: building a Deployment*, everything you do in a cluster goes
through its API server, the cluster's front door. To knock on that door you
need two things: the server's address, and something that proves who you are.

Ops engineers rarely have just one cluster. A typical laptop can reach
staging, production in Europe, production in the US, and a local test
cluster. A *kubeconfig* is a YAML file that lists all of them. It has three
lists and one pointer:

| Section | What each entry holds |
|---|---|
| `clusters` | A cluster's API server URL (`server`) and the *CA* certificate used to check that the server is genuine. |
| `users` | *Credentials*: a token, a client certificate, or a command that fetches a token. |
| `contexts` | A named pair of cluster + user, plus an optional default namespace. |
| `current-context` | The name of the context to use when you don't say otherwise. |

Here is a small one:

```yaml
apiVersion: v1
kind: Config
current-context: staging
clusters:
  - name: staging
    cluster:
      server: https://staging.k8s.example.com:6443
users:
  - name: alice
    user:
      token: fake-token-alice
contexts:
  - name: staging
    context:
      cluster: staging      # a name from the clusters list
      user: alice           # a name from the users list
      namespace: shop
```

A context only *refers* to a cluster and a user by name. To find the server
for a context, you look up the context first, then look up its cluster.

`kubectl` reads `~/.kube/config` unless the `KUBECONFIG` environment variable
names other files. Switching clusters means changing the current context:

```text
kubectl config get-contexts          # list contexts
kubectl config current-context       # print the current one
kubectl config use-context prod-eu   # switch
```

A kubeconfig holds real credentials, so treat it like a password. The one in
`testdata/` is made up.

### Reading a kubeconfig in Go

The package `k8s.io/client-go/tools/clientcmd` parses kubeconfig files.
`clientcmd.Load` takes the file's bytes, and `clientcmd.LoadFromFile` takes a
path. Both return a `*clientcmdapi.Config` (package
`k8s.io/client-go/tools/clientcmd/api`):

```go
cfg, err := clientcmd.Load(data)
if err != nil {
	return "", err
}
cfg.CurrentContext                // "staging"
cfg.Contexts["staging"].Cluster  // "staging"
cfg.Clusters["staging"].Server   // "https://staging.k8s.example.com:6443"
```

`Clusters`, `AuthInfos` (the users) and `Contexts` are maps from name to a
pointer (`map[string]*clientcmdapi.Context` and so on). Use the comma-ok form
to check whether a name exists:

```go
ctx, ok := cfg.Contexts[name]
if !ok {
	// no context with that name
}
```

### Sorted map keys

Go maps have no order. Ranging over the same map twice can give two different
orders, so tests and humans can't rely on it. To get the keys in a stable
order, collect them and sort them:

```go
names := slices.Sorted(maps.Keys(cfg.Contexts))
```

`maps.Keys` hands out the keys one at a time, and `slices.Sorted` collects
them into a new slice and sorts it. A `for name := range cfg.Contexts` loop
that appends to a slice, followed by `slices.Sort`, works just as well.

## Your task

You are writing a small "which cluster am I on?" helper for your team's
deploy scripts. The tests use `testdata/kubeconfig.yaml`. Open it and look
around first.

### 1. Read the current context

Make `CurrentContext` parse `data` with `clientcmd.Load` and return the
config's `CurrentContext`. If `Load` returns an error, return it. Add the
import `"k8s.io/client-go/tools/clientcmd"`.

```go
CurrentContext(contentsOfTestdataFile)
// => "staging", nil
```

### 2. List the contexts in order

Make `ContextNames` return the names of all contexts, sorted alphabetically.
Add the imports `"maps"` and `"slices"`. A config with no contexts gives an
empty slice.

```go
ContextNames(cfg)
// => []string{"old-dev", "prod-eu", "prod-us", "staging"}
```

### 3. Find the server for a context

Make `ServerFor` follow the context to its cluster and return the cluster's
`Server`. It returns these errors (use `fmt.Errorf` with `%q`):

| Problem | Exact error text |
|---|---|
| No context called `contextName` | `context "prod" not found` |
| The context names a cluster that isn't in `Clusters` | `cluster "dev" not found for context "old-dev"` |

```go
ServerFor(cfg, "prod-us")
// => "https://prod-us-east-1.k8s.example.com:6443", nil
ServerFor(cfg, "prod")
// => "", error `context "prod" not found`
ServerFor(cfg, "old-dev")
// => "", error `cluster "dev" not found for context "old-dev"`
```

## Learn more

- [Kubernetes: Organizing cluster access using kubeconfig files](https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/)
- [Kubernetes: Configure access to multiple clusters](https://kubernetes.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters/)
- [pkg.go.dev: k8s.io/client-go/tools/clientcmd](https://pkg.go.dev/k8s.io/client-go/tools/clientcmd)
- [pkg.go.dev: maps.Keys](https://pkg.go.dev/maps#Keys)
