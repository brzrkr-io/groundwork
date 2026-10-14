## Task 1: Finish DecodeObject

Each new kind needs its own `case`, shaped like the `"Deployment"` one. The
`default:` case runs for every other kind, so that is where the error goes.

### Closer

Add the import `corev1 "k8s.io/api/core/v1"`. The new cases set `obj` to
`&corev1.Service{}` or `&corev1.ConfigMap{}`. Use `%q` with `tm.Kind` in
the `default` error.

### Answer

```go
case "Deployment":
	obj = &appsv1.Deployment{}
case "Service":
	obj = &corev1.Service{}
case "ConfigMap":
	obj = &corev1.ConfigMap{}
default:
	return nil, fmt.Errorf("unsupported kind %q", tm.Kind)
```

A missing kind is `""`, so it also falls into `default`.

## Task 2: Fix ImageTag

Compare two positions: where the last `:` is, and where the last `/` is. A
colon only starts a tag if it comes after the last slash.

### Closer

Find both with `strings.LastIndex(image, ":")` and
`strings.LastIndex(image, "/")`. If the colon is not after the slash, return
`""`; otherwise the tag is `image[colon+1:]`.

### Answer

```go
colon := strings.LastIndex(image, ":")
if colon <= strings.LastIndex(image, "/") {
	return ""
}
return image[colon+1:]
```

`<=` also covers images with no colon at all, because `colon` is then `-1`.

## Task 3: Write Lint

One loop over the containers, with two checks each: the image tag first,
then the limits. Collect the lines in a slice.

### Closer

Loop with `for _, c := range d.Spec.Template.Spec.Containers`. A `switch` on
`ImageTag(c.Image)` handles `""` and `"latest"`. Build lines with
`fmt.Sprintf` and `%q`, for example
`fmt.Sprintf("container %q: no resource limits", c.Name)`.

### Answer

```go
var problems []string
for _, c := range d.Spec.Template.Spec.Containers {
	switch ImageTag(c.Image) {
	case "":
		problems = append(problems, fmt.Sprintf("container %q: image %q has no tag", c.Name, c.Image))
	case "latest":
		problems = append(problems, fmt.Sprintf("container %q: image %q uses the latest tag", c.Name, c.Image))
	}
	if len(c.Resources.Limits) == 0 {
		problems = append(problems, fmt.Sprintf("container %q: no resource limits", c.Name))
	}
}
return problems
```

A `nil` slice that never gets appended to is the clean result.
