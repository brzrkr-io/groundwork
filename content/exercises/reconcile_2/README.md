# Reconcile loops: idempotent updates

Now the reconcile loop talks to the API server. A real controller calls it
over and over, so it must create what is missing, fix what drifted, and
leave everything else alone.

## Learn

### Recap

A reconcile function observes the desired and the actual state, diffs them,
and acts only on the difference, so running it twice is safe (idempotent).
As in *The clientset: get, create and delete*, you read objects with
`cs.AppsV1().Deployments(ns).Get(ctx, name, metav1.GetOptions{})`, write them
with `Create` and `Update`, and spot a missing object with
`apierrors.IsNotFound(err)`. `ptr.Deref(d.Spec.Replicas, 1)` reads replicas,
counting `nil` as 1.

### Don't overwrite what you don't own

Many hands touch a live Deployment: other teams add labels and annotations,
and tools add env vars. If you build a fresh object and `Update` with it,
everything you didn't set is wiped out. Instead, update a copy of the object
you just read, and change only your own fields:

```go
updated := current.DeepCopy()
updated.Spec.Replicas = new(desired.Replicas)
```

The copy keeps the object's `ResourceVersion` too, which the API server uses
to reject updates based on stale data.

### Seeing what a fake clientset did

`fake.NewClientset` records every call as an *action*. The tests read them
with `cs.Actions()`. Each action has a *verb* (`a.GetVerb()`): `"get"`,
`"list"` and `"watch"` only read, while `"create"`, `"update"`, `"patch"`
and `"delete"` change something. `cs.ClearActions()` forgets the list.
The tests use this to check idempotence:

```go
Reconcile(ctx, cs, "payments", "worker") // may create or update
cs.ClearActions()
Reconcile(ctx, cs, "payments", "worker") // must only "get"
// a helper turns cs.Actions() into the write list, e.g. ["update deployments"]
```

## Your task

The `payments` team declares each app's desired state in a ConfigMap with
the same name as its Deployment, holding `replicas` and `image`. A teammate
wrote `Reconcile`, but it can't create Deployments, it writes on every run,
and it wipes out other people's changes.

`desiredFrom` and `newDeployment` are finished. Every Deployment here has
exactly one container.

### 1. Create the missing Deployment

When the Deployment doesn't exist (`apierrors.IsNotFound(err)`), create
`newDeployment(namespace, name, desired)` with
`cs.AppsV1().Deployments(namespace).Create(ctx, ..., metav1.CreateOptions{})`
and return. Return a create error wrapped with `fmt.Errorf`, like the other
errors in the function.

```go
// cluster: only the ConfigMap payments/worker (replicas "3", image ".../worker:2.4.0")
Reconcile(ctx, cs, "payments", "worker")
// => nil, writes: ["create deployments"], Deployment with 3 replicas of .../worker:2.4.0
```

### 2. Do nothing when in sync

`Reconcile` calls `Update` every time, even when nothing changed. Before
updating, check whether the Deployment is already in sync: its replicas
(`nil` counts as 1, so use `ptr.Deref`) and its first container's image both
match `desired`. If so, return `nil` without writing anything. Add the import
`"k8s.io/utils/ptr"`.

```go
// Deployment has 3 replicas of .../worker:2.3.0, ConfigMap says the same
Reconcile(ctx, cs, "payments", "worker")
// => nil, writes: none
```

### 3. Update a copy of the live Deployment

When the Deployment has drifted, `Reconcile` builds a brand-new object with
`newDeployment`, which wipes out labels, annotations and env vars that
others added. Replace that: update `current.DeepCopy()` after setting
`Spec.Replicas` to `new(desired.Replicas)` and the first container's
`Image` to `desired.Image`. You won't need the `ResourceVersion` line any
more.

```go
// Deployment: 2 replicas of .../worker:2.3.0, label team=payments,
// annotation gopherlings.dev/owner=payments-team, env LOG_LEVEL=debug
// ConfigMap: replicas "4", image ".../worker:2.3.0"
Reconcile(ctx, cs, "payments", "worker")
// => 4 replicas, and the label, annotation and env var are all still there
```

## Learn more

- [Kubernetes: Controllers](https://kubernetes.io/docs/concepts/architecture/controller/)
- [pkg.go.dev: k8s.io/client-go/testing (Action)](https://pkg.go.dev/k8s.io/client-go/testing#Action)
- [github.com/kubernetes/sample-controller](https://github.com/kubernetes/sample-controller)
