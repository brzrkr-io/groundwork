# Image security: non-root and no secrets

An image is shared: CI builds it, a registry stores it, and every cluster
that pulls it can read every layer. Two habits keep it safe to share: the
program inside runs as an ordinary user, and no password or token is ever
written into it.

**New here:** `USER`, UIDs and GIDs, why root in a container is risky,
`ARG` vs `ENV`, `docker history`, BuildKit secret mounts
(`RUN --mount=type=secret`), `.netrc`, `GOPRIVATE`.

## Learn

### Who a container runs as

Linux knows users by number, the *UID* (user ID), and groups by the *GID*.
UID 0 is root, which may do anything. Unless the image says otherwise, a
container's program runs as UID 0.

Containers share the node's kernel, so that root is the node's root, fenced
in only by the container's walls. If an attacker gets into your service and
finds one kernel or runtime bug, they are root on the node, next to every
other team's pods. As an ordinary user, the same bug gets them much less.

### USER

`USER` sets who the rest of the stage, and the container, run as:

```dockerfile
FROM gcr.io/distroless/static-debian13
COPY --from=build /out/reportd /reportd
USER 10001:10001
ENTRYPOINT ["/reportd"]
```

| You write | Means |
|---|---|
| `USER 10001` | UID 10001; the group defaults to 0, root's group |
| `USER 10001:10001` | UID 10001, GID 10001 |
| `USER nonroot` | a name, looked up in the image's `/etc/passwd` |
| `USER root` or `USER 0` | root again |

A UID doesn't need an entry in `/etc/passwd` to work. Distroless images do
have one ready: user `nonroot` is UID 65532 with group 65532.

`USER` is a setting, like `ENV`: in a multi-stage build only the final
stage's `USER` ends up in the image, and the last `USER` in that stage is the
one that counts.

> **Why is it written like this?**
> - **Kubernetes requires:** a number. A pod can demand `runAsNonRoot`, and
>   Kubernetes then refuses to start a container whose user it can't prove
>   is not root. It can't look a name up, so `USER nonroot` fails that check
>   and `USER 65532:65532` passes.
> - **Convention:** give the group too. Without it the process runs with
>   group 0, root's group.
> - **Author's choice:** `USER` just before `ENTRYPOINT`. Docker applies it
>   wherever it sits in the final stage, but there a reader sees who the
>   program runs as right before it starts.

### Where ENV and ARG end up

| | `ENV NAME=value` | `ARG NAME=default` |
|---|---|---|
| Set with | the Dockerfile | the Dockerfile, or `docker build --build-arg NAME=value` |
| Available | in later build steps **and** in the running container | in later build steps only |
| Recorded in the image | yes: `docker inspect` shows it | its value, in `docker history`, for every `RUN` that uses it |

`docker history` lists how each layer of an image was made: one row per
instruction, with its settings written out in full.

```text
$ docker history --no-trunc --format '{{.CreatedBy}}' mailer:dev
ENTRYPOINT ["/mailer"]
ENV SMTP_PASSWORD=correct-horse
ARG SMTP_PASSWORD=correct-horse
COPY /out/mailer /mailer # buildkit
...
```

Anyone who can pull `mailer:dev` can run that. A `LABEL` is no better:
`docker inspect` shows every label. So a secret in `ENV`, `ARG` or `LABEL`
is a secret given to everyone with pull access.

A build argument passed with `--build-arg` is recorded even in a build
stage that never ships: in that stage's history, in the build cache, and in
the *provenance*, a record of how the image was built that the GitHub
Actions for Docker attach to the images they push (in full for public
repositories).

### Secrets at run time

A program that needs a password reads it from its environment when it
starts, and whoever runs the container supplies it:

```bash
docker run --rm -e SMTP_PASSWORD mailer:dev        # copies $SMTP_PASSWORD from your shell
docker run --rm -e SMTP_PASSWORD=... mailer:dev    # or gives the value directly
```

In production a Kubernetes *Secret* does the same job: it stores the value in
the cluster and hands it to the container as an environment variable. The
image itself stays the same in staging and production.

If a password was ever in an image or a commit, deleting it isn't enough:
old images and Git history keep it. Change (*rotate*) the password too.

### Secrets at build time: secret mounts

Some build steps need a credential, such as downloading private code.
BuildKit, the engine behind `docker build`, can lend one `RUN` step a secret
file that is never written into a layer:

```dockerfile
RUN --mount=type=secret,id=aws,target=/root/.aws/credentials \
    aws s3 cp s3://acme-fixtures/rates.json /data/
```

```bash
docker build --secret id=aws,src=$HOME/.aws/credentials -t rates:dev .
```

