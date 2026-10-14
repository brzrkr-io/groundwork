# Shipping it for real: clusters, registries and your portfolio

**In one sentence:** shipping is the path from "the tests pass on my laptop" to "it runs on a cluster, from an image in a registry, and anyone can see how and why": a local cluster to try it on, a registry to hold the image, manifests that describe what should run, a pipeline that does it the same way every time, and a README that explains it.

**Why it matters:** a service nobody can run is a draft. The skills in this guide are the daily work of a platform engineer: getting an image into a cluster, reading `kubectl describe` when a pod won't start, keeping data alive across restarts, and wiring CI so a tag becomes a release. They're also exactly what an interviewer probes when they open your GitHub.

**You'll learn:**
- how kind, minikube and k3d give you a real Kubernetes cluster on a laptop, and which to pick
- how images are named, pushed to a registry such as ghcr.io, and pulled by a cluster
- what `kubectl apply` does, and how kustomize bases and overlays keep environments apart without copy-paste
- how to reach a service with `port-forward`, and how to debug a deploy with `get`, `describe`, events and `logs`
- how persistent volumes keep a database alive when its pod dies, and what that costs
- how CI turns a Git tag into a running release, and how to present the result as a portfolio project

## A cluster on your laptop

You don't need a cloud account to run Kubernetes. Three popular tools run a whole cluster on one machine:

| Tool | How it runs Kubernetes | Getting your image in | Good for |
|---|---|---|---|
| *kind* (Kubernetes in Docker) | each node is a Docker container running real, upstream Kubernetes | `kind load docker-image app:dev` | CI, testing manifests, multi-node clusters in seconds |
| *minikube* | a VM or a container, with many add-ons (dashboard, ingress, metrics) | `minikube image load app:dev` | learning, clicking around, add-ons |
| *k3d* | *k3s*, a slimmed-down Kubernetes distribution, in Docker | `k3d image import app:dev`, or its built-in registry | the lightest footprint, a local registry |

All three give you a kubeconfig context (see *kubeconfig: clusters, users and contexts*), so `kubectl` works the same as against a cloud cluster. This course uses kind: it's what the Kubernetes project itself uses to test Kubernetes, it starts in under a minute, and it runs unchanged in GitHub Actions.

```bash
kind create cluster --name demo      # creates the node container and switches kubectl to it
kubectl get nodes                    # demo-control-plane   Ready
kind delete cluster --name demo      # gone, with everything in it
```

**The image trap.** Your laptop's Docker has the image you just built, but the cluster's nodes have their own image store. A pod asking for `app:dev` gets `ErrImagePull` until you copy the image in with `kind load docker-image`. That's also why a local image needs a tag other than `latest`: `latest` makes Kubernetes default to `imagePullPolicy: Always`, which ignores the loaded copy and tries a registry that doesn't have it.

## Images and registries

An image reference has up to five parts:

```text
ghcr.io / octocat / golinks : v0.1.0 @ sha256:9f1c…
registry  namespace  repository  tag        digest
```

- **Registry:** the server. With no registry, Docker means Docker Hub (`docker.io`).
- **Namespace and repository:** usually the owner and the project. On ghcr.io they must be lowercase.
- **Tag:** a movable label (`v0.1.0`, `dev`). Anyone with push rights can point it at a new image.
- **Digest:** the image's content hash. It can't move, which is why production deploys pin it (see *Continuous integration and delivery*).

*ghcr.io*, the GitHub Container Registry, is a natural home for a project on GitHub. A workflow can push with its built-in `GITHUB_TOKEN` (with `packages: write`), so there's no password to store. From your laptop you log in with a personal access token that has the `write:packages` scope. A new package starts **private**; make it public in its settings if anyone else's cluster should pull it.

**How a node pulls.** When a pod is scheduled, the kubelet on that node checks its local image store, then pulls from the registry if needed. Failures show up as `ErrImagePull`, then `ImagePullBackOff` while it waits to retry. The causes are nearly always one of three: a typo in the name or tag, a private image without pull credentials (an `imagePullSecret`), or, on kind, an image you forgot to load.

## kubectl apply: describe, don't command

