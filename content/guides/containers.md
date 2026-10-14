# Containers and images

**In one sentence:** a *container* is an ordinary Linux process that the kernel fences in, so it sees its own filesystem, network and PIDs and may use only the CPU and memory it's given; an *image* is the read-only filesystem it starts from, built from a Dockerfile and pulled from a registry.

**Why it matters:** almost everything you'll deploy ships as an image. A 20-minute CI build, a deploy that rolled out the wrong code, a scan with 300 CVEs, a token leaked to a registry and a pod stuck in `ImagePullBackOff` are all image problems, each with a small, boring fix.

**You'll learn:**
- what a container really is: processes plus namespaces plus cgroups, not a small VM
- how images stack read-only layers, and how a Dockerfile builds them
- how the build context, `.dockerignore` and the build cache make builds fast or slow
- how a multi-stage build turns a Go program into a tiny, static image
- how tags, digests and registries name images, and why tags move
- how to run as non-root, keep secrets out of images, and what Kubernetes does with your image

## What a container really is

Run `docker run nginx` on a Linux server, then `ps aux` on the server itself. nginx is right there, with an ordinary PID on the host. There's no hidden machine: a container is just processes.

Two kernel features, which you met in *The operating system for programmers*, fence them in:

| Feature | Controls | Example |
|---|---|---|
| *namespaces* | what the process can **see** | its own PID 1, its own `/`, network and hostname |
| *cgroups* | what the process may **use** | at most 1 CPU and 256 MiB of memory |

Inside, nginx thinks it's PID 1 on a machine holding only its own files. Outside, it's PID 48213 in a cgroup with a memory limit. Go over the limit and the OOM killer ends it: exit code 137, `OOMKilled` in `kubectl describe pod`.

**A container is not a virtual machine.** A VM boots its own kernel; containers share the host's.

```text
  virtual machines                     containers
  ┌────────┐ ┌────────┐                ┌────────┐ ┌────────┐
  │  app   │ │  app   │                │  app   │ │  app   │
  │ libs   │ │ libs   │                │ libs   │ │ libs   │
  │ kernel │ │ kernel │                └────────┘ └────────┘
  └────────┘ └────────┘                 namespaces + cgroups
     hypervisor                         ─────────────────────
     host kernel                         one shared host kernel
```

So containers start in milliseconds, but isolation is weaker: a bug in the shared kernel can let a process escape. Remember that when you reach running as non-root.

## Images and containers

An *image* is a packaged, read-only filesystem plus settings: which program to start, as which user, with which environment. A container is a running instance of an image, the way a process is a running instance of a program file. One image `statusd:1.4.2` can run as 20 replicas.

### Layers

An image is a stack of read-only *layers*, each a set of file changes: "add `/statusd`", "add the CA certificates". When a container starts, the runtime adds one thin *writable layer* on top and shows the process all of them as one filesystem.

```text
  ┌──────────────────────────────┐  writable layer (this container only)
  ├──────────────────────────────┤
  │ COPY statusd /statusd        │  read-only layer
  │ COPY ca-certificates.crt     │  read-only layer
  │ base image files             │  read-only layer(s)
  └──────────────────────────────┘
```

- **Layers are shared.** Ten images on the same base download and store it once. A node that has most layers pulls a new version in seconds.
- **Writes don't survive.** The writable layer is deleted with the container, so every deploy starts clean. Logs go to stdout; data goes to a database or a volume.

## Writing a Dockerfile

A *Dockerfile* is the recipe for an image: instructions run top to bottom, most adding a layer or a setting.

```dockerfile
FROM golang:1.26
WORKDIR /src
COPY . .
RUN go build -o /usr/local/bin/statusd .
ENV APP_ENV=production
EXPOSE 8080
CMD ["statusd"]
```

