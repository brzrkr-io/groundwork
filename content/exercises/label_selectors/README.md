# Labels and selectors: matching in Go

Kubernetes connects objects with labels, not names. When a Service sends
traffic to the wrong Pods, or to none, a label is nearly always the cause. In
this exercise you answer "which Pods does this pick?" in Go.

## Learn

### Labels

*Labels* are short key/value tags in an object's `metadata`, as you saw in
*Kubernetes objects: building a Deployment*. Teams agree on a few keys and put them on
everything:

| Key | Example values | Answers |
|---|---|---|
| `app` | `checkout`, `api` | Which program is this? |
| `env` | `prod`, `staging`, `dev` | Which environment? |
| `tier` | `web`, `api`, `batch` | Which layer of the system? |
| `team` | `payments`, `shop` | Who do I page when it breaks? |

### Why labels instead of names

Pods come and go. A Deployment creates Pods with random names such as
`checkout-7d9f8-x2b4q`, and replaces them whenever it likes. So nothing in
Kubernetes points at a Pod by name. Objects describe the labels they want
instead:

- A Service sends traffic to every Pod whose labels match its
  `spec.selector`.
- A Deployment counts the Pods that match its `spec.selector` to know whether
  it has enough.

A new Pod with the right labels joins straight away. Remove a label and the
Pod drops out. That is how you take a broken Pod out of traffic for
debugging without killing it.

### Selectors

A *label selector* is a query over labels. `kubectl get pods -l '<selector>'`
uses the same syntax you will parse in Go:

| Selector | Matches Pods that... |
|---|---|
| `env=prod` | have `env` set to `prod` |
| `env!=dev` | don't have `env` set to `dev`, including Pods with no `env` label |
| `tier in (web,api)` | have `tier` set to `web` or `api` |
| `tier notin (batch)` | don't have `tier` set to `batch`, including Pods with no `tier` label |
| `canary` | have a `canary` label, whatever its value |
| `!canary` | have no `canary` label |

The first two are *equality-based*, the rest *set-based*. A comma means AND:
`env=prod,tier in (web,api)` needs both. Spaces between requirements are not
allowed; use a comma.

### Selectors in Go

The package `k8s.io/apimachinery/pkg/labels` does the matching:

```go
import "k8s.io/apimachinery/pkg/labels"
```

`labels.Set` is declared as `type Set map[string]string`. A Pod's labels are
a plain `map[string]string`, so convert them the same way you'd convert any
type: `labels.Set(pod.Labels)`. A `nil` map is fine; it means "no labels".

`labels.Selector` is an interface. You get one in two ways:

```go
// From a map, like a Service's spec.selector: every pair must match.
sel := labels.SelectorFromSet(labels.Set{"app": "checkout"})

// From a string, like `kubectl -l`. Strings can be wrong, so you get an error.
sel, err := labels.Parse("env=prod,tier in (web,api)")
if err != nil {
	// err.Error() is something like:
	// unable to parse requirement: found 'web' expected: '('
}
```

Then ask it about some labels:

```go
podLabels := labels.Set{"app": "checkout", "env": "prod", "tier": "web"}
sel.Matches(podLabels) // => true
sel.String()           // => "env=prod,tier in (api,web)"
```

`String()` writes the selector back out in a tidy, sorted form, which is
handy in log lines and error messages.

One trap: an empty selector, from `labels.SelectorFromSet(nil)` or
`labels.Parse("")`, matches **everything**.

## Your task

The `shop` team keeps asking "which Pods does this Service or selector
actually pick?". You're writing the Go behind a small tool that answers it
from a list of Pods, without asking the cluster.

### 1. Check whether a Service selects a Pod

A Service keeps its selector in `svc.Spec.Selector`, a `map[string]string`.
Make `ServiceSelects` return `true` when every key/value pair in that map is
also in the Pod's labels (`pod.Labels`). Extra labels on the Pod don't
matter.

```go
svc := corev1.Service{Spec: corev1.ServiceSpec{
	Selector: map[string]string{"app": "checkout", "tier": "web"},
}}
pod := corev1.Pod{ObjectMeta: metav1.ObjectMeta{
	Name:   "checkout-7d9f8-x2b4q",
	Labels: map[string]string{"app": "checkout", "tier": "web", "env": "prod"},
}}
ServiceSelects(svc, pod)
// => true
```

With the Pod's `tier` set to `"worker"`, or with no labels at all, the result
is `false`.

### 2. List the names of matching Pods

Make `MatchingPodNames` return the `Name` of every Pod in `pods` that `sel`
matches, in the same order as `pods`. When nothing matches, return an empty
or `nil` slice.

```go
pods := []corev1.Pod{web1, api1, web2} // web1 and web2 have team=shop
MatchingPodNames(pods, labels.SelectorFromSet(labels.Set{"team": "shop"}))
// => []string{"web-1", "web-2"}
```

### 3. Select Pods with a selector string

Make `SelectPods` parse `expr` with `labels.Parse`.

- If parsing fails, return `nil` and an error in this form (use `fmt.Errorf`
  with `%q` and `%w`, and add the `"fmt"` import):

  ```text
  invalid selector "tier in web": unable to parse requirement: found 'web' expected: '('
  ```

  The tests check everything up to and including `": "`, so the text after
  that is whatever `labels.Parse` returned.

- Otherwise, return the names that `MatchingPodNames` gives you, and a `nil`
  error.

```go
SelectPods(pods, "env=prod,tier in (web,api)")
// => []string{"api-1", "web-1", "web-canary"}, nil
SelectPods(pods, "!canary,app=web")
// => []string{"web-1", "web-dev"}, nil
```

## Learn more

- [Kubernetes: Labels and Selectors](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/)
- [Kubernetes: Recommended Labels](https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/)
- [pkg.go.dev: k8s.io/apimachinery/pkg/labels](https://pkg.go.dev/k8s.io/apimachinery/pkg/labels)
