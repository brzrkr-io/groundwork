# Zero values: layered timeouts and probes

Real platforms layer their settings: a request can override the service,
and the service can override a global default. Zero means "not set at this
layer, ask the next one".

**New here:** nothing; practice with zero values, `!=` and `fmt.Sprintf`.

## Learn

### Recap: checking layers in order

An unset `int` is `0`, an unset `string` is `""`. With several layers,
check them in order and return the first one that is set:

```go
if userRegion != "" {
	return userRegion
}
if teamRegion != "" {
	return teamRegion
}
return "us-east-1"
```

The first `return` ends the function, so later checks only run when earlier
layers were unset.

> **Why is it written like this?** `!=` ("is not equal") is Go's required
> spelling. The early-return style is convention. The fallback on the last
> line runs only when nothing above returned.

### Recap: building text

In `fmt.Sprintf`, fixed text goes inside the quotes. Values go after the
format, outside the quotes, matched in order to `%s` (string) and `%d`
(int):

```go
fmt.Sprintf("%s has %d pods", ns, count) // => "web has 3 pods"
```

Spaces and punctuation you want in the result must be inside the quotes too.

## Your task

You're writing the settings code for a service platform on Kubernetes.

### 1. Pick the effective timeout

**Goal:** `EffectiveTimeout(requestSeconds, serviceSeconds, globalSeconds)` returns the first value that is not `0`, or `30` if all are `0`.

Check in that order: request, then service, then global.

```go
EffectiveTimeout(5, 10, 60) // => 5
EffectiveTimeout(0, 10, 60) // => 10
EffectiveTimeout(0, 0, 60)  // => 60
EffectiveTimeout(0, 0, 0)   // => 30
EffectiveTimeout(7, 0, 0)   // => 7
EffectiveTimeout(0, 15, 0)  // => 15
```

**Think it through:**
- One check per parameter, in the order above. What does each check return?
- What does the last line return when no layer is set?

### 2. Describe a health probe

**Goal:** `ProbeConfig(path, port, periodSeconds)` fills in defaults, then returns `GET PATH on :PORT every Ns`.

- an empty `path` becomes `"/healthz"`
- a `port` of `0` becomes `8080`
- a `periodSeconds` of `0` becomes `10`

The starter's `fmt.Sprintf` format is a placeholder; change it to the
format above.

```go
ProbeConfig("", 0, 0)          // => "GET /healthz on :8080 every 10s"
ProbeConfig("/ready", 9090, 5) // => "GET /ready on :9090 every 5s"
ProbeConfig("/live", 0, 30)    // => "GET /live on :8080 every 30s"
ProbeConfig("", 3000, 0)       // => "GET /healthz on :3000 every 10s"
```

**Think it through:**
- First set each default by overwriting the parameter, then build the string once at the end.
- In `GET PATH on :PORT every Ns`, which pieces are fixed text (including spaces, `:` and `s`) and which come from parameters?

## Learn more

- [A Tour of Go: Zero values](https://go.dev/tour/basics/12)
- [Kubernetes: Configure liveness, readiness and startup probes](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/)
