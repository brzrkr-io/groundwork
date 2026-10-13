## Task 1: don't change the constant

You can't store a new value in `DefaultPort`. You don't need to: the answer is
just `DefaultPort` plus `n`. Delete the assignment line and return that sum.

## Task 2: constants join like strings

`ServiceName` and `Namespace` are strings, so join them with `+`. You need
dots between the parts and `".svc."` before the cluster domain.

## Task 3: subtract

The spare replicas are the limit minus the ones already running.

## Almost the answer

```go
func SidecarPort(n int) int {
	return DefaultPort + n
}

func DNSName(clusterDomain string) string {
	return ServiceName + "." + Namespace + ".svc." + clusterDomain
}

func SpareReplicas(running int) int {
	return MaxReplicas - running
}
```
