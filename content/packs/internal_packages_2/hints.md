## Task 1: Use deploy's public API in report

`report` lives outside `deploy/`, so Go won't let it import anything under `deploy/internal`. It may import `deploy` itself, which offers the same feature in public.

### Closer

Replace the import with `groundwork.local/internal_packages_2/deploy`, and replace `naming.Release(` with `deploy.ReleaseName(`. The arguments stay the same.

### Answer

```go
import "groundwork.local/internal_packages_2/deploy"

func Line(service, version string, healthy bool) string {
	status := "DOWN"
	if healthy {
		status = "healthy"
	}
	return deploy.ReleaseName(service, version) + ": " + status
}
```

## Task 2: Fix the bug inside naming

The name is built correctly but keeps capital letters. One more `strings` call around the result fixes it.

### Closer

`strings.ToLower(s)` turns a whole string into lower case. Wrap the existing expression in it.

### Answer

```go
func Release(service, version string) string {
	return strings.ToLower(strings.ReplaceAll(service+"-"+version, ".", "-"))
}
```

## Task 3: Write rollout.Steps

Work out the release name once, then build one line per host in a loop.

### Closer

Import `groundwork.local/internal_packages_2/deploy/internal/naming`, which is allowed because `rollout` is inside `deploy/`. Use `name := naming.Release(service, version)`, then `append` `"deploy "+name+" to "+host` for each host.

### Answer

```go
import "groundwork.local/internal_packages_2/deploy/internal/naming"

func Steps(service, version string, hosts []string) []string {
	name := naming.Release(service, version)
	var steps []string
	for _, host := range hosts {
		steps = append(steps, "deploy "+name+" to "+host)
	}
	return steps
}
```