1. `FROM` starts from the official Go image: Debian plus the Go toolchain.
2. `WORKDIR /src` creates `/src` and makes it the working directory.
3. `COPY . .` copies your project from the build context into `/src`.
4. `RUN` runs a command at build time and saves what it changed as a layer.
5. `ENV` sets an environment variable every container starts with.
6. `EXPOSE` only documents the port; `docker run -p 8080:8080` publishes it.
7. `CMD` says what to run when a container starts.

| Instruction | What it does | When |
|---|---|---|
| `ARG VERSION=dev` | a variable set with `--build-arg VERSION=1.4.2` | build only |
| `USER 65532:65532` | the user later steps and the container run as | build and run |
| `ENTRYPOINT ["/statusd"]` | the program the container always runs | run |
| `CMD ["--port=8080"]` | default arguments, or the default command | run |

### CMD, ENTRYPOINT and PID 1

**ENTRYPOINT is the program, CMD is the default arguments.** With both lines from the table, the container runs `/statusd --port=8080`, and `docker run statusd:1.4.2 --port=9090` replaces only the CMD part. In Kubernetes, `command` replaces ENTRYPOINT and `args` replaces CMD.

Both come in two forms:

| Form | Written as | PID 1 is |
|---|---|---|
| *exec form* | `CMD ["/statusd"]` | `/statusd` |
| *shell form* | `CMD /statusd` | `/bin/sh -c "/statusd"` |

In shell form, the shell is PID 1. When Kubernetes stops the pod it sends SIGTERM to PID 1, and the shell doesn't pass it on. Your service never hears it, the 30-second grace period runs out, and SIGKILL ends it: exit code 137 and dropped requests on every rollout. Use exec form. Images without a shell can't run shell form at all.

## Build context and .dockerignore

In `docker build -t statusd:dev .`, the final `.` is the *build context*: the folder whose files the build may `COPY`, sent to the builder before the first instruction runs.

If it holds `.git`, old binaries and a 2 GB fixture, every build ships all of it, and `COPY . .` puts it in the image, along with that `.env` file holding your staging password.

A *.dockerignore* file in the context's root lists what to leave out:

```text
.git
bin/
*.md
.env
```

A repository with several Dockerfiles can give each its own rules: `build.Dockerfile` uses `build.Dockerfile.dockerignore`, which wins over the root `.dockerignore`.

**A trade-off:** without `.git`, `go build` can't record the commit in its `runtime/debug` build info. Pass it in with `--build-arg VERSION=$(git rev-parse --short HEAD)` and `-ldflags "-X main.version=${VERSION}"`, as in *Packages and project layout*.

## The build cache: order matters

The builder caches every layer. Same instruction, same parent layer, and for `COPY` the same file contents? It reuses the saved layer instantly. The first miss rebuilds that step **and every step after it**.

The slow version, behind many 10-minute CI builds:

```dockerfile
COPY . .
RUN go mod download
RUN go build -o /out/statusd .
```

Change one line of `main.go`, `COPY . .` misses, and every dependency downloads again. Dependencies change rarely, code every commit, so put the rare things first:

```dockerfile
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /out/statusd .
```

```text
  edit main.go and rebuild:
  COPY go.mod go.sum   cached    (files unchanged)
  RUN go mod download  cached    (parent unchanged)
  COPY . .             rebuilt   (main.go changed)
  RUN go build         rebuilt
```

A module with no dependencies has no `go.sum` yet; copy just `go.mod`. You'll fix this in *Dockerfiles: fix the slow builds*.

## Multi-stage builds for Go

`golang:1.26` is hundreds of megabytes: compiler, Git, a whole Debian. Your service needs none of it at run time. A *multi-stage build* compiles in one stage and copies only the result into a clean final stage.

```dockerfile
# syntax=docker/dockerfile:1

FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=dev
RUN CGO_ENABLED=0 go build -ldflags "-X main.version=${VERSION}" -o /out/statusd .

FROM gcr.io/distroless/static-debian13
COPY --from=build /out/statusd /statusd
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/statusd"]
```

