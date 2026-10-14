# CI workflows

*Continuous integration* (CI) means a machine builds and tests every change
you push, so a broken commit is caught in minutes, not in production. On
GitHub this is a *workflow*: a YAML file that says when to run and what to
run.

**New here:** YAML, GitHub Actions workflows, `on`, `jobs`, `runs-on`,
`steps`, `uses`, `run`, `with`, `needs`.

## Learn

### Where workflows live

GitHub runs every YAML file in the repository's `.github/workflows/`
folder. Files whose names start with a dot can't be part of a Groundwork
exercise, so here the workflows sit next to the tests. In a real repository
`2_test.yml` would be `.github/workflows/test.yml`.

### YAML in two minutes

YAML is nested `key: value` pairs. **Indentation is the structure**: two
spaces deeper means "inside the line above". Tabs are not allowed.

```yaml
service: inventory        # key: value
owners:                   # a list: one "- " per item
  - platform-team
  - sre
limits:                   # a map inside a map
  cpu: "500m"
  memory: 256Mi
ports: [80, 443]          # a short list on one line
```

A list item can itself be a map: `- name: x` starts an item, and the keys
under it (indented to line up with `name`) belong to the same item.

> **Why is it written like this?**
> - **YAML requires:** spaces, not tabs, and the same indentation for keys
>   that belong together. One wrong space moves a key to another level.
> - **Convention:** two spaces per level, and quotes only where they are
>   needed. Version numbers are one such place: YAML reads `1.20` as the
>   number 1.2, so write `'1.20'`.

### The shape of a workflow

```yaml
name: lint                      # shown in the Actions tab

on: push                        # when: every push to any branch

jobs:
  shellcheck:                   # a job id, your choice
    runs-on: ubuntu-latest      # a fresh virtual machine for this job
    steps:
      - uses: actions/checkout@v7         # step 1: an action
      - run: shellcheck scripts/*.sh      # step 2: a shell command
```

| Key | Meaning |
|---|---|
| `name` | the workflow's display name |
| `on` | the *events* that start it: `push`, `pull_request`, ... |
| `jobs` | the jobs, each under an id you choose |
| `runs-on` | the kind of machine (*runner*) the job gets |
| `steps` | the job's steps, run in order, top to bottom |
| `uses: owner/repo@version` | run a ready-made *action* |
| `with:` | inputs for that action |
| `run:` | shell commands |

Every job starts on a **fresh machine**. Common tools such as Git, Docker
and a Go release are installed, but your code isn't there until the
`actions/checkout` step clones it, so checkout is almost always step one.

### Events with settings

`on` can list several events. Each can have settings, such as which
branches count:

```yaml
on:
  push:
    branches: [main]      # only pushes to main
  pull_request:           # every pull request, no settings
```

An event with no settings ends with a bare colon.

Remember the Norway problem from the guide *Data formats and parsing*?
Under YAML 1.1 rules a bare `on` is the boolean `true`. GitHub reads it as
the word `on`, so write it bare like everyone does; the tests here cope
with it too.

### Actions and versions

An *action* is a reusable step someone published. You always say which
version you want after `@`: `actions/checkout@v7` means "the newest
release of major version 7": bug fixes, but no breaking changes. A branch
such as `@main` can change under you overnight.

`actions/setup-go` installs Go. Tell it which version through `with`:

```yaml
- uses: actions/setup-go@v7
  with:
    go-version: '1.25'          # a fixed version
```

or let it read the `go` line of your `go.mod`, so CI and your laptop never
disagree:

```yaml
  with:
    go-version-file: go.mod
```

### Jobs that wait for other jobs

Jobs run **at the same time** unless one says it `needs` another. Then it
waits, and it is skipped if the other one fails:

```yaml
jobs:
  lint:
    runs-on: ubuntu-latest
    steps: [...]
  package:
    needs: lint             # waits for lint; skipped if lint fails
    runs-on: ubuntu-latest
    steps: [...]
```

> **Why is it written like this?**
> - **GitHub requires:** `runs-on` and `steps` in every job, and `uses`
>   *or* `run` in every step.
> - **GitHub requires:** `needs` to name a job id from the same file. With
>   one job it's a plain name; for several use a list, `needs: [lint, test]`.
> - **Author's choice:** job ids (`test`, `build`) and the workflow `name`.

## Your task

Your team's `statusd` service has no CI. You'll write three workflows. The
tests read each file and check the rules below.

### 1. Say hello on every push

**Goal:** in `1_hello.yml`, write a workflow named `hello` that runs
`go version` on every push.

- `name: hello`.
- It triggers on `push` (no settings needed).
- One job with the id `hello`, running on `ubuntu-latest`.
- That job has a step that runs `go version`.

**Think it through:**
- Which three top-level keys does every workflow have?
- Is `go version` an action (`uses`) or a shell command (`run`)?

### 2. Test every push and pull request

**Goal:** in `2_test.yml`, write a workflow named `test` that vets and
tests the code on pushes to `main` and on every pull request.

- `name: test`.
- Triggers: `push` limited to the branch `main`, and `pull_request`.
- One job with the id `test`, on `ubuntu-latest`, with these steps in
  order:
  1. `actions/checkout`, pinned to a version (such as `@v7`),
  2. `actions/setup-go`, pinned to a version, reading the Go version from
     `go.mod` with `go-version-file`,
  3. a step that runs `go vet ./...`,
  4. a step that runs `go test ./...`.

"Pinned" means `@` and a version tag (`@v7`, `@v7.0.1`) or a full commit
hash, never a branch like `@main`.

**Think it through:**
- What's on the runner before checkout? Why does that decide the order?
- `branches` takes a list, even for one branch.

### 3. Build after the tests pass

**Goal:** in `3_build.yml`, add a job `build` that only runs after the
`test` job succeeds.

The `test` job is already written. Add a second job with the id `build`:

- it `needs` the `test` job,
- it runs on `ubuntu-latest`,
- its steps: `actions/checkout` and `actions/setup-go` (pinned, as in
  task 2, with `go-version-file: go.mod`), then a step that runs
  `go build -o bin/statusd .`, in that order.

**Think it through:**
- The `build` job gets its own empty machine. What does it need before it
  can build?
- `build` goes under `jobs:`, at the same indentation as `test:`.

## Learn more

- [Understanding GitHub Actions](https://docs.github.com/en/actions/get-started/understand-github-actions)
- [Workflow syntax for GitHub Actions](https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax)
- [Building and testing Go](https://docs.github.com/en/actions/tutorials/build-and-test-code/go)
- [Learn YAML in Y minutes](https://learnxinyminutes.com/yaml/)
