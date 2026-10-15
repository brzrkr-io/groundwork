# The clientset: get, create and delete

Every tool that changes a cluster, from `kubectl` to a deploy pipeline, talks
to the API server. Here you do the same from Go: read a Deployment, create a
ConfigMap and delete a Deployment with client-go.

**New here:** the client-go clientset, `Get`/`Create`/`Delete`, `kubernetes.Interface`, the fake clientset, `apierrors.IsNotFound`.

## Learn

### The API server is a REST API

The API server is an HTTP server. Every object has a URL path built from its
API group, version, namespace, kind and name:

```text
/apis/apps/v1/namespaces/shop/deployments/web    # Deployment "web" in "shop"
/api/v1/namespaces/shop/configmaps/flags         # core group: /api, no group name
```

`GET` reads an object, `POST` creates one, `PUT` replaces one and `DELETE`
removes one. `kubectl get deployment web -n shop` is just a `GET` on the first
path. Controllers make the same requests: they are all *clients*.

### The clientset turns paths into method calls

*client-go* is the official Go library for these requests. Its *clientset*
has one typed client per API group and version. You pick the pieces of the
path with method calls, left to right:

```go
cs.CoreV1().Services("shop")   // /api/v1/namespaces/shop/services
cs.CoreV1().Secrets("billing") // /api/v1/namespaces/billing/secrets
```

| Group in the URL | Method | Kinds in it |
|---|---|---|
| core (`/api/v1`) | `CoreV1()` | Pods, Services, ConfigMaps, Secrets |
| `apps/v1` | `AppsV1()` | Deployments, StatefulSets, DaemonSets |

> **Why is it written like this?** The chain of calls is required by
> client-go: each method returns the next client. The namespace is a plain
> `string` argument. When your function receives the namespace as a
> parameter, pass that parameter, not quoted text.

### Get, Create, Delete

Every typed client has the same methods. The first argument is always a
`context.Context` (as in *Context and cancellation*). The last is an options
struct from `metav1`; the empty struct literal means "no special options".

```go
svc, err := cs.CoreV1().Services("shop").Get(ctx, "cart", metav1.GetOptions{})
s, err := cs.CoreV1().Secrets("billing").Create(ctx, secret, metav1.CreateOptions{})
err := cs.CoreV1().Services("shop").Delete(ctx, "cart", metav1.DeleteOptions{})
```

| Method | Returns |
|---|---|
| `Get(ctx, name, metav1.GetOptions{})` | the object, `error` |
| `Create(ctx, obj, metav1.CreateOptions{})` | the stored object, `error` |
| `Delete(ctx, name, metav1.DeleteOptions{})` | only `error` |

`Get` and `Create` return the object as the API server stored it, with extra
fields filled in (a unique ID, a creation time, and so on).

> **Why is it written like this?** The argument order and the options
> parameter are required: the method signatures demand them, even when the
> options are empty. `metav1.GetOptions{}` is a struct literal with no
> fields set.

Reading a field is just walking the structs from *Kubernetes objects:
building a Deployment*. A Deployment's containers live at
`d.Spec.Template.Spec.Containers`, a slice; index `[0]` is the first one.
Check `len(...)` first, because indexing an empty slice panics.

### Accept the interface, test with a fake

A real program builds its clientset from the `*rest.Config` you made in
*kubeconfig: building a rest.Config*. The exercise never does this, but here
is how:

```go
cs, err := kubernetes.NewForConfig(restConfig) // cs is a *kubernetes.Clientset
```

Your functions take `kubernetes.Interface` instead, the interface that
`*kubernetes.Clientset` implements. This is the trick from *Fakes and
dependency injection*: the tests pass a fake that keeps objects in memory, so
no cluster is needed.

```go
cs := fake.NewClientset(svc1, svc2) // package k8s.io/client-go/kubernetes/fake
```