1. `AS build` names the first stage.
2. `CGO_ENABLED=0` stops Go linking C code, so the binary is *static*: it needs no shared libraries from the image.
3. The second `FROM` starts over from *distroless*: CA certificates, time zones and a user list, but no shell and no package manager.
4. `COPY --from=build` takes one file from the first stage; the rest of that stage is thrown away.
5. The final image is your binary plus a couple of megabytes.

`FROM scratch`, the empty image, is smaller still, but you bring everything yourself. Forget the CA certificates and every HTTPS call fails with `x509: certificate signed by unknown authority`:

```dockerfile
FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /out/statusd /statusd
USER 65532:65532
ENTRYPOINT ["/statusd"]
```

Practise both in *Multi-stage builds*.

## Naming images: tags, digests and registries

A *registry* stores images: Docker Hub, GitHub's ghcr.io, AWS's ECR, or your company's own. A full *image reference* has four parts:

```text
  ghcr.io / acme/statusd : 1.4.2 @ sha256:9f2c...e41a
  ───┬───   ─────┬─────    ──┬──   ────────┬─────────
  registry   repository     tag         digest
```

No registry means Docker Hub: `golang:1.26` is `docker.io/library/golang:1.26`. An ECR reference looks like `123456789012.dkr.ecr.eu-west-1.amazonaws.com/statusd:1.4.2`.

**A tag is a movable label.** Anyone with push access can point `1.4.2` at new content, and `golang:1.26` moves to each patch release on purpose. `latest` isn't "the newest"; it's just the tag you get when you don't write one.

**A digest never moves.** It's the SHA-256 hash of the image's content: same digest, same bytes. In `golang:1.26@sha256:...` the tag is for humans; only the digest is used to pull.

| You deploy | A restart next week runs | Rollback |
|---|---|---|
| `statusd:latest` | whatever `latest` is now | guesswork |
| `statusd:1.4.2` | the same, unless someone re-pushed | redeploy `1.4.1` |
| `statusd@sha256:...` | exactly the same bytes | exact |

Tag releases with versions or commit SHAs. Pin base images by digest when builds must be reproducible, and let a bot propose digest bumps as pull requests.

## Running as non-root

By default a container runs as root, UID 0, and without user namespaces that's the same UID 0 as root on the node. An attacker who breaks into your service and finds one kernel or runtime bug lands on the node as root, next to every other team's pods.

The fix is one line: `USER 65532:65532`, an unprivileged user and group (distroless calls it `nonroot`).

- **Use numbers, not names.** Kubernetes' `runAsNonRoot: true` refuses to start a container running as UID 0, and it can only verify a numeric `USER`.
- **Mind permissions.** A non-root process can't write root-owned directories or bind ports below 1024. Listen on 8080; write to a volume or `/tmp`.

Practise this in *Image security: non-root and no secrets*.

## Secrets never go in images

Every layer is kept, and anyone who can pull the image can read all of them: CI, the registry, everyone you shared it with.

| How the secret got in | Where it's still visible |
|---|---|
| `ENV DB_PASSWORD=...` | `docker inspect` and every container's environment |
| `ARG TOKEN=...` used by `RUN` | `docker history` |
| `COPY .env .` then `RUN rm .env` | the `COPY` layer; `rm` only hides it |

**At build time**, use a BuildKit *secret mount*: a file available to one `RUN` step, never written to a layer.

```dockerfile
RUN --mount=type=secret,id=netrc,target=/root/.netrc go mod download
```

```text
docker build --secret id=netrc,src=$HOME/.netrc -t statusd:dev .
```

**At run time**, inject secrets as environment variables or mounted files from a Kubernetes Secret or a secrets manager. The same image then runs in staging and production with different credentials. If a secret was ever pushed in an image, deleting the tag isn't enough: rotate it.

## Small images, fewer CVEs

A scanner such as Trivy matches every package in your image against known vulnerabilities, *CVEs*. A full Debian base brings hundreds of packages, so its report always has findings, mostly in code you never call. Each one still costs triage time.

