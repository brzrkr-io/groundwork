# Reconcile loops: idempotent updates

Now the reconcile loop talks to the API server. A real controller calls it
over and over, so it must create what is missing, fix what drifted, and leave
everything else alone.

**New here:** create-or-update reconcile, updating a `DeepCopy` of the live object, checking writes with fake clientset actions.

## Learn

### Recap

A reconcile function observes the desired and actual state, diffs them, and
acts only on the difference, so running it twice is safe (*idempotent*).

From *The clientset: get, create and delete*:

| You want to | Call |
|---|---|
| read one object | `cs.AppsV1().Deployments(ns).Get(ctx, name, metav1.GetOptions{})` |
| make a new one | `...Create(ctx, obj, metav1.CreateOptions{})` |
| replace an existing one | `...Update(ctx, obj, metav1.UpdateOptions{})` |
| spot a missing object | `apierrors.IsNotFound(err)` |

`ptr.Deref(p, 1)` reads a pointer, using `1` when it is `nil`.

### The shape of Reconcile

A create-or-update reconcile has one branch per situation, and each branch
ends with a `return`:

```go
current, err := cs.CoreV1().Services(ns).Get(ctx, name, metav1.GetOptions{})
if apierrors.IsNotFound(err) {
	// missing: create it, then return
}
if err != nil {
	return fmt.Errorf("get service %s/%s: %w", ns, name, err)
}
// exists: in sync? return nil. Drifted? update, then return.
```

> **Why is it written like this?** Go requires you to handle `err` yourself;
> checking NotFound *before* the general `err != nil` is what makes "missing"
> a normal case instead of a failure. Wrapping with `"verb kind ns/name: %w"`
> is this file's convention.

### Don't overwrite what you don't own

Many hands touch a live Deployment: other teams add labels and annotations,
and tools add env vars. If you build a fresh object and `Update` with it,
everything you didn't set is wiped out.

Instead, update a copy of the object you just read, and change only your own
fields:

```go
updated := svc.DeepCopy()
updated.Spec.Type = corev1.ServiceTypeNodePort // only the field we own
```

The copy also keeps the object's `ResourceVersion`, which the API server uses
to reject updates based on stale data.

### Seeing what a fake clientset did

`fake.NewClientset` records every call as an *action*. Each action has a
*verb*: `"get"`, `"list"` and `"watch"` only read; `"create"`, `"update"`,
`"patch"` and `"delete"` write. The tests use this to check idempotence:

```go
Reconcile(ctx, cs, "payments", "worker") // may create or update
cs.ClearActions()
Reconcile(ctx, cs, "payments", "worker") // must only "get"
// the tests list the writes, e.g. ["update deployments"]
```

## Your task

The `payments` team declares each app's desired state in a ConfigMap with the
same name as its Deployment, holding `replicas` and `image`. A teammate wrote
`Reconcile`, but it can't create Deployments, it writes on every run, and it
wipes out other people's changes.

`desiredFrom` and `newDeployment` are finished. Every Deployment here has
exactly one container.

### 1. Create the missing Deployment

**Goal:** when the Deployment doesn't exist, make `Reconcile` create `newDeployment(namespace, name, desired)` and return `nil`.

- Use `Create` with `metav1.CreateOptions{}`.
- If `Create` fails, return its error wrapped with `fmt.Errorf` and `%w`, like
  the other errors in the function (e.g. `create deployment payments/worker: ...`).

```go
// cluster: only the ConfigMap payments/worker (replicas "3", image ".../worker:2.4.0")
Reconcile(ctx, cs, "payments", "worker")
// => nil, writes: ["create deployments"], Deployment with 3 replicas of .../worker:2.4.0
```

**Think it through:**
- Find the `if` that already detects "not found". What should happen inside it?
- `Create` returns two values. Which one do you need?

### 2. Do nothing when in sync

**Goal:** make `Reconcile` return `nil` without writing anything when the Deployment already matches `desired`.

"In sync" means both:

- its replicas equal `desired.Replicas` (`nil` counts as 1), **and**
- its first container's image equals `desired.Image`.

Add the import `"k8s.io/utils/ptr"`.

```go
// Deployment has 3 replicas of .../worker:2.3.0, ConfigMap says the same
Reconcile(ctx, cs, "payments", "worker")
// => nil, writes: none
```

**Think it through:**
- Which variable holds the live Deployment, and which the desired state?
- Which operator means "both must be true"?

### 3. Update a copy of the live Deployment

**Goal:** when the Deployment has drifted, make `Reconcile` update a `DeepCopy` of the live Deployment instead of a freshly built one.

On the copy, change only two fields:

- `Spec.Replicas` → a pointer to `desired.Replicas`
- the first container's `Image` → `desired.Image`

Then `Update` the copy. You won't need the `ResourceVersion` line any more.

```go
// Deployment: 2 replicas of .../worker:2.3.0, label team=payments,
// annotation gopherlings.dev/owner=payments-team, env LOG_LEVEL=debug
// ConfigMap: replicas "4", image ".../worker:2.3.0"
Reconcile(ctx, cs, "payments", "worker")
// => 4 replicas, and the label, annotation and env var are all still there
```

**Think it through:**
- Which line builds the fresh object that wipes the extra fields? Replace just that one.
- What is the full path from a Deployment to its first container's image?

## Learn more

- [Kubernetes: Controllers](https://kubernetes.io/docs/concepts/architecture/controller/)
- [pkg.go.dev: k8s.io/client-go/testing (Action)](https://pkg.go.dev/k8s.io/client-go/testing#Action)
- [github.com/kubernetes/sample-controller](https://github.com/kubernetes/sample-controller)
