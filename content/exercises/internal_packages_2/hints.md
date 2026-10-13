## Task 1: swap the import

`report` may import `deploy` but not `deploy/internal/naming`. Replace the
import line, then replace `naming.Release(` with `deploy.ReleaseName(`. The
arguments stay the same.

## Task 2: one more strings call

`strings.ToLower` turns a whole string into lower case. Wrap the existing
result with it.

## Task 3: loop over the hosts

```go
import "gopherlings.local/internal_packages_2/deploy/internal/naming"

func Steps(service, version string, hosts []string) []string {
	name := naming.Release(service, version)
	var steps []string
	for _, host := range hosts {
		steps = append(steps, "deploy "+name+" to "+host)
	}
	return steps
}
```