`kubectl create deployment` and `kubectl scale` are *imperative*: do this, now. `kubectl apply -f deployment.yaml` is *declarative*: here is what should exist, make it so. Apply compares the file with what's in the cluster and changes only the difference, so running it twice is harmless. That property is *idempotence*, and it's why manifests live in Git: the repository is the record of what should be running.

| Command | What it does |
|---|---|
| `kubectl apply -f file.yaml` | create or update the objects in the file |
| `kubectl apply -k dir/` | run kustomize on `dir/`, then apply the result |
| `kubectl diff -k dir/` | show what apply would change, without changing it |
| `kubectl rollout status deployment/app` | wait until the new pods are ready, or report why not |
| `kubectl rollout undo deployment/app` | go back to the previous pod template |
| `kubectl delete -k dir/` | delete everything the directory describes |

Under the hood, a Deployment controller sees the new desired state and works towards it, exactly the loop from *Declarative systems and control loops*.

## Kustomize: one base, many environments

Dev runs a local image with debug logs; production runs `v1.4.0` from ghcr.io with info logs and more memory. Copying the YAML per environment works for a week, then the copies drift.

*Kustomize* (built into `kubectl`) fixes this without templates. A *base* folder holds the manifests every environment shares, plus a `kustomization.yaml` listing them. An *overlay* folder lists the base and declares only the differences:

```text
deploy/
├── base/                     deployment, service, configmap, pvc
│   └── kustomization.yaml    resources: [deployment.yaml, ...]
└── overlays/
    ├── dev/                  images: newName app, newTag dev; debug logs
    └── prod/                 images: newTag v1.4.0; more replicas or memory
```

| Overlay field | Changes |
|---|---|
| `images` | the name, tag or digest of a matching image, wherever it appears |
| `patches` | any field, by merging a partial object with the same kind and name |
| `namespace`, `namePrefix` | where everything goes, and what it's called |
| `configMapGenerator` | a ConfigMap built from files or literals, with a hash in its name so pods restart when it changes |

`kubectl kustomize deploy/overlays/dev` prints the final YAML: read it before you apply it. The alternative you'll meet everywhere is *Helm*, which renders templates with values and packages them as charts. Helm suits software you hand to strangers with many options; kustomize suits your own service in a few environments.

## Reaching your service: port-forward

A Service of type `ClusterIP` (the default) has an address only inside the cluster. Your laptop can't reach it, so `curl` from your terminal fails.

`kubectl port-forward service/golinks 8080:80` opens port 8080 on your laptop and tunnels each connection through the API server to one pod behind the Service. It's a debugging tool: one pod, no load balancing, and it stops when you press Ctrl+C or the pod is replaced. Real traffic comes in through an *Ingress* or a `LoadBalancer` Service, which this course leaves for later.

## When a deploy goes wrong

A deploy fails in a handful of ways, and the same four commands find almost all of them. Go down the ladder; each step narrows the search.

```text
kubectl get pods                            1. what state is each pod in?
kubectl describe pod <name>                 2. why? read Events at the bottom
kubectl get events --sort-by=.lastTimestamp 3. what else happened around it?
kubectl logs <name> [--previous]            4. what did the program say before it died?
```

| What you see | Usual cause | Where to look |
|---|---|---|
| `Pending` | no node has the CPU or memory you requested, or a PVC can't bind | `describe pod`: `FailedScheduling`; `describe pvc` |
| `ErrImagePull`, `ImagePullBackOff` | wrong image name or tag, private image, image not loaded into kind | `describe pod`: the `Failed` pull event |
| `CrashLoopBackOff` | the program exits at startup: bad config, missing file, can't write | `logs --previous` |
| `Running` but `0/1` ready | the readiness probe fails | `describe pod`: `Readiness probe failed`, then `logs` |
| `OOMKilled`, exit code 137 | the container went over its memory limit | `describe pod`: Last State |
| `CreateContainerConfigError` | a ConfigMap or Secret it references doesn't exist | `describe pod` |

`--previous` matters for crash loops: by the time you look, the container has restarted and its current log is empty. Distroless images have no shell, so `kubectl exec` can't help; `kubectl debug` attaches a temporary container with tools instead.