> **Why is it written like this?** Taking an interface is a convention, not a
> rule. The compiler would accept `*kubernetes.Clientset`, but then the fake
> could not be passed in.

### API errors

When a request fails, the API server says why. The package
`k8s.io/apimachinery/pkg/api/errors` has helpers that check the reason:

```go
import apierrors "k8s.io/apimachinery/pkg/api/errors"

_, err := cs.CoreV1().Secrets("billing").Get(ctx, "stripe-key", metav1.GetOptions{})
if apierrors.IsNotFound(err) {
	fmt.Println("no such secret")
}
```

`apierrors.IsNotFound(nil)` is `false`, so you may call it before checking
`err != nil`. Like `errors.Is`, it still works if you wrap the error with
`%w`.

> **Why is it written like this?** The name `apierrors` before the import
> path is the author's choice (a common convention). Without it the package
> would be called `errors` and clash with the standard `errors` package.

## Your task

The `shop` team is writing a small cleanup tool. You write the three
functions that talk to the cluster.

### 1. Read a Deployment's image

**Goal:** make `DeploymentImage` return the `Image` of the first container of the Deployment `namespace/name`.

```go
DeploymentImage(ctx, cs, "shop", "web")
// => "registry.example.com/shop/web:2.0.1", nil
```

- Use the parameters `namespace` and `name`, not fixed text.
- If `Get` fails, return `""` and its error unchanged (or wrapped with `%w`),
  so `apierrors.IsNotFound` still works for a missing Deployment.
- If the Deployment has no containers, return `""` and an error with exactly
  this text, where `shop` and `empty` are the namespace and name. Add `"fmt"`
  to the imports.

```text
deployment shop/empty has no containers
```

**Think it through:**
- Deployments are in `apps/v1`. Which method on `cs` picks that group?
- Where in the Deployment struct do the containers live, and how do you know there is at least one?

### 2. Create a ConfigMap

**Goal:** make `CreateConfigMap` store `data` in a new ConfigMap and return what `Create` returns.

`cm` is already built with the right name and namespace. Give it the `data`
parameter in its `Data` field, send it with `Create`, and return both results.

```go
CreateConfigMap(ctx, cs, "shop", "feature-flags", map[string]string{"checkout_v2": "true"})
// => the stored ConfigMap, nil
```

If a ConfigMap with that name already exists, the API server refuses and
`Create` returns an error. Return that error.

**Think it through:**
- ConfigMaps are in the core group. Which typed client do you need?
- `Create` already returns `(*corev1.ConfigMap, error)`, the same as your function. Do you need extra variables?

### 3. Delete a Deployment, if it is there

**Goal:** make `DeleteDeploymentIfExists` delete the Deployment `namespace/name` and report whether it deleted something.

Cleanup tools are often run twice, so deleting something already gone is not
a failure:

| What happened | Return |
|---|---|
| The Deployment was deleted | `true, nil` |
| `Delete` failed with a NotFound error | `false, nil` |
| `Delete` failed any other way | `false, err` |

```go
DeleteDeploymentIfExists(ctx, cs, "shop", "web")    // web exists
// => true, nil
DeleteDeploymentIfExists(ctx, cs, "shop", "ghost")  // nothing called ghost
// => false, nil
```

Add `apierrors "k8s.io/apimachinery/pkg/api/errors"` to the imports. The last
test makes the fake fail every delete with `connection refused` (using
`cs.PrependReactor`) to check you pass other errors on.

**Think it through:**
- `Delete` returns only an error. Which of the three table rows should you check first?

## Learn more

- [Kubernetes API concepts](https://kubernetes.io/docs/reference/using-api/api-concepts/)
- [client-go kubernetes package](https://pkg.go.dev/k8s.io/client-go/kubernetes)
- [client-go fake clientset](https://pkg.go.dev/k8s.io/client-go/kubernetes/fake)
- [apimachinery api/errors](https://pkg.go.dev/k8s.io/apimachinery/pkg/api/errors)
