# The clientset: updates and create-or-update

Reading and creating objects is half the job. Day-to-day automation mostly
*changes* objects that already exist: scale an app, roll out new config. Here
you fix and finish two functions that do that.

## Learn

### Recap

The clientset groups typed clients by API group and version:
`cs.AppsV1().Deployments(ns)` and `cs.CoreV1().ConfigMaps(ns)`. Each call
takes `ctx` first and a `metav1` options struct last. Your functions accept
`kubernetes.Interface`, so the tests can pass `fake.NewClientset(objs...)`,
and `apierrors.IsNotFound(err)` tells you that an object is missing.

### Update replaces the whole object

`Update` is an HTTP `PUT`: you send the complete object, and it replaces
what is stored. Any field you leave empty is stored as empty. So the safe
pattern is always *read, modify, write*:

```go
d, err := cs.AppsV1().Deployments("shop").Get(ctx, "web", metav1.GetOptions{})
if err != nil {
	return err
}
d.Spec.Replicas = new(int32(4)) // change only what you mean to change
_, err = cs.AppsV1().Deployments("shop").Update(ctx, d, metav1.UpdateOptions{})
```

The object from `Get` also carries `metadata.resourceVersion`. If someone
else changed the object between your `Get` and your `Update`, a real API
server rejects the write with a *Conflict* error instead of silently
overwriting their change. Real controllers then simply try again. You don't
need to handle that here.

### Create or update

`Create` fails if the name is taken. `apierrors.IsAlreadyExists(err)`
recognises that error. A *create-or-update* helper uses it like this:

1. Try `Create`. If it works, you're done.
2. If the error is AlreadyExists, `Get` the existing object, change it and
   `Update` it.
3. Any other error is returned.

Why try `Create` first, instead of `Get` first and then `Create` only on
NotFound? Both work, but with `Get` first another program can still create
the object between your `Get` and your `Create`, and you'd get AlreadyExists
anyway. Starting with `Create` makes AlreadyExists the one signal you rely
on, and a brand-new object costs a single request.

### Skip writes that change nothing

Every write wakes up every controller watching that object. If the stored
data already matches, don't send an `Update`. `maps.Equal(a, b)` from the
standard `maps` package reports whether two maps hold the same keys and
values.

The fake clientset records every request it receives. The last test reads
that list with `cs.Actions()` and counts the requests whose verb is
`"update"`.

## Your task

The `payments` team's release bot scales Deployments and pushes config. Its
client code has a bug and an unfinished function.

### 1. Fix `ScaleDeployment`

`ScaleDeployment` should change only the replica count. Right now it builds
a brand-new Deployment for `Update`, which wipes out the labels and
containers. Use the object returned by `Get`, set its `Spec.Replicas` with
`new(replicas)`, and update that:

```go
ScaleDeployment(ctx, cs, "shop", "web", 5)
// => nil; shop/web now has 5 replicas, with the same labels and containers
```

If `Get` fails, return its error unchanged, so a missing Deployment gives a
NotFound error. Once the fix is done, `appsv1` is no longer used in
`exercise.go`. Delete that import.

### 2. Finish `ApplyConfigMap`

`ApplyConfigMap` already tries `Create`. Finish the create-or-update steps
above. When `Create` fails with AlreadyExists, `Get` the existing ConfigMap,
set its `Data` to `data`, `Update` it, and return what `Update` returns. Keep
the existing object's other fields: the tests check that its `team` label
survives. Return any other error unchanged. Add
`apierrors "k8s.io/apimachinery/pkg/api/errors"` to the imports.

```go
// payments/gateway-config holds {"currency": "EUR"}
ApplyConfigMap(ctx, cs, "payments", "gateway-config", map[string]string{"currency": "GBP"})
// => the ConfigMap with data {"currency": "GBP"}, nil
```

### 3. Skip no-op updates

If the existing ConfigMap's `Data` already equals `data`, return the
existing ConfigMap and `nil` without calling `Update`. Add `"maps"` to the
imports.

```go
// payments/gateway-config holds {"currency": "EUR"}
ApplyConfigMap(ctx, cs, "payments", "gateway-config", map[string]string{"currency": "EUR"})
// => the existing ConfigMap, nil, and no update request was sent
```

## Learn more

- [Kubernetes API concepts: resource versions](https://kubernetes.io/docs/reference/using-api/api-concepts/#resource-versions)
- [client-go typed DeploymentInterface](https://pkg.go.dev/k8s.io/client-go/kubernetes/typed/apps/v1#DeploymentInterface)
- [apimachinery api/errors](https://pkg.go.dev/k8s.io/apimachinery/pkg/api/errors)
- [Go maps package](https://pkg.go.dev/maps)
