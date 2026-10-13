# YAML manifests: reading them into Go

Nobody types Kubernetes objects into Go by hand all day. Teams keep them in
YAML files in Git and run `kubectl apply -f`. Platform tools (linters, CI
checks, deploy bots) read those same files, so your Go code needs to read them too.

## Learn

### What a manifest is

A *manifest* is a text file that describes one or more Kubernetes objects.
You hand it to the cluster with:

```text
kubectl apply -f checkout.yaml
```

It has the same parts you built in Go in *Kubernetes objects: building a Deployment*:
`apiVersion`, `kind`, `metadata` and `spec`.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: checkout
  namespace: shop
  labels:
    app: checkout
spec:
  replicas: 3
  template:
    spec:
      containers:
        - name: checkout
          image: registry.example.com/shop/checkout:2.1.0
          ports:
            - containerPort: 8080
```

### YAML in five rules

*YAML* holds the same data as JSON, but uses indentation instead of braces.

| YAML | JSON | Meaning |
|---|---|---|
| `name: checkout` | `"name": "checkout"` | a key and its value |
| an indented block under `metadata:` | `"metadata": {...}` | a map (object) |
| lines starting with `- ` | `[...]` | a list |
| `replicas: 3` | `"replicas": 3` | numbers stay numbers |
| `# note` | (none) | a comment |

Indentation must use **spaces, never tabs**. Items in the same map must line
up exactly. `containers:` above is a list holding one map, and that map has
the keys `name`, `image` and `ports`.

### Reading YAML into Go types

The package `sigs.k8s.io/yaml` is the YAML library Kubernetes itself uses.
Its `yaml.Unmarshal` works in two steps:

1. It converts the YAML text to JSON.
2. It calls `json.Unmarshal` (from *Decoding JSON*) on the result.

So the Go types' `json` struct tags decide which key fills which field. That
is why manifests use camelCase names such as `apiVersion` and
`containerPort`: they are the `json` names of the Go fields.

```go
import (
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/yaml"
)

var d appsv1.Deployment
if err := yaml.Unmarshal(data, &d); err != nil {
	return fmt.Errorf("read manifest: %w", err)
}
fmt.Println(d.Name, *d.Spec.Replicas) // checkout 3
```

`data` is a `[]byte`, just like with `json.Unmarshal`. Keys that match no field
are silently ignored, so a typo like `replica: 3` leaves `Replicas` as `nil`.

### Peeking at the kind first

A file could hold a Deployment, a Service or anything else, and each kind
needs a different Go type. So tools first read only `apiVersion` and `kind`.
The type `metav1.TypeMeta` (from `k8s.io/apimachinery/pkg/apis/meta/v1`) has
exactly those two fields:

```go
var tm metav1.TypeMeta
err := yaml.Unmarshal(data, &tm) // every other key is ignored
fmt.Println(tm.APIVersion, tm.Kind) // apps/v1 Deployment
```

Once you know the kind, you pick the right type and unmarshal again.

### Many objects in one file

One file often holds several objects, called *documents*, separated by a line
that is exactly `---`:

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: shop
---
apiVersion: apps/v1
kind: Deployment
...
```

`yaml.Unmarshal` reads only one document, so you split the stream first.
These `strings` functions help:

```go
strings.Split("a\nb", "\n")         // []string{"a", "b"}
strings.Join([]string{"a", "b"}, "\n") // "a\nb"
strings.TrimSpace("\n  hi \n")       // "hi"
```

When you only want to loop over the pieces, `strings.SplitSeq` is the
better choice. It hands you one piece at a time in a `for range` loop
instead of building the whole slice first:

```go
for line := range strings.SplitSeq("a\nb", "\n") {
	fmt.Println(line) // "a", then "b"
}
```

## Your task

You're writing the first pieces of `shiplint`, a tool that checks the
manifests in your team's deploy repo before they reach the cluster.

### 1. Parse a Deployment

Finish `ParseDeployment(data []byte) (*appsv1.Deployment, error)`. Unmarshal
`data` into an `appsv1.Deployment` with `yaml.Unmarshal` and return a pointer
to it. If unmarshalling fails, return `nil` and the error wrapped with `%w`
and the prefix `parse deployment: `.

```go
d, err := ParseDeployment(checkoutYAML) // testdata/checkout.yaml
// d.Name == "checkout", d.Namespace == "shop", *d.Spec.Replicas == 3

_, err = ParseDeployment([]byte("replicas: [3"))
// err.Error() starts with "parse deployment: "
```

### 2. Read the type

`ReadType(data []byte) (metav1.TypeMeta, error)` already unmarshals `data`
into a `metav1.TypeMeta` (look at it for a working `yaml.Unmarshal` call).
Finish its error handling:

- If unmarshalling fails, return the error wrapped with `%w` and the prefix
  `read type: `, instead of returning it bare.
- If the document has no `kind`, return the error `read type: missing kind`.

```go
ReadType(serviceYAML) // testdata/service.yaml
// => metav1.TypeMeta{APIVersion: "v1", Kind: "Service"}, nil

ReadType([]byte("metadata:\n  name: web\n"))
// => error "read type: missing kind"
```

### 3. Split a stream into documents

Finish `SplitDocuments(stream string) []string`:

- A line that is exactly `---` separates two documents. Any other line
  (for example `name: a---b`) is ordinary text.
- Trim each document with `strings.TrimSpace`.
- Skip documents that are empty after trimming (for example before a leading
  `---`, or between two `---` lines).

The helper `addDocument` is already written: it joins a document's lines,
trims it, and appends it unless it is empty. The final call that adds the last
document is in place too. You write the loop over the lines, using
`strings.SplitSeq`.

```go
SplitDocuments("---\nkind: Namespace\n---\n\n---\nkind: Service\n")
// => []string{"kind: Namespace", "kind: Service"}
```

## Learn more

- [Kubernetes: Objects in Kubernetes (describing an object)](https://kubernetes.io/docs/concepts/overview/working-with-objects/)
- [Kubernetes: Managing resources (multiple resources in one file)](https://kubernetes.io/docs/concepts/workloads/management/)
- [sigs.k8s.io/yaml on pkg.go.dev](https://pkg.go.dev/sigs.k8s.io/yaml)
- [YAML in Y minutes](https://learnxinyminutes.com/docs/yaml/)
