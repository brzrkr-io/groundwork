# Informers: listers and the read-only cache

Once an informer's cache is full, your code can answer "what is running?"
without asking the API server at all. In this exercise you read from that
cache with a *lister*, and fix two classic informer bugs.

## Learn

### Recap

An informer lists one kind of object, watches for changes, and keeps every
object in a local cache. A shared informer factory hands out one informer per
kind. Ask it for informers *before* `StartWithContext`, because it only
starts the ones it already knows about. Then wait for
`WaitForCacheSyncWithContext(ctx).AsError()`, and on the way out call
`cancel()` and `factory.Shutdown()`.

### Listers

A *lister* reads from an informer's cache. It never makes a network call:

```go
import appslisters "k8s.io/client-go/listers/apps/v1"

var lister appslisters.DeploymentLister = factory.Apps().V1().Deployments().Lister()

d, err := lister.Deployments("shop").Get("web")       // one object
all, err := lister.List(labels.Everything())           // every namespace
mine, err := lister.Deployments("shop").List(selector) // one namespace, filtered
```

`List` takes a `labels.Selector`, as in *Labels and selectors: matching in Go*.
`labels.Everything()` matches every object, and
`labels.SelectorFromSet(labels.Set{"team": "storefront"})` matches objects
carrying that label. Both come from `k8s.io/apimachinery/pkg/labels`.

When `Get` finds nothing, it returns the same kind of NotFound error as the
clientset, so `apierrors.IsNotFound(err)` works.

Asking for `Lister()` also asks the factory for the informer behind it. If
you ask after `StartWithContext`, that informer never runs, and the lister
stays empty forever.

### The cache is shared and read-only

`Get` and `List` hand you pointers to the objects *inside* the cache. Every
other handler and lister in the program sees those same objects. Never change
them. To make a change, edit a copy. Every Kubernetes type has a `DeepCopy`
method that copies all the nested maps, slices and pointers too:

```go
cached, err := lister.Deployments("shop").Get("web")
d := cached.DeepCopy() // now safe to edit
d.Spec.Replicas = new(int32(5))
_, err = cs.AppsV1().Deployments("shop").Update(ctx, d, metav1.UpdateOptions{})
```

After that `Update`, the cache still shows the old object for a moment. It
catches up when the watch event arrives.

## Your task

The on-call capacity bot answers questions from the cache, so it doesn't load
the API server. Its code has bugs and one unfinished function.

### 1. Fix StartLister

`StartLister` should start the factory, wait for the caches, and return a
Deployment lister that sees every Deployment. Right now the lister is always
empty. Fix the order of the steps.

```go
lister, err := StartLister(ctx, factory) // cluster has shop/web and shop/api
lister.List(labels.Everything())
// => web and api, err == nil
```

### 2. Finish TeamReplicas

Make `TeamReplicas` list, from the lister, the Deployments in `namespace`
with the label `team=<team>`, and return the sum of their replicas. A nil
`Spec.Replicas` counts as `1` (the Kubernetes default). No matches gives
`0`. Return any error from `List`. Add the import
`"k8s.io/apimachinery/pkg/labels"`.

```go
// shop:     web (team=storefront, 3), api (team=storefront, nil), worker (team=fulfilment, 2)
// payments: checkout (team=storefront, 5)
TeamReplicas(lister, "shop", "storefront")     // => 4
TeamReplicas(lister, "payments", "storefront") // => 5
TeamReplicas(lister, "shop", "nobody")         // => 0
```

### 3. Fix Pause

`Pause` reads a Deployment from the cache, sets `Spec.Paused` to `true` and
sends it to the API server with `Update`. But it edits the cached object
itself. Make it edit a `DeepCopy` instead, so the cached object stays
untouched. A missing Deployment must still return the lister's NotFound error.

```go
Pause(ctx, cs, lister, "shop", "web")
// => nil; the API server now has Spec.Paused == true,
//    and the object you got from the lister before the call still has false

Pause(ctx, cs, lister, "shop", "ghost")
// => an error where apierrors.IsNotFound(err) is true
```

## Learn more

- [pkg.go.dev: k8s.io/client-go/listers/apps/v1](https://pkg.go.dev/k8s.io/client-go/listers/apps/v1)
- [pkg.go.dev: k8s.io/apimachinery/pkg/labels](https://pkg.go.dev/k8s.io/apimachinery/pkg/labels)
- [GitHub: kubernetes/sample-controller](https://github.com/kubernetes/sample-controller)
