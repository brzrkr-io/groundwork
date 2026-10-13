## Task 1: copy the Deployment case

Each new case looks like the `"Deployment"` one, with `&corev1.Service{}` or
`&corev1.ConfigMap{}`. The `default:` case runs for every other kind, so it is
where you `return nil, fmt.Errorf("unsupported kind %q", tm.Kind)`.

## Task 2: compare two positions

Find the last `:` and the last `/` with `strings.LastIndex`. If the colon
comes after the slash, the tag is `image[colon+1:]`. Otherwise (including
when the colon is `-1`) there is no tag.

## Task 3: one loop, two checks

Loop with `for _, c := range d.Spec.Template.Spec.Containers`. Use
`fmt.Sprintf` with `%q` for the quoted names: `fmt.Sprintf("container %q: no resource limits", c.Name)`.

## Almost the answer

```go
var problems []string
for _, c := range d.Spec.Template.Spec.Containers {
	switch ImageTag(c.Image) {
	case "":
		problems = append(problems, fmt.Sprintf("container %q: image %q has no tag", c.Name, c.Image))
	case "latest":
		// same, with "uses the latest tag"
	}
	if len(c.Resources.Limits) == 0 {
		problems = append(problems, fmt.Sprintf("container %q: no resource limits", c.Name))
	}
}
return problems
```
