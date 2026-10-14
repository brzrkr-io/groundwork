# Containers and images

**In one sentence:** a *container* is an ordinary Linux process that the kernel fences in, so it sees its own filesystem, network and PIDs and may use only the CPU and memory it's given; an *image* is the read-only filesystem it starts from, built from a Dockerfile and pulled from a registry.

**Why it matters:** almost everything you'll deploy ships as an image. A 20-minute CI build, a deploy that rolled out the wrong code, a scan with 300 CVEs, a token leaked to a registry and a pod stuck in `ImagePullBackOff` are all image problems, each with a small, boring fix.

**You'll learn:**
- what a container really is: processes plus namespaces plus cgroups, not a small VM
- how images stack read-only layers, and how a Dockerfile builds them
- how `.dockerignore` and the build cache make builds fast or slow
- how a multi-stage build turns a Go program into a tiny, static image
- how tags and digests name images, and why tags move
- how to run as non-root and keep secrets out of images

## What a container really is

Run `docker run nginx` on a Linux server, then `ps aux` on the server itself. nginx is right there, with an ordinary PID on the host. There's no hidden machine: a container is just processes.

Two Linux kernel features fence them in (you'll meet them again in *The operating system for programmers*):

| Feature | Controls | Example |
|---|---|---|
| *namespaces* | what the process can **see** | its own PID 1, its own `/`, network and hostname |
| *cgroups* | what the process may **use** | at most 1 CPU and 256 MiB of memory |

Inside, nginx thinks it's PID 1 on a machine holding only its own files. Outside, it's PID 48213 in a cgroup with a memory limit. Go over the limit and the kernel's *OOM killer* (out of memory) ends it at once: exit code 137, and `OOMKilled` in `kubectl describe pod`.

**A container is not a virtual machine.** A VM boots its own kernel; containers share the host's. So containers start in milliseconds, but isolation is weaker: a bug in the shared kernel can let a process escape.

## Images and layers

An *image* is a read-only filesystem plus settings: which program to start, as which user, with which environment. A container is a running instance of an image, the way a process is a running instance of a program file. One image `statusd:1.4.2` can run as 20 replicas.

An image is a stack of read-only *layers*, each a set of file changes: "add the CA certificates", "add `/statusd`". A running container gets one thin writable layer on top.

- **Layers are shared.** Ten images on the same base download and store it once.
- **Writes don't survive.** The writable layer is deleted with the container. Logs go to stdout; data goes to a database or a volume.

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
2. `WORKDIR /src` creates `/src` and moves into it.
3. `COPY . .` copies your project into `/src`.
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

**ENTRYPOINT is the program, CMD is the default arguments.** With both lines from the table, the container runs `/statusd --port=8080`, and `docker run statusd:1.4.2 --port=9090` replaces only the CMD part.

**How a container is stopped.** `docker stop` and Kubernetes send *signals*, short messages the kernel delivers to a process. SIGTERM means "please stop" and the program can react. SIGKILL, sent when SIGTERM is ignored too long (30 seconds by default in Kubernetes), can't be caught. A process killed by signal n exits with 128 + n, so SIGKILL (9) shows as **137**.

Signals go to the container's *PID 1*, its first process. ENTRYPOINT and CMD each come in two forms, and the form decides what PID 1 is:

| Form | Written as | PID 1 is |
|---|---|---|
| *exec form* | `CMD ["/statusd"]` | `/statusd` |
| *shell form* | `CMD /statusd` | `/bin/sh -c "/statusd"` |

In shell form, the shell doesn't pass SIGTERM on to your service. The grace period runs out, SIGKILL ends it mid-request, and every rollout shows exit code 137. **Use exec form.** Images without a shell, like distroless, can't run shell form at all.

**PID 1 is special:** the kernel won't apply the default "die on SIGTERM" to it. Go's runtime handles SIGTERM itself, so a Go binary exits; finishing in-flight requests first is code you'll write in *Signals: graceful stop for a rolling drain*.

## Build context and .dockerignore

In `docker build -t statusd:dev .`, the final `.` is the *build context*: the folder the build may `COPY` from. If it holds `.git`, old binaries and a `.env` file with your staging password, `COPY . .` sends them all to the builder and into the image.

A *.dockerignore* file in the context's root lists what to leave out:

```text
.git
bin/
.env
```

It also helps the cache: every commit changes `.git`, so without this file `COPY . .` misses the cache even when no code changed.

**A trade-off:** without `.git`, `go build` can't stamp the commit into the binary. Pass it in with `--build-arg VERSION=$(git rev-parse --short HEAD)` and `-ldflags "-X main.version=${VERSION}"`, as in *Packages and project layout*.

## The build cache: order matters

The builder caches every layer. Same instruction, same parent layer, and for `COPY` the same file contents? It reuses the saved layer. The first miss rebuilds that step **and every step after it**.

`RUN` is matched by its text, not its result: a cached `RUN apt-get update` is reused even after the package lists online have changed. `--no-cache` forces a fresh build.

The slow version, behind many 10-minute CI builds:

```dockerfile
COPY . .
RUN go mod download
RUN go build -o /out/statusd .
```

Edit `main.go`, `COPY . .` misses, and every dependency downloads again. Put the things that change rarely first:

```dockerfile
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /out/statusd .
```

Now editing `main.go` reruns only the last two steps; the download stays cached until `go.mod` or `go.sum` changes. (A module with no dependencies has no `go.sum`; copy just `go.mod`.) You'll reorder a slow Dockerfile in *Dockerfiles: fix the slow builds*.

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
3. `ARG VERSION` sits after the download, so a new version doesn't invalidate the cached modules.
4. The second `FROM` starts over from *distroless*: CA certificates, time zones and a user list, but no shell and no package manager.
5. `COPY --from=build` takes one file from the first stage; the rest is thrown away. The final image is a few megabytes.

*Scratch*, the empty image (`FROM scratch`), is smaller still, but you bring everything yourself. Forget the CA certificates and every HTTPS call fails with `x509: certificate signed by unknown authority`:

```dockerfile
FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /out/statusd /statusd
USER 65532:65532
ENTRYPOINT ["/statusd"]
```

**Smaller is safer.** A scanner such as Trivy lists the known vulnerabilities, *CVEs*, in every package in your image. A full Debian base always has findings; distroless and scratch have almost nothing to find, and pull faster when 40 pods restart on new nodes. The price: no shell to `docker exec` into. Practise both in *Multi-stage builds*.

## Naming images: tags and digests

A *registry* stores images: Docker Hub, GitHub's ghcr.io, AWS's ECR. A full image reference has four parts:

```text
  ghcr.io / acme/statusd : 1.4.2 @ sha256:9f2c...e41a
  registry  repository     tag     digest
```

No registry means Docker Hub: `golang:1.26` is `docker.io/library/golang:1.26`.

A *tag* is a movable label. Anyone with push access can point `1.4.2` at new content, and `golang:1.26` moves to each patch release on purpose. `latest` isn't "the newest"; it's just the tag you get when you don't write one.

A *digest* never moves. It's a SHA-256 hash of the image's content (settings and every layer): change one byte and the digest changes. In `golang:1.26@sha256:...` the tag is for humans; only the digest is used to pull.

| You deploy | A restart next week runs | Rollback |
|---|---|---|
| `statusd:latest` | whatever `latest` is now | guesswork |
| `statusd:1.4.2` | the same, unless someone re-pushed | redeploy `1.4.1` |
| `statusd@sha256:...` | exactly the same bytes | exact |

Tag releases with versions or commit SHAs, and pin base images by digest when builds must be reproducible.

## Running as non-root

By default a container runs as root, UID 0, and without user namespaces that's root on the node too. One kernel or runtime bug and an attacker in your service owns the node.

The fix is one line: `USER 65532:65532`, an unprivileged user and group (distroless calls it `nonroot`).

- **Use numbers, not names.** Kubernetes' `runAsNonRoot: true` refuses to start a root container, and can only check a numeric `USER`.
- **Mind permissions.** Non-root can't write root-owned directories, and on some runtimes can't listen below port 1024. Listen on 8080; write to a volume or `/tmp`.

## Secrets never go in images

Anyone who can pull an image can read every layer.

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

**At run time**, inject secrets as environment variables or mounted files, so one image runs in staging and production with different credentials. If a secret was ever pushed in an image, rotate it; deleting the tag isn't enough. Practise this in *Image security: non-root and no secrets*.

## Containers in Kubernetes

Kubernetes runs images; it doesn't build them. A *Pod* is one or more containers on one node, each started from an image reference. The node's kubelet pulls the image unless it's already there, creates the namespaces and a cgroup with the Pod's limits, and starts ENTRYPOINT + CMD as PID 1.

- **`imagePullPolicy`**: `IfNotPresent` is the default for a version tag; `Always` for `latest` or no tag.
- **`ImagePullBackOff`**: the pull failed: a typo, a tag never pushed, or missing registry credentials. Read the events in `kubectl describe pod`.
- **Limits are cgroups.** Go over `memory: 128Mi` and the container is OOM-killed with exit code 137.

## Common mistakes

- **`COPY . .` before `go mod download`.** Every code change re-downloads all dependencies.
- **No `.dockerignore`.** Builds upload `.git`, and `.env` files land in the image.
- **Shipping the builder.** A final stage `FROM golang` carries a compiler and hundreds of CVEs.
- **Forgetting `CGO_ENABLED=0` on scratch or distroless static.** The binary fails with "no such file or directory".
- **Shell-form `CMD`.** The shell swallows SIGTERM; every rollout ends in SIGKILL.
- **Deploying `latest`.** You can't tell what's running or roll back cleanly.
- **Running as root.** One escape bug away from owning the node.
- **Secrets in `ENV`, `ARG` or a copied file.** They stay in the layers forever.

## Check yourself

1. What two Linux kernel features turn an ordinary process into a container, and what does each control?
2. A Go service's Dockerfile runs `COPY . .`, then `RUN go mod download`, then `RUN go build`. Every CI build re-downloads all dependencies, even for a one-line change. Why, and how do you reorder it?
3. A Dockerfile copies `.env`, uses it in a `RUN` step, then runs `RUN rm .env`. Is the secret gone from the image?
4. Two pods started a week apart from `statusd:latest` behave differently. How can that happen, and what should the deployment reference instead?
5. A Go service in a `FROM scratch` image fails every HTTPS call with `x509: certificate signed by unknown authority`. What's missing?
6. A Dockerfile ends with `CMD /statusd` (shell form). Every Kubernetes rollout waits 30 seconds per pod, and old pods exit with code 137. Why?
7. Why is a container running as root riskier than it looks, and what Dockerfile line fixes it?
8. A Go project has no `.dockerignore`, and its Dockerfile runs `COPY . .`. Name three problems this causes.

## Answers

1. *Namespaces* control what it can see (PIDs, filesystem, network); *cgroups* control what it may use (CPU, memory).
2. Any code change misses the cache at `COPY . .`, so every later step reruns. Copy `go.mod` and `go.sum`, run `go mod download`, then `COPY . .` and build.
3. No: the `COPY` layer still holds `.env`, and `rm` only hides it. Use a secret mount and rotate the secret.
4. Someone pushed a new image under the movable `latest` tag in between. Deploy a version tag, or better the digest (`statusd@sha256:...`).
5. CA certificates: copy `/etc/ssl/certs/ca-certificates.crt` from the build stage, or use distroless static.
6. The shell is PID 1 and doesn't forward SIGTERM, so after the grace period SIGKILL (128 + 9 = 137) ends it. Use exec form, `CMD ["/statusd"]`.
7. Containers share the host kernel, so root inside is root on the node after one escape bug. `USER 65532:65532` runs as an unprivileged numeric user.
8. Slow builds and big images (`.git`, binaries), secrets like `.env` baked into a layer, and a cache miss on every commit because `.git` changed.

## Key terms

| Term | Meaning |
|---|---|
| container | ordinary processes isolated by namespaces and limited by cgroups, sharing the host's kernel |
| namespace | kernel feature controlling what a process can see: PIDs, filesystem, network, hostname |
| cgroup | kernel feature limiting what processes may use: CPU, memory, process count |
| OOM killer | the kernel feature that kills a process using more memory than allowed; exit code 137 |
| image | a read-only stack of filesystem layers plus settings, used to start containers |
| layer | one read-only set of file changes in an image, made by an instruction like `COPY` or `RUN` |
| Dockerfile | the recipe for an image: instructions like `FROM`, `COPY` and `RUN`, run top to bottom |
| signal | a short message the kernel delivers to a process: SIGTERM asks it to stop, SIGKILL forces it (exit 137) |
| PID 1 | a container's first process; stop signals go to it, and SIGTERM doesn't kill it by default |
| exec form | `CMD ["/app"]`: the program runs directly as PID 1 and receives signals |
| shell form | `CMD /app`: `/bin/sh` runs as PID 1 and doesn't pass SIGTERM on |
| build context | the folder sent to the image builder, which `COPY` may read |
| .dockerignore | a file listing paths to leave out of the build context |
| multi-stage build | a Dockerfile with several `FROM` stages that copies only the build result into the final image |
| static binary | a program that needs no shared libraries at run time; Go builds one with `CGO_ENABLED=0` |
| distroless | minimal base images with no shell or package manager, just what a program needs to run |
| scratch | the empty base image: nothing but what you copy in |
| CVE | a public ID for a known security vulnerability; scanners list the ones in your image's packages |
| registry | a server that stores and serves images, such as Docker Hub, ghcr.io or ECR |
| tag | a movable label for an image, like `1.4.2` or `latest` |
| digest | `sha256:...`, a hash of an image's content; the same digest always means the same bytes |
| secret mount | `RUN --mount=type=secret`: a secret available to one build step, never stored in a layer |
| Pod | the Kubernetes unit that runs one or more containers together on one node |

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