**Structured logs pay off here.** A JSON log line with `status`, `route` and `error` fields can be filtered with `jq` or a log system; a line of free text can only be read.

## State on Kubernetes: persistent volumes

A container's filesystem dies with the container. For a SQLite file, that means every link lost on each deploy. Kubernetes splits durable storage into three objects:

| Object | Who writes it | What it is |
|---|---|---|
| *PersistentVolumeClaim* (PVC) | you | "I need 1Gi I can read and write", in your namespace |
| *PersistentVolume* (PV) | usually the cluster | an actual disk: a cloud volume, a network share, a node directory |
| *StorageClass* | the cluster admin | how to make a PV on demand: which provisioner, which disk type |

When you create a PVC, the StorageClass's *provisioner* creates a matching PV and *binds* them. kind ships a class called `standard` that stores volumes as directories on the node, and waits until a pod uses the claim before creating it (`WaitForFirstConsumer`), so a claim shows `Pending` until then. That's normal.

**Access modes** say how many nodes may mount the volume:

| Mode | Meaning | Typical backing |
|---|---|---|
| `ReadWriteOnce` (RWO) | read-write, by one node at a time | a cloud block disk, kind's local path |
| `ReadOnlyMany` (ROX) | read-only, by many nodes | shared reference data |
| `ReadWriteMany` (RWX) | read-write, by many nodes | a network filesystem such as NFS or EFS |

A SQLite file on an RWO volume means one writer, so one replica. It also means the `Recreate` strategy: the default rolling update starts the new pod before stopping the old one, and on a real cloud disk the new pod waits forever for a volume the old pod still holds. The price is a few seconds of downtime per deploy, and no high availability. That's a fine trade for an internal tool, and a good one to write down; a service that must survive a node failure needs a database server with replicas instead.

**Volumes aren't backups.** The default *reclaim policy*, `Delete`, removes the PV with its PVC, and `kind delete cluster` deletes everything. Back up the data somewhere else, such as a nightly copy to S3.

## From CI to cluster

The whole path, for a project like *Capstone: go links, built and shipped*:

```text
git push main ──► test (vet, test -race)
git tag v0.2.0 ─► test ──► build image ──► push ghcr.io/you/golinks:v0.2.0
                                                   │
               push style: CI runs kubectl apply ◄─┤  (needs cluster credentials in CI)
               pull style: bump the tag in Git  ◄──┘  (Argo CD or Flux applies it)
```

- **Test on every change, publish on tags.** An `if:` on the image job keeps pull requests from publishing anything.
- **Tag the image with the Git tag**, and stamp the same version into the binary with `-ldflags`, so a running pod tells you exactly which commit it is.
- **Push or pull.** In *push* style the pipeline holds credentials for the cluster and runs `kubectl apply`. In *pull* style (GitOps) the pipeline only commits the new tag to Git, and an agent inside the cluster applies it. Pull keeps cluster credentials out of CI; see *Continuous integration and delivery*.
- **kind works in CI too.** A job can create a kind cluster, load the image it just built, apply the manifests and run a smoke test, catching a broken manifest before any real cluster sees it.

## Presenting a portfolio project

Hiring managers spend about two minutes on a repository. In that time they want proof that you can **finish** something and **run** it, not a list of tools. A small service that builds, tests, deploys and explains itself beats an ambitious one that stops halfway.

| They ask | Your repository answers with |
|---|---|
| Does it work? | a green CI badge at the top, and a "Run it" section that works on a fresh machine |
| Can I see it? | a 20–30 second demo GIF or terminal recording (`asciinema`, VHS): create a link, follow it |
| Can they design? | a diagram and a paragraph: the parts, the request path, where data lives |
| Do they think about operations? | probes, metrics, structured logs, graceful shutdown, resource limits, non-root images |
| Do they know the costs? | a Trade-offs section: what you chose, what you gave up, why |
| Can I work with them? | small commits with clear messages, a tidy layout, tests that read like documentation |

