## Task 1: Write the base

A kustomization starts like any object, with `apiVersion` and `kind`, then
lists its manifests under `resources`.

### Closer

`apiVersion: kustomize.config.k8s.io/v1beta1`, `kind: Kustomization`, and
`resources:` as a list of the two file names, written as they sit next to
the kustomization.

### Answer

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - deployment.yaml
  - service.yaml
```

## Task 2: Write the staging overlay

Three things: where to start (the base folder), the namespace, and an
image change. The image's `name` is how the base writes it, minus the tag.

### Closer

`resources:` with `- ../../base`, `namespace: staging`, and `images:` with
one item: `name: ghcr.io/stefanprodan/podinfo` and `newTag: "6.9.2"`.

### Answer

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - ../../base
namespace: staging
images:
  - name: ghcr.io/stefanprodan/podinfo
    newTag: "6.9.2"
```

## Task 3: Write the replicas patch

A patch is a cut-down Deployment: the three fields that identify it, and
the field that changes. Nothing else.

### Closer

`apiVersion: apps/v1`, `kind: Deployment`, `metadata.name: web`, then
`spec:` with `replicas: 4` under it.

### Answer

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: web
spec:
  replicas: 4
```

## Task 4: Write the production overlay

Task 2's file, with another namespace and tag, plus a `patches` list.
The patch file sits next to the kustomization.

### Closer

Add `patches:` with one item, `- path: replicas.yaml`. Paths are relative
to the folder of the kustomization file.

### Answer

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - ../../base
namespace: production
images:
  - name: ghcr.io/stefanprodan/podinfo
    newTag: "6.9.1"
patches:
  - path: replicas.yaml
```
