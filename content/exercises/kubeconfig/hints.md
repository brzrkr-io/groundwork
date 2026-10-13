## Task 1: Read the current context

First turn the raw bytes into a config with `clientcmd.Load`, then read one field off the result. Check the error before you use the config.

### Closer

`clientcmd.Load(data)` returns `(cfg, err)`, and the field you want is `cfg.CurrentContext`. Add the import `"k8s.io/client-go/tools/clientcmd"`.

```go
cfg, err := clientcmd.Load(data)
if err != nil {
	// ...
}
```

### Answer

```go
cfg, err := clientcmd.Load(data)
if err != nil {
	return "", err
}
return cfg.CurrentContext, nil
```

`Load` parses the YAML into a `*clientcmdapi.Config`, so the current context is just a field on it.

## Task 2: List the contexts in order

You want the keys of the `cfg.Contexts` map, not its values. Maps have no order, so you must sort the keys yourself.

### Closer

`maps.Keys(cfg.Contexts)` hands out the keys, and `slices.Sorted(...)` collects them into a sorted `[]string`. Add `"maps"` and `"slices"` to the imports.

### Answer

```go
return slices.Sorted(maps.Keys(cfg.Contexts))
```

`slices.Sorted` builds a new slice from the keys and sorts it, so an empty map gives an empty slice.

## Task 3: Find the server for a context

This takes two lookups: first the context by name, then the cluster that the context names. Each lookup can fail, and each failure has its own error text.

### Closer

Use the comma-ok form for both maps. The context's `Cluster` field is a cluster *name*, the key into `cfg.Clusters`:

```go
ctx, ok := cfg.Contexts[contextName]
if !ok {
	return "", fmt.Errorf("context %q not found", contextName)
}
// now look up ctx.Cluster in cfg.Clusters
```

### Answer

```go
ctx, ok := cfg.Contexts[contextName]
if !ok {
	return "", fmt.Errorf("context %q not found", contextName)
}
cluster, ok := cfg.Clusters[ctx.Cluster]
if !ok {
	return "", fmt.Errorf("cluster %q not found for context %q", ctx.Cluster, contextName)
}
return cluster.Server, nil
```

A context only refers to its cluster by name, so you follow that name to find the server.