**Write trade-offs, not features.** "Uses SQLite" is a fact. "SQLite instead of Postgres: no server to run, but one replica, so no high availability; fine below a few hundred requests a second" shows judgement, and judgement is what a junior engineer is hired for.

**Keep it honest.** Interviewers will ask about any line in your README. A "What I'd do next" list ("move to Postgres and run two replicas", "add a latency histogram") shows you know the limits, and gives the interview something to talk about.

Then pin the repository on your GitHub profile, link it from your CV, and be ready to run the demo live.

## Common mistakes

- **Forgetting to load the image into kind.** The pod sits in `ImagePullBackOff`; run `kind load docker-image` again after every rebuild.
- **Deploying `:latest`.** It moves, it hides what's running, and it switches the pull policy to `Always`.
- **Looking only at `kubectl get pods`.** The answer is almost always in `describe`'s Events or in `logs --previous`.
- **Copy-pasting manifests per environment.** Use a base and overlays; the copies drift.
- **A rolling update on an RWO volume.** The new pod waits for a disk the old pod holds. Use `Recreate`, or a database server.
- **Treating a PVC as a backup.** Deleting the claim, or the cluster, deletes the data.
- **A shutdown grace longer than `terminationGracePeriodSeconds`.** Kubernetes kills the process mid-drain.
- **A README that lists tools.** Show the problem, the design, how to run it and the trade-offs.

## Check yourself

1. A pod on a kind cluster shows `ImagePullBackOff` for the image `golinks:dev`, which `docker images` lists on your laptop. What's the most likely cause, and the fix?
2. Why should a Kubernetes Deployment use an image tag like `v1.2.0` rather than `latest`?
3. What is the difference between `kubectl create` and `kubectl apply`, and why do teams keep manifests in Git and use `apply`?
4. In kustomize, what is the difference between a base and an overlay?
5. A pod is in `CrashLoopBackOff` and `kubectl logs` prints nothing. Which command shows why it crashed?
6. A pod is `Running` but shows `0/1` under READY. What is failing, and where do you look?
7. Why does a Deployment that keeps a SQLite file on a ReadWriteOnce volume need one replica and the `Recreate` strategy?
8. What do a PersistentVolumeClaim, a PersistentVolume and a StorageClass each represent?
9. `kubectl port-forward service/app 8080:80` works, so a teammate wants to use it for production traffic. Why is that a bad idea?
10. In a pull-style (GitOps) deployment, what does the CI pipeline change, and who applies it to the cluster?
11. Name three things a hiring manager looks for in a portfolio project's README.

## Answers

1. The image exists in your laptop's Docker, but the kind node has its own image store and tries to pull `golinks:dev` from Docker Hub. Run `kind load docker-image golinks:dev --name <cluster>`, and make sure the pull policy isn't `Always`.
2. `latest` moves, so you can't tell which code is running or roll back to a known image, and it makes Kubernetes default to `imagePullPolicy: Always`. A version tag (or a digest) names one exact image.
3. `create` makes an object and fails if it exists; `apply` makes the cluster match a file, creating or updating, and is safe to repeat. Manifests in Git plus `apply` make the repository the record of what should run, with history and review.
4. A base holds the manifests every environment shares. An overlay lists the base and declares only what differs in one environment, such as the image tag, a setting or the replica count.
5. `kubectl logs <pod> --previous` shows the log of the container that crashed; the current container has just restarted. `kubectl describe pod` adds the exit code and reason.
6. The readiness probe is failing, so the Service sends it no traffic. `kubectl describe pod` shows `Readiness probe failed` events; then read `kubectl logs` for why.
7. Only one node can mount a ReadWriteOnce volume read-write, and SQLite allows one writer. A rolling update starts the new pod before stopping the old one, so the new pod waits for a disk the old one still holds. `Recreate` stops the old pod first, at the cost of brief downtime.
8. A PVC is a request for storage made by an app ("1Gi, read-write once"); a PV is an actual piece of storage in the cluster; a StorageClass describes how to create PVs on demand.
9. It's a debugging tunnel through the API server to a single pod: no load balancing, no failover, and it stops when the command ends or the pod is replaced. Production traffic uses an Ingress or a LoadBalancer Service.
10. CI builds and pushes the image, then commits the new tag or digest to the Git repository that describes the cluster. An agent in the cluster, such as Argo CD or Flux, notices the change and applies it.
11. Any three of: a green CI badge and a working "Run it" section; a demo GIF; a design diagram with an explanation; operational care (probes, metrics, logs, shutdown); honest trade-offs; a "what I'd do next" list; clean commits and tests.

