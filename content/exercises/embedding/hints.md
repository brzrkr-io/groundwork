## Task 1: Build a pod

Promoted fields can't be used in a struct literal, so `Name: name` won't
compile inside `Pod{...}`. Fill in the embedded struct as a field named after
its type.

### Closer

Add a `Metadata: Metadata{...}` line next to `Image: image`. A map with one
entry is written `map[string]string{"app": name}`.

### Answer

```go
return Pod{
	Metadata: Metadata{
		Name:      name,
		Namespace: namespace,
		Labels:    map[string]string{"app": name},
	},
	Image: image,
}
```

## Task 2: Describe the metadata

Inside `Describe`, `m` is the `Metadata`, so its namespace and name are right
there. Join them with a slash.

### Closer

Strings join with `+`: `m.Namespace + "/" + ...`.

### Answer

```go
return m.Namespace + "/" + m.Name
```

## Task 3: Override Describe for pods

Task 2 already builds `default/api`. Call that method from the `Pod` version
and wrap the result with `pod ` and the image.

### Closer

`p.Metadata.Describe()` runs `Metadata`'s method. Don't write `p.Describe()`:
that calls the `Pod` method itself, forever. For `fmt.Sprintf`, add
`import "fmt"` below the `package` line and use `"pod %s (%s)"`.

### Answer

```go
return fmt.Sprintf("pod %s (%s)", p.Metadata.Describe(), p.Image)
```
