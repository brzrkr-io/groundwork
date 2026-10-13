# YAML manifests: decoding kinds and linting images

A deploy repo holds more than Deployments, and some manifests in it are risky.
Here you decode any supported kind and flag containers that would cause
trouble in production.

## Learn

### Recap

`yaml.Unmarshal` from `sigs.k8s.io/yaml` turns YAML into JSON and then fills a
Go value through its `json` tags. To find out which Go type a document needs,
you first unmarshal it into a `metav1.TypeMeta` and look at `Kind`.

### One return type for many kinds

A function that can return a Deployment, a Service or a ConfigMap needs one
return type for all of them. Kubernetes uses the interface `runtime.Object`
from `k8s.io/apimachinery/pkg/runtime`. Every API type implements it through
its pointer (`*appsv1.Deployment`, `*corev1.Service`, ...). The caller gets
the concrete type back with a type switch, as in *Type assertions and switches*:

```go
switch o := obj.(type) {
case *appsv1.Deployment:
	fmt.Println("deployment", o.Name)
case *corev1.Service:
	fmt.Println("service", o.Name)
}
```

You can unmarshal straight into a `runtime.Object` variable, because it holds
a pointer:

```go
var obj runtime.Object = &corev1.ConfigMap{}
err := yaml.Unmarshal(doc, obj) // fills the ConfigMap
```

### Image tags

A container image name looks like `registry/path/name:tag`. The *tag* picks
the version. Two kinds of tag cause trouble:

- No tag at all, like `nginx`, means `latest`.
- `latest` is just whatever was pushed most recently. You can't tell which
  version is running, and a rollback may pull the same broken image again.

A registry can also have a port, as in `registry.example.com:5000/ops/shipper`.
That `:5000` is **not** a tag. A colon only starts a tag when it comes after the
last `/`. `strings.LastIndex` finds the last match, or returns `-1` when there
is none:

```go
strings.LastIndex("a/b/c", "/") // 3
strings.LastIndex("abc", "/")   // -1
```

### Resource limits

As in *Kubernetes objects: building a Deployment*, `resources.limits` caps how much CPU and
memory a container may use. In Go it is `c.Resources.Limits`, a map. A
container without limits can starve its neighbours, so `len(c.Resources.Limits) == 0`
is worth a warning.

## Your task

`shiplint` grows up: it should decode every supported kind and report risky
containers.

### 1. Finish DecodeObject

`DecodeObject(doc []byte) (runtime.Object, error)` reads the kind and then
unmarshals into the right type. It is half-written: only `"Deployment"` has a
case, and the `default` error doesn't say which kind it got. Make it handle:

- `"Service"` → `&corev1.Service{}`
- `"ConfigMap"` → `&corev1.ConfigMap{}`
- any other kind (including a missing one) → return `nil` and an error whose
  text is exactly `unsupported kind "<kind>"` (use `%q`).

You'll need to add the import `corev1 "k8s.io/api/core/v1"`.

```go
DecodeObject([]byte("kind: Ingress\n"))
// => nil, error `unsupported kind "Ingress"`

DecodeObject([]byte("metadata:\n  name: web\n"))
// => nil, error `unsupported kind ""`
```

### 2. Fix ImageTag

`ImageTag(image string) string` returns the tag of an image, or `""` when it
has none. It cuts at the **first** colon, so it mistakes a registry port for a
tag. Fix it: the tag is the text after the last `:`, but only if that `:`
comes after the last `/`.

```go
ImageTag("nginx:1.27")                                // => "1.27"
ImageTag("nginx")                                     // => ""
ImageTag("registry.example.com:5000/ops/shipper")     // => ""
ImageTag("registry.example.com:5000/ops/shipper:2.0") // => "2.0"
```

### 3. Write Lint

`Lint(d *appsv1.Deployment) []string` returns one line per problem. Go through
`d.Spec.Template.Spec.Containers` in order. For each container, check the
image first, then the limits:

| Problem | Line (for a container named `web`) |
|---|---|
| `ImageTag` returns `""` | `container "web": image "nginx" has no tag` |
| `ImageTag` returns `"latest"` | `container "web": image "nginx:latest" uses the latest tag` |
| `Resources.Limits` is empty | `container "web": no resource limits` |

A clean Deployment returns an empty (`nil`) slice. For `testdata/web.yaml`:

```go
[]string{
	`container "web": image "nginx:latest" uses the latest tag`,
	`container "web": no resource limits`,
	`container "log-shipper": image "registry.example.com:5000/ops/shipper" has no tag`,
}
```

## Learn more

- [Kubernetes: Images (image names and tags)](https://kubernetes.io/docs/concepts/containers/images/)
- [Kubernetes: Resource management for pods and containers](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/)
- [k8s.io/apimachinery/pkg/runtime on pkg.go.dev](https://pkg.go.dev/k8s.io/apimachinery/pkg/runtime#Object)
- [strings.LastIndex on pkg.go.dev](https://pkg.go.dev/strings#LastIndex)
