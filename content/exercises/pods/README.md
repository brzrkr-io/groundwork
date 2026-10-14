# Pods: your first manifest

You can build an image and run it with Docker. Kubernetes runs the same
images, but on a whole group of machines, and keeps them running for you.
Here you write your first Kubernetes *manifests*: three Pods.

**New here:** Kubernetes, clusters and nodes, Pods, manifests (`apiVersion`, `kind`, `metadata`, `spec`), `command`/`args`/`env`/`ports` in a container, `kubectl apply`/`get`/`logs`/`delete`.

## Learn

Read the *Kubernetes from the user's side* guide first, or alongside: it
draws the whole picture this module fills in, one object at a time.

### What Kubernetes does

A *cluster* is a group of machines, called *nodes*, plus a *control plane*
that manages them. You never start containers on a node yourself. You send
the control plane's *API server* a description of what you want, and
Kubernetes picks a node and runs it there.

```text
you: kubectl apply -f pod.yaml
        |
        v
   API server  --stores-->  "a Pod named clock should exist"
        |
   scheduler   --picks-->   node-2
        |
   kubelet on node-2 --runs--> the container, and reports back
```

That description is a *manifest*: a YAML file, like the CI workflows you
wrote in *CI workflows*. Kubernetes stores it as an *object*.

### Pods

A *Pod* is the smallest thing Kubernetes runs: one or more containers that
share a network address and run together on one node. Most Pods hold one
container, so for now "a Pod" is "a running container, plus Kubernetes'
notes about it".

Here is a Pod that prints the time every five seconds:

```yaml
apiVersion: v1          # which API defines this kind; Pods are in the core "v1"
kind: Pod               # what sort of object this is
metadata:               # who it is
  name: clock
  labels:
    app: clock
spec:                   # what you want
  containers:
    - name: clock
      image: busybox:1.37
      command: ["sh", "-c", "while true; do date; sleep 5; done"]
      env:
        - name: TZ
          value: UTC
```

Every object has those four top-level keys. Kubernetes adds a fifth,
`status`, to report what is really happening. You never write `status`.

> **Why is it written like this?**
> - **Kubernetes requires** `apiVersion`, `kind` and `metadata.name`. The
>   API server rejects a manifest without them, and also rejects any field
>   it doesn't know, such as a misspelled `contaners`.
> - **YAML requires** the `- ` before `name: clock`: `containers` is a list,
>   and each item is a map. `env` is a list of maps too.
> - **Convention:** a label `app: <name>` on everything. You'll see why in
>   *Deployments: replicas, labels and selectors*.

### Inside a container

| Field | What it holds | Docker equivalent |
|---|---|---|
| `name` | The container's name inside the Pod | (none) |
| `image` | The image to run | `docker run IMAGE` |
| `command` | Replaces the image's `ENTRYPOINT` | `--entrypoint` |
| `args` | Replaces the image's `CMD` | the words after the image name |
| `env` | Environment variables, each `name` and `value` | `-e NAME=value` |
| `ports` | Ports the container listens on, each with an optional `name` | `EXPOSE` |

`command` and `args` follow the Dockerfile rules from *Dockerfiles*.
Check the image before you use them: `docker inspect IMAGE` shows its
`Entrypoint` and `Cmd`. podinfo, the image in this module, has no
`ENTRYPOINT` and `CMD ["./podinfo"]`. So `args: ["--level=debug"]` alone
would replace `./podinfo` itself, and the Pod fails to start with:

```text
exec: "--level=debug": stat --level=debug: no such file or directory
```

A port's `name` is a short label, such as `http`, that other objects can
use instead of the number. Listing a port documents it; it doesn't open or
close anything.

### podinfo

*podinfo* is a small web app written in Go, built for trying out
Kubernetes. Its image is `ghcr.io/stefanprodan/podinfo:6.9.2`. It:

- listens for HTTP on port `9898`;
- answers `curl localhost:9898` with JSON that includes a `message`;
- reads settings from flags such as `--level=debug` and from environment
  variables that start with `PODINFO_`: `PODINFO_UI_MESSAGE` sets the
  message.

### kubectl in five commands