## Key terms

| Term | Meaning |
|---|---|
| kind | Kubernetes in Docker: a tool that runs each cluster node as a Docker container |
| minikube | a tool that runs a local Kubernetes cluster in a VM or container, with add-ons |
| k3d | a tool that runs k3s, a lightweight Kubernetes distribution, in Docker |
| container registry | a server that stores and serves images, such as ghcr.io or Docker Hub |
| ghcr.io | the GitHub Container Registry; workflows push to it with `GITHUB_TOKEN` |
| image digest | an image's `sha256:` content hash; unlike a tag, it can't be moved |
| ImagePullBackOff | a pod status: the node can't pull the image and is waiting before it retries |
| declarative configuration | describing the desired state in files and letting a tool make the system match |
| idempotent | safe to repeat: running it twice gives the same result as running it once |
| kustomize | a tool built into kubectl that combines and patches manifests without templates |
| base (kustomize) | the folder of manifests every environment shares |
| overlay (kustomize) | a folder that lists a base and declares only one environment's differences |
| port-forward | a kubectl tunnel from a local port to a pod, for debugging |
| CrashLoopBackOff | a pod status: the container keeps exiting, and Kubernetes waits longer before each restart |
| PersistentVolumeClaim (PVC) | an app's request for durable storage of a size and access mode |
| PersistentVolume (PV) | a piece of durable storage in the cluster, bound to one claim |
| StorageClass | a recipe for creating PersistentVolumes on demand |
| ReadWriteOnce | an access mode: the volume can be mounted read-write by one node at a time |
| Recreate strategy | a Deployment strategy that stops all old pods before starting new ones |
| reclaim policy | what happens to a PersistentVolume when its claim is deleted: `Delete` or `Retain` |
| GitOps | desired cluster state lives in Git, and an agent in the cluster applies changes from it |

## Watch

- [Testing your K8s apps with KIND](https://www.youtube.com/watch?v=8KtmevMFfxA) · CNCF [Cloud Native Computing Foundation] · 36 min: kind's creator, Benjamin Elder, on how kind works and how to test your own Kubernetes apps with it, on a laptop and in CI.
- [Kustomize - How to Simplify Kubernetes Configuration Management](https://www.youtube.com/watch?v=Twtbg6LFnAg) · DevOps & AI Toolkit · 33 min: Viktor Farcic on kustomize's template-free approach, with a base and overlays built live.
- [Kubernetes Volumes explained | Persistent Volume, Persistent Volume Claim & Storage Class](https://www.youtube.com/watch?v=0swOh5C3OVM) · TechWorld with Nana · 21 min: how PVs, PVCs and StorageClasses fit together, with diagrams.
- [Troubleshooting Kubernetes Pod Errors | How to find them and fix them](https://www.youtube.com/watch?v=uIYIreo2Pu4) · Drewbernetes · 17 min: invalid commands, bad and private images, then logs and events: the troubleshooting ladder, demonstrated.

## Go deeper

- [kind Quick Start](https://kind.sigs.k8s.io/docs/user/quick-start/): creating clusters, loading images, and multi-node configs.
- [A visual guide on troubleshooting Kubernetes deployments](https://learnk8s.io/troubleshooting-deployments): a flowchart from "it doesn't work" to the cause, worth printing.
- [Declarative management of Kubernetes objects using Kustomize](https://kubernetes.io/docs/tasks/manage-kubernetes-objects/kustomization/): every kustomize field with examples.
- [Persistent Volumes](https://kubernetes.io/docs/concepts/storage/persistent-volumes/): binding, access modes and reclaim policies in full.
- [Working with the Container registry](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry): logging in to ghcr.io, pushing, and linking images to a repository.
