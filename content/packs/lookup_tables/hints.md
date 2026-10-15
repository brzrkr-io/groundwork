## Task 1: Turn the routing switch into a table

Each `case` is one row: the alert name is the key, and the team and severity are the value. A struct with two fields holds both.

### Closer

`type route struct { team, severity string }` and `var routes = map[string]route{"DiskFull": {"infra", "critical"}, ...}`. In `Route`, use `r, ok := routes[alert]`, and when `!ok`, use the default route instead. Giving the default its own variable (`unknownRoute`) keeps `"platform"` in one place.

### Answer

```go
// route says who gets paged for an alert, and how urgently.
type route struct {
	team     string
	severity string
}

// routes maps an alert's name to its route.
var routes = map[string]route{
	"DiskFull":      {"infra", "critical"},
	"DiskFillingUp": {"infra", "warning"},
	// ... one row per case
}

// unknownRoute is where alerts that aren't in routes go.
var unknownRoute = route{"platform", "info"}

func Route(alert string) (team, severity string) {
	r, ok := routes[alert]
	if !ok {
		r = unknownRoute
	}
	return r.team, r.severity
}
```

## Task 2: Turn the exit code switch into a table

The key is the `int` code and the value is the `string` reason. The `default` case becomes what you return when the lookup fails.

### Closer

`var exitReasons = map[int]string{0: "success", 1: "general error", ...}`. In `ExitReason`: `if reason, ok := exitReasons[code]; ok { return reason }`, then the `fmt.Sprintf` line.

### Answer

```go
// exitReasons explains the exit codes that show up in deploy logs.
var exitReasons = map[int]string{
	0:   "success",
	1:   "general error",
	// ... one row per case
	143: "terminated (SIGTERM)",
}

func ExitReason(code int) string {
	if reason, ok := exitReasons[code]; ok {
		return reason
	}
	return fmt.Sprintf("exit code %d", code)
}
```

## Task 3: List the teams

The new route is one row in `routes`. For `Teams`, start the list with the team for unknown alerts, loop over `routes`, add each team you haven't seen, then sort.

### Closer

`slices.Contains(teams, r.team)` tells you if a team is already in the list, and `slices.Sort(teams)` sorts it in place. Add `"slices"` to the imports.

### Answer

```go
"PodCrashLooping": {"apps", "critical"},
```

```go
func Teams() []string {
	teams := []string{unknownRoute.team}
	for _, r := range routes {
		if !slices.Contains(teams, r.team) {
			teams = append(teams, r.team)
		}
	}
	slices.Sort(teams)
	return teams
}
```

`unknownRoute` is the default route from Task 1's Answer. If you wrote the default as `return "platform", "info"` instead, start the list with `[]string{"platform"}`, or declare `unknownRoute` so `"platform"` lives in one place.