`kubectl` is the command-line client for the API server.

| Command | What it does |
|---|---|
| `kubectl apply -f clock.yaml` | Send the manifest; create the object or update it |
| `kubectl get pods` | List Pods with their status |
| `kubectl logs clock` | Print the container's output |
| `kubectl describe pod clock` | Everything about one Pod, with recent events |
| `kubectl delete -f clock.yaml` | Delete what the manifest describes |

## Your task

Your team is moving to Kubernetes. Write the first three Pods, one per file.
The tests read each file into the real Kubernetes Go type, as strictly as
the API server does.

### 1. Write your first Pod

**Goal:** fill in the four blanks (`___`) in `1_hello.yaml`.

- `kind`: `Pod`.
- `metadata.name`: `hello`.
- One container named `hello` running `ghcr.io/stefanprodan/podinfo:6.9.2`.

```yaml
kind: Pod
metadata:
  name: hello
# ... and the container's name and image
```

**Think it through:**
- Which blank is the object's name, and which is the container's?
- Is the image name exactly as written above, tag included?

### 2. Label it and name its port

**Goal:** give the Pod in `2_web.yaml` two labels and one named port.

- Labels on the Pod: `app: web` and `tier: frontend`.
- One port on the container: name `http`, `containerPort` `9898`.

```yaml
# kubectl get pod web --show-labels
# NAME   READY   STATUS    RESTARTS   AGE   LABELS
# web    1/1     Running   0          8s    app=web,tier=frontend
```

**Think it through:**
- Labels belong to the Pod. Which of the four top-level parts describes
  who an object is?
- `ports` is a list, like `containers`. What does one item look like?

### 3. Configure it with a command and env

**Goal:** make the container in `3_greeter.yaml` log at debug level and
greet with `hello from Kubernetes`.

- Run `./podinfo` with the flag `--level=debug`, using `command` (and
  `args`, if you like).
- Set the environment variable `PODINFO_UI_MESSAGE` to
  `hello from Kubernetes`.

```bash
kubectl port-forward pod/greeter 9898:9898 &
curl -s localhost:9898 | grep message
# => "message": "hello from Kubernetes",
```

**Think it through:**
- podinfo has no `ENTRYPOINT`. What happens to `./podinfo` if you set only
  `args`?
- `env` is a list of maps, each with a `name` and a `value`.

## Run it for real

Optional, and the best way to make this stick. You need Docker and
`kubectl`; *kind* runs a whole Kubernetes cluster inside one Docker
container. Run everything from this exercise folder.

```bash
nix shell nixpkgs#kind
kind create cluster --name gw-pods --wait 60s  # waits until the node is Ready
kubectl get nodes                           # gw-pods-control-plane   Ready
kubectl apply -f 1_hello.yaml -f 2_web.yaml -f 3_greeter.yaml
kubectl get pods --show-labels              # wait until all three are Running
kubectl logs greeter                        # podinfo's JSON log lines
kubectl port-forward pod/greeter 9898:9898  # leave it running
```

In a second terminal, `curl -s localhost:9898` shows your message. Stop the
port-forward with Ctrl+C, then:

```bash
kubectl delete pod hello
kubectl get pods                            # hello is gone, and stays gone
```

Nothing brings `hello` back: a bare Pod has no one looking after it.
*Deployments: replicas, labels and selectors* fixes that. Clean up when
you're done:

```bash
kind delete cluster --name gw-pods
```

If `kind create cluster` fails with `could not find a log line that matches
"Reached target .*Multi-User System.*"`, Docker logs to journald and kind
can't read the node's output back. On NixOS, add
`virtualisation.docker.logDriver = "json-file";` to your configuration, run
`sudo nixos-rebuild switch`, and try again.

## Learn more

- [Kubernetes: Pods](https://kubernetes.io/docs/concepts/workloads/pods/)
- [Kubernetes: Define a command and arguments for a container](https://kubernetes.io/docs/tasks/inject-data-application/define-command-argument-container/)
- [kind: Quick start](https://kind.sigs.k8s.io/docs/user/quick-start/)
- [podinfo on GitHub](https://github.com/stefanprodan/podinfo)
