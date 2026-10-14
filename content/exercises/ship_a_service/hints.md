## Task 1: Write the Dockerfile

Two stages, one `FROM` each. Write the build stage top to bottom in the
order of the numbered rules, then the final stage, which only copies the
binary in and says how to run it.

### Closer

The build stage: `FROM ... AS build`, `WORKDIR`, `COPY go.mod go.sum ./`,
`RUN go mod download`, `COPY . .`, `ARG VERSION=dev`, then the exact
`RUN CGO_ENABLED=0 go build ...` line. The final stage needs
`COPY --from=build /out/inventoryd /inventoryd`, the numeric `USER`,
`EXPOSE`, and `ENTRYPOINT ["..."]` with the same path.

### Answer

```dockerfile
FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=dev
RUN CGO_ENABLED=0 go build -ldflags "-X main.version=${VERSION}" -o /out/inventoryd .

FROM gcr.io/distroless/static-debian13
COPY --from=build /out/inventoryd /inventoryd
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/inventoryd"]
```

## Task 2: Keep the build context small

One line per group in the table: the Git folder, the env files, the `bin`
folder, and the Markdown files. A folder's name alone leaves out
everything inside it.

### Closer

`*` after a name catches every file that starts with it, so one pattern
covers `.env` and `.env.local`. For files in any folder, start the pattern
with `**/`.

### Answer

```text
.git
.env*
bin/
**/*.md
```

## Task 3: Test every change

This is the workflow from *CI workflows*, with `-race` added to the tests:
`name`, `on` with two events, and one job whose four steps set up the code
and Go, then vet and test.

### Closer

`push:` gets `branches: [main]` under it, and `pull_request:` stays bare.
The setup-go step takes `with:` and `go-version-file: go.mod`. The last two
steps are `- run:` lines.

### Answer

```yaml
name: ci

on:
  push:
    branches: [main]
  pull_request:

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v7
      - uses: actions/setup-go@v7
        with:
          go-version-file: go.mod
      - run: go vet ./...
      - run: go test -race ./...
```

## Task 4: Release on tags

Start from your `ci.yml`: change the name and the trigger, keep the `test`
job as it is, and add the `image` job beside it under `jobs:`.

### Closer

The trigger is `push:` with `tags: ['v*']` under it. The `image` job gets
`needs: test`, a `permissions:` block with the two settings, and three
steps: checkout, `docker/login-action@v4` with a `with:` block, and
`docker/build-push-action@v7` with `context`, `push`, `tags` and
`build-args` under its `with:`.

### Answer

```yaml
name: release

on:
  push:
    tags: ['v*']

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v7
      - uses: actions/setup-go@v7
        with:
          go-version-file: go.mod
      - run: go vet ./...
      - run: go test -race ./...

  image:
    needs: test
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write
    steps:
      - uses: actions/checkout@v7
      - uses: docker/login-action@v4
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      - uses: docker/build-push-action@v7
        with:
          context: .
          push: true
          tags: ghcr.io/acme/inventoryd:${{ github.ref_name }}
          build-args: VERSION=${{ github.ref_name }}
```
