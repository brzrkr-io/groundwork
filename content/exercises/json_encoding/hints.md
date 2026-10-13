## Task 1: one tag per field

Each field gets a tag in backquotes after its type, with no spaces inside:

```go
Service string `json:"service"`
```

`Error` needs `,omitempty` after its key, and `AuthToken` needs `json:"-"`.

## Task 2: start with both keys

Create the map with both keys set to zero, then loop over the deployments and
add one to the right key:

```go
counts := map[string]int{"healthy": 0, "unhealthy": 0}
```

## Task 3: nil becomes null

A nil slice encodes as `null`. Check `deps == nil` first and replace it with
`[]Deployment{}`. Then fill in a `Report` value and pass it to
`json.MarshalIndent`.

## Task 3: almost the answer

```go
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
