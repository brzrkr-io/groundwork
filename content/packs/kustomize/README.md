# Kustomize: one base, many environments

Staging and production run the same app with small differences: another
namespace, another image tag, more replicas. Copying the YAML for each
means the copies drift apart. *Kustomize*, built into `kubectl`, keeps one
*base* and a small *overlay* per environment that says only what differs.

**New here:** kustomize, `kustomization.yaml` (`resources`, `namespace`, `images`, `patches`), bases and overlays, strategic merge patches, `kubectl kustomize`, `kubectl apply -k`.

## Learn

### The layout

```text
deploy/
├── base/
│   ├── kustomization.yaml     lists the manifests
│   ├── deployment.yaml
│   └── service.yaml
└── overlays/
    ├── staging/
    │   └── kustomization.yaml   "the base, but…"
    └── production/
        ├── kustomization.yaml
        └── replicas.yaml        a patch
```

Every folder kustomize works on has a `kustomization.yaml`. The plain
manifests stay plain: no templates, no placeholders. Kustomize reads them,
changes them as the kustomization says, and prints the result.

### A base

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:          # files (or folders) to include, relative to this file
  - configmap.yaml
  - deployment.yaml
```

```bash
kubectl kustomize deploy/base      # print the combined YAML, change nothing
```

### An overlay

An overlay lists the base as a resource, then says what to change:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - ../../base                     # a folder with its own kustomization.yaml
namespace: wiki-dev                # set on every object
images:
  - name: registry.example.com/wiki/docs   # the image as the base writes it, without the tag
    newTag: "2.4.0-rc1"                    # quoted: a tag like 2.10 would be a number in YAML
patches:
  - path: resources.yaml           # a patch file next to this one
```

| Field | Changes |
|---|---|
| `resources` | What to start from: files, or folders with a kustomization |
| `namespace` | `metadata.namespace` of every object (the namespace itself must exist) |
| `images` | The image of every container whose image `name` matches: `newTag`, `newName` or both |
| `patches` | Anything else, with a patch |

### Patches

The simplest patch is a *strategic merge patch*: a piece of an object with
just enough to say which object (`apiVersion`, `kind`, `metadata.name`)
and the fields to change. Kustomize merges it into the matching object:

```yaml
# resources.yaml: give the docs container more memory
apiVersion: apps/v1
kind: Deployment
metadata:
  name: docs
spec:
  template:
    spec:
      containers:
        - name: docs          # list items are matched by name, not position
          resources:
            limits:
              memory: 512Mi
```

Everything the patch doesn't mention stays as the base has it. That's why
a patch never repeats the image or the ports.

### Building and applying

```bash
kubectl kustomize deploy/overlays/production   # print what would be applied
kubectl diff -k deploy/overlays/production     # compare with the cluster
kubectl apply -k deploy/overlays/production    # apply it
```

`-k` means "a kustomization folder" where `-f` means "a file". Paths in a
kustomization are relative to its own folder: from `overlays/staging`, the
base is `../../base`.

> **Why is it written like this?**
> - **Kustomize requires** `resources` paths relative to the kustomization
>   file, and patches that name their target's kind and name.
> - **Convention:** the `base` and `overlays/<env>` folder names. Kustomize
>   doesn't care what they're called.
> - **Your choice:** what goes in the base. Put there what every
>   environment shares, and nothing environment-specific.

## Your task

Set up `web` (podinfo) for two environments. `base/deployment.yaml` and
`base/service.yaml` are written; the four kustomize files are yours. The
tests read your files the way kustomize would.

### 1. Write the base

**Goal:** make `base/kustomization.yaml` a Kustomization that lists
`deployment.yaml` and `service.yaml`.

```bash
kubectl kustomize base    # prints the Service, then the Deployment
```

**Think it through:**
- Which `apiVersion` and `kind` does a kustomization have?
- The files sit next to the kustomization. What do their paths look like?

### 2. Write the staging overlay

**Goal:** make `overlays/staging/kustomization.yaml` build the base in the
namespace `staging`, with podinfo's tag changed to `6.9.2`.

```text
$ kubectl kustomize overlays/staging | grep -E 'namespace|image:'
  namespace: staging
  namespace: staging
      - image: ghcr.io/stefanprodan/podinfo:6.9.2
```

**Think it through:**
- How do you get from `overlays/staging` to `base`?
- What is the image's `name`: the whole reference, or the part before
  the `:`?

### 3. Write the replicas patch

**Goal:** make `overlays/production/replicas.yaml` a patch that sets the
Deployment `web` to 4 replicas.

**Think it through:**
- Which fields say which object the patch is for?
- What is the least the patch can contain?

### 4. Write the production overlay

**Goal:** make `overlays/production/kustomization.yaml` build the base in
the namespace `production`, with podinfo's tag changed to `6.9.1`, and
apply the patch in `replicas.yaml`.

```text
$ kubectl kustomize overlays/production | grep -E 'replicas|image:'
  replicas: 4
      - image: ghcr.io/stefanprodan/podinfo:6.9.1
```

**Think it through:**
- Which parts can you copy from task 2?
- Is the patch path relative to the overlay or to the exercise folder?

## Run it for real

Optional. Set up kind as in *Pods: your first manifest*. From the exercise
folder:

```bash
nix shell nixpkgs#kind
kind create cluster --name gw-kustomize
kubectl kustomize overlays/production          # read it first
kubectl create namespace staging
kubectl create namespace production
kubectl apply -k overlays/staging
kubectl apply -k overlays/production
kubectl get deploy -A -l app=web -o wide       # 2 in staging on 6.9.2, 4 in production on 6.9.1
kubectl diff -k overlays/production            # nothing: the cluster matches
kind delete cluster --name gw-kustomize
```

## Learn more

- [Kubernetes: Declarative management of objects using Kustomize](https://kubernetes.io/docs/tasks/manage-kubernetes-objects/kustomization/)
- [Kustomize documentation](https://kubectl.docs.kubernetes.io/references/kustomize/kustomization/)
