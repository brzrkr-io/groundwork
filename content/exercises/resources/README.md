# Resources: requests, limits and QoS

Every container shares a node's CPU and memory with its neighbours. Tell
Kubernetes too little and your app is killed or never scheduled; tell it
too much and you pay for idle machines. Two numbers per resource decide
it: the *request* and the *limit*.

**New here:** `resources.requests` and `resources.limits`, CPU millicores, scheduling on requests, OOMKilled and exit code 137, CPU throttling, QoS classes (Guaranteed, Burstable, BestEffort), `kubectl describe node`.

## Learn

### Requests and limits

```yaml
      containers:
        - name: docs
          image: nginx:1.27
          resources:
            requests:          # what the scheduler reserves on a node
              cpu: 100m
              memory: 64Mi
            limits:            # the most the container may use
              cpu: 500m
              memory: 128Mi
```

| | Request | Limit |
|---|---|---|
| Used by | the scheduler, to pick a node | the node's kernel, while running |
| CPU | a guaranteed share | above it the container is slowed down (*throttled*) |
| Memory | reserved for the container | above it the container is killed: `OOMKilled` |

The units are the ones from *Custom types: Kubernetes memory quantities*:
`128Mi` is 128 × 1024 × 1024 bytes, `128M` is 128 × 1000 × 1000. CPU is
counted in cores: `1` is one core, `500m` (500 *millicores*) is half of
one. The request may not be bigger than the limit.

### Out of memory: exit code 137

A container that goes over its memory limit is killed by the kernel with
`SIGKILL` (signal 9). Exit codes above 128 mean "killed by signal
code − 128", so 137 = 128 + 9:

```text
    Last State:     Terminated
      Reason:       OOMKilled
      Exit Code:    137
```

Kubernetes restarts it, it grows again, it's killed again: the Pod shows
`OOMKilled`, then `CrashLoopBackOff`. The fix is a limit above what the
app really uses, or an app that uses less. CPU never kills: an app at its
CPU limit just runs slower.

### Scheduling happens on requests

The scheduler doesn't look at how busy a node really is. It adds up the
**requests** of the Pods already there and only places a Pod where its
request still fits in the node's *allocatable* capacity.
`kubectl describe node` shows both:

```text
Allocatable:
  cpu:                1930m
...
Allocated resources:
  Resource           Requests      Limits
  cpu                1380m (71%)   4500m (233%)
```

This node has 1930m − 1380m = 550m of CPU left to promise. A Pod that
requests more stays `Pending`, with an event like:

```text
Warning  FailedScheduling  default-scheduler  0/3 nodes are available: 3 Insufficient cpu.
```

Limits can add up to more than 100%: they are ceilings, not reservations.

### QoS classes

From the requests and limits, Kubernetes gives each Pod a *Quality of
Service* class. When a node runs out of memory, it evicts BestEffort Pods
first and Guaranteed Pods last.

| Class | When | Eviction |
|---|---|---|
| Guaranteed | every container has CPU and memory limits, and requests equal to them | last |
| Burstable | some requests or limits, but not Guaranteed | middle |
| BestEffort | no requests or limits at all | first |

`kubectl describe pod` prints it as `QoS Class:`.

> **Why is it written like this?**
> - **Kubernetes requires** request ≤ limit. If you set only a limit, the
>   request defaults to it.
> - **Convention:** always set memory requests and limits; many teams leave
>   out the CPU limit, since throttling hurts more than it helps.
> - **Your choice:** the numbers. Start from what the app uses under real
>   load, plus headroom.

## Your task

Three Deployments in `shop` need their resources fixed. The tests compare
quantities by value, so `0.25` and `250m` are the same.

### 1. Make web Guaranteed

**Goal:** in `1_web.yaml`, request **and** limit `250m` of CPU and `128Mi`
of memory.

```text
$ kubectl -n shop describe pod -l app=web | grep QoS
QoS Class:                   Guaranteed
```

**Think it through:**
- Where does `resources` go: on the Deployment, the Pod or the container?
- What must be true of requests and limits for Guaranteed?

### 2. Stop the OOM kills

**Goal:** give report in `2_report.yaml` enough memory: a memory request of
**at least 300Mi**, a memory limit of **at least 384Mi**, and the request
no bigger than the limit.

report crashes over and over. The evidence is in
`testdata/report_describe.txt`.

**Think it through:**
- What does exit code 137 say, and what was the memory limit?
- The log says `"memory":300000000`. How many `Mi` is that, roughly?

### 3. Get indexer scheduled

**Goal:** lower **only the CPU request** in `3_indexer.yaml` so the Pod
fits on a node: more than 0 and no more than the free CPU of the node with
the most room. Leave the memory request and both limits as they are.

indexer has been `Pending` for an hour. The evidence is in
`testdata/indexer_describe.txt` and `testdata/nodes.txt`.

**Think it through:**
- For each node: allocatable CPU minus the CPU already requested.
- Which node has the most left, and how much is that?

## Run it for real

Optional. Set up kind as in *Pods: your first manifest*. A kind node is
your whole machine, so `3_indexer.yaml` fits there either way; to see
`Pending`, set its CPU request to `"64"` and apply it.

```bash
nix shell nixpkgs#kind
kind create cluster --name gw-resources
kubectl create namespace shop
kubectl apply -f 1_web.yaml -f 2_report.yaml -f 3_indexer.yaml
kubectl -n shop get pods                       # all Running
kubectl -n shop describe pod -l app=web | grep QoS
kubectl describe node gw-resources-control-plane | grep -A8 "Allocated resources"
```

Put report's limit back to `256Mi`, apply, and watch it go `OOMKilled` and
then `CrashLoopBackOff` with `kubectl -n shop get pods -w`. Then clean up:

```bash
kind delete cluster --name gw-resources
```

## Learn more

- [Kubernetes: Resource management for Pods and containers](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/)
- [Kubernetes: Pod quality of service classes](https://kubernetes.io/docs/concepts/workloads/pods/pod-qos/)
- [Kubernetes: Assign memory resources to containers and Pods](https://kubernetes.io/docs/tasks/configure-pod-container/assign-memory-resource/)
