## Task 1: Fix DecodePods

Look at the second argument to `json.Unmarshal`. It has to change `list`, so
what must you pass when a function needs to change your variable?

### Closer

`Unmarshal` needs a pointer to the value it fills in. Add one character in
front of `list`.

### Answer

```go
if err := json.Unmarshal(data, &list); err != nil {
	return nil, fmt.Errorf("decode pods: %w", err)
}
```

Passing `&list` lets `Unmarshal` write into your variable instead of a copy.

## Task 2: Fix the Pod struct

The json package can only fill exported fields (capital first letter), and
the tag must match the JSON key exactly.

### Closer

Rename `namespace` to `Namespace` in the struct **and** in `Ref`. Change the
tag on `Phase` from `json:"state"` to the key the JSON really uses.

### Answer

```go
type Metadata struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type Status struct {
	Phase string `json:"phase"`
}

func (p Pod) Ref() string {
	return p.Metadata.Namespace + "/" + p.Metadata.Name
}
```

## Task 3: Count phases in a stream

Create one `json.NewDecoder(r)` before a loop, and call `Decode` once per pod
until it reports `io.EOF`. Keep a counter so you know which pod failed.

### Closer

Use `for n := 1; ; n++` and declare `var pod Pod` inside the loop. Check
`errors.Is(err, io.EOF)` first (return the counts), then any other error
(wrap it with `fmt.Errorf("pod %d: %w", n, err)`), then do
`counts[pod.Status.Phase]++`. Add `"errors"` to the imports.

### Answer

```go
dec := json.NewDecoder(r)
for n := 1; ; n++ {
	var pod Pod
	err := dec.Decode(&pod)
	if errors.Is(err, io.EOF) {
		return counts, nil
	}
	if err != nil {
		return nil, fmt.Errorf("pod %d: %w", n, err)
	}
	counts[pod.Status.Phase]++
}
```

A fresh `pod` each time means no fields leak over from the previous object.
