# Project: Kubernetes controller

You build a small but real Kubernetes controller. Whenever a team labels a
Deployment `gopherlings.dev/expose: "true"`, your controller gives it a
Service, keeps that Service correct, and removes it when the label goes away.
Platform teams write controllers like this all the time to take repetitive
chores off people's hands.

## Learn

### What you are building

A developer runs:

```text
kubectl label deployment web -n shop gopherlings.dev/expose=true
```

A moment later a Service called `web` appears in the `shop` namespace. It
sends traffic to the `web` Pods on every port their containers listen on. If
someone changes a container port, the Service follows. If the label is
removed, the Service disappears. Nobody has to write a Service by hand.

### The architecture

Every piece is something you met in *Informers: list, watch and cache*,
*Workqueues: keys, dedup and the worker loop* and *Reconcile loops: desired
vs actual*. This project plugs them together:

```text
                       API server
          watch          │       ▲  get / create / update / delete
     Deployments         ▼       │  Services (clientset)
    ┌─────────────────────────┐  │
    │ informer + local cache  │  │
    └─────────────────────────┘  │
          │ add/update events     │
          ▼                       │
    handlers.go  onAdd, onUpdate  │   filter: only exposed Deployments
          │ "shop/web"            │
          ▼                       │
    workqueue[string]             │   dedups keys, retries with backoff
          │ Get                   │
          ▼                       │
    worker.go  processNextItem ───┤
          │                       │
          ▼                       │
    reconcile.go  Reconcile ──────┘   reads the Deployment from the cache,
                                      makes the Service match
```

- The *informer* lists and watches Deployments and keeps a local copy of all
  of them. The *lister* reads that copy without calling the API server.
- The *event handlers* only decide "does this Deployment need attention?" and,
  if so, put its key (`"namespace/name"`) on the queue. They never call the
  API.
- The *workqueue* holds keys. A key that is added twice before a worker picks
  it up is only processed once, and one key is never handled by two workers at
  the same time.
- The *workers* take keys off the queue and call `Reconcile`. A failed key goes
  back on the queue with a growing delay.
- `Reconcile` is level-triggered: it ignores what the event was, looks at the
  Deployment as it is now, and makes the Service match.

### Services, from zero

Pods come and go, and each new Pod gets a new IP address. A *Service* is a
stable address in front of a changing set of Pods:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: web
  namespace: shop
spec:
  type: ClusterIP          # a virtual IP reachable inside the cluster
  selector:
    app: web               # send traffic to Pods with this label
  ports:
    - name: http
      port: 8080           # the port clients connect to on the Service
      targetPort: 8080     # the port on the Pod that receives the traffic
      protocol: TCP
