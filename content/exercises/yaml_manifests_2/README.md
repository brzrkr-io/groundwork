# YAML manifests: decoding kinds and linting images

A deploy repo holds more than Deployments, and some manifests in it are risky.
Here you decode any supported kind and flag containers that would cause
trouble in production.

**New here:** `runtime.Object`, image tags, `strings.LastIndex`.

## Learn

### Recap

`yaml.Unmarshal` from `sigs.k8s.io/yaml` turns YAML into JSON, then fills a
Go value through its `json` tags. To learn which Go type a document needs,
unmarshal it into a `metav1.TypeMeta` first and look at `Kind`.

### One return type for many kinds

A function that can return a Deployment, a Service or a ConfigMap needs one
return type for all of them. Kubernetes uses the interface `runtime.Object`
(package `k8s.io/apimachinery/pkg/runtime`). Every API type implements it
through its **pointer**: `*appsv1.Deployment`, `*corev1.Service`, and so on.

```go
var obj runtime.Object = &corev1.Secret{}
err := yaml.Unmarshal(doc, obj) // fills the Secret
```

That works because `obj` already holds a pointer. The caller gets the
concrete type back with a type switch, as in *Type assertions and switches*.

> **Why is it written like this?**
> - **Go requires** the `&`: only `*corev1.Secret` has the methods
>   `runtime.Object` asks for, so a plain `corev1.Secret{}` won't compile here.
> - **Convention:** the import name `corev1` for `k8s.io/api/core/v1`.

### Image tags

A container image looks like `registry/path/name:tag`. The *tag* picks the
version. Two cases cause trouble:

- No tag, like `nginx`: Kubernetes uses `latest`.
- `latest`: whatever was pushed most recently. You can't tell which version
  runs, and a rollback may pull the same broken image.

A registry can have a port: `registry.example.com:5000/ops/shipper`. That
`:5000` is **not** a tag. A colon starts a tag only when it comes **after
the last `/`**.

### Finding and slicing

`strings.LastIndex(s, sub)` gives the position of the last match, or `-1`
if there is none. A *slice expression* `s[i:]` gives the text from position
`i` to the end:

```go
path := "docs/guide/intro.md"
dot := strings.LastIndex(path, ".") // 16
fmt.Println(path[dot+1:])            // "md"
strings.LastIndex("README", ".")     // -1
```

> **Why is it written like this?** **Go requires** positions to count from
> 0, so `dot` is the dot itself and `dot+1` is the first letter after it.
> Comparing two positions with `<` or `>` tells you which comes first; `-1`
> is smaller than every real position.

### Resource limits

As in *Kubernetes objects: building a Deployment*, `c.Resources.Limits` is a
map that caps a container's CPU and memory. A container with no entries in
that map can starve its neighbours.

## Your task

`shiplint` grows up: it should decode every supported kind and report risky
containers.

### 1. Finish DecodeObject

**Goal:** make `DecodeObject` handle `"Service"` and `"ConfigMap"`, and name the kind in its error.

The `switch tm.Kind` has only a `"Deployment"` case. Make it:

- `"Service"` → `&corev1.Service{}`
- `"ConfigMap"` → `&corev1.ConfigMap{}`
- any other kind, including a missing one → return `nil` and an error whose
  text is exactly `unsupported kind "<kind>"` (use `%q`).

Add the import `corev1 "k8s.io/api/core/v1"`. Broken YAML already returns an
error starting with `decode: `; keep that.

```go
DecodeObject([]byte("kind: Ingress\n"))
// => nil, error `unsupported kind "Ingress"`

DecodeObject([]byte("metadata:\n  name: web\n"))
// => nil, error `unsupported kind ""`
```

**Think it through:**
- Copy the shape of the existing `"Deployment"` case.
- In the `default` error, which variable holds the kind? `%q` adds the quotes.

### 2. Fix ImageTag

**Goal:** make `ImageTag(image)` return the text after the last `:`, but only if that `:` comes after the last `/`; else `""`.

Right now it cuts at the **first** colon, so it mistakes a registry port for
a tag.

```go
ImageTag("nginx:1.27")                                // => "1.27"
ImageTag("nginx")                                     // => ""
ImageTag("registry.example.com:5000/ops/shipper")     // => ""
ImageTag("registry.example.com:5000/ops/shipper:2.0") // => "2.0"
```

**Think it through:**
- Find two positions: the last `:` and the last `/`. Which must be bigger?
- What do the positions look like for `nginx` (no `:` and no `/`)?

### 3. Write Lint

**Goal:** make `Lint(d)` return one line per problem in the Deployment's containers.

Go through `d.Spec.Template.Spec.Containers` in order. For each container,
check the image first, then the limits:

| Problem | Line (for a container named `web`) |
|---|---|
| `ImageTag` returns `""` | `container "web": image "nginx" has no tag` |
| `ImageTag` returns `"latest"` | `container "web": image "nginx:latest" uses the latest tag` |
| `Resources.Limits` is empty | `container "web": no resource limits` |

A clean Deployment returns `nil`. For `testdata/web.yaml`:

```go
[]string{
	`container "web": image "nginx:latest" uses the latest tag`,
	`container "web": no resource limits`,
	`container "log-shipper": image "registry.example.com:5000/ops/shipper" has no tag`,
}
```

**Think it through:**
- Which parts of each line are fixed text and which are `c.Name` / `c.Image`? The quotes come from `%q`.
- One container can add zero, one or two lines. Check image and limits separately.

## Learn more

- [Kubernetes: Images (image names and tags)](https://kubernetes.io/docs/concepts/containers/images/)
- [Kubernetes: Resource management for Pods and containers](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/)
- [pkg.go.dev: runtime.Object](https://pkg.go.dev/k8s.io/apimachinery/pkg/runtime#Object)
- [pkg.go.dev: strings.LastIndex](https://pkg.go.dev/strings#LastIndex)
