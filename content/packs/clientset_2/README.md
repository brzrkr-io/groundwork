# The clientset: updates and create-or-update

Day-to-day automation mostly *changes* objects that already exist: scale an
app, roll out new config. Here you fix and finish two functions that do that.

**New here:** `Update`, read-modify-write, `apierrors.IsAlreadyExists`, `maps.Equal`.

## Learn

### Recap

As in *The clientset: get, create and delete*: `cs.AppsV1().Deployments(ns)`
and `cs.CoreV1().ConfigMaps(ns)` are typed clients. Each call takes `ctx`
first and a `metav1` options struct last. Tests pass `fake.NewClientset(...)`,
and `apierrors.IsNotFound(err)` spots a missing object.

### Update replaces the whole object

`Update` is an HTTP `PUT`. You send the complete object and it replaces what
is stored. Any field you leave empty is stored as empty.

So the safe pattern is *read, modify, write*: `Get` the object, change one
field on that same value, and pass it to `Update`.

```go
svc, err := cs.CoreV1().Services("shop").Get(ctx, "cart", metav1.GetOptions{})
if err != nil {
	return err
}
svc.Spec.Type = corev1.ServiceTypeNodePort // change only what you mean to
_, err = cs.CoreV1().Services("shop").Update(ctx, svc, metav1.UpdateOptions{})
```

> **Why is it written like this?** `err =` (not `:=`) on the last line is
> required: `err` already exists, and `:=` needs at least one new name on the
> left. The `_` is required too: `Update` returns two values, and Go makes you
> say what to do with each.

The object from `Get` also carries `metadata.resourceVersion`. If someone
else changed it between your `Get` and `Update`, a real API server rejects
the write with a *Conflict* error instead of overwriting their change. You
don't need to handle that here.

### Pointer fields and `new`

`Spec.Replicas` is a `*int32`, a pointer, so "not set" (nil) differs from 0.
`new(value)` makes a pointer to a copy of a value:

```go
var limit int32 = 3
p := new(limit) // p is a *int32 pointing at 3
```

### Create or update

`Create` fails if the name is taken, and `apierrors.IsAlreadyExists(err)`
recognises that error. A *create-or-update* helper:

1. Try `Create`. If it works, you're done.
2. If the error is AlreadyExists, `Get` the existing object, change it and
   `Update` it.
3. Return any other error.

Why `Create` first? With `Get` first, another program could create the object
between your `Get` and your `Create`. Starting with `Create` makes
AlreadyExists the one signal you rely on.

### Skip writes that change nothing

Every write wakes every controller watching that object. If the stored data
already matches, don't send an `Update`. `maps.Equal(a, b)` (package `maps`)
reports whether two maps hold the same keys and values:

```go
maps.Equal(map[string]string{"a": "1"}, map[string]string{"a": "1"}) // => true
```

The fake clientset records every request. The last test reads them with
`cs.Actions()` and counts the `"update"` requests.

## Your task

The `payments` team's release bot scales Deployments and pushes config. Its
client code has a bug and an unfinished function.

### 1. Fix `ScaleDeployment`

**Goal:** make `ScaleDeployment` change only the replica count of the Deployment `namespace/name`, keeping its labels and containers.

Right now it throws away the result of `Get` and sends a brand-new, nearly
empty Deployment to `Update`, which wipes out the labels and containers.

```go
ScaleDeployment(ctx, cs, "shop", "web", 5)
// => nil; shop/web now has 5 replicas, with the same labels and containers
```

- Set the replicas from the `replicas` parameter (it may be `0`).
- If `Get` fails, return its error unchanged, so a missing Deployment gives a
  NotFound error.
- After the fix, `appsv1` is unused in `exercise.go`. Delete that import (Go
  refuses to compile unused imports).

**Think it through:**
- Which value already holds the full Deployment, labels and all? Right now it is thrown away with `_`.
- Which field holds the replica count, and what type is it?

### 2. Finish `ApplyConfigMap`

**Goal:** when `Create` fails because the ConfigMap already exists, make `ApplyConfigMap` update the existing one to hold `data` and return what `Update` returns.

- AlreadyExists: `Get` the existing ConfigMap, set its `Data` to the `data`
  parameter, `Update` it, and return both results of `Update`.
- Keep the existing object's other fields: the tests check that its `team`
  label survives.
- Any other error from `Create` (or from `Get`): return `nil` and the error.
- Add `apierrors "k8s.io/apimachinery/pkg/api/errors"` to the imports.

```go
// payments/gateway-config holds {"currency": "EUR"}
ApplyConfigMap(ctx, cs, "payments", "gateway-config", map[string]string{"currency": "GBP"})
// => the ConfigMap with data {"currency": "GBP"}, nil
```

**Think it through:**
- After `Create` fails, which question do you ask about `err` first?
- Why update the object from `Get` instead of `desired`? (Hint: labels.)

### 3. Skip no-op updates

**Goal:** if the existing ConfigMap's `Data` already equals `data`, make `ApplyConfigMap` return the existing ConfigMap and `nil` without calling `Update`.

Add `"maps"` to the imports.

```go
// payments/gateway-config holds {"currency": "EUR"}
ApplyConfigMap(ctx, cs, "payments", "gateway-config", map[string]string{"currency": "EUR"})
// => the existing ConfigMap, nil, and no update request was sent
```

**Think it through:**
- The comparison must happen *before* you overwrite the existing `Data`. Where is that in your code?

## Learn more

- [Kubernetes API concepts: resource versions](https://kubernetes.io/docs/reference/using-api/api-concepts/#resource-versions)
- [client-go typed DeploymentInterface](https://pkg.go.dev/k8s.io/client-go/kubernetes/typed/apps/v1#DeploymentInterface)
- [apimachinery api/errors](https://pkg.go.dev/k8s.io/apimachinery/pkg/api/errors)
- [Go maps package](https://pkg.go.dev/maps)
