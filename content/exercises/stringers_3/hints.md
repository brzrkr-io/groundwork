## Tasks 1 and 2: value receivers and Sprintf

Each method looks like `func (r Replicas) String() string { ... }` and
returns a `fmt.Sprintf`. In `Deployment.String`, print `d.Replicas` and
`d.Status` with `%v` so their own `String` methods do the work.

## Task 3: two loops or one

First count the deployments whose `Status` is `Degraded`. Then write the
summary line into a `strings.Builder` with `fmt.Fprintf`, and add each
deployment with `fmt.Fprintln(&b, d)`.

## Almost the answer

```go
func (d Deployment) String() string {
	return fmt.Sprintf("%s/%s %v %v", d.Namespace, d.Name, d.Replicas, d.Status)
}
```

```go
var b strings.Builder
fmt.Fprintf(&b, "deployments: %d, degraded: %d\n", len(deps), degraded)
for _, d := range deps {
	fmt.Fprintln(&b, d)
}
return b.String()
```
