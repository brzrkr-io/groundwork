# Image security: fix the audit findings

A security audit reads your Dockerfiles the way an attacker would, and its
findings land on your team as tickets. This exercise closes three common
ones: base images nobody can reproduce, images nobody can trace back to
their source, and a build context full of secrets.

**New here:** pinning by digest (`name:tag@sha256:...`), `LABEL` and the
OCI labels, `ARG` scope in multi-stage builds, ignore files for one
Dockerfile, `.dockerignore` patterns (`*`, `?`, `**`, `!`).

## Learn

### Tags move, digests don't

As the *Containers and images* guide explains, a tag is a label someone can
point at new content. `golang:1.26` moves with every patch release, and
`latest` moves whenever anyone pushes. A *digest* is the SHA-256 hash of the
image's content, so it always means the same bytes.

| You write | Docker pulls | Next month |
|---|---|---|
| `alpine` | `alpine:latest` | whatever `latest` is then |
| `alpine:3.24` | the current 3.24 | a newer 3.24 patch |
| `alpine:3.24@sha256:<64 hex>` | exactly that digest | the same bytes |

With both a tag and a digest, Docker pulls by the digest alone. The tag
is there for people: a reviewer sees "3.24", not 64 hex characters.

To find a digest you'd normally ask the registry:

```bash
docker buildx imagetools inspect alpine:3.24
# Name:      docker.io/library/alpine:3.24
# Digest:    sha256:...   <- the multi-platform digest to pin
```

> **Why is it written like this?**
> - **Docker requires:** `@sha256:` followed by all 64 lowercase hex
>   characters, after the tag. A shortened digest is an error.
> - **Convention:** keep the tag next to the digest. Update bots such as
>   Dependabot and Renovate read it, and propose new digests as pull
>   requests, so pinning doesn't mean never updating.

### Labels

`LABEL` stores `key=value` metadata in the image. `docker inspect` shows it,
and so do registries and scanners.

```dockerfile
LABEL org.opencontainers.image.title="reportd" \
      org.opencontainers.image.licenses="Apache-2.0"
```

One `LABEL` can set several pairs, on one line or continued with `\`.
Quote a value with spaces in it. Keys are reverse domain names; the
`org.opencontainers.image.*` keys are a standard every tool understands:

| Key | Value |
|---|---|
| `org.opencontainers.image.source` | the URL of the source repository |
| `org.opencontainers.image.revision` | the commit the image was built from |
| `org.opencontainers.image.version` | the release version, such as `1.8.0` |
| `org.opencontainers.image.title` | a human-readable name |

GitHub's registry, ghcr.io, also uses `source` to link the image to its
repository.

### ARG in multi-stage builds

A label's value can come from a build argument: `LABEL build.channel="$CHANNEL"`.
But an `ARG` only exists in the stage that declares it, from that line on:

```dockerfile
ARG CHANNEL=stable              # global: only FROM lines can see it...

FROM alpine:3.24 AS fetch
ARG CHANNEL                     # ...until a stage declares it again
RUN echo "fetching $CHANNEL"    # => fetching stable

FROM scratch
COPY --from=fetch /etc/os-release /
LABEL build.channel="$CHANNEL"  # empty: this stage never declared CHANNEL
```

| Declared | Usable in |
|---|---|
| before the first `FROM` | `FROM` lines only, until a stage repeats `ARG NAME` |
| in a stage | that stage, from that line on, and stages built `FROM` it |

`--build-arg CHANNEL=beta` sets the value for every stage that declares
`CHANNEL`. A stage that only does `COPY --from=fetch` inherits nothing.

> **Why is it written like this?**
> - **Docker requires:** double quotes or none around `$NAME`. Single
>   quotes keep the text as written: `'$CHANNEL'` is the 8 characters
>   `$CHANNEL`.
> - **Docker requires:** labels in the final stage. Labels from a stage
>   you only `COPY --from` never reach the image.
> - **Convention:** `${NAME}` and `$NAME` mean the same; braces make it
>   clear where the name ends.

### An ignore file for one Dockerfile

Everything in the *build context* (the `.` in `docker build .`) is sent to
the builder, and `COPY . .` copies it into a layer. In a multi-stage build
that layer isn't in the final image, but it still matters: CI builders are
often shared, the build cache keeps every layer (and `--cache-to` pushes the
cache to a registry), and the files are one careless `COPY --from` away from
shipping.

An ignore file lists what to leave out. Next to `api.Dockerfile`, a file
named `api.Dockerfile.dockerignore` is that Dockerfile's own ignore file,
and BuildKit uses it instead of the context's `.dockerignore`.

### Ignore patterns

One pattern per line. Patterns start at the root of the context.

| Pattern | Leaves out |
|---|---|
| `tmp` | the file or folder `tmp` at the root, and everything inside it |
| `/tmp/` | the same: leading and trailing slashes don't matter |
| `*.log` | `app.log`, but **not** `logs/app.log`: `*` never crosses a `/` |
| `logs/*.log` | `.log` files directly in `logs/` |
| `**/*.log` | `.log` files in any folder, at any depth, the root included |
| `report-?.csv` | `report-1.csv`, `report-a.csv`: `?` is one character |
| `!keep.log` | brings `keep.log` back, if an earlier line left it out |

- A line starting with `#` is a comment; blank lines are skipped.
- Every line is checked in order, and **the last line that matches wins**.
  So an exception (`!`) goes after the pattern it makes an exception to.
