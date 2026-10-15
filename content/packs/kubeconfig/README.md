# kubeconfig: clusters, users and contexts

Before a tool can talk to Kubernetes it has to know *which* cluster to talk to
and who you are. That information lives in a kubeconfig file. In this
exercise you read one with client-go, the same way `kubectl` does.

**New here:** kubeconfig files, `clientcmd.Load`, the `clientcmdapi.Config` maps.

## Learn

### What a kubeconfig is

Everything you do in a cluster goes through its API server (as in
*Kubernetes objects: building a Deployment*). To talk to it you need its
address and something that proves who you are.

Ops engineers often reach several clusters: staging, production EU,
production US. A *kubeconfig* is a YAML file that lists all of them. It has
three lists and one field that picks the default:

| Section | What each entry holds |
|---|---|
| `clusters` | The API server URL (`server`) and the *CA* certificate that proves the server is genuine. |
| `users` | *Credentials*: a token, a client certificate, or a command that fetches a token. |
| `contexts` | A named pair of cluster + user, plus an optional default namespace. |
| `current-context` | The context to use when you don't say otherwise. |

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

A context only *refers* to a cluster by name. To find a context's server,
look up the context first, then look up its cluster. The context name and
the cluster name may differ.

`kubectl` reads `~/.kube/config` unless `KUBECONFIG` names other files.
`kubectl config use-context prod-eu` changes the current context. Treat a
kubeconfig like a password; the one in `testdata/` is made up.

### Reading a kubeconfig in Go

Package `k8s.io/client-go/tools/clientcmd` parses kubeconfigs:

| Function | Takes | Returns |
|---|---|---|
| `clientcmd.Load(data)` | the file's bytes (`[]byte`) | `*clientcmdapi.Config`, `error` |
| `clientcmd.LoadFromFile(path)` | a file path | `*clientcmdapi.Config`, `error` |

`clientcmdapi` is the import name for `k8s.io/client-go/tools/clientcmd/api`.
The YAML key `current-context` becomes the field `CurrentContext`.

```go
cfg, err := clientcmd.LoadFromFile("/tmp/ci.kubeconfig")
if err != nil {
	return err
}
fmt.Println(len(cfg.AuthInfos), "users")
```

> **Why is it written like this?**
> - **Go requires** you to handle both return values: `cfg, err :=` names them.
> - **Convention:** check `err` right away and return early; the happy path
>   stays at the left edge.
> - **Convention:** import `.../clientcmd/api` as `clientcmdapi`, because its
>   real package name, `api`, is too vague.

### The three maps

`Clusters`, `AuthInfos` (the users) and `Contexts` are maps from a name to a
pointer, for example `map[string]*clientcmdapi.Context`. The *comma-ok* form
tells you whether a name exists:

```go
user, ok := cfg.AuthInfos["deploy-bot"]
if !ok {
	return errors.New("no user called deploy-bot")
}
fmt.Println(user.Token)
```

A `Context` has fields `Cluster`, `AuthInfo` and `Namespace`. A `Cluster`
has a field `Server`.

### Sorted map keys

Map order is random, so tests and humans can't rely on it. To get the keys
in a stable order, `maps.Keys(m)` hands out the keys and `slices.Sorted(...)`
collects them into a sorted slice:

```go
sizes := map[string]int{"xl": 4, "m": 2, "s": 1}
slices.Sorted(maps.Keys(sizes)) // []string{"m", "s", "xl"}
```

A `for` loop that appends each key, followed by `slices.Sort`, works too.

## Your task

You are writing a small "which cluster am I on?" helper for your team's
deploy scripts. The tests use `testdata/kubeconfig.yaml`. Open it first.

### 1. Read the current context

**Goal:** make `CurrentContext(data)` parse the kubeconfig bytes and return its current context name.

- Parse `data` with `clientcmd.Load`.
- If it returns an error, return `""` and that error.
- Otherwise return the config's `CurrentContext` and `nil`.

Add the import `"k8s.io/client-go/tools/clientcmd"`.

```go
CurrentContext(contentsOfTestdataFile)
// => "staging", nil
CurrentContext([]byte("apiVersion: v1\nkind: Config\ncurrent-context: prod-eu\n"))
// => "prod-eu", nil
```

**Think it through:**
- `data` is bytes, not a path. Which of the two functions takes bytes?
- Two return values each time: what goes in each on the error path?

### 2. List the contexts in order

**Goal:** make `ContextNames(cfg)` return the names of all contexts, sorted alphabetically.

Add the imports `"maps"` and `"slices"`. A config with no contexts gives an
empty slice.

```go
ContextNames(cfg)
// => []string{"old-dev", "prod-eu", "prod-us", "staging"}
```

**Think it through:**
- Which map on `cfg` holds the contexts? Are the names its keys or its values?

### 3. Find the server for a context

**Goal:** make `ServerFor(cfg, contextName)` follow the context to its cluster and return that cluster's `Server`.

1. Look up `contextName` in `cfg.Contexts`. Missing → error.
2. Look up the context's `Cluster` name in `cfg.Clusters`. Missing → error.
3. Return the cluster's `Server` and `nil`.

On error return `""`. Build errors with `fmt.Errorf` and `%q`:

| Problem | Exact error text |
|---|---|
| No context called `contextName` | `context "prod" not found` |
| The context names a cluster not in `Clusters` | `cluster "dev" not found for context "old-dev"` |

```go
ServerFor(cfg, "prod-us")
// => "https://prod-us-east-1.k8s.example.com:6443", nil
ServerFor(cfg, "prod")
// => "", error `context "prod" not found`
ServerFor(cfg, "old-dev")
// => "", error `cluster "dev" not found for context "old-dev"`
```

**Think it through:**
- Use comma-ok twice. Which key do you use for the second lookup: `contextName`, or a field of the context you found?
- In the second error, which value is the cluster name and which is the context name?

## Learn more

- [Kubernetes: Organizing cluster access using kubeconfig files](https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/)
- [Kubernetes: Configure access to multiple clusters](https://kubernetes.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters/)
- [pkg.go.dev: k8s.io/client-go/tools/clientcmd](https://pkg.go.dev/k8s.io/client-go/tools/clientcmd)
- [pkg.go.dev: maps.Keys](https://pkg.go.dev/maps#Keys)
