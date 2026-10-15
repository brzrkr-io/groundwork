# Kubernetes from the user's side

**In one sentence:** Kubernetes runs your containers on a group of machines: you describe the result you want in YAML objects, and it keeps making the cluster match, restarting, rescheduling and replacing things as needed.

**Why it matters:** most platform teams run their services on Kubernetes, so "the deploy is stuck", "the pod keeps restarting" and "nothing can reach the service" are everyday tickets. Each one is solved by reading `kubectl` output and fixing a few lines of YAML, once you know what the objects are and how they connect.

**You'll learn:**
- what a cluster is made of, and what happens between `kubectl apply` and a running container
- the handful of objects you'll use every day, and how labels tie them together
- the life of a Pod: scheduling, pulling, starting, probing, restarting and stopping
- how a rolling update replaces Pods without downtime, and how to roll back
- a debugging loop with `get`, `describe`, `logs` and events, and what the common statuses mean
- how to run all of it on your own machine with kind

## The problem Kubernetes solves

You have a Go service in an image (the *Containers and images* guide). On one server, `docker run` is enough. Now you need three copies on three servers, restarted when they crash, moved when a server dies, updated one at a time, and reachable at one address while they come and go.

You could script that. The script would have to watch every server forever, and handle every failure in between. Kubernetes is that script, written once, very carefully, and run for you.

The trick is that you never tell it *how*. You tell it *what*: "three copies of this image, reachable on port 80". Kubernetes keeps comparing that with reality and fixes the difference. The *Declarative systems and control loops* guide explains that loop in depth; this guide is about using it.

## A cluster, from the outside

A *cluster* is a set of machines, called *nodes*, plus a *control plane* that runs the show.

```text
 you ── kubectl ──> API server ──> etcd (the database of every object)
                        │
          ┌─────────────┼───────────────────┐
          v             v                   v
     scheduler     controllers         kubelet on each node
   picks a node   Deployment, Job…    starts containers via the
   for each Pod   keep objects true    container runtime, runs
                                       probes, reports status
```

- The *API server* is the front door. `kubectl` and every other component talk only to it, over HTTPS.
- *etcd* stores every object. Nothing else keeps state.
- The *scheduler* gives each new Pod a node.
- *Controllers* each watch one kind of object and do its job: the Deployment controller makes ReplicaSets, the ReplicaSet controller makes Pods.
- The *kubelet* is an agent on every node. It starts the containers it is assigned, runs their probes, and reports back.

None of these call each other. They all read and write objects through the API server, which is why every problem shows up somewhere in an object you can `kubectl get`.

## Objects and manifests

Everything in Kubernetes is an *object* with the same outline:

```yaml
apiVersion: apps/v1        # which API defines this kind
kind: Deployment           # what it is
metadata:                  # who it is
  name: web
  namespace: shop
  labels:
    app: web
spec:                      # what you want
  replicas: 3
  # ...
status:                    # what is really happening (Kubernetes writes this)
  readyReplicas: 2
```

You write a *manifest*, a YAML file with one or more objects, and send it with `kubectl apply -f`. Apply is declarative: run it twice and nothing changes the second time; edit the file and apply again, and Kubernetes changes only what differs. Keep manifests in Git and the repository says what should be running.

A *namespace* is a folder for objects: `shop`, `monitoring`, `kube-system`. Names only have to be unique inside one. Most commands take `-n shop`; without it, `kubectl` uses the namespace `default`.

## The objects you'll use

| Object | What it is | You'll practise it in |
|---|---|---|
| Pod | One or more containers that run together on one node, sharing an IP | *Pods: your first manifest* |
| Deployment | "Keep N copies of this Pod template running, and update them gradually" | *Deployments: replicas, labels and selectors* |
| ReplicaSet | Made by a Deployment, one per version of the template; keeps the count | (you rarely touch it) |
| Service | One stable name and virtual IP in front of a set of Pods | *Services: selectors, ports and endpoints* |
| ConfigMap, Secret | Settings and credentials, handed to Pods as variables or files | *ConfigMaps and Secrets: settings from the cluster* |
| Namespace | A folder for objects, and a boundary for access rules | |
| Job | Run a Pod to completion once, such as a database migration | |

### Labels tie it together

Objects don't point at each other by name. They use *labels*, `key: value` tags on any object, and *selectors*, queries over labels:

```text
Deployment web ── selector app=web ──> ReplicaSet web-6d9f7c8b5d
                                          ├── Pod web-6d9f7c8b5d-4kzrt  app=web
                                          └── Pod web-6d9f7c8b5d-q8wtn  app=web
                                                 ^
Service web ──── selector app=web ───────────────┘  (every ready Pod with app=web)
```

