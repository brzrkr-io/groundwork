# YAML manifests: reading them into Go

You've kept Kubernetes objects in YAML files and run `kubectl apply -f` all
through *Kubernetes fundamentals*. Platform tools (linters, CI checks,
deploy bots) read those same files, so your Go code needs to read them too.

**New here:** `sigs.k8s.io/yaml` in Go, `metav1.TypeMeta`, splitting multi-document streams.

## Learn

### What a manifest is

A *manifest* is a text file that describes one or more Kubernetes objects,
such as the ones you wrote in *Pods: your first manifest*. In Go, its parts
are the ones you built in *Kubernetes objects: building a Deployment*:
`apiVersion`, `kind`, `metadata` and `spec`.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: checkout
  namespace: shop
spec:
  replicas: 3
  template:
    spec:
      containers:
        - name: checkout
          image: registry.example.com/shop/checkout:2.1.0
```

### Recap: YAML in five rules

*YAML*, from *CI workflows*, holds the same data as JSON, using indentation
instead of braces.

| YAML | JSON | Meaning |
|---|---|---|
| `name: checkout` | `"name": "checkout"` | a key and its value |
| an indented block under `metadata:` | `"metadata": {...}` | a map (object) |
| lines starting with `- ` | `[...]` | a list |
| `replicas: 3` | `"replicas": 3` | numbers stay numbers |
| `# note` | (none) | a comment |

Indent with **spaces, never tabs**. Keys in the same map must line up
exactly.

### Reading YAML into Go types

`sigs.k8s.io/yaml` is the YAML library Kubernetes itself uses.
`yaml.Unmarshal` converts the YAML to JSON, then calls `json.Unmarshal` (as
in *Decoding JSON*). So the Go types' `json` tags decide which key fills
which field. Here it reads a Service:

```go
var svc corev1.Service
if err := yaml.Unmarshal(data, &svc); err != nil {
	return fmt.Errorf("load service: %w", err)
}
fmt.Println(svc.Name, svc.Spec.Ports[0].Port) // checkout 80
```

`data` is a `[]byte`. Keys that match no field are silently ignored, so a
typo like `replica: 3` leaves `Replicas` as `nil`.

> **Why is it written like this?**
> - **Go requires** `&svc`: `Unmarshal` must change your variable, so it
>   needs its address.
> - **Convention:** declare an empty `var` first, fill it, then return it
>   (or `&` it if the function returns a pointer).
> - **Author's choice:** the text before `%w` (`"load service: "`) is up to
>   you; it tells the reader where the error came from.

### Peeking at the kind first

A file could hold any kind, and each kind needs a different Go type. So
tools first read only `apiVersion` and `kind`, using `metav1.TypeMeta`,
which has exactly those two fields (`APIVersion`, `Kind`). Every other key is
ignored. Then they unmarshal again into the right type.

A missing key leaves its field at the zero value, `""`. That is not an
error to `Unmarshal`; you must check for it yourself.

### Many objects in one file

One file often holds several objects, called *documents*, separated by a
line that is exactly `---`:

```yaml
apiVersion: v1
kind: Namespace
---
apiVersion: apps/v1
kind: Deployment
```

`yaml.Unmarshal` reads one document, so you split the stream first.
`strings.SplitSeq` hands you one piece at a time in a `for range` loop:

```go
for word := range strings.SplitSeq("red,green,blue", ",") {
	fmt.Println(word) // "red", then "green", then "blue"
}
```

> **Why is it written like this?** `SplitSeq` vs `strings.Split` is a
> **convention**: `Split` builds a whole slice first; `SplitSeq` is preferred
> when you only loop over the pieces.

## Your task

You're writing the first pieces of `shiplint`, a tool that checks manifests
in your team's deploy repo before they reach the cluster.

### 1. Parse a Deployment

**Goal:** make `ParseDeployment(data)` return a pointer to the `appsv1.Deployment` read from `data`.

- Use `yaml.Unmarshal` into an `appsv1.Deployment`.
- On success: return a pointer to it and `nil`.
- On failure: return `nil` and the error wrapped with `%w`, prefix
  `parse deployment: `. Add the `"fmt"` import.

```go
d, err := ParseDeployment(checkoutYAML) // testdata/checkout.yaml
// d.Name == "checkout", d.Namespace == "shop", *d.Spec.Replicas == 3

_, err = ParseDeployment([]byte("spec:\n  replicas: three\n"))
// err.Error() starts with "parse deployment: "
```

**Think it through:**
- Compare with the Service example above: what changes for a Deployment?
- The function returns `*appsv1.Deployment`. How do you get a pointer to your variable?

### 2. Read the type

**Goal:** finish the error handling in `ReadType`.

`ReadType` already unmarshals into a `metav1.TypeMeta`. Change it so:

- If unmarshalling fails, it returns the error wrapped with `%w` and the
  prefix `read type: ` (not bare).
- If `tm.Kind` is empty, it returns the error `read type: missing kind`
  exactly.

Add the `"errors"` import (or use `fmt.Errorf` without verbs).

```go
ReadType(serviceYAML) // testdata/service.yaml
// => metav1.TypeMeta{APIVersion: "v1", Kind: "Service"}, nil

ReadType([]byte("apiVersion: v1\n"))
// => error "read type: missing kind"
```

**Think it through:**
- Which line returns the bare `err` now?
- Where does the "missing kind" check go: before or after a successful unmarshal?

### 3. Split a stream into documents

**Goal:** write the loop in `SplitDocuments` that splits `stream` on `---` lines.

- Loop over the lines with `strings.SplitSeq(stream, "\n")`.
- A line that is **exactly** `---` ends a document. Other lines, like
  `name: a---b`, are ordinary text.
- `addDocument(docs, current)` is written for you: it joins, trims, and
  appends unless empty. Call it on each `---`, then start `current` fresh.
- Any other line is appended to `current`.

The final `addDocument` for the last document is already in place. Empty
input or input of only `---` lines returns `nil`.

```go
SplitDocuments("---\nkind: Namespace\n---\n\n---\nkind: Service\n")
// => []string{"kind: Namespace", "kind: Service"}
```

**Think it through:**
- What two things happen on a `---` line? What happens on any other line?
- How do you "start fresh" with a slice variable?

## Learn more

- [Kubernetes: Objects in Kubernetes](https://kubernetes.io/docs/concepts/overview/working-with-objects/)
- [Kubernetes: Managing resources (multiple resources in one file)](https://kubernetes.io/docs/concepts/workloads/management/)
- [pkg.go.dev: sigs.k8s.io/yaml](https://pkg.go.dev/sigs.k8s.io/yaml)
- [pkg.go.dev: strings.SplitSeq](https://pkg.go.dev/strings#SplitSeq)