- A pattern that matches a folder leaves out everything in it.

> **Why is it written like this?**
> - **Docker requires:** the `#` in the first column. Docker reads
>   `  # note` as a pattern.
> - **Convention:** a comment above each group saying why it's left out,
>   so nobody deletes a line without knowing what it protects.
> - It looks like `.gitignore`, but differs: in `.gitignore`, `*.log`
>   matches at any depth and `tmp/` means "folders only".

## Your task

The security team audited the three images of the ingest pipeline: `api`
receives uploads, `worker` processes them, and `ingest` loads the results
into the database. Each task closes one finding. You don't need Docker:
the tests read your files.

### 1. Pin the base images

**Goal:** in `1_api.Dockerfile`, pin both base images by digest, keeping a
readable tag.

These are the exercise's approved digests. They're made up, so you can't
look them up or pull them; copy them from here:

| Image | Tag | Approved digest |
|---|---|---|
| `golang` | `1.26` | `sha256:727841129a8d3bbd6b99dc9f63427b62bb5a98f6e6e856a5b1399bf19152c03e` |
| `gcr.io/distroless/static-debian13` | `nonroot` | `sha256:f0b02dd57038ad9b032fcb6ad271ebeaf4cc1804bd067250270908b22f4b0114` |

- Every `FROM` that names an image (not an earlier stage, such as
  `FROM build`) is written `name:tag@digest`, with exactly that image's tag
  and digest from the table.
- No `latest`, and no image without a tag.
- Both stages stay. The final stage moves from `latest` to the `nonroot`
  tag, whose user is the same UID 65532 that `USER` already sets.

```text
FROM <image>:<tag>@sha256:<64 hex characters> AS build
```

**Think it through:**
- Which part of the reference does Docker pull by, and which part is for
  the reviewer?
- Copy the digest from the table; a single wrong character is a different
  image.

### 2. Label the image

**Goal:** in `2_worker.Dockerfile`, label the final image with its source
repository, commit and version.

The final stage sets these three labels (one `LABEL` or several, quoted
with double quotes or not at all):

| Key | Value |
|---|---|
| `org.opencontainers.image.source` | exactly `https://github.com/acme/worker` |
| `org.opencontainers.image.revision` | the build argument `REVISION`: `$REVISION` or `${REVISION}` |
| `org.opencontainers.image.version` | the build argument `VERSION`: `$VERSION` or `${VERSION}` |

- The final stage declares `ARG REVISION` and `ARG VERSION` (a default is
  optional) before the `LABEL` that uses them.
- No `LABEL` in the build stage.

```bash
docker inspect --format '{{json .Config.Labels}}' ghcr.io/acme/worker:1.8.0
# => {"org.opencontainers.image.revision":"9f2c41e...","org.opencontainers.image.source":"https://github.com/acme/worker","org.opencontainers.image.version":"1.8.0"}
```

**Think it through:**
- The build stage already has `ARG VERSION=dev`. Can the final stage see
  it?
- The build command at the top of the file passes both values. Which
  stage receives them?

### 3. Keep secrets out of the build context

**Goal:** write `3_ingest.Dockerfile.dockerignore` so that building
`3_ingest.Dockerfile`, which does `COPY . .`, leaves the secrets and junk
out and keeps the source.

`testdata/context.txt` lists every file in the ingest repository. Your
patterns must:

- **leave out** everything under `.git/`, `.env`, `.env.production`,
  `deploy/tls/server.key`, `deploy/tls/server.pem`, everything under
  `bin/`, and `secrets/db.txt`;
- **keep** `go.mod`, `go.sum`, `main.go`, `README.md` (`main.go` embeds it
  as its `--help` text, as in *Embedding files*),
  `internal/parser/parser.go` and `internal/store/store.go`.

Every other file is your call. The test decides like Docker does (its rules
are in `dockerignore_test.go`) and names each file that is wrongly kept or
wrongly left out.

```text
--- FAIL: TestKeepSecretsOutOfContext
    3_ingest.Dockerfile.dockerignore keeps deploy/tls/server.key in the build context, want it excluded: it's the TLS private key
```

**Think it through:**
- Does `*.key` reach a file two folders down?
- One pattern per folder can cover everything inside it.
- Would `*.md` leave out a file the build needs?

## Learn more

- [Build context and .dockerignore files (Docker docs)](https://docs.docker.com/build/concepts/context/)
- [Pin base image versions (Docker docs)](https://docs.docker.com/build/building/best-practices/#pin-base-image-versions)
- [Build variables: scoping (Docker docs)](https://docs.docker.com/build/building/variables/#scoping)
- [OCI image annotations](https://github.com/opencontainers/image-spec/blob/main/annotations.md)
