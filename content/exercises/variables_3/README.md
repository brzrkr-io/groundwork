# Variables: build ops commands

Ops engineers build names and commands from smaller pieces all day. Here you
write three builders from scratch: a Kubernetes pod name, an SSH target and a
`kubectl` command.

## Learn

### Recap

Parameters are a function's inputs. Join strings with `+` and add every
separator yourself. Create a variable with `:=` when a value is used more than
once or a line gets long, and change it with `=` or `+=`.

Nothing new this time. Each function starts as `return ""`.

## Your task

Write the three builders. Read each example carefully: every dash, `@`, slash
and space counts.

### 1. Name a pod

A Kubernetes *Deployment* names its pods: the deployment name, a dash, a hash,
another dash, then a random suffix. Write `PodName`:

```go
PodName("checkout", "7d9f8b6c5", "x2k4p")
// => "checkout-7d9f8b6c5-x2k4p"

PodName("api", "5f6d7c8b9", "q8w7e")
// => "api-5f6d7c8b9-q8w7e"
```

### 2. Build an SSH target

`ssh` takes a target of the form user, `@`, host, a dot, then domain. Write
`SSHTarget`:

```go
SSHTarget("deploy", "web-01", "prod.example.com")
// => "deploy@web-01.prod.example.com"

SSHTarget("admin", "db-02", "staging.example.com")
// => "admin@db-02.staging.example.com"
```

### 3. Build a kubectl set-image command

To roll out a new image, you run `kubectl set image`. In this cluster each
deployment has one container, with the same name as the deployment. Write
`SetImageCommand` so it returns:

```text
kubectl -n <namespace> set image deployment/<deployment> <deployment>=<image>
```

Each `<...>` is replaced by a parameter. Note that the deployment name appears
twice.

```go
SetImageCommand("payments", "checkout", "registry.internal/checkout:v2.1.0")
// => "kubectl -n payments set image deployment/checkout checkout=registry.internal/checkout:v2.1.0"

SetImageCommand("web", "api", "ghcr.io/acme/api:v1.4.2")
// => "kubectl -n web set image deployment/api api=ghcr.io/acme/api:v1.4.2"
```

## Learn more

- [Kubernetes: Deployments](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/)
- [kubectl set image](https://kubernetes.io/docs/reference/kubectl/generated/kubectl_set/kubectl_set_image/)
- [Go by Example: Variables](https://gobyexample.com/variables)
