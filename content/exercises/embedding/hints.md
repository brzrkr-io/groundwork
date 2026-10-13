## Task 1: fill in the embedded struct

In a struct literal, the embedded struct is a field named after its type. Add a
`Metadata: Metadata{...}` line next to `Image: image`. A map with one entry is
written `map[string]string{"app": name}`.

## Task 2: join with a slash

Inside `Describe`, `m` is the `Metadata`. Join `m.Namespace`, `"/"` and
`m.Name` with `+`.

## Task 3: call the inner Describe

Inside the `Pod` method, `p.Metadata.Describe()` gives you `"default/api"`.
Don't write `p.Describe()`: that calls the `Pod` method itself, forever.

## Almost the answer

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

For task 3, add `import "fmt"` below the `package` line, then:

```go
return fmt.Sprintf("pod %s (%s)", p.Metadata.Describe(), p.Image)
```
