# Variables: build ops commands

Ops engineers build names and commands from smaller pieces all day. Here you
write three builders from scratch: a Kubernetes pod name, an SSH target and a
`kubectl` command.

**New here:** nothing; practice turning an example into code.

## Learn

### Recap

- Use parameters by name, outside quotes. Join with `+`.
- Every separator (dash, dot, space) is fixed text you add yourself.
- `:=` creates a variable; `=` or `+=` changes it.

Each function starts as `return ""`. You replace the `""` with your
expression. The `func` line above it stays exactly as it is.

### From an example to code

Building a string from an example takes two steps. First split the example
into pieces. Then decide for each piece whether it comes from a parameter or
is fixed text:

- **Fixed text goes inside quotes**, exactly as it appears, spaces included:
  `"-"`, `"@"`, `"git clone "`.
- **Parameters go outside quotes.** Go swaps in their values.

Take `api-v2`, returned by `ServiceTag("api", "v2")`, where the `func` line
is `func ServiceTag(name, version string) string`:

| piece | comes from | Go code |
|---|---|---|
| `api` | the `name` parameter | `name` |
| `-` | fixed text | `"-"` |
| `v2` | the `version` parameter | `version` |

Join the pieces with `+` and return the result.

Common mistakes:

| Wrong | Why |
|---|---|
| `"name-"` | Just the letters n-a-m-e. Go doesn't look inside quotes for names. |
| `"name + "` | The `+` is inside the quotes, so it's text, not joining. |
| `ServiceTag` | That's the function's name, not a value. Use the parameters. |
| `"git clone" + repo` | Missing space: gives `git clonerepo`. |

> **Why is it written like this?** **Go requires** each `+` to sit between
> two complete values, and each quoted piece to open and close its own
> quotes. **Your choice:** one long `return` line, or build a variable with
> `:=` and `+=` over several lines and return it. Both are fine.

## Your task

Write the three builders. Every dash, `@`, slash and space in the examples
counts.

### 1. Name a pod

**Goal:** make `PodName` return the deployment, a dash, the hash, a dash, then
the suffix.

```go
PodName("checkout", "7d9f8b6c5", "x2k4p")
// => "checkout-7d9f8b6c5-x2k4p"

PodName("api", "5f6d7c8b9", "q8w7e")
// => "api-5f6d7c8b9-q8w7e"
```

A Kubernetes *Deployment* names its pods this way. All three pieces are
parameters; the two dashes are fixed text.

**Think it through:**
- How many pieces is that, and how many `+` join them?

### 2. Build an SSH target

**Goal:** make `SSHTarget` return the user, `@`, the host, a dot, then the
domain.

```go
SSHTarget("deploy", "web-01", "prod.example.com")
// => "deploy@web-01.prod.example.com"

SSHTarget("admin", "db-02", "staging.example.com")
// => "admin@db-02.staging.example.com"
```

`ssh` takes a target in this form. `user`, `host` and `domain` are
parameters; `@` and `.` are fixed text.

**Think it through:**
- Write the pieces in order first, then mark which ones need quotes.

### 3. Build a kubectl set-image command

**Goal:** make `SetImageCommand` return this command, with each `<...>`
replaced by that parameter:

```text
kubectl -n <namespace> set image deployment/<deployment> <deployment>=<image>
```

```go
SetImageCommand("payments", "checkout", "registry.internal/checkout:v2.1.0")
// => "kubectl -n payments set image deployment/checkout checkout=registry.internal/checkout:v2.1.0"

SetImageCommand("web", "api", "ghcr.io/acme/api:v1.4.2")
// => "kubectl -n web set image deployment/api api=ghcr.io/acme/api:v1.4.2"
```

In this cluster each deployment has one container with the same name as the
deployment, so `deployment` appears twice. Everything outside the `<...>`
(words, spaces, `-n`, `/`, `=`) is fixed text.

**Think it through:**
- Where does each fixed-text piece start and end? Each one usually ends with a space.
- It's long. Would building it over a few lines with `+=` be easier to check?

## Learn more

- [Kubernetes: Deployments](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/)
- [kubectl set image](https://kubernetes.io/docs/reference/kubectl/generated/kubectl_set/kubectl_set_image/)
- [Go by Example: Variables](https://gobyexample.com/variables)
