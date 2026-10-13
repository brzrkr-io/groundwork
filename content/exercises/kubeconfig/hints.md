## Task 1: load, then read a field

`clientcmd.Load(data)` returns two values: the config and an error. Check the
error first, then return `cfg.CurrentContext`. Don't forget the import
`"k8s.io/client-go/tools/clientcmd"`.

## Task 2: keys, then sort

You want the keys of the `cfg.Contexts` map, not its values.
`maps.Keys(cfg.Contexts)` gives you the keys, and `slices.Sorted(...)` turns
them into a sorted `[]string`. Add `"maps"` and `"slices"` to the imports.

## Task 3: two lookups

Look up the context with comma-ok. Its `Cluster` field is a cluster *name*,
so you need a second comma-ok lookup in `cfg.Clusters`:

```go
ctx, ok := cfg.Contexts[contextName]
if !ok {
	return "", fmt.Errorf("context %q not found", contextName)
}
cluster, ok := cfg.Clusters[ctx.Cluster]
if !ok {
	return "", fmt.Errorf("cluster %q not found for context %q", ctx.Cluster, contextName)
}
// return the cluster's server
```
