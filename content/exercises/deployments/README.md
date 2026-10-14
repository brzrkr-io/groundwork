# Deployments: replicas, labels and selectors

A bare Pod that dies stays dead. A *Deployment* is a standing order, "keep
N copies of this Pod running", that Kubernetes enforces all the time. It is
how almost every app runs on Kubernetes.

**New here:** Deployments, ReplicaSets, replicas, labels and `selector.matchLabels`, the Pod `template`, immutable fields, `kubectl get deploy`/`rs`, `kubectl rollout status`.

## Learn

### From Pod to Deployment

In *Pods: your first manifest* you deleted a Pod and nothing replaced it.
A Deployment wraps a Pod *template* and a count:

```yaml
apiVersion: apps/v1          # Deployments live in the "apps" API group
kind: Deployment
metadata:
  name: clock
  labels:
    app: clock
spec:
  replicas: 2                # how many copies
  selector:                  # how the Deployment finds its Pods
    matchLabels:
      app: clock
  template:                  # the Pod to copy: a Pod manifest without
    metadata:                # apiVersion and kind
      labels:
        app: clock
    spec:
      containers:
        - name: clock
          image: busybox:1.37
          command: ["sh", "-c", "while true; do date; sleep 5; done"]
```

Everything under `template` is the Pod from the last exercises, moved one
level in. Leave out `replicas` and you get 1.

### Who does the work

A Deployment doesn't run Pods itself. It creates a *ReplicaSet*, and the
ReplicaSet keeps the right number of Pods alive:

```text
Deployment clock            (you write this)
  └── ReplicaSet clock-6d9f7c8b5d     (one per version of the template)
        ├── Pod clock-6d9f7c8b5d-4kzrt
        └── Pod clock-6d9f7c8b5d-q8wtn
```

That is where the random parts of Pod names come from. Delete a Pod and
the ReplicaSet sees 1 of 2 and makes a new one within seconds. Each of
these is a *controller*: a loop that compares what you asked for with what
exists and fixes the difference. The guide
*Declarative systems and control loops* explains the idea.

### Labels and selectors

*Labels* are `key: value` tags on any object. A *selector* is a query over
labels: `matchLabels` picks every object that has all of the listed
labels (and possibly more).

A Deployment uses its selector to count its Pods, so **the selector must
match the template's labels**. Otherwise its own Pods wouldn't count, and
it would create more forever. The API server refuses that up front:

```text
The Deployment "clock" is invalid: spec.template.metadata.labels: Invalid
value: {"app":"clock-v2"}: `selector` does not match template `labels`
```

The template may carry extra labels, such as `track: stable`. The selector
only needs its own labels to be there.

### Some fields can't change

The selector is *immutable*: once the Deployment exists, the API server
rejects any change to it.

```text
* spec.selector: Invalid value: {"matchLabels":{"app":"clock2"}}: field is immutable
```

So choose a selector you'll never need to change, usually just `app`.
Everything else (replicas, the template, the Deployment's own labels) you
change by editing the file and applying it again.

> **Why is it written like this?**
> - **Kubernetes requires** `selector` in `apps/v1` Deployments, and that
>   it matches the template's labels.
> - **Convention:** the same `app` label on the Deployment, in the selector
>   and on the Pods, so `kubectl get pods -l app=clock` finds them.
> - **Your choice:** which extra labels to add. Labels on the Deployment
>   describe the Deployment; labels in the template land on every Pod.

### kubectl for Deployments

| Command | What it shows |
|---|---|
| `kubectl get deploy` | READY `2/2`, UP-TO-DATE, AVAILABLE per Deployment |
| `kubectl get rs` | The ReplicaSets and their counts |
| `kubectl get pods -l app=clock` | Only the Pods with that label |
| `kubectl rollout status deployment/clock` | Waits until all copies are up |

`kubectl scale deployment/clock --replicas=5` also works, but then the
cluster no longer matches your file. Change the file instead, and apply it.

## Your task

Three Deployments for the shop. The tests read each file into the real
`appsv1.Deployment` type and check what the API server would check.

### 1. Keep three copies running

**Goal:** complete `1_web.yaml` so it runs 3 Pods labelled `app: web`.

- `spec.replicas`: `3`.
- `spec.selector.matchLabels`: `app: web`.
- `spec.template.metadata.labels`: `app: web`.

```text
$ kubectl get deploy web
NAME   READY   UP-TO-DATE   AVAILABLE   AGE
web    3/3     3            3           20s
```

**Think it through:**
- Which of the three is the Deployment's own setting, and which belongs
  to the Pods it makes?
- What goes wrong if the selector and the Pod labels differ?

### 2. Fix the rejected Deployment

**Goal:** make `2_api.yaml` valid, keeping its selector `app: api`.

The error from `kubectl apply` is at the top of the file. Other objects in
the cluster already select `app: api`, so change the other side.

**Think it through:**
- Which two label sets does the error compare?
- After your fix, does every label in the selector appear in the template?

### 3. Scale and relabel a running Deployment

**Goal:** edit `3_worker.yaml` so a re-apply scales `worker` to 4 replicas,
labels the Deployment `team: payments`, and labels its Pods
`track: stable`.

- `replicas`: `4`.
- `team: payments` on the Deployment only (not on the Pods).
- `track: stable` on the Pods, as well as `app: worker`.
- The selector stays exactly `app: worker`.

```bash
kubectl get pods -l app=worker,track=stable   # 4 Pods
```

**Think it through:**
- Where does a label go to land on the Deployment, and where to land on
  each Pod?
- Why would adding `track: stable` to the selector break the apply?

## Run it for real

Optional. Set up kind as in *Pods: your first manifest*, then run this from
the exercise folder:

```bash
nix shell nixpkgs#kind
kind create cluster --name gw-deployments
kubectl apply -f 1_web.yaml
kubectl rollout status deployment/web       # waits for 3/3
kubectl get deploy,rs,pods -l app=web
kubectl delete pod -l app=web --wait=false  # delete all three Pods...
kubectl get pods -l app=web                 # ...and three new ones appear
kubectl apply -f 2_api.yaml -f 3_worker.yaml
kubectl get pods -l app=worker --show-labels # 4 Pods, each with track=stable
```

To see the immutable selector for yourself, add `track: stable` to the
selector in `3_worker.yaml`, apply again, read the error, and undo the
change. Clean up when you're done:

```bash
kind delete cluster --name gw-deployments
```

## Learn more

- [Kubernetes: Deployments](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/)
- [Kubernetes: ReplicaSet](https://kubernetes.io/docs/concepts/workloads/controllers/replicaset/)
- [Kubernetes: Labels and selectors](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/)
