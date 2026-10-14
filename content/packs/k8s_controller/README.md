# Project: Kubernetes controller

You build a small but real Kubernetes controller. Whenever a team labels a
Deployment `groundwork.dev/expose: "true"`, your controller gives it a
Service, keeps that Service correct, and removes it when the label goes away.

**New here:** owner references and garbage collection, `metav1.IsControlledBy`, writing your own methods; plus practice wiring informers, workqueues and reconcile together.

## Learn

### What you are building

A developer runs:

```text
kubectl label deployment web -n shop groundwork.dev/expose=true
```

A moment later a Service called `web` appears in the `shop` namespace. It
sends traffic to the `web` Pods on every port their containers listen on.

If someone changes a container port, the Service follows. If the label is
removed, the Service disappears. Nobody writes a Service by hand.

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

| Piece | Its one job |
|---|---|
| informer + lister | keep a local copy of all Deployments; read it without calling the API |
| event handlers | decide "does this Deployment need attention?"; if so, queue its key. Never call the API. |
| workqueue | hold `"namespace/name"` keys; dedup them; retry failures with a growing delay |
| workers | take a key, call `Reconcile`, report success or failure to the queue |
| `Reconcile` | level-triggered: look at the Deployment *now* and make the Service match |

### File map

| File | What it holds | You change it? |
|---|---|---|
| `controller.go` | The `Controller` struct, `New`, and `Run`, which starts everything and shuts it down | No, but read it first |
| `keys.go` | `ExposeLabel`, `Key`, `Exposed` | Task 1 |
| `service.go` | `DesiredService`, and `ServiceInSync`, which compares selector and ports | Task 2 |
| `reconcile.go` | `ErrNotOwned`, `Reconcile`; you add `ensureService` and `removeService` | Tasks 3 and 4 |
| `handlers.go` | `addHandlers`, `onAdd`, `onUpdate` | Task 5 |
| `worker.go` | `processNextItem`, `runWorker` | Tasks 6 and 7 |
| `controller_test.go` | One test per task, plus helpers | No |

Do the tasks in order. Tasks 1 to 6 each have a test you can pass on its own.
Task 7's test runs the whole controller, so it needs everything else working.

### Recap: Services

You wrote Services in *Services: selectors, ports and endpoints* and built
them in Go in *Kubernetes objects: Services and selectors*. The one this
controller makes looks like this:

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

- `selector` picks the Pods. A Deployment's `Spec.Selector.MatchLabels`
  already picks exactly its own Pods, so the Service reuses it.
- `port` is what clients dial; `targetPort` is where the container listens.
  This controller makes them the same number.
- `targetPort` has the type `intstr.IntOrString` (a number *or* a port name).
  Build a number one with `intstr.FromInt32(n)`.

In Go, the type is `corev1.Service` with `Spec corev1.ServiceSpec`, and each
port is a `corev1.ServicePort`. A container's ports are in
`container.Ports`; each has `Name` and `ContainerPort`.

### Building a slice from nested slices

A Deployment has several containers, each with several ports. Two `range`
loops, one inside the other, visit every port in order:

```go
var names []string
for _, team := range teams {
	for _, member := range team.Members {
		names = append(names, team.Name+"/"+member)
	}
}
```

### Owner references and garbage collection

Every object's metadata can list *owner references*: "this object belongs to
that one". Kubernetes' *garbage collector* deletes an object once its owner
is gone. So deleting the `web` Deployment also deletes the `web` Service. You
need no delete handler.

An owner reference names the owner by API version, kind, name and *UID* (a
unique ID the API server gives every object). One owner per object may be
marked as the *controller*. `metav1.NewControllerRef` builds that reference
and returns a pointer:

```go
ref := metav1.NewControllerRef(sts, appsv1.SchemeGroupVersion.WithKind("StatefulSet"))
// ref is a *metav1.OwnerReference: APIVersion "apps/v1", Kind "StatefulSet",
// Name sts.Name, UID sts.UID, Controller true.
pvc.OwnerReferences = []metav1.OwnerReference{*ref} // * turns the pointer into a value
```

Owner references also answer "is this Service mine?". A Service called `web`
might have been made by hand, and your controller must never touch it.
`metav1.IsControlledBy(obj, owner)` is true only when `obj`'s controller
reference has `owner`'s UID:

```go
if !metav1.IsControlledBy(pvc, sts) {
	// not ours: hands off
}
```

### Writing your own method

In tasks 3 and 4 you write new methods on `*Controller`. Here is a different
one, with every part labelled:

```go
func (c *Controller) countPorts(d *appsv1.Deployment) int {
//   ^receiver      ^name      ^parameters               ^result type
	n := 0
	for _, ct := range d.Spec.Template.Spec.Containers {
		n += len(ct.Ports)
	}
	return n
}
```

