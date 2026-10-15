## Task 1: Tag the Deployment struct

Each field needs a struct tag that tells `encoding/json` its key. Two fields
are special: one is dropped when empty, one is never encoded.

### Closer

A tag goes in backticks after the type, with no spaces inside:
``Service string `json:"service"` ``. Add `,omitempty` after the key for
`Error`, and use `json:"-"` for `AuthToken`.

### Answer

```go
type Deployment struct {
	Service   string    `json:"service"`
	Version   string    `json:"version"`
	Replicas  int       `json:"replicas"`
	Healthy   bool      `json:"healthy"`
	Started   time.Time `json:"started"`
	Error     string    `json:"error,omitempty"`
	AuthToken string    `json:"-"`
}
```

Keys come out in field order, and only `error` uses `omitempty`.

## Task 2: Count healthy deployments

Both keys must be present even when nothing is counted, so put them in the
map before the loop. Then add one to the right key for each deployment.

### Closer

Start with `counts := map[string]int{"healthy": 0, "unhealthy": 0}`, loop
with `for _, d := range deps`, and use `counts["healthy"]++` or
`counts["unhealthy"]++` depending on `d.Healthy`.

### Answer

```go
counts := map[string]int{"healthy": 0, "unhealthy": 0}
for _, d := range deps {
	if d.Healthy {
		counts["healthy"]++
	} else {
		counts["unhealthy"]++
	}
}
return counts
```

Ranging over a nil slice runs zero times, so `nil` gives two zeros.

## Task 3: Encode the report

Fill in a `Report` value and hand it to `json.MarshalIndent`. Watch out: a
nil slice encodes as `null`, not `[]`.

### Closer

Check `deps == nil` first and replace it with `[]Deployment{}`. Then call
`b, err := json.MarshalIndent(report, "", "  ")` and return `string(b)`.
Add `"encoding/json"` to the imports.

### Answer

```go
if deps == nil {
	deps = []Deployment{}
}
report := Report{
	Environment: env,
	GeneratedAt: generatedAt,
	Counts:      HealthCounts(deps),
	Deployments: deps,
}
b, err := json.MarshalIndent(report, "", "  ")
if err != nil {
	return "", err
}
return string(b), nil
```

An empty, non-nil slice is what makes the output say `"deployments": []`.
