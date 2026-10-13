# The clientset: get, create and delete

Every tool that changes a cluster, from `kubectl` to a deploy pipeline, talks
to the API server. In this exercise you do the same from Go: read a
Deployment, create a ConfigMap and delete a Deployment with client-go.

## Learn

### The API server is a REST API

The API server is an HTTP server. Every object has a URL path, built from its
API group, version, namespace, kind and name:

```text
/apis/apps/v1/namespaces/shop/deployments/web    # Deployment "web" in "shop"
/api/v1/namespaces/shop/configmaps/flags         # core group: /api, no group name
```

You work with objects using ordinary HTTP methods: `GET` reads one, `POST`
creates one, `PUT` replaces one and `DELETE` removes one. `kubectl get
deployment web -n shop` is just a `GET` on that first path. Controllers are
programs that make the same requests. Nothing is special about any of them:
they are all *clients* of the API server.

### client-go and the clientset

*client-go* is the official Go library for those requests. Its *clientset*
holds one typed client per API group and version, so the path pieces become
method calls:

```go
cs.AppsV1().Deployments("shop")     // /apis/apps/v1/namespaces/shop/deployments
cs.CoreV1().ConfigMaps("shop")      // /api/v1/namespaces/shop/configmaps
```

Each of those has the same methods. The first argument is always a
`context.Context`, as in *Context and cancellation*. The last is an options
struct from `metav1`. The empty struct `metav1.GetOptions{}` means "no special
options":

```go
d, err := cs.AppsV1().Deployments("shop").Get(ctx, "web", metav1.GetOptions{})
cm, err := cs.CoreV1().ConfigMaps("shop").Create(ctx, cm, metav1.CreateOptions{})
err := cs.AppsV1().Deployments("shop").Delete(ctx, "web", metav1.DeleteOptions{})
```

`Get` and `Create` return the object as the API server stored it, with extra
fields filled in (a unique ID, a creation time, and so on).

Reading a field is just walking the structs from *Kubernetes objects: building a Deployment*.
A Deployment's containers live at `d.Spec.Template.Spec.Containers`.

### Accept the interface, test with a fake

A real program builds its clientset from the `*rest.Config` you made in
*kubeconfig: building a rest.Config*. This is reading material only. The exercise never
does it:

```go
cs, err := kubernetes.NewForConfig(restConfig) // cs is a *kubernetes.Clientset
```

Your functions should not ask for `*kubernetes.Clientset`, though. Ask for
`kubernetes.Interface`, the interface that `*kubernetes.Clientset`
implements:

```go
func DeploymentImage(ctx context.Context, cs kubernetes.Interface, namespace, name string) (string, error)
```

This is the trick from *Fakes and dependency injection*. client-go ships a
fake that implements the same interface and keeps objects in memory. The
tests build one with the objects they need, and no cluster is involved:

```go
cs := fake.NewClientset(deployment1, deployment2) // package k8s.io/client-go/kubernetes/fake
```

### API errors

When a request fails, the API server says why. Package
`k8s.io/apimachinery/pkg/api/errors` (imported as `apierrors`) has helpers to
check the reason:

```go
_, err := cs.AppsV1().Deployments("shop").Get(ctx, "web", metav1.GetOptions{})
if apierrors.IsNotFound(err) {
	// there is no Deployment shop/web
}
```

`apierrors.IsNotFound(nil)` is `false`, so you can call it before checking
`err != nil`. Like `errors.Is`, it still works if you wrap the error with
`%w`.

## Your task

The `shop` team is writing a small cleanup tool. You write the three
functions that talk to the cluster.

### 1. Read a Deployment's image

Make `DeploymentImage` get the Deployment `namespace/name` and return the
`Image` of its first container:

```go
DeploymentImage(ctx, cs, "shop", "web")
// => "registry.example.com/shop/web:2.0.1", nil
```

- If `Get` fails, return its error unchanged (or wrapped with `%w`), so that
  `apierrors.IsNotFound` still works for a missing Deployment.
- If the Deployment has no containers, return `""` and an error with exactly
  this text (add `"fmt"` to the imports):

```text
deployment shop/empty has no containers
```

### 2. Create a ConfigMap

`CreateConfigMap` already builds a ConfigMap with the right name and
namespace. Set its `Data` field to `data`, create it with `Create`, and return
what `Create` returns:

```go
CreateConfigMap(ctx, cs, "shop", "feature-flags", map[string]string{"checkout_v2": "true"})
// => the stored ConfigMap, nil
```

If a ConfigMap with that name already exists, the API server refuses and
`Create` returns an error. Return that error.

### 3. Delete a Deployment, if it is there

Cleanup tools are often run twice, so deleting something that is already
gone should not count as a failure. Make `DeleteDeploymentIfExists` delete the
Deployment `namespace/name` and return:

| What happened | Return |
|---|---|
| The Deployment was deleted | `true, nil` |
| `Delete` failed with a NotFound error | `false, nil` |
| `Delete` failed any other way | `false, err` |

Add `apierrors "k8s.io/apimachinery/pkg/api/errors"` to the imports. The last
test makes the fake fail every delete with a `connection refused` error,
using `cs.PrependReactor`, to check that you pass other errors on.

## Learn more

- [Kubernetes API concepts](https://kubernetes.io/docs/reference/using-api/api-concepts/)
- [client-go kubernetes package](https://pkg.go.dev/k8s.io/client-go/kubernetes)
- [client-go fake clientset](https://pkg.go.dev/k8s.io/client-go/kubernetes/fake)
- [apimachinery api/errors](https://pkg.go.dev/k8s.io/apimachinery/pkg/api/errors)