| Part | Meaning |
|---|---|
| `--secret id=aws,src=...` | on the `docker build` command: send this file, under the name `aws` |
| `--mount=type=secret` | on the `RUN` line: mount a secret for this step only |
| `id=aws` | which secret, matching the `id` of `--secret` |
| `target=/root/.aws/credentials` | where the file appears while the step runs |

The options after `--mount=` are separated by commas, in any order. After the
step the file is gone, and the layer, the cache and `docker history` hold
nothing but the command.

### Private Go modules

`go mod download` normally fetches modules through the public proxy,
`proxy.golang.org`, which can't see private repositories. `GOPRIVATE` lists
module paths to fetch straight from their repository instead:
`GOPRIVATE=github.com/acme/*` means "everything under github.com/acme".

For those, `go` runs `git`, and `git` logs in to HTTPS servers with the
credentials in `~/.netrc`, a small text file:

```text
machine git.example.com
login deploy-bot
password s3cr3t-t0ken
```

For root, `~` is `/root`, so a container's build steps look for
`/root/.netrc`.

> **Why is it written like this?**
> - **Docker requires:** `# syntax=docker/dockerfile:1` only as the very
>   first line. It asks BuildKit for the newest 1.x Dockerfile syntax, which
>   understands `RUN --mount`. Anywhere else it's an ordinary comment.
> - **Convention:** a secret's `id` names what it is (`aws`, `netrc`), and
>   its `target` is where the tool already looks, so the command needs no
>   extra options.

## Your task

Three services want into the production cluster, and the platform review
blocked all three. You don't need Docker: the tests read your Dockerfiles.

### 1. Run as a non-root user

**Goal:** in `1_api.Dockerfile`, make the final stage run the api as a
numeric non-root user.

- The last `USER` in the final stage has a numeric, non-zero UID. A group
  is optional; if you give one, it's a number too.
- That `USER` comes before `ENTRYPOINT`.
- Nothing after it switches back to root (`USER root` or `USER 0`).
- A `USER` in the build stage doesn't count.

```bash
docker build -f 1_api.Dockerfile -t api:dev .
docker inspect --format '{{.Config.User}}' api:dev
# => the UID:GID you chose
```

**Think it through:**
- Which stage's settings end up in the image?
- Distroless already has an unprivileged user. What is its number?

### 2. Keep the password out of the image

**Goal:** in `2_migrate.Dockerfile`, delete the baked-in database password,
so it arrives only at run time.

- No `ENV`, `ARG` or `LABEL` anywhere in the file has a name containing
  `PASSWORD`, `TOKEN`, `SECRET` or `KEY` (in any case).
- The password `hunter2` appears nowhere in the file, **comments
  included**: the Dockerfile is in Git, and anyone who can read the
  repository can read a comment. Delete the lines; don't comment them out.
- Keep the settings that aren't secret: the final stage still sets
  `DATABASE_HOST=db.internal` and `DATABASE_USER=migrate`.

```bash
DATABASE_PASSWORD=... docker run --rm -e DATABASE_PASSWORD migrate:dev
docker history --no-trunc migrate:dev | grep -c PASSWORD
# => 0
```

**Think it through:**
- Two instructions put the password in, and a comment repeats it.
- Where does the program get `DATABASE_PASSWORD` from once the image no
  longer sets it?

### 3. Use a build secret for private modules

**Goal:** in `3_private_deps.Dockerfile`, download billing's private modules
with a secret mount instead of a token in a build argument.

- No `ARG` or `ENV` has a name containing `TOKEN`, and no other instruction
  mentions a token any more (no `GITHUB_TOKEN`, no `x-access-token`).
- In the build stage, a `RUN` has the flag
  `--mount=type=secret,id=netrc,target=/root/.netrc` (options in any order;
  extra ones such as `required=true` are fine), and its command is exactly
  `go mod download`.
- Keep `ENV GOPRIVATE=github.com/acme/*` in the build stage, before that
  `RUN`. It's already there; the test checks it stays.

The netrc file lives on the build machine and holds the GitHub login:

```text
machine github.com
login x-access-token
password <a GitHub token>
```

```bash
docker build -f 3_private_deps.Dockerfile --secret id=netrc,src=$HOME/.netrc -t billing:dev .
```

**Think it through:**
- Where does a build argument's value end up? Where does a mounted secret's?
- The `git config` line only existed to hand git the token. Who hands it
  over now?

## Learn more

- [Build secrets (Docker docs)](https://docs.docker.com/build/building/secrets/)
- [Dockerfile reference: USER](https://docs.docker.com/reference/dockerfile/#user)
- [Pod Security Standards (Kubernetes docs)](https://kubernetes.io/docs/concepts/security/pod-security-standards/)
- [Go Modules Reference: private modules](https://go.dev/ref/mod#private-modules)
