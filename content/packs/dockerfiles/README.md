# Dockerfiles

*Project: inventory service* closed *Production services* with a service
that's ready to run. This module packages it as a container, so CI/CD can
build and ship it automatically.

Almost everything you deploy as a platform engineer ships as a *container
image*. A `Dockerfile` is the recipe for one: a short list of instructions
that says what to start from, what to copy in, what to run, and what the
container does when it starts.

**New here:** images and containers, `FROM`, `WORKDIR`, `COPY`, `RUN`,
`CMD`, `ENV`, `EXPOSE`, exec form.

## Learn

### Images and containers

An *image* is a frozen filesystem plus a little metadata: which program to
start, which environment variables to set. A *container* is a running copy
of an image: an ordinary process that sees only the image's files.

You build an image from a Dockerfile, then run it:

```bash
docker build -f 2_build.Dockerfile -t statusd:dev .
docker run --rm -p 8080:8080 statusd:dev
```

`-f` picks the Dockerfile, `-t` gives the image a name and a *tag*
(`statusd:dev`), and the `.` at the end is the *build context*: the folder
whose files `COPY` can see. You don't need Docker for this exercise. The
tests read your Dockerfiles; they don't build them.

### The shape of a Dockerfile

One instruction per line: a keyword in capitals, then its arguments.
Lines starting with `#` are comments.

```dockerfile
# A throwaway image that prints the date.
FROM alpine:3.24
CMD ["date", "-u"]
```

`FROM` always comes first. It names the *base image* you start from, as
`name:tag`. The tag picks the version: `alpine:3.24` is Alpine Linux 3.24.
The official `golang` images are tagged by Go version, so `golang:1.26` has
Go 1.26 and its tools installed.

`CMD` says what the container runs when it starts.

### The instructions you need

| Instruction | What it does | Example |
|---|---|---|
| `FROM image:tag` | start from a base image | `FROM alpine:3.24` |
| `WORKDIR /path` | create a folder and `cd` into it for every later instruction | `WORKDIR /app` |
| `COPY src dest` | copy files from the build context into the image | `COPY config.yaml /etc/app/` |
| `RUN command` | run a command *while building*, and keep the files it makes | `RUN apk add curl` |
| `CMD ["prog", "arg"]` | the command the container runs *when it starts* | `CMD ["date", "-u"]` |
| `ENV KEY=value` | set an environment variable for later instructions and for the container | `ENV TZ=UTC` |
| `EXPOSE port` | document which port the program listens on | `EXPOSE 9090` |

`COPY . .` copies the whole build context into the current `WORKDIR`. The
first `.` is "everything in the context", the second is "here".

`RUN` and `CMD` are easy to mix up. `RUN` happens once, on the build
machine, and its result is saved in the image. `CMD` happens every time a
container starts.

> **Why is it written like this?**
> - **Docker requires:** `FROM` before anything else (only `ARG` may come
>   earlier). An image must start from something, even if that something is
>   the empty image `scratch`.
> - **Docker requires:** `CMD ["date", "-u"]` with square brackets and
>   double quotes is JSON, called the *exec form*. Docker starts the program
>   directly. Without brackets (`CMD date -u`, the *shell form*), Docker wraps
>   it in `/bin/sh -c`, so the shell, not your program, receives stop signals
>   such as the SIGTERM you handled in *Signals: graceful stop for a rolling drain*.
> - **Convention:** instructions in capitals. Docker accepts `from`, but
>   every Dockerfile you'll read uses `FROM`.
> - **Convention:** `WORKDIR` instead of `RUN cd /app`. Each `RUN` starts a
>   fresh shell, so a `cd` is forgotten at the end of its line.

### Images build on images

`FROM` can name *any* image, including one your team built. If
`statusd:dev` already has the program and its `CMD`, this is a complete
Dockerfile:

```dockerfile
FROM statusd:dev
ENV LOG_LEVEL=debug
```

The new image has everything from `statusd:dev`, the extra setting, and the
same `CMD`, because settings are inherited from the base image.

### Settings a program reads

Containers are usually configured with environment variables. `ENV` bakes a
default into the image; `docker run -e KEY=value` or a Kubernetes manifest
can still override it.

```dockerfile
ENV TZ=UTC LOG_FORMAT=json
```

One `ENV` can set several variables, separated by spaces. `EXPOSE 9090`
changes nothing at run time: it is documentation for people and tools,
saying "the program inside listens on 9090".

## Your task

`statusd` (in `main.go`) is the small health service behind the team's
deploy dashboard: one handler, as in *HTTP servers: handlers and routes*.
It listens on the port in `$PORT` (default `8080`) and answers `/healthz`
with `ok (<APP_ENV>)`. You don't need to change it. Your job is three
Dockerfiles. Each starts with a comment saying what it's for.

### 1. Check the Go version

**Goal:** in `1_go_version.Dockerfile`, write an image that starts from
`golang:1.26` and runs `go version` when it starts.

- Exactly one `FROM`, with the image `golang:1.26`.
- A `CMD` in exec form that runs `go` with the argument `version`.

```bash
docker build -f 1_go_version.Dockerfile -t goversion .
docker run --rm goversion
# => go version go1.26.2 linux/amd64
```

**Think it through:**
- Which instruction picks the starting image, and which one says what runs
  when the container starts?
- In exec form, the program and each argument are separate strings in the
  list.

### 2. Build statusd

**Goal:** in `2_build.Dockerfile`, build `statusd` from source and run it.

In this order:

1. Start from `golang:1.26`.
2. Set the working directory to `/src`.
3. Copy the whole build context into it.
4. Build the program with `go build -o /usr/local/bin/statusd .`
5. Run `statusd` when the container starts, in exec form. Either
   `statusd` or the full path `/usr/local/bin/statusd` is fine.

**Think it through:**
- Which step happens while building (`RUN`) and which when the container
  starts (`CMD`)?
- `go build` needs the source files. What has to happen before it?
- `/usr/local/bin` is on the `PATH`, which is why the short name works.

### 3. Configure it for production

**Goal:** in `3_configure.Dockerfile`, start from the `statusd:dev` image
that task 2 builds and set its production defaults.

The `FROM statusd:dev` line is already there. Add:

- the environment variable `APP_ENV` with the value `production`,
- the environment variable `PORT` with the value `8080`,
- an `EXPOSE` for port `8080`.

Don't add a `CMD`: the one from `statusd:dev` is inherited.

```bash
docker build -f 3_configure.Dockerfile -t statusd:prod .
docker run --rm -p 8080:8080 statusd:prod
curl localhost:8080/healthz
# => ok (production)
```

**Think it through:**
- Which instruction sets environment variables? Can one line set both?
- `EXPOSE` takes just the port number.

## Learn more

- [Dockerfile reference](https://docs.docker.com/reference/dockerfile/)
- [Writing a Dockerfile (Docker docs)](https://docs.docker.com/get-started/docker-concepts/building-images/writing-a-dockerfile/)
- [The official golang image](https://hub.docker.com/_/golang)
