# Dockerfiles: fix the slow builds

The team's images build correctly, but CI spends minutes rebuilding things
that didn't change, and one build fails at random. The fix is almost always
the order of the lines.

**New here:** layers, the build cache, cache-friendly ordering,
`go mod download`, one-`RUN` package installs.

## Learn

### Every instruction makes a layer

An image is a stack of *layers*. Each `RUN`, `COPY` and `ADD` saves what it
changed as one new layer on top of the ones before it (`ENV`, `CMD` and
friends only change settings).

```text
  RUN go build ...        layer 4
  COPY . .                layer 3
  WORKDIR /src            layer 2
  FROM golang:1.26        layers of the base image
```

### The build cache

The builder keeps every layer it makes. On the next build, for each
instruction in order, it asks: *same parent layer, same instruction, and for
`COPY` the same file contents?* If yes, it reuses the saved layer instantly
(`CACHED` in the build output). The first time the answer is no, it reruns
that step **and every step after it**, because each later layer sits on top
of one that changed.

```text
  edit one .go file, rebuild:

  FROM golang:1.26     cached
  WORKDIR /src         cached
  COPY . .             MISS   (a file in the context changed)
  RUN go mod download  rerun  (parent changed: 40 s of downloads, again)
  RUN go build ...     rerun
```

`COPY . .` copies the whole build context, so it misses on **every
commit**. Anything below it reruns every time.

### Finding the step that breaks the cache

When a build is slow, read the step list it prints, top to bottom:

```text
 => CACHED [2/5] WORKDIR /src
 => [3/5] COPY . .                    0.4s
 => [4/5] RUN go mod download        41.2s
```

1. Find the first step **without** `CACHED`. That's where the cache broke.
2. Ask: *which input of this step changed?* For `COPY . .`, any file.
3. Look at the slow steps below it. Could they move above it?

Here `go mod download` (41 s) sits below `COPY . .`, so it reruns on every
commit. That's task 1.

### Order: what changes least goes first

Put slow steps whose inputs rarely change *above* the steps whose inputs
change on every commit:

| Changes... | Examples | Put it |
|---|---|---|
| almost never | base image, OS packages, tools | first |
| when dependencies change | `go.mod`, `go.sum`, `go mod download` | next |
| on every commit | `COPY . .`, `go build` | last |

`go mod download` reads only `go.mod` and `go.sum`, so it can run before
the code is there. Copy just those two files first:

```dockerfile
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
```

That's the same trick for a Node.js project: the slow install stays cached
until the dependency files change.

> **Why is it written like this?**
> - **Docker requires:** a `COPY` is cached by the *contents* of the files
>   it copies. Copying two files that rarely change gives a layer that
>   rarely misses.
> - **Convention:** `COPY go.mod go.sum ./` with the trailing `/`, so the
>   destination is clearly a folder (the current `WORKDIR`).

### Installing OS packages: one `RUN`

On Debian images, `apt-get update` downloads the list of available
packages, and `apt-get install` installs from that list.

```dockerfile
RUN apt-get update
RUN apt-get install -y git
```

Each line is cached on its own. Weeks later you add a package to the second
line: it reruns, but `apt-get update` is still cached, so it installs from a
weeks-old list whose files the mirror has deleted: `404 Not Found` or
`Unable to locate package`. Do both in **one** `RUN`, so they are always
cached (or rerun) together:

```dockerfile
RUN apt-get update \
    && apt-get install -y --no-install-recommends git \
    && rm -rf /var/lib/apt/lists/*
```

| Part | Why |
|---|---|
| `\` at the end of a line | continues the instruction on the next line |
| `&&` | runs the next command only if the last one worked |
| `-y` | answers "yes" to the install prompt; there is no one to type it |
| `--no-install-recommends` | installs what you asked for, not the "recommended" extras |
| `rm -rf /var/lib/apt/lists/*` | deletes the package lists, which the running container never needs |

> **Why is it written like this?**
> - **Docker requires:** the `rm` must be in the same `RUN`. A file deleted
>   in a *later* layer is hidden, but still stored in the earlier one.
> - **Convention:** one package step per image, with `&&` at the start of
>   each continued line so they line up.

### Tools: install them before the code

Some builds need an extra tool, such as `stringer`, which `go generate` runs
to write `String` methods. `go install tool@version` needs only the version,
not your code, so it belongs above `COPY . .`, where its layer stays cached
until you change the version.

## Your task

Three of the team's Dockerfiles work but waste CI time or fail at random.
Each `TODO` says what CI is seeing. Change the order and the steps, not
what the image ends up containing.

### 1. Download modules before copying the code

**Goal:** in `1_api.Dockerfile`, make a code change reuse the cached
`go mod download`.

- `COPY go.mod go.sum ./` (just those two files, not `.`).
- Then `RUN go mod download`.
- Then `COPY . .`.
- Then the build line, unchanged: `RUN go build -o /usr/local/bin/api ./cmd/api`.

```text
  edit one .go file, rebuild:
  COPY go.mod go.sum ./   cached
  RUN go mod download     cached
  COPY . .                MISS
  RUN go build ...        rerun
```

**Think it through:**
- Which files does `go mod download` read?
- Which line misses on every commit, and what's below it now?

### 2. Install packages in one step

**Goal:** in `2_worker.Dockerfile`, install `ca-certificates` and `curl`
reliably and without leftovers.

- No `RUN` that runs `apt-get update` without `apt-get install`.
- One `RUN` that runs `apt-get update`, then
  `apt-get install -y --no-install-recommends ca-certificates curl`, then
  `rm -rf /var/lib/apt/lists/*`, joined with `&&`.
- Splitting it over several lines with `\` is fine (and easier to read).

**Think it through:**
- Why does a cached `apt-get update` on its own break later installs?
- Why must the `rm` be in the same `RUN`?

### 3. Install tools before copying the code

**Goal:** in `3_tools.Dockerfile`, make code changes reuse the installed
`stringer`.

- Move `RUN go install golang.org/x/tools/cmd/stringer@v0.52.0` above
  `COPY . /app/`, unchanged.
- `go generate ./...` and the `go build` line stay below the `COPY`: they
  need the code.

**Think it through:**
- Does `go install tool@version` read anything from your project?
- Where does it go relative to `WORKDIR`? (Either place works.)

## Learn more

- [Docker build cache](https://docs.docker.com/build/cache/)
- [Optimize cache usage in builds](https://docs.docker.com/build/cache/optimize/)
- [Building best practices](https://docs.docker.com/build/building/best-practices/)