That's flexible and fragile at once. A typo in a label isn't an error; it's a Service with no endpoints and an app nobody can reach.

## The life of a Pod

Following one Pod from `kubectl apply` to deletion explains most statuses you'll see:

```text
 created ─> Pending ─────────────> ContainerCreating ─> Running ─> Ready
            (waiting for a node)   (pulling the image,   (started)  (readiness
                                    mounting volumes)                probe passes)
```

1. **Scheduling.** The scheduler looks for a node with enough room for the Pod's *requests*: the CPU and memory it reserves. Real usage doesn't matter, only the sum of requests. No node with room: the Pod stays `Pending` with a `FailedScheduling` event.
2. **Pulling.** The kubelet pulls the image. A wrong name or tag, or a private registry without credentials, gives `ErrImagePull`, then `ImagePullBackOff` while it waits to retry.
3. **Starting.** A missing ConfigMap, Secret or key stops it here with `CreateContainerConfigError`. Otherwise the container starts: `Running`.
4. **Probing.** The *readiness probe* decides whether the Pod gets traffic; until it passes, READY shows `0/1` and Services skip the Pod. The *liveness probe* restarts a container that is stuck. A *startup probe* holds both back while a slow app starts.
5. **Crashing.** If the container exits, the kubelet restarts it, waiting longer each time (10s, 20s, 40s, … up to five minutes): `CrashLoopBackOff`. Going over the memory *limit* gets it killed by the kernel: `OOMKilled`, exit code 137.
6. **Stopping.** On delete, or when a rollout replaces it, the Pod is removed from Services and its containers get `SIGTERM`. After a grace period (30 seconds by default) anything still running gets `SIGKILL`. An app that drains in-flight requests on `SIGTERM`, as in *Graceful shutdown: drain in-flight requests*, loses nothing.

Your health handlers from *Health checks: liveness and readiness* are the other half of step 4, and your graceful shutdown code is the other half of step 6. Kubernetes only calls them; the app has to answer honestly.

### Requests, limits and QoS

Requests are for the scheduler; *limits* are ceilings enforced on the node. Above its CPU limit a container is slowed down; above its memory limit it is killed. From the two, each Pod gets a *QoS class*: Guaranteed (requests equal limits), Burstable (only some set) or BestEffort (none set). When a node runs short of memory, BestEffort Pods are evicted first. *Resources: requests, limits and QoS* goes through the arithmetic.

## Rollouts and rollbacks

Change the Pod template of a Deployment (usually the image) and the Deployment controller starts a *rollout*: a new ReplicaSet scales up while the old one scales down. Two settings bound it: `maxSurge` (extra Pods allowed, rounded up) and `maxUnavailable` (Pods allowed missing, rounded down), both 25% by default.

New Pods must pass their readiness probe before old ones go. So a broken release usually stalls with one bad Pod next to the old ones, instead of taking the service down. `kubectl rollout status` waits and fails if the rollout stops making progress; `kubectl rollout undo` goes back to an earlier revision. Kubernetes never rolls back by itself. *Rollouts: surge, unavailable and rollback* covers the numbers.

## Reading the cluster

Almost every problem is solved with four commands, in this order:

```bash
kubectl get pods                       # which Pods, what STATUS, how many restarts
kubectl describe pod NAME              # State, Last State, and Events at the bottom
kubectl logs NAME --previous           # what the crashed container printed
kubectl get events --sort-by=.lastTimestamp   # the whole namespace, newest last
```

| You see | It usually means | Look at |
|---|---|---|
| `Pending` | No node has room for the requests | describe: `FailedScheduling` |
| `ImagePullBackOff` | Wrong image name or tag, or no access | describe: "Failed to pull image" |
| `CreateContainerConfigError` | A ConfigMap, Secret or key is missing | describe: the `Failed` event |
| `CrashLoopBackOff` | The app keeps exiting | `logs --previous`, Last State's exit code |
| `OOMKilled`, exit 137 | Over the memory limit | the limit vs real usage |
| `Running`, READY `0/1` | Readiness probe failing | describe: "Readiness probe failed" |
| Restarts climbing on a `Running` Pod | Liveness probe failing | describe: "failed liveness probe" |
| Pods fine, Service unreachable | Selector or targetPort wrong | `describe service`: Endpoints |

*Pod status: reading kubectl get* and *Debugging pods: describe, events and exit codes* turn this table into Go code, and *Practice: deploy and debug a service* walks through one incident from start to finish.

## Config and security