You call it with values only: `c.countPorts(d)`. The `func`, the receiver and
the types appear only where you *declare* it, never in a call or a `return`.

> **Why is it written like this?** Go requires `func`, the parentheses and
> the types in a declaration. The receiver name `c` and the pointer receiver
> `*Controller` follow the convention of the rest of this package. The method
> names `ensureService` and `removeService` are the author's choice; use
> them so the hints match.

### What `Reconcile` already does

`Reconcile` (in `reconcile.go`) splits the key and reads the Deployment from
the lister. If it is NotFound, the Deployment was deleted after its key was
queued: `Reconcile` returns `nil`, and the garbage collector removes the
Service.

If the Deployment exists, `Reconcile` must return `ensureService`'s result
when it is exposed, and `removeService`'s result when it is not.

### Testing without a cluster

The tests use `fake.NewClientset(...)`, as in *The clientset: get, create
and delete*. Helpers at the bottom of `controller_test.go`:

- `newTestController` puts Deployments straight into the informer's cache
  without starting it, so a test can call `Reconcile`, `onAdd` or
  `processNextItem` directly.
- `writes(cs)` lists your writes, such as `["create services"]`. Reads
  (`get`) don't count.
- The queue gets a fast rate limiter: retries after 1ms, 2ms, 4ms…

## Your task

The `shop` and `payments` teams want their Deployments exposed without
writing Services. Finish the controller.

### 1. Keys and the expose label

**Goal:** in `keys.go`, make `Key(d)` return `"<namespace>/<name>"` and `Exposed(d)` report whether `d` has the expose label set to `"true"`.

- `Key(d)`: the namespace, then `/`, then the name. Same format as
  `cache.MetaNamespaceKeyFunc`.
- `Exposed(d)`: true only when `d`'s labels contain `ExposeLabel`
  (`"groundwork.dev/expose"`) with exactly the value `"true"`. A missing
  label, `"false"`, `"yes"` or a nil labels map all mean false.

```go
Key(web)       // => "shop/web"
Exposed(web)   // labels {"groundwork.dev/expose": "true"}  => true
Exposed(api)   // labels {"groundwork.dev/expose": "yes"}   => false
```

**Think it through:**
- `ExposeLabel` is a constant: use its name, not its text in quotes.
- What does reading a missing key from a map (even a nil one) give you?

### 2. Build the desired Service

**Goal:** in `service.go`, make `DesiredService(d)` return the `*corev1.Service` that should exist for `d`.

| Field | Value |
|---|---|
| `Name`, `Namespace` | the Deployment's |
| `OwnerReferences` | exactly one: a controller reference to `d`, kind `"Deployment"` |
| `Spec.Type` | `corev1.ServiceTypeClusterIP` |
| `Spec.Selector` | a copy (`maps.Clone`) of `d.Spec.Selector.MatchLabels` |
| `Spec.Ports` | one `corev1.ServicePort` per port of every container in `d.Spec.Template.Spec.Containers`, in order |

Each port has exactly four fields set:

- `Name`: the container port's `Name` (even if empty)
- `Port`: its `ContainerPort`
- `TargetPort`: `intstr.FromInt32` of its `ContainerPort`
- `Protocol`: `corev1.ProtocolTCP`

Clone the selector because `d` lives in the informer cache and must never
share memory with objects you send to the API. Import
`metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"` and
`"k8s.io/apimachinery/pkg/util/intstr"`.

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
// svc.OwnerReferences => one, apps/v1 Deployment "checkout", its UID, Controller true
```

**Think it through:**
- Build the ports slice first with two loops (see *Building a slice from nested slices*).
- The owner reference example in Learn uses a StatefulSet. What changes for a Deployment?
- `Name` and `Namespace` live in `ObjectMeta`; `Type`, `Selector`, `Ports` in `Spec`.

### 3. Create or update the Service

**Goal:** in `reconcile.go`, add the method below and return its result from `Reconcile` at `TODO(3)` (the branch for an exposed `d`).

```go
func (c *Controller) ensureService(ctx context.Context, d *appsv1.Deployment) error
```

Steps:

1. Build the desired Service with `DesiredService(d)`. Get the live Service
   with the same name and namespace from the API, through `c.clientset`.
2. NotFound (`apierrors.IsNotFound`): create the desired Service and return
   the error from `Create` (nil on success).
3. Any other error: return it.
4. If `d` does not control the live Service, return an error that wraps
   `ErrNotOwned`, such as `service shop/web: service is not owned by the deployment`. Write
   nothing.
5. If `ServiceInSync(live, desired)` is true, return nil and write nothing.
   That makes the controller *idempotent*.
6. Otherwise copy the desired `Spec.Selector` and `Spec.Ports` onto the
   **live** Service, `Update` it, and return the error.

Updating the live object (it came from the API, not the cache, so you may
change it) keeps fields the API server filled in, such as `resourceVersion`
and the cluster IP. Import `"fmt"`, `appsv1 "k8s.io/api/apps/v1"` and
`metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"`.

