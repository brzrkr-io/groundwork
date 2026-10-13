# Zero values: layered timeouts and probes

Real platforms layer their settings: a request can override the service, and
the service can override a global default. Zero means "not set at this
layer, ask the next one".

## Learn

### Recap

An unset `int` or `string` holds its zero value (`0` or `""`). Check for the
zero value with `==` and fall back to a default. With several layers, check
them in order and return the first one that is set:

```go
if userRegion != "" {
	return userRegion
}
if teamRegion != "" {
	return teamRegion
}
return "us-east-1"
```

## Your task

You're writing the settings code for a service platform that runs on
Kubernetes.

### 1. Pick the effective timeout

`EffectiveTimeout(requestSeconds, serviceSeconds, globalSeconds)` returns the
first value that is not `0`, checking in that order: request, then service,
then global. If all three are `0`, return `30`.

```go
EffectiveTimeout(5, 10, 60) // => 5
EffectiveTimeout(0, 10, 60) // => 10
EffectiveTimeout(0, 0, 60)  // => 60
EffectiveTimeout(0, 0, 0)   // => 30
```

### 2. Describe a health probe

`ProbeConfig(path, port, periodSeconds)` describes a Kubernetes-style HTTP
health check as `GET PATH on :PORT every Ns`. Fill in defaults first:

- an empty `path` becomes `"/healthz"`
- a `port` of `0` becomes `8080`
- a `periodSeconds` of `0` becomes `10`

```go
ProbeConfig("", 0, 0)        // => "GET /healthz on :8080 every 10s"
ProbeConfig("/ready", 9090, 5) // => "GET /ready on :9090 every 5s"
ProbeConfig("/live", 0, 30)  // => "GET /live on :8080 every 30s"
```

## Learn more

- [A Tour of Go: Zero values](https://go.dev/tour/basics/12)
- [Kubernetes: Configure liveness, readiness and startup probes](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/)