*ConfigMaps* hold settings and *Secrets* hold credentials. Both reach a container as environment variables or as files in a mounted volume. A Secret's values are only base64-encoded, not encrypted: what protects them is who may read them.

A *security context* takes away powers the app doesn't need: run as a non-root UID, no privilege escalation, no Linux capabilities, a read-only root filesystem. Namespaces can enforce the *restricted* Pod Security Standard, so a Pod that asks for too much is refused. *Security contexts: least privilege for pods* practises it.

Who may do what in the cluster is decided by *RBAC*: Roles that list allowed actions, bound to users and *ServiceAccounts*, the identities Pods run as. That gets its own guide, *Cloud and cluster access: IAM, RBAC and short-lived credentials*, in *Cloud fundamentals*.

## Running it on your machine

*kind* ("Kubernetes in Docker") runs a whole cluster inside a Docker container, in about a minute:

```bash
nix shell nixpkgs#kind
kind create cluster --name demo        # also points kubectl at it
kubectl get nodes                      # demo-control-plane   Ready
kubectl apply -f deployment.yaml
kind delete cluster --name demo        # everything gone
```

Every exercise in this module has an optional *Run it for real* section built on it. On NixOS, kind needs Docker's log driver set to `json-file`: add `virtualisation.docker.logDriver = "json-file";` to your configuration.

When the same app runs in several environments, *kustomize* (built into `kubectl`) keeps one *base* of plain manifests and a small *overlay* per environment with only the differences: `kubectl apply -k overlays/production`. *Kustomize: one base, many environments* sets one up.

## Common mistakes

- **A selector that doesn't match.** A Service with a typo in its selector has no endpoints; nothing errors. Check `Endpoints:` in `kubectl describe service`.
- **Mixing up `port` and `targetPort`.** `port` is what clients call; `targetPort` is where the app listens. A wrong `targetPort` gives endpoints that refuse connections.
- **Liveness probes that check dependencies.** If the database is down, restarting every app makes it worse. Liveness asks "is this process stuck?"; readiness can say "I can't serve right now".
- **No requests or limits.** BestEffort Pods are evicted first and can starve their neighbours. Set at least memory requests and limits.
- **Using the `latest` tag.** You can't tell what's running, and a rollback goes nowhere. Use a version tag or a digest.
- **Deleting Pods to "fix" them.** The ReplicaSet makes new ones with the same problem. Fix the Deployment and apply it.
- **Reading `logs` after a crash and seeing nothing useful.** That's the new container. Use `--previous`.
- **Treating "applied" as "deployed".** `kubectl apply` returns at once; wait for `kubectl rollout status`.

## Check yourself

1. A Deployment is applied, but its Pods show `Pending` for ten minutes. What is the most likely cause, and which command shows it?
2. A Pod is `Running` with READY `0/1` and zero restarts. What is wrong, and does it get traffic from its Service?
3. A container's Last State says `OOMKilled`, exit code 137. What does 137 mean, and what do you change?
4. The Pods behind Service `cart` are all `1/1 Running`, but `curl http://cart/` fails at once and `kubectl describe service cart` shows an empty `Endpoints:`. What do you compare?
5. You delete a Pod that belongs to a Deployment. What happens next, and why?
6. A Pod shows `CreateContainerConfigError`. What kind of problem is it, and where is the exact reason?
7. You change a Deployment's image to a tag that doesn't exist. With 4 replicas and the default rollout settings, does the service go down? Why or why not?
8. Why does a container restarted by a failing liveness probe often show exit code 137 instead of 143?

## Answers

