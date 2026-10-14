## Task 1: Fix the compile error in SidecarPort

You can't store a new value in the constant `DefaultPort`. You don't need to:
the port you want is just `DefaultPort` with `n` added.

### Closer

Delete the line that assigns to `DefaultPort`, and return the sum directly:
`return DefaultPort + ...`.

### Answer

```go
func SidecarPort(n int) int {
	return DefaultPort + n
}
```

## Task 2: Build the in-cluster DNS name

`ServiceName` and `Namespace` are string constants, so you join them with `+`
just like variables. You need dots between the parts and `svc` before the
cluster domain.

### Closer

Fixed text goes inside quotes (`"."` and `".svc."`); constants and parameters
(`ServiceName`, `Namespace`, `clusterDomain`) don't:
`ServiceName + "." + ...`.

### Answer

```go
return ServiceName + "." + Namespace + ".svc." + clusterDomain
```

## Task 3: Count the spare replicas

The spare replicas are the limit minus the ones already running. Which
constant holds the limit?

### Closer

`-` subtracts: take `MaxReplicas` and subtract the `running` parameter.

### Answer

```go
return MaxReplicas - running
```