```

- `selector` picks the Pods. A Deployment's `spec.selector.matchLabels`
  already picks exactly its own Pods, so the Service reuses it.
- `port` is what clients dial; `targetPort` is where the container listens.
  This controller makes them the same number. `targetPort` has the type
  `intstr.IntOrString` (it may be a number or a port name), so you build it
  with `intstr.FromInt32(8080)`.
- `ClusterIP` is the default type: reachable only from inside the cluster.

In Go, the type is `corev1.Service` with `Spec corev1.ServiceSpec`, and each
port is a `corev1.ServicePort`.

### Owner references and garbage collection

Every object's metadata can list *owner references*: "this object belongs to
that one". Kubernetes runs a *garbage collector* that deletes an object once
its owner is gone. So when someone deletes the `web` Deployment, the cluster
deletes the `web` Service for you. Your controller doesn't even need a delete
handler.

An owner reference names the owner by API version, kind, name and *UID*. The
UID is a unique ID the API server gives every object. A Deployment that is
deleted and recreated with the same name gets a new UID, so it is a different
owner. One owner per object may be marked as the *controller*: the one that
manages it. `metav1.NewControllerRef` builds that reference:

```go
ref := metav1.NewControllerRef(d, appsv1.SchemeGroupVersion.WithKind("Deployment"))
// ref is a *metav1.OwnerReference: APIVersion "apps/v1", Kind "Deployment",
// Name d.Name, UID d.UID, Controller true.
svc.OwnerReferences = []metav1.OwnerReference{*ref}
```

Owner references also answer "is this Service mine?". A Service called `web`
might already exist because a person created it by hand. Your controller must
never change or delete a Service it doesn't own.
`metav1.IsControlledBy(svc, d)` returns true only when `svc`'s controller
reference has `d`'s UID:

```go
if !metav1.IsControlledBy(svc, d) {
	// Not ours: hands off.
}
```

### What happens to deleted Deployments

`Reconcile` (already written, in `reconcile.go`) reads the Deployment from the
lister. If the lister returns a NotFound error, the Deployment was deleted
after its key was queued. `Reconcile` returns `nil`, because there is nothing
to do: the owner reference lets the garbage collector delete the Service. If
the Deployment exists, `Reconcile` should call your `ensureService` when it is
exposed and your `removeService` when it is not.

### Testing without a cluster

The tests use `fake.NewClientset(...)`, as in *The clientset: get, create and delete*. Read
the helpers at the bottom of `controller_test.go`:

- `newTestController` builds a Controller and puts Deployments straight into
  the informer's cache, without starting the informer. That lets a test call
  `Reconcile`, `onAdd` or `processNextItem` directly, with no goroutines in
  the background.
- `writes(cs)` lists the writes your code made, such as
  `["create services"]`. Reads (`get`) are not counted.
- The tests give the work queue a fast rate limiter: retries come back after
  1ms, 2ms, 4ms… instead of the production delays.

### File map

| File | What it holds | You change it? |
|---|---|---|
| `controller.go` | The `Controller` struct, `New`, and `Run`, which starts everything and shuts it down | No, but read it |
| `keys.go` | `ExposeLabel`, `Key`, `Exposed` | Task 1 |
| `service.go` | `DesiredService`, and `ServiceInSync`, which compares selector and ports | Task 2 |
| `reconcile.go` | `ErrNotOwned`, `Reconcile`; you add `ensureService` and `removeService` | Tasks 3 and 4 |
| `handlers.go` | `addHandlers`, `onAdd`, `onUpdate` | Task 5 |
| `worker.go` | `processNextItem`, `runWorker` | Tasks 6 and 7 |
| `controller_test.go` | One test per task, plus helpers | No |

Suggested order: read `controller.go` first, then do the tasks in order. Each
task has its own test, and tasks 1 to 6 can be checked on their own. Task 7's
test runs the whole controller, so it needs everything else working.

## Your task

The `shop` and `payments` teams want their Deployments exposed without
writing Services. Finish the controller.

### 1. Keys and the expose label

In `keys.go`:

- `Key(d)` returns the work-queue key: the namespace, `/`, then the name.
  That is the same format `cache.MetaNamespaceKeyFunc` produces in
  *Workqueues: keys, dedup and the worker loop*.
- `Exposed(d)` returns true only when `d`'s labels contain `ExposeLabel`
  (`"gopherlings.dev/expose"`) with exactly the value `"true"`. A missing
  label, `"false"`, `"yes"` or a nil labels map all mean false.

```go
Key(web)                                 // => "shop/web"
Exposed(web)   // labels {"gopherlings.dev/expose": "true"}  => true
Exposed(api)   // labels {"gopherlings.dev/expose": "yes"}   => false
```

### 2. Build the desired Service

In `service.go`, `DesiredService(d)` returns a `*corev1.Service` with:

- `Name` and `Namespace` equal to the Deployment's.
- `OwnerReferences`: exactly one, built with `metav1.NewControllerRef(d,
  appsv1.SchemeGroupVersion.WithKind("Deployment"))` (dereference it with
  `*`).
- `Spec.Type`: `corev1.ServiceTypeClusterIP`.
- `Spec.Selector`: the Deployment's `Spec.Selector.MatchLabels`. Copy the
  map with `maps.Clone`, because `d` lives in the informer cache and must
  never share memory with objects you send to the API.
- `Spec.Ports`: one `corev1.ServicePort` for every port of every container
  in `d.Spec.Template.Spec.Containers`, in order (first container's ports
  first). Each one has `Name` = the container port's `Name` (even if empty),
  `Port` = `ContainerPort`, `TargetPort` = `intstr.FromInt32(ContainerPort)`
  and `Protocol` = `corev1.ProtocolTCP`. Leave every other field empty.

For `payments/checkout`, whose `checkout` container has ports `http` 8080 and
`grpc` 9000, and whose `sidecar` container has `metrics` 9090:

```go
svc := DesiredService(checkout)
// svc.Name, svc.Namespace => "checkout", "payments"
// svc.Spec.Selector       => {"app": "checkout"}
// svc.Spec.Ports =>
//   {Name: "http",    Port: 8080, TargetPort: intstr.FromInt32(8080), Protocol: "TCP"}
//   {Name: "grpc",    Port: 9000, TargetPort: intstr.FromInt32(9000), Protocol: "TCP"}
//   {Name: "metrics", Port: 9090, TargetPort: intstr.FromInt32(9090), Protocol: "TCP"}
```

### 3. Create or update the Service

In `reconcile.go`, add a method
`func (c *Controller) ensureService(ctx context.Context, d *appsv1.Deployment) error`
and return its result from `Reconcile` where `TODO(3)` is, the branch for an
exposed `d`.

1. Build `want := DesiredService(d)` and get the live Service with the same
   name from the API: `c.clientset.CoreV1().Services(d.Namespace).Get(ctx,
   d.Name, metav1.GetOptions{})`.
2. If that returns a NotFound error (`apierrors.IsNotFound`), create `want`
   and return the error from `Create` (nil on success).
3. Any other error: return it.
4. If `d` does not control the live Service (`metav1.IsControlledBy` is
   false), return an error that wraps `ErrNotOwned`, such as
   `fmt.Errorf("service %s/%s: %w", ns, name, ErrNotOwned)`. Write nothing.
5. If `ServiceInSync(live, want)` is true, return nil and write nothing. This
   is what makes the controller *idempotent*: reconciling twice in a row
   makes no second write.
6. Otherwise copy `want`'s `Spec.Selector` and `Spec.Ports` onto the live
   Service and `Update` it. The live Service came from the API server (not the
   cache), so you may change it. Updating the live object, rather than
   sending `want`, keeps the fields the API server filled in, such as
   `resourceVersion` and the cluster IP.

| Live Service `shop/web` | What `Reconcile("shop/web")` does | Writes |
|---|---|---|
| missing | creates it | `create services` |
| ours, port 80 instead of 8080 | updates it | `update services` |
| ours, selector `app: old-web` | updates it | `update services` |
| ours, already matches | nothing | none |
| no owner, or owned by another Deployment | returns an error wrapping `ErrNotOwned` | none |

### 4. Remove the Service when the label goes

Still in `reconcile.go`, add a method `removeService(ctx, d)` with the same
signature and return its result at the end of `Reconcile`, where `TODO(4)`
is. That line runs when `d` exists but is not exposed (label missing or not
`"true"`).

1. Get the Service with `d`'s name. If it is NotFound, return nil.
2. Any other error: return it.
3. If `d` does not control it, return nil and leave it alone. Someone else
   made it; it is not an error.
4. Otherwise `Delete` it with `metav1.DeleteOptions{}`. If the delete fails
   with NotFound, someone beat you to it: return nil. Return any other error.

| Situation | Result | Writes |
|---|---|---|
| label `"false"` or missing, our Service exists | Service deleted, nil | `delete services` |
| label missing, Service with no owner or another owner | nil | none |
| label missing, no Service | nil | none |
| Deployment already deleted (handled by `Reconcile`) | nil | none |

### 5. Queue the right Deployments

In `handlers.go`, the informer calls `onAdd` for every Deployment it first
sees (including all existing ones at start-up) and `onUpdate` whenever one
changes. Queue keys with `c.queue.Add(Key(d))`.

- `onAdd(d)`: queue `Key(d)` only if `Exposed(d)`.
- `onUpdate(oldD, newD)`: queue `Key(newD)` if `Exposed(oldD)` **or**
  `Exposed(newD)`. Checking the old version is what catches the label being
  removed: the new version is not exposed, but its Service still has to go.

```go
c.onAdd(web)                // exposed => queue has "shop/web"
c.onAdd(worker)             // no label => queue stays empty
c.onUpdate(web, webNoLabel) // label removed => queue has "shop/web"
c.onUpdate(worker, worker)  // never exposed => queue stays empty
```

### 6. Process one item

In `worker.go`, `processNextItem(ctx)` handles one key, as in *Workqueues: keys, dedup and the worker loop*.
The first steps are already written:

1. `key, shutdown := c.queue.Get()`. If `shutdown` is true, return false.
2. `defer c.queue.Done(key)`, so the queue knows you've finished with this
   key, whatever happens.
3. Call `c.Reconcile(ctx, key)`.

But the result of `Reconcile` is thrown away, so a key that fails is never
tried again. Fix that:

- On error, call `c.queue.AddRateLimited(key)` so the key comes back after a
  delay that grows with each failure.
- On success, call `c.queue.Forget(key)` to reset its failure count.
- Return true either way.

The test checks each step: after a success `NumRequeues(key)` is 0; after
two failures in a row it is 2 and the key returns to the queue; after the
next success it is back to 0; adding the key again after a run puts it back
on the queue (only possible if you called `Done`); and a shut-down queue
makes `processNextItem` return false.

### 7. Run the worker loop

In `worker.go`, `runWorker(ctx)` calls `processNextItem(ctx)` again and again
until it returns false. A `for` loop with only a condition is enough.

`Run` starts `runWorker` in several goroutines. `TestRun` then runs the whole
controller against a fake clientset holding `shop/web` and
`payments/checkout` (both exposed) and `shop/worker` (not exposed), and
checks, waiting up to 5 seconds for each:

1. Services `shop/web` and `payments/checkout` appear, and no third one.
2. Changing `checkout`'s first container port to 9001 updates its Service.
3. Removing the label from `web` deletes its Service.

Finally it cancels the context and expects `Run` to return nil within 5
seconds. Run the tests with `-race`, as Gopherlings does, to be sure your
workers don't share data unsafely.

## Learn more

- [Kubernetes: Service](https://kubernetes.io/docs/concepts/services-networking/service/)
- [Kubernetes: Garbage collection and owner references](https://kubernetes.io/docs/concepts/architecture/garbage-collection/)
- [sample-controller, the official example controller](https://github.com/kubernetes/sample-controller)
- [pkg.go.dev: k8s.io/client-go/util/workqueue](https://pkg.go.dev/k8s.io/client-go/util/workqueue)