| Live Service `shop/web` | What `Reconcile("shop/web")` does | Writes |
|---|---|---|
| missing | creates it | `create services` |
| ours, port 80 instead of 8080 | updates it | `update services` |
| ours, selector `app: old-web` | updates it | `update services` |
| ours, already matches | nothing | none |
| no owner, or owned by another Deployment | returns an error wrapping `ErrNotOwned` | none |

**Think it through:**
- Which variable is the live Service and which is the desired one? Keep them apart.
- Steps 2–6 are a chain of `if`s, each ending in a `return`.

### 4. Remove the Service when the label goes

**Goal:** in `reconcile.go`, add a method `removeService` with the same signature as `ensureService`, and return its result at `TODO(4)` (the line that runs when `d` exists but is not exposed).

1. Get the Service with `d`'s name. NotFound: return nil.
2. Any other error: return it.
3. If `d` does not control it, return nil and leave it alone. Someone else
   made it; that is not an error.
4. Otherwise `Delete` it with `metav1.DeleteOptions{}`. If the delete fails
   with NotFound, someone beat you to it: return nil. Return any other error.

| Situation | Result | Writes |
|---|---|---|
| label `"false"` or missing, our Service exists | Service deleted, nil | `delete services` |
| label missing, Service with no owner or another owner | nil | none |
| label missing, no Service | nil | none |
| Deployment already deleted (handled by `Reconcile`) | nil | none |

**Think it through:**
- Which checks from `ensureService` can you reuse?
- How is "not ours" handled differently here than in task 3?

### 5. Queue the right Deployments

**Goal:** in `handlers.go`, make `onAdd` and `onUpdate` queue a Deployment's key only when it needs attention.

The informer calls `onAdd` for every Deployment it first sees (including all
existing ones at start-up), and `onUpdate` whenever one changes.

- `onAdd(d)`: queue `d`'s key only if `d` is exposed.
- `onUpdate(oldD, newD)`: queue `newD`'s key if `oldD` **or** `newD` is
  exposed. Checking the old version catches the label being removed.

```go
c.onAdd(web)                // exposed => queue has "shop/web"
c.onAdd(worker)             // no label => queue stays empty
c.onUpdate(web, webNoLabel) // label removed => queue has "shop/web"
c.onUpdate(worker, worker)  // never exposed => queue stays empty
```

**Think it through:**
- Your two functions from task 1 do the work. Which one decides, and which one makes the key?
- The queue is the `c.queue` field. Which queue method adds a key?

### 6. Process one item

**Goal:** in `worker.go`, make `processNextItem` tell the queue whether `Reconcile` succeeded, and return `true`.

Getting the key, the shutdown check and `defer ...Done(key)` are already
written. But the result of `Reconcile` is thrown away, so a failed key is
never retried. Fix that:

- On error: add the key back rate-limited, so it returns after a growing delay.
- On success: forget the key, to reset its failure count.
- Return `true` either way (`false` stays only for a shut-down queue).

The test checks: after a success `NumRequeues(key)` is 0; after two failures
in a row it is 2 and the key is back in the queue; after the next success it
is 0 again; and the key can be added again after a run.

**Think it through:**
- Which value does `Reconcile` return, and where is it being dropped?
- Which two queue methods from *Workqueues: retries with backoff* match "failed" and "succeeded"?

### 7. Run the worker loop

**Goal:** in `worker.go`, make `runWorker(ctx)` call `processNextItem(ctx)` again and again until it returns `false`.

`Run` starts `runWorker` in several goroutines. `TestRun` then runs the whole
controller against a fake cluster with `shop/web` and `payments/checkout`
(exposed) and `shop/worker` (not exposed), and checks, waiting up to 5
seconds for each:

1. Services `shop/web` and `payments/checkout` appear, and no third one.
2. Changing `checkout`'s first container port to 9001 updates its Service.
3. Removing the label from `web` deletes its Service.

Finally it cancels the context and expects `Run` to return nil within 5
seconds. Groundwork runs tests with `-race`, so workers must not share data
unsafely.

**Think it through:**
- Which form of `for` loops while a condition stays true?
- Does the loop body need anything in it?

## Learn more

- [Kubernetes: Service](https://kubernetes.io/docs/concepts/services-networking/service/)
- [Kubernetes: Garbage collection and owner references](https://kubernetes.io/docs/concepts/architecture/garbage-collection/)
- [sample-controller, the official example controller](https://github.com/kubernetes/sample-controller)
- [pkg.go.dev: k8s.io/client-go/util/workqueue](https://pkg.go.dev/k8s.io/client-go/util/workqueue)
