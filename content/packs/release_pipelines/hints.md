## Task 1: Release on version tags

Two changes: replace `on: push` with a push that has only a tag filter, and
add a `release` job beside `test`. The version comes from the github
context; look for the row that holds `v1.4.2`.

### Closer

`push:` gets `tags: ['v*']` indented under it. The `release` job is shaped
like `test`, plus `needs: test`. Its last step is the command from
*Stamping the version*, with `main.version`, `-o bin/statusd`, and
`${{ github.... }}` where the value goes.

### Answer

```yaml
on:
  push:
    tags: ['v*']
```

```yaml
  release:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1 # v7.0.1
      - uses: actions/setup-go@b7ad1dad31e06c5925ef5d2fc7ad053ef454303e # v7.0.0
        with:
          go-version-file: go.mod
      - run: go build -ldflags "-X main.version=${{ github.ref_name }}" -o bin/statusd .
```

## Task 2: Build and push the image

The job reads the code and writes a package, so its `permissions:` block
names both. Then three steps: get the code, log in, build and push.

### Closer

`permissions:` sits beside `runs-on`, with `contents: read` and
`packages: write` indented under it. The login step's `with:` has
`registry`, `username` and `password`; the build step's has `context`,
`push`, `tags` and `build-args`, each value from the task list.

### Answer

```yaml
    permissions:
      contents: read
      packages: write
    steps:
      - uses: actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1 # v7.0.1
      - uses: docker/login-action@dbcb813823bdd20940b903addbd779551569679f # v4.6.0
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      - uses: docker/build-push-action@c3c9e263c25d99ce0380d002d59b67737d91b0dc # v7.4.0
        with:
          context: .
          push: true
          tags: ghcr.io/acme/statusd:${{ github.ref_name }}
          build-args: VERSION=${{ github.ref_name }}
```

## Task 3: Promote from staging to production

Two jobs in a chain: staging waits for the image, production waits for
staging. Each checks out the repository to get `deploy.sh` and passes it
the environment and the tag. Only production needs `concurrency`.

### Closer

`environment: staging` and `needs: image` sit beside `runs-on`. The
production job is the same shape with `needs: deploy-staging`,
`environment: production` and `concurrency: production`. The deploy step
is `- run: ./deploy.sh <environment> ${{ github.ref_name }}`.

### Answer

```yaml
  deploy-staging:
    needs: image
    runs-on: ubuntu-latest
    environment: staging
    steps:
      - uses: actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1 # v7.0.1
      - run: ./deploy.sh staging ${{ github.ref_name }}

  deploy-production:
    needs: deploy-staging
    runs-on: ubuntu-latest
    environment: production
    concurrency: production
    steps:
      - uses: actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1 # v7.0.1
      - run: ./deploy.sh production ${{ github.ref_name }}
```
