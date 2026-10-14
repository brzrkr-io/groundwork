# Security contexts: least privilege for pods

In *Image security: non-root and no secrets* you made images run as a
normal user. A *security context* is where the Pod spec says so too, and
takes away everything else the app doesn't need. If an attacker gets into
the container, they get as little as possible.

**New here:** `securityContext` on the Pod and the container, `runAsNonRoot`/`runAsUser`, `allowPrivilegeEscalation`, `capabilities.drop`, `seccompProfile`, `readOnlyRootFilesystem`, `emptyDir` volumes, Pod Security Standards.

## Learn

### Two places for the same idea

```yaml
    spec:
      securityContext:               # Pod level: applies to every container
        runAsNonRoot: true
        runAsUser: 10001
        seccompProfile:
          type: RuntimeDefault
      containers:
        - name: docs
          image: registry.example.com/wiki/docs:2.3.0
          securityContext:           # container level: wins over the Pod's
            allowPrivilegeEscalation: false
            readOnlyRootFilesystem: true
            capabilities:
              drop: ["ALL"]
```

Some fields exist in both places; a value on the container overrides the
Pod's. Some are container-only: `allowPrivilegeEscalation`,
`capabilities` and `readOnlyRootFilesystem`.

| Field | What it stops |
|---|---|
| `runAsNonRoot: true` | Starting at all if the process would run as root (UID 0) |
| `runAsUser: 10001` | Running as whatever user the image says: this UID instead |
| `allowPrivilegeEscalation: false` | A process gaining more rights than its parent (`sudo`, setuid programs) |
| `capabilities.drop: ["ALL"]` | Root-like powers Linux grants piece by piece: binding port 80, changing file owners, … |
| `seccompProfile.type: RuntimeDefault` | Rare, risky system calls, using the container runtime's standard filter |
| `readOnlyRootFilesystem: true` | Writing anywhere in the image: no dropped tools, no edited binaries |

### Why `runAsNonRoot` needs a number

Kubernetes checks `runAsNonRoot` **before** starting the container, using
the image's `USER`. A numeric user like `10001` can be checked. A user
*name* like `app` can't: the name is only looked up inside the image, in
`/etc/passwd`. So the container never starts:

```text
Error: container has runAsNonRoot and image has non-numeric user (app), cannot verify user is non-root
```

The fix is to say the UID yourself with `runAsUser`, or to build the image
with `USER 10001` instead of a name.

### Writable folders on a read-only image

With `readOnlyRootFilesystem: true` nothing in the image can be written,
and apps that use `/tmp` or a cache folder fail with `read-only file
system`. Give them an *emptyDir*: a volume that starts empty when the Pod
starts and is deleted with it.

```yaml
    spec:
      volumes:
        - name: cache
          emptyDir: {}
      containers:
        - name: docs
          volumeMounts:
            - name: cache
              mountPath: /var/cache/docs
```

`{}` is an empty map in YAML: "an emptyDir, with default settings".

### Pod Security Standards

Kubernetes defines three levels, and a namespace label makes the API
server enforce one of them:

| Level | Meaning |
|---|---|
| `privileged` | No rules |
| `baseline` | Blocks the known escalations, such as privileged containers |
| `restricted` | Baseline, plus everything in the table above except `runAsUser` and the read-only root |

```bash
kubectl label namespace wiki pod-security.kubernetes.io/enforce=restricted
```

After that, a Pod that breaks the rules is refused, and the message lists
every rule it broke. For a Deployment, `kubectl apply` prints it as a
`Warning:`, and the ReplicaSet's events say `FailedCreate`.

> **Why is it written like this?**
> - **Kubernetes requires** `capabilities` and `allowPrivilegeEscalation`
>   on the container; they don't exist at Pod level.
> - **Convention:** shared settings (user, seccomp) on the Pod, so a
>   sidecar container gets them too; per-container rules on the container.
> - **Your choice:** the UID. Above 10000 avoids clashing with users on
>   the node, but any non-zero UID passes.

## Your task

Three Deployments in `shop`, all running podinfo, whose image runs as the
user `app`. `testdata/passwd.txt` shows the image's users.

### 1. Run as a user Kubernetes can check

**Goal:** fix `1_web.yaml` so web starts: keep `runAsNonRoot: true` and
add the UID to run as.

web is stuck in `CreateContainerConfigError`; the evidence is in
`testdata/web_describe.txt`. Use the UID of `app` from
`testdata/passwd.txt` (the third field of its line).

**Think it through:**
- Why can't Kubernetes tell whether `app` is root?
- Which field gives it a number, and at which level?

### 2. Pass the restricted standard

**Goal:** add security contexts to `2_api.yaml` so it passes every rule in
the warning at the top of the file.

- `runAsNonRoot: true` and a non-root `runAsUser` (podinfo's is 100).
- `seccompProfile` with `type: RuntimeDefault`.
- On the container: `allowPrivilegeEscalation: false` and
  `capabilities.drop: ["ALL"]`.

```text
$ kubectl apply -f 2_api.yaml
deployment.apps/api configured        (and no Warning)
```

**Think it through:**
- Which settings may go on the Pod, and which must go on the container?
- Read the warning one rule at a time. Each names the field it wants.

### 3. Give upload a writable /tmp

**Goal:** in `3_upload.yaml`, add an `emptyDir` volume and mount it at
`/tmp`, writable, keeping `readOnlyRootFilesystem: true`.

Uploads fail with a 500 since the root filesystem became read-only. The
evidence is in `testdata/upload_logs.txt`.

**Think it through:**
- Where does the app try to write, and why does that fail now?
- Which part goes in the Pod's `volumes`, and which in the container?

## Run it for real

Optional. Set up kind as in *Pods: your first manifest*. From the exercise
folder:

```bash
nix shell nixpkgs#kind
kind create cluster --name gw-security-context
kubectl create namespace shop
kubectl apply -f 1_web.yaml
kubectl label namespace shop pod-security.kubernetes.io/enforce=restricted
#   Warning: existing pods in namespace "shop" violate the new PodSecurity enforce level ...
kubectl apply -f 2_api.yaml -f 3_upload.yaml        # no warnings
kubectl -n shop exec deployment/api -- id           # uid=100(app) gid=101 ...
kubectl -n shop port-forward deployment/upload 9898:9898   # leave it running
```

In a second terminal, `curl -s -X POST -d hello localhost:9898/store`
returns a hash instead of an error. Then clean up:

```bash
kind delete cluster --name gw-security-context
```

## Learn more

- [Kubernetes: Configure a security context for a Pod or container](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/)
- [Kubernetes: Pod Security Standards](https://kubernetes.io/docs/concepts/security/pod-security-standards/)
- [Kubernetes: Pod Security Admission](https://kubernetes.io/docs/concepts/security/pod-security-admission/)