| Final stage | Rough size | What an attacker finds |
|---|---|---|
| `golang:1.26` | hundreds of MB | compiler, shell, Git, package manager |
| `distroless/static` | a few MB | your binary, certificates, user list |
| `scratch` | just your binary | your binary |

Smaller also pulls faster. When a node dies and 40 pods restart elsewhere, each new node downloads your image first; less to pull is a shorter outage. The price: no shell to `docker exec` into when debugging.

## Containers in Kubernetes

Kubernetes doesn't build images; it runs them. A *Pod* is one or more containers run together on one node, each from an image reference:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: statusd
spec:
  containers:
    - name: statusd
      image: ghcr.io/acme/statusd:1.4.2
      imagePullPolicy: IfNotPresent
      securityContext:
        runAsNonRoot: true
      resources:
        limits:
          memory: 128Mi
```

```text
  scheduler picks a node
    └─► kubelet reads the image reference
          ├─► container runtime pulls it, unless already there
          ├─► creates namespaces, and a cgroup with the limits
          └─► starts ENTRYPOINT + CMD as PID 1
```

- **`imagePullPolicy`**: `IfNotPresent` (default for a version tag), `Always` (default for `latest` or no tag) or `Never`.
- **`ImagePullBackOff`**: the pull failed: a typo, a tag never pushed, or a private registry with no credentials. Check the events in `kubectl describe pod`.
- **Limits are cgroups.** `memory: 128Mi` becomes a cgroup limit; go over it and the container is OOM-killed with exit code 137.

## Common mistakes

- **`COPY . .` before `go mod download`.** Every code change re-downloads all dependencies.
- **No `.dockerignore`.** Builds upload `.git`, and `.env` files land in the image.
- **Shipping the builder.** A final stage `FROM golang` carries a compiler and hundreds of CVEs.
- **Forgetting `CGO_ENABLED=0` on scratch or distroless static.** The binary needs C libraries that aren't there and fails with "no such file or directory".
- **Shell-form `CMD`.** The shell swallows SIGTERM; every rollout ends in SIGKILL.
- **Deploying `latest`.** You can't tell what's running or roll back cleanly.
- **Running as root.** One escape bug away from owning the node.
- **Secrets in `ENV`, `ARG` or a copied file.** They stay in the layers forever.

## Check yourself

1. What two Linux kernel features turn an ordinary process into a container, and what does each control?
2. A Go service's Dockerfile runs `COPY . .`, then `RUN go mod download`, then `RUN go build`. Every CI build re-downloads all dependencies, even for a one-line change. Why, and how do you reorder it?
3. A Dockerfile copies `.env`, uses it in a `RUN` step, then runs `RUN rm .env`. Is the secret gone from the image?
4. Two pods started a week apart from `statusd:latest` behave differently. How, and what should the deployment reference instead?
5. A Go service in a `FROM scratch` image fails every HTTPS call with `x509: certificate signed by unknown authority`. What's missing?
6. A Dockerfile ends with `CMD /statusd` (shell form). Every Kubernetes rollout waits 30 seconds per pod, and old pods exit with code 137. Why?
7. Why is a container running as root riskier than it looks, and what Dockerfile line fixes it?

## Answers

1. *Namespaces* control what it can see (PIDs, filesystem, network, hostname); *cgroups* control what it may use (CPU, memory). It's still an ordinary process on the host's shared kernel, not a VM.
2. The build cache reuses a layer only if everything before it is unchanged. A code change invalidates `COPY . .`, so every later step reruns. Copy `go.mod` and `go.sum`, run `go mod download`, then `COPY . .` and build: the download stays cached until dependencies change.
3. No. The `COPY` layer still contains `.env`; `rm` adds a newer layer that only hides it, and anyone who pulls the image can extract it. Use a secret mount (`RUN --mount=type=secret`) and rotate the leaked secret.
4. `latest` is just a tag, and tags move: someone pushed a new image under it in between. Deploy a version tag, or better the digest (`statusd@sha256:...`), which always means the same bytes.
5. CA certificates. `scratch` is empty, so Go can't verify servers. Copy `/etc/ssl/certs/ca-certificates.crt` from the build stage, or use distroless static, which includes them.
6. Shell form makes `/bin/sh -c` PID 1. Kubernetes sends SIGTERM to PID 1, the shell doesn't forward it, and after the grace period SIGKILL ends everything (128 + 9 = 137). Exec form, `CMD ["/statusd"]`, makes the service PID 1.
7. Containers share the host kernel, and root inside is root on the node unless user namespaces remap it, so one kernel or runtime bug hands an attacker the node. `USER 65532:65532` runs as an unprivileged numeric user that `runAsNonRoot` can verify.

## Key terms

| Term | Meaning |
|---|---|
| container | ordinary processes isolated by namespaces and limited by cgroups, sharing the host's kernel |
| namespace | kernel feature controlling what a process can see: PIDs, filesystem, network, hostname |
| cgroup | kernel feature limiting what processes may use: CPU, memory, process count |
| image | a read-only stack of filesystem layers plus settings, used to start containers |
| layer | one read-only set of file changes in an image, made by an instruction like `COPY` or `RUN` |
| build context | the files sent to the image builder, which `COPY` may read |
| .dockerignore | a file listing paths to leave out of the build context |
| build cache | saved layers reused while an instruction and its inputs are unchanged; the first miss rebuilds everything after |
| exec form | `CMD ["/app"]`: the program runs directly as PID 1 and receives signals; shell form (`CMD /app`) runs it under `/bin/sh` |
| ENTRYPOINT vs CMD | ENTRYPOINT is the program a container runs; CMD is its default, overridable arguments |
| multi-stage build | a Dockerfile with several `FROM` stages that copies only the build result into the final image |
| distroless | minimal base images with no shell or package manager, just what a program needs to run |
| registry | a server that stores and serves images, such as Docker Hub, ghcr.io or ECR |
| tag | a movable label for an image, like `1.4.2` or `latest` |
| digest | `sha256:...`, the hash of an image's content, which never changes |
| secret mount | `RUN --mount=type=secret`: a secret available to one build step, never stored in a layer |
| imagePullPolicy | the Kubernetes setting for when the kubelet pulls an image: `IfNotPresent`, `Always` or `Never` |

## Watch

- [Docker in 100 Seconds](https://www.youtube.com/watch?v=Gjnup-PuquQ) · Fireship · 2 min: Dockerfile, image and container in one quick picture.
- [Cgroups, namespaces, and beyond: what are containers made from?](https://www.youtube.com/watch?v=sK5i-N34im8) · Docker · 54 min: Jérôme Petazzoni's classic tour of the kernel features behind containers; the first half covers namespaces and cgroups.
- [Dockerfile Best Practices](https://www.youtube.com/watch?v=JofsaZ3H1qM) · Docker · 40 min: Docker engineers on cache ordering, multi-stage builds, small images and build secrets.
- [Dockerfile: From Start to Optimized (DockerCon 2023)](https://www.youtube.com/watch?v=saSJa9YVroA) · Docker · 38 min: one Dockerfile improved step by step, from slow and large to cached, small and safe.

## Go deeper

- [Building best practices](https://docs.docker.com/build/building/best-practices/): Docker's checklist for base images, multi-stage builds, cache ordering and each instruction.
- [Build context](https://docs.docker.com/build/concepts/context/): what gets sent to the builder, `.dockerignore` syntax and Dockerfile-specific ignore files.
- [Build secrets](https://docs.docker.com/build/building/secrets/): secret and SSH mounts, and why `ARG` and `ENV` are the wrong place for credentials.
- [Images](https://kubernetes.io/docs/concepts/containers/images/): how Kubernetes names, pulls and caches images, `imagePullPolicy` defaults and `ImagePullBackOff`.
- [GoogleContainerTools/distroless](https://github.com/GoogleContainerTools/distroless): what's inside distroless images, the `nonroot` user, and Go examples.
