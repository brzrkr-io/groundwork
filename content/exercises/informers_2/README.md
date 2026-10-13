# Informers: listers and the read-only cache

Once an informer's cache is full, your code can answer "what is running?"
without asking the API server. Here you read that cache with a *lister*, and
fix two classic informer bugs.

**New here:** listers, label selectors on a lister, `DeepCopy`.

## Learn

### Recap

An informer lists one kind of object, watches for changes, and keeps every
object in a local cache. A shared informer factory hands out one informer per
kind.

The order matters: request informers, then `StartWithContext`, then
`WaitForCacheSyncWithContext(ctx).AsError()`.

### Listers

A *lister* reads from an informer's cache. It never makes a network call.
Here is the Pod version:

```go
import corelisters "k8s.io/client-go/listers/core/v1"

var pods corelisters.PodLister = factory.Core().V1().Pods().Lister()

p, err := pods.Pods("monitoring").Get("prometheus-0")   // one object
all, err := pods.List(labels.Everything())               // every namespace
some, err := pods.Pods("monitoring").List(selector)      // one namespace, filtered
```

The Deployment lister works the same way, with `Deployments(...)` in place of
`Pods(...)`. It is `appslisters.DeploymentLister`.

When `Get` finds nothing, it returns a NotFound error, the same kind the
clientset returns, so `apierrors.IsNotFound(err)` works.

**Watch out:** calling `Lister()` also tells the factory you want that
informer. If you call it after `StartWithContext`, the informer never runs,
and the lister stays empty forever.

> **Why is it written like this?** `corelisters` and `appslisters` are
> import *aliases*: you pick the name after `import`. Go requires an alias
> only when two packages share a name (both are `v1`); the names themselves
> are convention.

### Label selectors

`List` takes a `labels.Selector`, as in *Labels and selectors: matching in Go*.
Both helpers come from `k8s.io/apimachinery/pkg/labels`:

```go
labels.Everything()                                   // matches every object
labels.SelectorFromSet(labels.Set{"app": "prometheus"}) // matches app=prometheus
```

In `labels.Set{"app": "prometheus"}` the key `"app"` is fixed text, so it goes
in quotes. If the value came from a parameter, you'd write the parameter's
name with no quotes.

### Pointers that may be nil

`Spec.Replicas` is a `*int32`: a pointer that may be `nil` (not set). Check
before you read it, and use `*` to get the number:

```go
if p.Spec.Priority != nil {
	total += *p.Spec.Priority
}
```

### The cache is shared and read-only

`Get` and `List` hand you pointers to objects *inside* the cache. Every other
handler and lister sees those same objects. Never change them.

To make a change, edit a copy. Every Kubernetes type has a `DeepCopy` method
that copies all the nested maps, slices and pointers too:

```go
cached, err := pods.Pods("monitoring").Get("prometheus-0")
p := cached.DeepCopy() // safe to edit
p.Labels["checked"] = "yes"
```

After you `Update` the API server, the cache still shows the old object for a
moment. It catches up when the watch event arrives.

> **Why is it written like this?** Go has no read-only pointers, so nothing
> stops you editing the cache. Using `DeepCopy` is a Kubernetes rule you
> follow by hand. The names `cached` and `p` are the author's choice.

## Your task

The on-call capacity bot answers questions from the cache, so it doesn't load
the API server. Its code has two bugs and one unfinished function.

### 1. Fix StartLister

**Goal:** fix the order of steps in `StartLister` so the Deployment lister it returns sees every Deployment.

`StartLister` must start the factory, wait for the caches (returning `nil` and
the error if waiting fails), and return the lister with a `nil` error. Right
now the lister is always empty. It must also keep seeing Deployments created
later.

```go
lister, err := StartLister(ctx, factory) // cluster has shop/web and shop/api
lister.List(labels.Everything())
// => web and api, err == nil
```

**Think it through:**
- When does the starter call `Lister()`: before or after `StartWithContext`?
- You need the lister at the end too. Where can you keep it until then?

### 2. Finish TeamReplicas

**Goal:** make `TeamReplicas` return the total replicas of the Deployments in `namespace` that carry the label `team=<team>`.

- Read them from the lister, not the API server.
- The label key is the fixed text `team`; the value is the `team` parameter.
- A nil `Spec.Replicas` counts as `1` (the Kubernetes default).
- No matches gives `0`.
- If `List` returns an error, return `0` and that error.

Add the import `"k8s.io/apimachinery/pkg/labels"`.

```go
// shop:     web (team=storefront, 3), api (team=storefront, nil), worker (team=fulfilment, 2)
// payments: checkout (team=storefront, 5)
TeamReplicas(lister, "shop", "storefront")     // => 4
TeamReplicas(lister, "payments", "storefront") // => 5
TeamReplicas(lister, "shop", "nobody")         // => 0
```

**Think it through:**
- Which parameter picks the namespace, and which goes into the selector?
- Why is `api` worth 1 in the first example?

### 3. Fix Pause

**Goal:** make `Pause` edit a `DeepCopy` of the cached Deployment, so the cached object stays untouched.

`Pause` reads a Deployment from the lister, sets `Spec.Paused` to `true`, and
sends it with `Update`. Keep all of that. A missing Deployment must still
return the lister's NotFound error.

```go
Pause(ctx, cs, lister, "shop", "web")
// => nil; the API server now has Spec.Paused == true,
//    and the object you got from the lister before the call still has false

Pause(ctx, cs, lister, "shop", "ghost")
// => an error where apierrors.IsNotFound(err) is true
```

**Think it through:**
- Which variable holds the object from the cache?
- Which variable do the later lines change and send? It should be the copy.

## Learn more

- [pkg.go.dev: k8s.io/client-go/listers/apps/v1](https://pkg.go.dev/k8s.io/client-go/listers/apps/v1)
- [pkg.go.dev: k8s.io/apimachinery/pkg/labels](https://pkg.go.dev/k8s.io/apimachinery/pkg/labels)
- [GitHub: kubernetes/sample-controller](https://github.com/kubernetes/sample-controller)