1. No node has enough unreserved CPU or memory for the Pod's requests. `kubectl describe pod` shows a `FailedScheduling` event such as `0/3 nodes are available: 3 Insufficient cpu`. Lower the request or add capacity.
2. Its readiness probe is failing: the app runs but says it can't serve, or the probe checks the wrong path or port. Not-ready Pods are left out of the Service's endpoints, so it gets no traffic. Nothing restarts it, because readiness never restarts.
3. 137 is 128 + 9: the process was killed by `SIGKILL`, here by the kernel's out-of-memory killer because the container went over its memory limit. Raise the limit above what the app really uses, or make it use less.
4. The Service's selector with the Pods' labels (`kubectl get pods --show-labels`). An empty endpoint list means no ready Pod has every label the selector asks for, usually a typo or a label from a different app.
5. The ReplicaSet controller sees one Pod fewer than it wants and creates a replacement within seconds. The Deployment is a standing order; to really remove Pods, scale or delete the Deployment.
6. Something the container needs from the cluster is missing, usually a ConfigMap, Secret or one of their keys (or a security setting Kubernetes can't satisfy). The `Failed` event in `kubectl describe pod` names exactly what.
7. No. With 4 replicas, the defaults allow 1 extra Pod (`maxSurge` 25%, rounded up) and 1 missing Pod (`maxUnavailable` 25%, rounded down). The new Pods never become ready (`ImagePullBackOff`), so the rollout stalls and at least 3 old Pods keep serving. Fix the tag or run `kubectl rollout undo`.
8. The kubelet first sends `SIGTERM` (which would give 143). An app that doesn't stop within the grace period, or a shell script that ignores the signal, is then killed with `SIGKILL`: 128 + 9 = 137. An app that handles `SIGTERM` and exits cleanly shows 0 instead.

## Key terms

| Term | Meaning |
|---|---|
| Cluster | A set of machines (nodes) run together by a control plane |
| Node | One machine in the cluster that runs Pods |
| Control plane | The API server, etcd, scheduler and controllers that manage the cluster |
| API server | The front door of a cluster; every client and component talks only to it |
| kubelet | The agent on each node that starts containers, runs probes and reports status |
| Manifest | A YAML file describing one or more Kubernetes objects |
| Namespace | A folder that groups objects and scopes their names |
| Pod | One or more containers that run together on one node and share an IP address |
| Deployment | An object that keeps N copies of a Pod template running and updates them gradually |
| ReplicaSet | The object a Deployment makes for each version of its Pod template, keeping the Pod count |
| Service | A stable name and virtual IP that routes to every ready Pod matching its selector |
| Label | A key/value tag on an object, such as `app: web` |
| Selector | A query over labels that picks a set of objects |
| Endpoints | The Pod addresses a Service currently sends traffic to |
| ConfigMap | An object holding plain settings for Pods |
| Secret | An object holding credentials for Pods, base64-encoded but not encrypted |
| Readiness probe | A check that decides whether a Pod receives traffic |
| Liveness probe | A check that restarts a container when it fails |
| Request | The CPU or memory the scheduler reserves for a container |
| Limit | The most CPU or memory a container may use; over the memory limit it is killed |
| CrashLoopBackOff | A container that keeps exiting, with Kubernetes waiting longer before each restart |
| Rollout | Replacing a Deployment's Pods with a new version, a few at a time |
| kind | A tool that runs a Kubernetes cluster inside Docker containers on your machine |

## Watch

- [The Illustrated Children's Guide to Kubernetes](https://www.youtube.com/watch?v=4ht22ReBjno) · Deis, Inc. · 8 min: the friendliest picture of Pods, ReplicaSets, Services and volumes there is; watch it first.
- [Kubernetes Explained in 6 Minutes | k8s Architecture](https://www.youtube.com/watch?v=TlHvYWVUZyc) · ByteByteGo · 6 min: the control plane and nodes, matching "A cluster, from the outside".
- [Kubernetes Deconstructed: Understanding Kubernetes by Breaking It Down - Carson Anderson, DOMO](https://www.youtube.com/watch?v=90kZRyPcRZw) · CNCF [Cloud Native Computing Foundation] · 33 min: builds a cluster up one component at a time, from a single kubelet to Deployments and Services.
- [Kubernetes for SysAdmins | Kelsey Hightower at PuppetConf | Talk & Demo](https://www.youtube.com/watch?v=HlAXp0-M6SY) · Perforce Puppet · 45 min: the classic live demo of scheduling, self-healing and rolling updates, told for people who run servers.
- [10 Ways to Shoot Yourself in the Foot with Kubernetes, #9 Will Surprise You - Laurent Bernaille](https://www.youtube.com/watch?v=QKI-JRs2RIE) · CNCF [Cloud Native Computing Foundation] · 37 min: real production incidents from Datadog, many of them the mistakes listed above.

## Go deeper

- [Kubernetes: Overview](https://kubernetes.io/docs/concepts/overview/) (kubernetes.io): the official introduction to what Kubernetes is and isn't.
- [Kubernetes: Cluster architecture](https://kubernetes.io/docs/concepts/architecture/) (kubernetes.io): the control plane and node components in more detail.
- [Kubernetes: Pod lifecycle](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/) (kubernetes.io): phases, container states, restart policy and termination, exactly as "The life of a Pod" sketches them.
- [Kubernetes: Troubleshooting applications](https://kubernetes.io/docs/tasks/debug/debug-application/) (kubernetes.io): the official debugging guides for Pods, Services and StatefulSets.
- [kind: Quick start](https://kind.sigs.k8s.io/docs/user/quick-start/) (kind.sigs.k8s.io): installing kind, multi-node clusters and loading local images.
