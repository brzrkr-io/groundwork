# Labels and selectors: matching in Go

Kubernetes connects objects with labels, not names. When a Service sends
traffic to the wrong Pods, or to none, a label is nearly always the cause.
Here you answer "which Pods does this pick?" in Go.

**New here:** labels, label selectors, the `labels` package (`labels.Set`, `SelectorFromSet`, `Parse`, `Matches`).

## Learn

### Labels

*Labels* are short key/value tags in an object's `metadata`, as in
*Kubernetes objects: building a Deployment*. Teams agree on a few keys:

| Key | Example values | Answers |
|---|---|---|
| `app` | `checkout`, `api` | Which program is this? |
| `env` | `prod`, `staging`, `dev` | Which environment? |
| `tier` | `web`, `api`, `batch` | Which layer? |
| `team` | `payments`, `shop` | Who do I page? |

### Why labels instead of names

A Deployment creates Pods with random names like `checkout-7d9f8-x2b4q` and
replaces them whenever it likes. So objects describe the labels they want:

- A Service sends traffic to every Pod matching its `spec.selector`.
- A Deployment counts the Pods matching its `spec.selector`.

Remove a label and the Pod drops out: a handy way to take a broken Pod out of
traffic without killing it.

### Selectors

A *label selector* is a query over labels, the same syntax as
`kubectl get pods -l '<selector>'`:

| Selector | Matches Pods that... |
|---|---|
| `env=prod` | have `env` set to `prod` |
| `env!=dev` | don't have `env=dev`, **including** Pods with no `env` label |
| `tier in (web,api)` | have `tier` set to `web` or `api` |
| `tier notin (batch)` | don't have `tier=batch`, including Pods with no `tier` |
| `canary` | have a `canary` label, any value |
| `!canary` | have no `canary` label |

A comma means AND: `env=prod,tier in (web,api)` needs both. Spaces between
requirements are not allowed: `env=prod tier=web` is invalid, and so is a
trailing comma (`env=prod,`).

### Selectors in Go

Package `k8s.io/apimachinery/pkg/labels` does the matching.

`labels.Set` is declared as `type Set map[string]string`. A Pod's labels are a
plain `map[string]string`, so you *convert* them by writing the type name like
a function call:

```go
m := map[string]string{"app": "cart"}
s := labels.Set(m) // same data, now a labels.Set
```

A `nil` map converts fine and means "no labels".

> **Why is it written like this?** Go requires the conversion: a
> `map[string]string` and a `labels.Set` are different types, even though one
> is defined as the other. `Type(value)` is Go's conversion syntax.

`labels.Selector` is an interface. You get one in two ways:

```go
// From a set: every pair must match.
sel := labels.SelectorFromSet(labels.Set{"app": "cart"})

// From a string. Strings can be wrong, so you also get an error.
sel, err := labels.Parse("env=staging,tier notin (batch)")
```

Then ask it about some labels:

```go
sel.Matches(labels.Set{"app": "cart", "env": "staging"}) // => true or false
sel.String()                                             // tidy, sorted text form
```

One trap: an empty selector (from `labels.SelectorFromSet(nil)` or
`labels.Parse("")`) matches **everything**.

## Your task

The `shop` team keeps asking "which Pods does this Service or selector
actually pick?". You write the Go behind a small tool that answers from a
list of Pods, without asking the cluster.

### 1. Check whether a Service selects a Pod

**Goal:** make `ServiceSelects` return `true` when every key/value pair in `svc.Spec.Selector` is also in `pod.Labels`.

Extra labels on the Pod don't matter. A Pod with no labels is not selected.

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

With the Pod's `tier` set to `"worker"`, or with no labels at all: `false`.

**Think it through:**
- Which of the two maps becomes the selector, and which is the thing being matched?
- Both are `map[string]string`. What type do the `labels` functions want?

### 2. List the names of matching Pods

**Goal:** make `MatchingPodNames` return the `Name` of every Pod in `pods` that `sel` matches, in the same order as `pods`.

When nothing matches, return an empty or `nil` slice.

```go
pods := []corev1.Pod{web1, api1, web2} // web1 and web2 have team=shop
MatchingPodNames(pods, labels.SelectorFromSet(labels.Set{"team": "shop"}))
// => []string{"web-1", "web-2"}
```

**Think it through:**
- You already have the selector as a parameter. What do you hand to `Matches` for each Pod?
- How do you keep the original order?

### 3. Select Pods with a selector string

**Goal:** make `SelectPods` parse `expr` with `labels.Parse` and return the names of the matching Pods.

- If parsing fails, return `nil` and an error in this form. `expr` goes inside
  the quotes (use `%q`), and the parse error follows, wrapped with `%w`. Add
  `"fmt"` to the imports.

  ```text
  invalid selector "tier in web": unable to parse requirement: found 'web' expected: '('
  ```

  The tests check everything up to and including `": "`; the rest is whatever
  `labels.Parse` returned.

- Otherwise return the names and a `nil` error.

```go
SelectPods(pods, "env=prod,tier in (web,api)")
// => []string{"api-1", "web-1", "web-canary"}, nil
SelectPods(pods, "!canary,app=web")
// => []string{"web-1", "web-dev"}, nil
```

**Think it through:**
- After parsing, which function from task 2 already does the rest?
- `%q` adds the quotes for you. Should you type `"` in the format string too?

## Learn more

- [Kubernetes: Labels and Selectors](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/)
- [Kubernetes: Recommended Labels](https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/)
- [pkg.go.dev: k8s.io/apimachinery/pkg/labels](https://pkg.go.dev/k8s.io/apimachinery/pkg/labels)
